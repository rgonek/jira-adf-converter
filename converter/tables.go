package converter

import (
	"strings"
)

// convertTable converts a table node to GFM table
func (c *Converter) convertTable(node Node) (string, error) {
	if len(node.Content) == 0 {
		return "", nil
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
				return "", err
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
		return "", nil
	}

	// Determine column count
	colCount := len(rows[0])
	for _, row := range rows {
		if len(row) > colCount {
			colCount = len(row)
		}
	}

	var sb strings.Builder

	// Prepare header row and data rows
	var headerRow []string
	var dataRows [][]string

	if hasHeader {
		headerRow = rows[0]
		dataRows = rows[1:]
	} else {
		// Create empty header row
		headerRow = make([]string, colCount)
		for i := 0; i < colCount; i++ {
			headerRow[i] = ""
		}
		dataRows = rows
	}

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
	return sb.String(), nil
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
			// Process list and convert to single line with <br> between items
			listContent, err := c.convertNode(child)
			if err != nil {
				return "", err
			}
			listContent = strings.TrimRight(listContent, "\n")
			if listContent != "" {
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
					parts = append(parts, strings.Join(cleanLines, sep))
				}
			}

		case "codeBlock":
			// Process code block
			// Tables don't support fenced code blocks, so we must handle this manually
			var sb strings.Builder
			for _, grandChild := range child.Content {
				if grandChild.Type == "text" {
					sb.WriteString(grandChild.Text)
				}
			}
			rawCode := sb.String()
			if strings.TrimSpace(rawCode) == "" {
				continue
			}

			if c.config.AllowHTML {
				// Escape HTML special chars
				safeCode := strings.ReplaceAll(rawCode, "&", "&amp;")
				safeCode = strings.ReplaceAll(safeCode, "<", "&lt;")
				safeCode = strings.ReplaceAll(safeCode, ">", "&gt;")
				safeCode = strings.ReplaceAll(safeCode, "\"", "&quot;")
				// Replace newlines with <br>
				safeCode = strings.ReplaceAll(safeCode, "\n", "<br>")
				parts = append(parts, "<code>"+safeCode+"</code>")
			} else {
				// Flatten and use backticks if HTML is not allowed
				flatCode := strings.ReplaceAll(rawCode, "\n", " ")
				parts = append(parts, "`"+flatCode+"`")
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
