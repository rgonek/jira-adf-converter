package converter

import (
	"fmt"
	"html"
	"strings"
)

// convertTable converts a table node to markdown/HTML depending on config.
func (s *state) convertTable(node Node) (string, error) {
	mode := s.config.TableMode
	if mode == TableAuto {
		if s.isComplexTable(node) {
			mode = TableHTML
		} else {
			mode = TablePipe
		}
	}
	if mode == TableAutoPandoc {
		if s.isComplexTable(node) {
			mode = TablePandoc
		} else {
			mode = TablePipe
		}
	}

	switch mode {
	case TableHTML:
		return s.renderTableHTML(node)
	case TablePandoc:
		rows, colCount, err := s.buildTableGrid(node)
		if err != nil {
			return "", err
		}
		if colCount == 0 {
			return "", nil
		}
		return s.renderGridTable(rows, colCount), nil

	default:
		rows, err := s.extractTableRows(node)
		if err != nil {
			return "", err
		}
		if len(rows) == 0 {
			return "", nil
		}
		return s.renderTableGFM(rows), nil
	}
}

func (s *state) isComplexTable(node Node) bool {
	for _, rowNode := range node.Content {
		if rowNode.Type != "tableRow" {
			continue
		}
		for _, cellNode := range rowNode.Content {
			if cellNode.Type != "tableCell" && cellNode.Type != "tableHeader" {
				continue
			}
			if cellNode.GetIntAttr("colspan", 1) > 1 || cellNode.GetIntAttr("rowspan", 1) > 1 {
				return true
			}
			for _, child := range cellNode.Content {
				if isComplexTableBlockNode(child.Type) {
					return true
				}
			}
		}
	}

	return false
}

func (s *state) hasTableSpans(node Node) bool {
	for _, rowNode := range node.Content {
		if rowNode.Type != "tableRow" {
			continue
		}
		for _, cellNode := range rowNode.Content {
			if cellNode.Type != "tableCell" && cellNode.Type != "tableHeader" {
				continue
			}
			if cellNode.GetIntAttr("colspan", 1) > 1 || cellNode.GetIntAttr("rowspan", 1) > 1 {
				return true
			}
		}
	}

	return false
}

func isComplexTableBlockNode(nodeType string) bool {
	switch nodeType {
	case "bulletList", "orderedList", "taskList", "codeBlock", "table":
		return true
	default:
		return false
	}
}

type adfCell struct {
	content  []string
	isHeader bool
	colspan  int
	rowspan  int
}

type gridPos struct {
	cell     *adfCell
	isOrigin bool
}

func (s *state) buildTableGrid(node Node) ([][]gridPos, int, error) {
	var adfRows []Node
	for _, child := range node.Content {
		if child.Type == "tableRow" {
			adfRows = append(adfRows, child)
		}
	}
	if len(adfRows) == 0 {
		return nil, 0, nil
	}

	// First pass: find dimensions
	rowCount := len(adfRows)
	colCount := 0
	for _, rowNode := range adfRows {
		colsInThisRow := 0
		for _, cellNode := range rowNode.Content {
			colsInThisRow += cellNode.GetIntAttr("colspan", 1)
		}
		if colsInThisRow > colCount {
			colCount = colsInThisRow
		}
	}

	// Initialize grid
	grid := make([][]gridPos, rowCount)
	for i := range grid {
		grid[i] = make([]gridPos, colCount)
	}

	// Second pass: fill grid
	for r, rowNode := range adfRows {
		c := 0
		for _, cellNode := range rowNode.Content {
			// Skip already filled cells (from rowspan above)
			for c < colCount && grid[r][c].cell != nil {
				c++
			}
			if c >= colCount {
				break
			}

			colspan := cellNode.GetIntAttr("colspan", 1)
			rowspan := cellNode.GetIntAttr("rowspan", 1)
			if colspan < 1 {
				colspan = 1
			}
			if rowspan < 1 {
				rowspan = 1
			}

			cellContent, err := s.convertCellContent(cellNode)
			if err != nil {
				return nil, 0, err
			}
			// Split content into lines for rendering
			lines := strings.Split(cellContent, "<br>")
			// In Grid tables, we want actual newlines
			var finalLines []string
			for _, line := range lines {
				finalLines = append(finalLines, strings.Split(line, "\n")...)
			}

			cell := &adfCell{
				content:  finalLines,
				isHeader: cellNode.Type == "tableHeader",
				colspan:  colspan,
				rowspan:  rowspan,
			}

			for dr := 0; dr < rowspan; dr++ {
				for dc := 0; dc < colspan; dc++ {
					if r+dr < rowCount && c+dc < colCount {
						grid[r+dr][c+dc] = gridPos{
							cell:     cell,
							isOrigin: dr == 0 && dc == 0,
						}
					}
				}
			}
			c += colspan
		}
	}

	return grid, colCount, nil
}

