package converter

import (
	"strings"
)

// convertTable converts a table node to GFM table
func (c *Converter) convertTable(node Node) (string, error) {
	rows, err := c.extractTableRows(node)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", nil
	}

	return c.renderTableGFM(rows), nil
}

// extractTableRows extracts and normalizes table rows from the node
func (c *Converter) extractTableRows(node Node) ([][]string, error) {
	if len(node.Content) == 0 {
		return nil, nil
	}

	var rows [][]string
	hasHeader := false

	// Process all rows
	for i, rowNode := range node.Content {
		if rowNode.Type != "tableRow" {
			continue
		}

		var row []string
		isHeaderRow := false

		// Process cells in this row
		for _, cellNode := range rowNode.Content {
			if cellNode.Type == "tableHeader" {
				isHeaderRow = true
			}
			cellContent, err := c.convertCellContent(cellNode)
			if err != nil {
				return nil, err
			}
			row = append(row, cellContent)
		}

		// Check if first row has headers
		if i == 0 && isHeaderRow {
			hasHeader = true
		}

		rows = append(rows, row)
	}

	if len(rows) == 0 {
		return nil, nil
	}

	// Normalize rows based on whether we have headers or not
	if !hasHeader {
		// Create empty header row if missing
		colCount := 0
		for _, r := range rows {
			if len(r) > colCount {
				colCount = len(r)
			}
		}
		headerRow := make([]string, colCount)
		// Prepend header row
		rows = append([][]string{headerRow}, rows...)
	}

	return rows, nil
}

// renderTableGFM renders a matrix of strings as a GFM table
func (c *Converter) renderTableGFM(rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}

	// Determine column count
	colCount := 0
	for _, row := range rows {
		if len(row) > colCount {
			colCount = len(row)
		}
	}

	var sb strings.Builder

	// Header is always row 0 after normalization
	headerRow := rows[0]
	dataRows := rows[1:]

	// Write header row
	sb.WriteString("|")
	for i := 0; i < colCount; i++ {
		sb.WriteString(" ")
		if i < len(headerRow) {
			sb.WriteString(headerRow[i])
		} else {
			sb.WriteString(" ")
		}
		sb.WriteString(" |")
	}
	sb.WriteString("\n")

	// Write separator
	sb.WriteString("|")
	for i := 0; i < colCount; i++ {
		sb.WriteString(" --- |")
	}
	sb.WriteString("\n")

	// Write data rows
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

// convertTableCell processes a table cell (header or data)
// Note: isHeader parameter is currently unused since GFM tables don't require
// different content processing for header vs data cells. The only difference
// is in the separator row. Kept for API consistency and future extensibility.
func (c *Converter) convertTableCell(node Node, isHeader bool) (string, error) {
	return c.convertCellContent(node)
}

// convertCellContent processes the content of a table cell, preserving block-level content
func (c *Converter) convertCellContent(node Node) (string, error) {
	if len(node.Content) == 0 {
		return "", nil
	}

	var parts []string
	for _, child := range node.Content {
		switch child.Type {
		case "paragraph":
			// Process paragraph inline content without the trailing newlines
			content, err := c.convertInlineContent(child.Content)
			if err != nil {
				return "", err
			}
			if content != "" {
				parts = append(parts, content)
			}

		case "bulletList", "orderedList", "taskList":
			res, err := c.convertListInTable(child)
			if err != nil {
				return "", err
			}
			if res != "" {
				parts = append(parts, res)
			}

		case "codeBlock":
			res, err := c.convertCodeBlockInTable(child)
			if err != nil {
				return "", err
			}
			if res != "" {
				parts = append(parts, res)
			}

		case "panel":
			// Process panel
			panelContent, err := c.convertNode(child)
			if err != nil {
				return "", err
			}
			panelContent = strings.TrimRight(panelContent, "\n")
			if panelContent != "" {
				parts = append(parts, panelContent)
			}

		case "blockquote":
			// Process blockquote
			quoteContent, err := c.convertNode(child)
			if err != nil {
				return "", err
			}
			quoteContent = strings.TrimRight(quoteContent, "\n")
			if quoteContent != "" {
				parts = append(parts, quoteContent)
			}

		default:
			// For other node types, try to convert them
			content, err := c.convertNode(child)
			if err != nil {
				return "", err
			}
			content = strings.TrimRight(content, "\n")
			if content != "" {
				parts = append(parts, content)
			}
		}
	}

	// Join with <br> (or space) for multi-paragraph or block content
	sep := "<br>"
	if !c.config.AllowHTML {
		sep = " "
	}
	result := strings.Join(parts, sep)
	// Escape pipe characters as they break GFM tables
	// Note: Child converters must NOT pre-escape pipes as this would cause
	// double-escaping. Only this final output should escape pipes.
	return strings.ReplaceAll(result, "|", "\\|"), nil
}

func (c *Converter) convertListInTable(node Node) (string, error) {
	listContent, err := c.convertNode(node)
	if err != nil {
		return "", err
	}
	listContent = strings.TrimRight(listContent, "\n")
	if listContent == "" {
		return "", nil
	}

	// Split by newlines and join with <br>
	lines := strings.Split(listContent, "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimRight(line, "\n")
		if trimmed != "" {
			cleanLines = append(cleanLines, trimmed)
		}
	}
	if len(cleanLines) > 0 {
		sep := "<br>"
		if !c.config.AllowHTML {
			sep = " "
		}
		return strings.Join(cleanLines, sep), nil
	}
	return "", nil
}

func (c *Converter) convertCodeBlockInTable(node Node) (string, error) {
	rawCode := c.extractTextFromContent(node.Content)
	if strings.TrimSpace(rawCode) == "" {
		return "", nil
	}

	if c.config.AllowHTML {
		// Escape HTML special chars
		safeCode := strings.ReplaceAll(rawCode, "&", "&amp;")
		safeCode = strings.ReplaceAll(safeCode, "<", "&lt;")
		safeCode = strings.ReplaceAll(safeCode, ">", "&gt;")
		safeCode = strings.ReplaceAll(safeCode, "\"", "&quot;")
		// Replace newlines with <br>
		safeCode = strings.ReplaceAll(safeCode, "\n", "<br>")
		return "<code>" + safeCode + "</code>", nil
	} else {
		// Flatten and use backticks if HTML is not allowed
		flatCode := strings.ReplaceAll(rawCode, "\n", " ")
		return "`" + flatCode + "`", nil
	}
}
