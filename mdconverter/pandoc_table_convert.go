package mdconverter

import (
	"strings"

	"github.com/rgonek/jira-adf-converter/converter"
)

// gridCell holds the parsed text content and colspan for a single table cell.
type gridCell struct {
	text    string
	colspan int
}

func (s *state) convertPandocGridTableNode(node *PandocGridTableNode) (converter.Node, bool, error) {
	literalFallback := pandocLiteralParagraph(node.Literal())
	if !s.config.TableGridDetection {
		return literalFallback, true, nil
	}

	headerRows, dataRows, colCount, ok := parsePandocGridTableLines(node.RawLines())
	if !ok || colCount == 0 {
		s.addWarning(converter.WarningDroppedFeature, "pandocGridTable", "invalid pandoc grid table; preserved as text")
		return literalFallback, true, nil
	}

	table := converter.Node{
		Type: "table",
	}

	for _, row := range headerRows {
		converted, err := s.convertPandocGridTableRow(row, true)
		if err != nil {
			return converter.Node{}, false, err
		}
		table.Content = append(table.Content, converted)
	}
	for _, row := range dataRows {
		converted, err := s.convertPandocGridTableRow(row, false)
		if err != nil {
			return converter.Node{}, false, err
		}
		table.Content = append(table.Content, converted)
	}

	if len(table.Content) == 0 {
		return literalFallback, true, nil
	}
	return table, true, nil
}

func (s *state) convertPandocGridTableRow(cells []gridCell, header bool) (converter.Node, error) {
	row := converter.Node{
		Type: "tableRow",
	}

	for _, cell := range cells {
		inlineContent, err := s.convertInlineFragment(cell.text)
		if err != nil {
			return converter.Node{}, err
		}

		cellType := "tableCell"
		if header {
			cellType = "tableHeader"
		}
		cellNode := converter.Node{
			Type: cellType,
			Content: []converter.Node{
				{
					Type:    "paragraph",
					Content: inlineContent,
				},
			},
		}
		if cell.colspan > 1 {
			cellNode.Attrs = map[string]interface{}{
				"colspan": float64(cell.colspan),
			}
		}
		row.Content = append(row.Content, cellNode)
	}

	return row, nil
}

// parsePandocGridTableLines parses all raw lines captured from a Pandoc grid
// table block into header and data rows of gridCell slices.
//
// The canonical column layout is derived from the border line with the most
// column segments (the "widest" border). This handles tables where the first
// border is a collapsed colspan row (e.g., a single-column border for a row
// that spans all columns).
func parsePandocGridTableLines(lines []string) ([][]gridCell, [][]gridCell, int, bool) {
	if len(lines) < 3 {
		return nil, nil, 0, false
	}

	// Find the canonical column layout: the border with the maximum number of
	// column segments. This is necessary because the first border may be a
	// single-column colspan border rather than the full-width border.
	var colWidths []int
	for _, line := range lines {
		if !strings.HasPrefix(line, "+") {
			continue
		}
		widths, _, ok := parsePandocGridBorder(line)
		if !ok {
			continue
		}
		if len(widths) > len(colWidths) {
			colWidths = widths
		}
	}
	if len(colWidths) == 0 {
		return nil, nil, 0, false
	}
	columns := len(colWidths)

	headerRows := make([][]gridCell, 0, 1)
	dataRows := make([][]gridCell, 0, 2)

	var pending []gridCell
	headerMode := true
	headerSeparatorSeen := false

	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "|"):
			cells, rowOK := parsePandocGridRow(line, colWidths)
			if !rowOK {
				return nil, nil, 0, false
			}
			if pending == nil {
				pending = cells
				continue
			}
			// Merge continuation lines: append non-empty text per cell.
			// Both pending and cells must have the same length (same colspan
			// layout).
			if len(cells) != len(pending) {
				return nil, nil, 0, false
			}
			for idx := range cells {
				part := strings.TrimSpace(cells[idx].text)
				if part == "" {
					continue
				}
				if pending[idx].text != "" {
					pending[idx].text += " "
				}
				pending[idx].text += part
			}

		case strings.HasPrefix(line, "+"):
			_, separatorChar, borderOK := parsePandocGridBorder(line)
			if !borderOK {
				return nil, nil, 0, false
			}
			if pending != nil {
				if headerMode {
					headerRows = append(headerRows, pending)
				} else {
					dataRows = append(dataRows, pending)
				}
				pending = nil
			}
			if separatorChar == '=' {
				headerSeparatorSeen = true
				headerMode = false
			}

		default:
			return nil, nil, 0, false
		}
	}

	if pending != nil {
		if headerMode {
			headerRows = append(headerRows, pending)
		} else {
			dataRows = append(dataRows, pending)
		}
	}

	if !headerSeparatorSeen {
		dataRows = append(dataRows, headerRows...)
		headerRows = nil
	}

	return headerRows, dataRows, columns, len(headerRows)+len(dataRows) > 0
}

