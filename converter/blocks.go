package converter

import (
	"strings"
)

// convertParagraph converts a paragraph node to markdown
func (c *Converter) convertParagraph(node Node) (string, error) {
	// Process paragraph content with mark continuity
	return c.convertParagraphContent(node.Content)
}

// convertText converts a text node (standalone, not within paragraph)
func (c *Converter) convertText(node Node) (string, error) {
	// Text nodes should be processed within paragraph context
	// This case handles standalone text (shouldn't normally occur)
	return node.Text, nil
}

// convertParagraphContent processes all content nodes in a paragraph
// while maintaining mark continuity across adjacent text nodes
func (c *Converter) convertParagraphContent(content []Node) (string, error) {
	res, err := c.convertInlineContent(content)
	if err != nil {
		return "", err
	}
	if res == "" {
		return "", nil
	}
	// Standard paragraph has two newlines to separate from next block
	return res + "\n\n", nil
}

// convertHeading converts a heading node to markdown
func (c *Converter) convertHeading(node Node) (string, error) {
	// Extract level from attributes (default to 1 if missing/invalid)
	level := 1
	if node.Attrs != nil {
		if lvl, ok := node.Attrs["level"].(float64); ok {
			level = int(lvl)
		}
	}

	// Clamp level to valid range (1-6)
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}

	// Process content with mark continuity
	content, err := c.convertInlineContent(node.Content)
	if err != nil {
		return "", err
	}

	if content == "" {
		return "", nil
	}
	// Edge case: if content ends with a hard break (backslash), remove it as headings don't support them at the end
	content = strings.TrimSuffix(content, "\\")

	// Build heading
	var sb strings.Builder
	sb.WriteString(strings.Repeat("#", level))
	if len(content) > 0 {
		sb.WriteString(" ")
		sb.WriteString(content)
	}
	sb.WriteString("\n\n") // Newline after heading + blank line after

	return sb.String(), nil
}

// convertBlockquote converts a blockquote node to markdown
func (c *Converter) convertBlockquote(node Node) (string, error) {
	// Handle empty blockquote
	if len(node.Content) == 0 {
		return "", nil
	}

	// Process child content recursively
	var sb strings.Builder
	for _, child := range node.Content {
		result, err := c.convertNode(child)
		if err != nil {
			return "", err
		}
		sb.WriteString(result)
	}

	// Get the content and prefix every line with ">"
	content := strings.TrimRight(sb.String(), "\n")
	lines := strings.Split(content, "\n")

	var quotedLines []string
	for _, line := range lines {
		// If line already starts with ">", don't add a space (for nested blockquotes)
		if strings.HasPrefix(line, ">") {
			quotedLines = append(quotedLines, ">"+line)
		} else {
			quotedLines = append(quotedLines, "> "+line)
		}
	}

	return strings.Join(quotedLines, "\n") + "\n\n", nil
}

// convertRule converts a horizontal rule node to markdown
func (c *Converter) convertRule() (string, error) {
	return "---\n\n", nil
}

// convertHardBreak converts a hard line break to markdown (backslash + newline)
func (c *Converter) convertHardBreak() (string, error) {
	return "\\\n", nil
}

// blockquoteContent converts content to blockquoted format with optional first-line prefix
func (c *Converter) blockquoteContent(content, firstLinePrefix string) string {
	content = strings.TrimRight(content, "\n")
	if content == "" {
		return ""
	}

	lines := strings.Split(content, "\n")

	var quotedLines []string
	for i, line := range lines {
		if i == 0 && firstLinePrefix != "" {
			// First line gets the prefix
			quotedLines = append(quotedLines, "> "+firstLinePrefix+line)
		} else {
			// Subsequent lines
			if line == "" {
				quotedLines = append(quotedLines, "> ")
			} else if strings.HasPrefix(line, ">") {
				// Already a blockquote (nested)
				quotedLines = append(quotedLines, ">"+line)
			} else {
				quotedLines = append(quotedLines, "> "+line)
			}
		}
	}

	return strings.Join(quotedLines, "\n")
}