// extractTableRows extracts and normalizes table rows from the node.
func (s *state) extractTableRows(node Node) ([][]string, error) {
	if len(node.Content) == 0 {
		return nil, nil
	}

	var rows [][]string
	var rowNodes []Node
	hasHeader := false

	// Process all rows.
	for i, rowNode := range node.Content {
		if rowNode.Type != "tableRow" {
			continue
		}

		var row []string
		isHeaderRow := false

		// Process cells in this row.
		for _, cellNode := range rowNode.Content {
			if cellNode.Type == "tableHeader" {
				isHeaderRow = true
			}
			cellContent, err := s.convertCellContent(cellNode)
			if err != nil {
				return nil, err
			}
			row = append(row, cellContent)
		}

		// Check if first row has headers.
		if i == 0 && isHeaderRow {
			hasHeader = true
		}

		rows = append(rows, row)
		rowNodes = append(rowNodes, rowNode)
	}

	if len(rows) == 0 {
		return nil, nil
	}

	// Normalize rows based on whether we have headers or not.
	if !hasHeader {
		// For a simple single-row table, treat that row as the header row.
		if len(rows) == 1 && len(rowNodes) == 1 && s.singleRowAsHeaderCandidate(rowNodes[0]) {
			return rows, nil
		}

		// Create empty header row if missing.
		colCount := 0
		for _, r := range rows {
			if len(r) > colCount {
				colCount = len(r)
			}
		}
		headerRow := make([]string, colCount)
		// Prepend header row.
		rows = append([][]string{headerRow}, rows...)
	}

	return rows, nil
}

func (s *state) singleRowAsHeaderCandidate(row Node) bool {
	if len(row.Content) == 0 {
		return false
	}

	for _, cellNode := range row.Content {
		if cellNode.Type != "tableCell" || len(cellNode.Content) != 1 {
			return false
		}
		if cellNode.Content[0].Type != "paragraph" {
			return false
		}
		for _, inlineNode := range cellNode.Content[0].Content {
			if inlineNode.Type != "text" || len(inlineNode.Marks) > 0 {
				return false
			}
		}
	}

	return true
}