// parsePandocGridBorder parses a Pandoc grid-table border line such as
// "+------+--------+" or "+-----------------+" (single column, used for
// colspan rows). It returns the width of each column segment (the number of
// '-' or '=' characters between two '+' signs), the fill character, and
// whether parsing succeeded.
func parsePandocGridBorder(line string) ([]int, byte, bool) {
	if len(line) < 3 || line[0] != '+' || line[len(line)-1] != '+' {
		return nil, 0, false
	}

	inner := line[1 : len(line)-1]
	parts := strings.Split(inner, "+")
	if len(parts) == 0 {
		return nil, 0, false
	}

	widths := make([]int, len(parts))
	separator := byte(0)
	for idx, part := range parts {
		if part == "" {
			return nil, 0, false
		}
		for i := 0; i < len(part); i++ {
			ch := part[i]
			if ch != '-' && ch != '=' {
				return nil, 0, false
			}
			if separator == 0 {
				separator = ch
			}
		}
		widths[idx] = len(part)
	}
	return widths, separator, true
}

// parsePandocGridRow parses a data line of the form "| cell1 | cell2 |".
// colWidths is the canonical per-column width array (from the first border
// line). The function handles colspan: if a segment's width matches the
// combined width of multiple canonical columns, a gridCell with colspan > 1
// is emitted.
//
// In a Pandoc grid table, the border line "+---+----+" has column widths
// [3, 4] (the dash counts, which include the surrounding spaces). The
// matching content line "| a | bc |" splits on "|" to give segments
// [" a ", " bc "] of lengths 3 and 4 respectively — exactly equal to the
// corresponding column widths.
//
// For a colspan=N merged cell, the content segment length equals:
//
//	colWidths[c] + colWidths[c+1] + ... + colWidths[c+N-1] + (N-1)
//
// where the +(N-1) accounts for the "+" junction characters that are now
// interior to the merged cell and appear as spaces in the content line.
func parsePandocGridRow(line string, colWidths []int) ([]gridCell, bool) {
	if len(line) < 2 || line[0] != '|' || line[len(line)-1] != '|' {
		return nil, false
	}

	// Split on "|" — each segment has width equal to one or more column widths.
	parts := strings.Split(line[1:len(line)-1], "|")
	if len(parts) == 0 {
		return nil, false
	}

	columns := len(colWidths)
	cells := make([]gridCell, 0, columns)
	colIdx := 0

	for _, part := range parts {
		if colIdx >= columns {
			return nil, false
		}

		segLen := len(part)

		// Try to match this segment against consecutive columns starting at colIdx.
		// For colspan=N, the expected segment width is:
		//   sum(colWidths[colIdx..colIdx+N-1]) + (N-1)
		// The +(N-1) accounts for the "+" junction characters between columns
		// that are now interior to the merged cell (shown as spaces).
		matched := false
		accumulated := 0
		for span := 1; colIdx+span-1 < columns; span++ {
			if span > 1 {
				accumulated++ // account for the "+" junction between columns
			}
			accumulated += colWidths[colIdx+span-1]
			if segLen == accumulated {
				cells = append(cells, gridCell{
					text:    strings.TrimSpace(part),
					colspan: span,
				})
				colIdx += span
				matched = true
				break
			}
			if segLen < accumulated {
				// Went past — this segment is malformed.
				break
			}
		}
		if !matched {
			return nil, false
		}
	}

	if colIdx != columns {
		// Not all columns were accounted for.
		return nil, false
	}

	return cells, true
}