// convertCodeBlock converts a code block node to markdown
func (c *Converter) convertCodeBlock(node Node) (string, error) {
	// Check if code block is empty or contains only whitespace
	if len(node.Content) == 0 {
		return "", nil
	}

	var sb strings.Builder
	for _, child := range node.Content {
		if child.Type == "text" {
			sb.WriteString(child.Text)
		}
	}

	content := sb.String()
	// If only whitespace, ignore per Core Principle #5
	if strings.TrimSpace(content) == "" {
		return "", nil
	}

	// Extract language attribute
	language := ""
	if node.Attrs != nil {
		if lang, ok := node.Attrs["language"].(string); ok {
			language = lang
		}
	}

	var result strings.Builder
	result.WriteString("```")
	result.WriteString(language)
	result.WriteString("\n")
	result.WriteString(strings.TrimRight(content, "\n"))
	result.WriteString("\n```\n\n")

	return result.String(), nil
}

// indent applies uniform indentation to content within a list item.
// The first line is prefixed with the marker, subsequent lines with spaces matching marker length.
func (c *Converter) indent(content, marker string) string {
	content = strings.TrimRight(content, "\n")
	if content == "" {
		return ""
	}

	lines := strings.Split(content, "\n")
	indentStr := strings.Repeat(" ", len(marker))

	var result []string
	for i, line := range lines {
		if i == 0 {
			result = append(result, marker+line)
		} else {
			if line != "" {
				result = append(result, indentStr+line)
			} else {
				result = append(result, "")
			}
		}
	}

	return strings.Join(result, "\n")
}

// convertPanel converts a panel node to blockquote with semantic label
func (c *Converter) convertPanel(node Node) (string, error) {
	// Handle empty panel
	if len(node.Content) == 0 {
		return "", nil
	}

	// Check if panel has actual content or just whitespace
	var sb strings.Builder
	for _, child := range node.Content {
		result, err := c.convertNode(child)
		if err != nil {
			return "", err
		}
		sb.WriteString(result)
	}

	fullContent := sb.String()
	if strings.TrimSpace(fullContent) == "" {
		return "", nil
	}

	// Get panel type
	panelType := ""
	if node.Attrs != nil {
		if pt, ok := node.Attrs["panelType"].(string); ok {
			panelType = pt
		}
	}

	// Map panel type to prefix
	prefix := ""
	switch panelType {
	case "info":
		prefix = "**Info**: "
	case "note":
		prefix = "**Note**: "
	case "success":
		prefix = "**Success**: "
	case "warning":
		prefix = "**Warning**: "
	case "error":
		prefix = "**Error**: "
	}

	quoted := c.blockquoteContent(fullContent, prefix)
	if quoted == "" {
		return "", nil
	}

	return quoted + "\n\n", nil
}

// convertDecisionList converts a decision list to a single continuous blockquote
func (c *Converter) convertDecisionList(node Node) (string, error) {
	if len(node.Content) == 0 {
		return "", nil
	}

	var items []string
	for _, child := range node.Content {
		if child.Type != "decisionItem" {
			continue
		}

		itemContent, err := c.convertDecisionItemContent(child)
		if err != nil {
			return "", err
		}
		if itemContent != "" {
			items = append(items, itemContent)
		}
	}

	if len(items) == 0 {
		return "", nil
	}

	// Join items with blank quoted line
	result := strings.Join(items, "\n> \n")
	return result + "\n\n", nil
}

// convertDecisionItem is a helper that should not be called directly
func (c *Converter) convertDecisionItem(node Node) (string, error) {
	return c.convertDecisionItemContent(node)
}

// convertDecisionItemContent processes a decision item's content
func (c *Converter) convertDecisionItemContent(node Node) (string, error) {
	if len(node.Content) == 0 {
		return "", nil
	}

	// Get decision state
	state := ""
	if node.Attrs != nil {
		if s, ok := node.Attrs["state"].(string); ok {
			state = s
		}
	}

	// Map state to prefix
	prefix := ""
	switch state {
	case "DECIDED":
		prefix = "**✓ Decision**: "
	case "UNDECIDED":
		prefix = "**? Decision**: "
	default:
		prefix = "**Decision**: "
	}

	// Process content
	var sb strings.Builder
	for _, child := range node.Content {
		result, err := c.convertNode(child)
		if err != nil {
			return "", err
		}
		sb.WriteString(result)
	}

	quoted := c.blockquoteContent(sb.String(), prefix)
	if quoted == "" {
		return "", nil
	}

	return quoted, nil
}