// renderTableGFM renders a matrix of strings as a GFM table.
func (s *state) renderTableGFM(rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}

	// Determine column count.
	colCount := 0
	for _, row := range rows {
		if len(row) > colCount {
			colCount = len(row)
		}
	}

	var sb strings.Builder

	// Header is always row 0 after normalization.
	headerRow := rows[0]
	dataRows := rows[1:]

	// Write header row.
	sb.WriteString("|")
	for i := 0; i < colCount; i++ {
		sb.WriteString(" ")
		if i < len(headerRow) {
			sb.WriteString(headerRow[i])
		}
		sb.WriteString(" |")
	}
	sb.WriteString("\n")

	// Write separator.
	sb.WriteString("|")
	for i := 0; i < colCount; i++ {
		sb.WriteString(" --- |")
	}
	sb.WriteString("\n")

	// Write data rows.
	for _, row := range dataRows {
		sb.WriteString("|")
		for i := 0; i < colCount; i++ {
			sb.WriteString(" ")
			if i < len(row) {
				sb.WriteString(row[i])
			}
			sb.WriteString(" |")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

func (s *state) renderGridTable(grid [][]gridPos, colCount int) string {
	rowCount := len(grid)
	colWidths := make([]int, colCount)

	// Pass 1: compute widths from cells that span exactly one column.
	for r := 0; r < rowCount; r++ {
		for c := 0; c < colCount; c++ {
			pos := grid[r][c]
			if pos.cell != nil && pos.isOrigin && pos.cell.colspan == 1 {
				for _, line := range pos.cell.content {
					if len(line) > colWidths[c] {
						colWidths[c] = len(line)
					}
				}
			}
		}
	}
	// Minimum column width of 1 so borders are valid.
	for c := range colWidths {
		if colWidths[c] < 1 {
			colWidths[c] = 1
		}
	}

	// Pass 2: expand columns to fit multi-column (colspan > 1) cells.
	// combinedWidth(c, span) = sum(colWidths[c..c+span-1]) + 3*(span-1)
	// The +3 per separator accounts for " | " between columns.
	combinedWidth := func(c, span int) int {
		w := 0
		for i := 0; i < span; i++ {
			w += colWidths[c+i]
		}
		w += 3 * (span - 1)
		return w
	}
	for r := 0; r < rowCount; r++ {
		for c := 0; c < colCount; c++ {
			pos := grid[r][c]
			if !pos.isOrigin || pos.cell == nil || pos.cell.colspan <= 1 {
				continue
			}
			colspan := pos.cell.colspan
			maxLen := 0
			for _, line := range pos.cell.content {
				if len(line) > maxLen {
					maxLen = len(line)
				}
			}
			cur := combinedWidth(c, colspan)
			if maxLen > cur {
				extra := maxLen - cur
				// Distribute extra width evenly across spanned columns.
				perCol := extra / colspan
				rem := extra % colspan
				for i := 0; i < colspan; i++ {
					colWidths[c+i] += perCol
					if i < rem {
						colWidths[c+i]++
					}
				}
			}
		}
	}

	// Compute row heights: each logical row height = max content lines across
	// all origin cells whose rowspan == 1 that start on this row.
	// For rowspan > 1 cells, we only need enough total rows to fit all lines;
	// assign all overflow to the cell's first row (top-aligned).
	rowHeights := make([]int, rowCount)
	for r := 0; r < rowCount; r++ {
		for c := 0; c < colCount; c++ {
			pos := grid[r][c]
			if !pos.isOrigin || pos.cell == nil {
				continue
			}
			h := len(pos.cell.content)
			if h < 1 {
				h = 1
			}
			if pos.cell.rowspan == 1 {
				if h > rowHeights[r] {
					rowHeights[r] = h
				}
			}
		}
	}
	// Rowspan > 1: ensure the spanned rows collectively have enough height.
	for r := 0; r < rowCount; r++ {
		for c := 0; c < colCount; c++ {
			pos := grid[r][c]
			if !pos.isOrigin || pos.cell == nil || pos.cell.rowspan <= 1 {
				continue
			}
			span := pos.cell.rowspan
			h := len(pos.cell.content)
			if h < 1 {
				h = 1
			}
			totalHeight := 0
			for i := 0; i < span && r+i < rowCount; i++ {
				totalHeight += rowHeights[r+i]
			}
			if h > totalHeight {
				rowHeights[r] += h - totalHeight
			}
		}
	}
	for i := range rowHeights {
		if rowHeights[i] < 1 {
			rowHeights[i] = 1
		}
	}

	// Pre-compute the starting line index (within the cell content) for each
	// (row, col) pair so the inner render loop is simple.
	// lineStart[r][c] = the content-line index to use for the *first* rendered
	// line of logical row r for the cell at (r,c).
	lineStart := make([][]int, rowCount)
	for r := range lineStart {
		lineStart[r] = make([]int, colCount)
	}
	for r := 0; r < rowCount; r++ {
		for c := 0; c < colCount; c++ {
			pos := grid[r][c]
			if pos.cell == nil || pos.isOrigin {
				// Origin rows start at 0 within the cell content.
				lineStart[r][c] = 0
				continue
			}
			// Non-origin: find origin row and accumulate row heights above.
			originRow := r
			for originRow > 0 && grid[originRow-1][c].cell == pos.cell {
				originRow--
			}
			offset := 0
			for i := originRow; i < r; i++ {
				offset += rowHeights[i]
			}
			lineStart[r][c] = offset
		}
	}

	var sb strings.Builder

	// writeBorder draws a separator line between logical rows r-1 and r
	// (r==0 → top border, r==rowCount → bottom border).
	// ch is "-" or "=" for the horizontal fill character.
	//
	// Pandoc grid-table rules for the border:
	//   - Between two columns whose cells are *different* at that boundary:
	//     write "+" as the column-junction.
	//   - Within a colspan span (same cell left and right of the column
	//     junction): write ch instead of "+" (no vertical bar crosses here).
	writeBorder := func(r int, ch string) {
		sb.WriteString("+")
		for c := 0; c < colCount; c++ {
			// Determine whether this column has a horizontal segment drawn.
			var cellAbove, cellBelow *adfCell
			if r > 0 {
				cellAbove = grid[r-1][c].cell
			}
			if r < rowCount {
				cellBelow = grid[r][c].cell
			}
			drawHoriz := (r == 0 || r == rowCount || cellAbove != cellBelow)
			if drawHoriz {
				sb.WriteString(strings.Repeat(ch, colWidths[c]+2))
			} else {
				sb.WriteString(strings.Repeat(" ", colWidths[c]+2))
			}

			// Column junction symbol after this column segment.
			// "+" unless both sides of the junction share the same cell
			// (colspan crossing this boundary) AND we are actually drawing
			// the horizontal segment (otherwise a space is already there).
			if c < colCount-1 {
				var leftAbove, rightAbove, leftBelow, rightBelow *adfCell
				if r > 0 {
					leftAbove = grid[r-1][c].cell
					rightAbove = grid[r-1][c+1].cell
				}
				if r < rowCount {
					leftBelow = grid[r][c].cell
					rightBelow = grid[r][c+1].cell
				}
				// If the same cell spans across the c/c+1 boundary on both
				// sides, the junction is interior to a colspan — use ch or " ".
				colspanAbove := (r > 0) && (leftAbove == rightAbove) && leftAbove != nil
				colspanBelow := (r < rowCount) && (leftBelow == rightBelow) && leftBelow != nil
				if colspanAbove && colspanBelow {
					// Both rows have the same cell spanning — use fill char.
					sb.WriteString(ch)
				} else if colspanAbove && r == rowCount {
					sb.WriteString(ch)
				} else if colspanBelow && r == 0 {
					sb.WriteString(ch)
				} else if !drawHoriz && colspanAbove {
					// No horizontal drawn on left side but spans above.
					sb.WriteString(" ")
				} else {
					sb.WriteString("+")
				}
			} else {
				sb.WriteString("+")
			}
		}
		sb.WriteString("\n")
	}

	for r := 0; r < rowCount; r++ {
		// Border above this row.
		if r == 0 {
			writeBorder(0, "-")
		} else {
			// Use "=" if the *previous* row contained header cells.
			prevIsHeader := false
			for c := 0; c < colCount; c++ {
				if grid[r-1][c].cell != nil && grid[r-1][c].cell.isHeader {
					prevIsHeader = true
					break
				}
			}
			if prevIsHeader {
				writeBorder(r, "=")
			} else {
				writeBorder(r, "-")
			}
		}

		// Content lines for this logical row.
		for lh := 0; lh < rowHeights[r]; lh++ {
			sb.WriteString("|")
			for c := 0; c < colCount; c++ {
				pos := grid[r][c]
				cell := pos.cell

				// Skip non-origin columns of a colspan (already handled by the
				// origin column writing wider content).
				if cell != nil && !pos.isOrigin && c > 0 && grid[r][c-1].cell == cell {
					// This column is interior to a colspan — no leading "|" or
					// content; the previous column already wrote past it.
					continue
				}

				// Compute the total display width for this cell (may span columns).
				displayWidth := colWidths[c]
				if cell != nil && cell.colspan > 1 {
					displayWidth = combinedWidth(c, cell.colspan)
				}

				lineIdx := lineStart[r][c] + lh
				content := ""
				if cell != nil && lineIdx < len(cell.content) {
					content = cell.content[lineIdx]
				}

				sb.WriteString(" ")
				sb.WriteString(content)
				sb.WriteString(strings.Repeat(" ", displayWidth-len(content)))
				sb.WriteString(" ")

				// Trailing separator: "|" unless a subsequent column in this
				// row belongs to the same cell (colspan continuation).
				nextC := c + 1
				if cell != nil && cell.colspan > 1 {
					nextC = c + cell.colspan
				}
				if nextC >= colCount {
					sb.WriteString("|")
				} else if grid[r][nextC].cell == cell {
					// Still inside the colspan (shouldn't normally happen with
					// the skip-logic above, but guard anyway).
					sb.WriteString(" ")
				} else {
					sb.WriteString("|")
				}
			}
			sb.WriteString("\n")
		}
	}

	// Bottom border.
	writeBorder(rowCount, "-")
	sb.WriteString("\n")

	return sb.String()
}

func (s *state) renderTableHTML(node Node) (string, error) {
	var rows []Node
	for _, child := range node.Content {
		if child.Type == "tableRow" {
			rows = append(rows, child)
		}
	}
	if len(rows) == 0 {
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString("<table>\n")

	if s.rowHasHeaders(rows[0]) {
		sb.WriteString("  <thead>\n")
		headerRow, err := s.renderHTMLRow(rows[0])
		if err != nil {
			return "", err
		}
		sb.WriteString(headerRow)
		sb.WriteString("  </thead>\n")

		if len(rows) > 1 {
			sb.WriteString("  <tbody>\n")
			for _, rowNode := range rows[1:] {
				rendered, err := s.renderHTMLRow(rowNode)
				if err != nil {
					return "", err
				}
				sb.WriteString(rendered)
			}
			sb.WriteString("  </tbody>\n")
		}
	} else {
		sb.WriteString("  <tbody>\n")
		for _, rowNode := range rows {
			rendered, err := s.renderHTMLRow(rowNode)
			if err != nil {
				return "", err
			}
			sb.WriteString(rendered)
		}
		sb.WriteString("  </tbody>\n")
	}

	sb.WriteString("</table>\n\n")
	return sb.String(), nil
}

func (s *state) rowHasHeaders(row Node) bool {
	for _, cell := range row.Content {
		if cell.Type == "tableHeader" {
			return true
		}
	}
	return false
}

func (s *state) renderHTMLRow(row Node) (string, error) {
	var sb strings.Builder
	sb.WriteString("    <tr>\n")

	for _, cell := range row.Content {
		switch cell.Type {
		case "tableHeader":
			rendered, err := s.renderHTMLCell(cell, "th")
			if err != nil {
				return "", err
			}
			sb.WriteString(rendered)
		case "tableCell":
			rendered, err := s.renderHTMLCell(cell, "td")
			if err != nil {
				return "", err
			}
			sb.WriteString(rendered)
		}
	}

	sb.WriteString("    </tr>\n")
	return sb.String(), nil
}

func (s *state) renderHTMLCell(cell Node, tag string) (string, error) {
	content, err := s.convertCellContentForHTML(cell)
	if err != nil {
		return "", err
	}

	var attrs strings.Builder
	if colspan := cell.GetIntAttr("colspan", 1); colspan > 1 {
		attrs.WriteString(fmt.Sprintf(` colspan="%d"`, colspan))
	}
	if rowspan := cell.GetIntAttr("rowspan", 1); rowspan > 1 {
		attrs.WriteString(fmt.Sprintf(` rowspan="%d"`, rowspan))
	}

	var sb strings.Builder
	sb.WriteString("      <")
	sb.WriteString(tag)
	sb.WriteString(attrs.String())
	sb.WriteString(">\n")

	if content != "" {
		for _, line := range strings.Split(content, "\n") {
			if line == "" {
				continue
			}
			sb.WriteString("        ")
			sb.WriteString(html.EscapeString(line))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("      </")
	sb.WriteString(tag)
	sb.WriteString(">\n")

	return sb.String(), nil
}

func (s *state) convertCellContentForHTML(node Node) (string, error) {
	if len(node.Content) == 0 {
		return "", nil
	}

	var parts []string
	for _, child := range node.Content {
		content, err := s.convertNode(child)
		if err != nil {
			return "", err
		}
		content = strings.TrimRight(content, "\n")
		if content != "" {
			parts = append(parts, content)
		}
	}

	return strings.Join(parts, "\n"), nil
}

// convertTableCell processes a table cell (header or data).
// Note: isHeader parameter is currently unused since GFM tables don't require
// different content processing for header vs data cells. The only difference
// is in the separator row. Kept for API consistency and future extensibility.
func (s *state) convertTableCell(node Node, isHeader bool) (string, error) {
	return s.convertCellContent(node)
}

// convertCellContent processes the content of a table cell, preserving block-level content.
func (s *state) convertCellContent(node Node) (string, error) {
	if len(node.Content) == 0 {
		return "", nil
	}

	var parts []string
	for _, child := range node.Content {
		switch child.Type {
		case "paragraph":
			// Process paragraph inline content without the trailing newlines.
			content, err := s.convertInlineContent(child.Content)
			if err != nil {
				return "", err
			}
			if content != "" {
				parts = append(parts, content)
			}

		case "bulletList", "orderedList", "taskList":
			res, err := s.convertListInTable(child)
			if err != nil {
				return "", err
			}
			if res != "" {
				parts = append(parts, res)
			}

		case "codeBlock":
			res, err := s.convertCodeBlockInTable(child)
			if err != nil {
				return "", err
			}
			if res != "" {
				parts = append(parts, res)
			}

		case "panel":
			panelContent, err := s.convertNode(child)
			if err != nil {
				return "", err
			}
			panelContent = strings.TrimRight(panelContent, "\n")
			if panelContent != "" {
				parts = append(parts, panelContent)
			}

		case "blockquote":
			quoteContent, err := s.convertNode(child)
			if err != nil {
				return "", err
			}
			quoteContent = strings.TrimRight(quoteContent, "\n")
			if quoteContent != "" {
				parts = append(parts, quoteContent)
			}

		default:
			content, err := s.convertNode(child)
			if err != nil {
				return "", err
			}
			content = strings.TrimRight(content, "\n")
			if content != "" {
				parts = append(parts, content)
			}
		}
	}

	sep := "<br>"
	if s.config.HardBreakStyle != HardBreakHTML {
		sep = " "
	}
	result := strings.Join(parts, sep)
	// Escape pipe characters as they break GFM tables.
	// Note: Child converters must NOT pre-escape pipes as this would cause
	// double-escaping. Only this final output should escape pipes.
	return strings.ReplaceAll(result, "|", "\\|"), nil
}

func (s *state) convertListInTable(node Node) (string, error) {
	listContent, err := s.convertNode(node)
	if err != nil {
		return "", err
	}
	listContent = strings.TrimRight(listContent, "\n")
	if listContent == "" {
		return "", nil
	}

	// Split by newlines and join with style-aware separator.
	lines := strings.Split(listContent, "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimRight(line, "\n")
		if trimmed != "" {
			cleanLines = append(cleanLines, trimmed)
		}
	}
	if len(cleanLines) == 0 {
		return "", nil
	}

	sep := "<br>"
	if s.config.HardBreakStyle != HardBreakHTML {
		sep = " "
	}
	return strings.Join(cleanLines, sep), nil
}

func (s *state) convertCodeBlockInTable(node Node) (string, error) {
	rawCode := s.extractTextFromContent(node.Content)
	if strings.TrimSpace(rawCode) == "" {
		return "", nil
	}

	if s.config.HardBreakStyle == HardBreakHTML {
		// Escape HTML special chars.
		safeCode := strings.ReplaceAll(rawCode, "&", "&amp;")
		safeCode = strings.ReplaceAll(safeCode, "<", "&lt;")
		safeCode = strings.ReplaceAll(safeCode, ">", "&gt;")
		safeCode = strings.ReplaceAll(safeCode, "\"", "&quot;")
		// Replace newlines with <br>.
		safeCode = strings.ReplaceAll(safeCode, "\n", "<br>")
		return "<code>" + safeCode + "</code>", nil
	}

	// Flatten and use backticks if HTML is not preferred.
	flatCode := strings.ReplaceAll(rawCode, "\n", " ")
	return "`" + flatCode + "`", nil
}
