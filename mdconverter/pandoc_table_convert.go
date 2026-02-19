package mdconverter

import (
	"strings"

	"github.com/rgonek/jira-adf-converter/converter"
)

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

func (s *state) convertPandocGridTableRow(cells []string, header bool) (converter.Node, error) {
	row := converter.Node{
		Type: "tableRow",
	}

	for _, cell := range cells {
		inlineContent, err := s.convertInlineFragment(cell)
		if err != nil {
			return converter.Node{}, err
		}

		cellType := "tableCell"
		if header {
			cellType = "tableHeader"
		}
		row.Content = append(row.Content, converter.Node{
			Type: cellType,
			Content: []converter.Node{
				{
					Type:    "paragraph",
					Content: inlineContent,
				},
			},
		})
	}

	return row, nil
}

func parsePandocGridTableLines(lines []string) ([][]string, [][]string, int, bool) {
	if len(lines) < 3 {
		return nil, nil, 0, false
	}

	widths, _, ok := parsePandocGridBorder(lines[0])
	if !ok {
		return nil, nil, 0, false
	}
	columns := len(widths)
	if columns == 0 {
		return nil, nil, 0, false
	}

	headerRows := make([][]string, 0, 1)
	dataRows := make([][]string, 0, 2)

	pending := []string(nil)
	headerMode := true
	headerSeparatorSeen := false

	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "|"):
			cells, rowOK := parsePandocGridRow(line, columns)
			if !rowOK {
				return nil, nil, 0, false
			}
			if pending == nil {
				pending = cells
				continue
			}
			for idx := 0; idx < columns; idx++ {
				part := strings.TrimSpace(cells[idx])
				if part == "" {
					continue
				}
				if pending[idx] != "" {
					pending[idx] += " "
				}
				pending[idx] += part
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

func parsePandocGridBorder(line string) ([]int, byte, bool) {
	if !pandocGridBorderRe.MatchString(line) || len(line) < 3 {
		return nil, 0, false
	}

	parts := strings.Split(line[1:len(line)-1], "+")
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
			if part[i] != '-' && part[i] != '=' {
				return nil, 0, false
			}
			if separator == 0 {
				separator = part[i]
			}
		}
		widths[idx] = len(part)
	}

	return widths, separator, true
}

func parsePandocGridRow(line string, columns int) ([]string, bool) {
	if len(line) < 2 || !strings.HasPrefix(line, "|") || !strings.HasSuffix(line, "|") {
		return nil, false
	}

	parts := strings.Split(line[1:len(line)-1], "|")
	if len(parts) != columns {
		return nil, false
	}

	cells := make([]string, columns)
	for idx, part := range parts {
		cells[idx] = strings.TrimSpace(part)
	}

	return cells, true
}
