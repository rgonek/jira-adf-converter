package converter

import (
	"fmt"
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

	switch mode {
	case TableHTML:
		return s.renderTableHTML(node)
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

func isComplexTableBlockNode(nodeType string) bool {
	switch nodeType {
	case "bulletList", "orderedList", "taskList", "codeBlock", "table":
		return true
	default:
		return false
	}
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
			sb.WriteString(line)
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
