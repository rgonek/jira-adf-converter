package converter

import (
	"fmt"
	"html"
	"strings"
)

// convertParagraph converts a paragraph node to markdown
func (s *state) convertParagraph(node Node) (string, error) {
	// Process paragraph content with mark continuity.
	content, err := s.convertParagraphContent(node.Content)
	if err != nil {
		return "", err
	}
	if content == "" {
		return "", nil
	}

	if alignment := s.getNodeAlignment(node); alignment != "" {
		trimmed := strings.TrimSuffix(content, "\n\n")
		switch s.config.AlignmentStyle {
		case AlignHTML:
			return fmt.Sprintf(`<div align="%s">%s</div>`+"\n\n", alignment, trimmed), nil
		case AlignPandoc:
			return fmt.Sprintf(":::{ style=\"text-align: %s;\" }\n\n%s\n\n:::\n\n", alignment, trimmed), nil
		}
	}

	return content, nil
}

// convertText converts a text node (standalone, not within paragraph)
func (s *state) convertText(node Node) (string, error) {
	// Text nodes should be processed within paragraph context
	// This case handles standalone text (shouldn't normally occur)
	return node.Text, nil
}

// convertParagraphContent processes all content nodes in a paragraph
// while maintaining mark continuity across adjacent text nodes
func (s *state) convertParagraphContent(content []Node) (string, error) {
	res, err := s.convertInlineContent(content)
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
func (s *state) convertHeading(node Node) (string, error) {
	// Extract level from attributes (default to 1 if missing/invalid).
	level := node.GetIntAttr("level", 0)
	if level <= 0 {
		level = node.Level
	}
	if level <= 0 {
		level = 1
	}

	level += s.config.HeadingOffset

	// Clamp level to valid range (1-6)
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}

	// Process content with mark continuity
	content, err := s.convertInlineContent(node.Content)
	if err != nil {
		return "", err
	}

	if content == "" {
		return "", nil
	}
	// Edge case: if content ends with a hard break (backslash), remove it as headings don't support them at the end
	content = strings.TrimSuffix(content, "\\")

	// Build heading
	heading := strings.Repeat("#", level)
	if len(content) > 0 {
		heading += " " + content
	}

	if alignment := s.getNodeAlignment(node); alignment != "" {
		switch s.config.AlignmentStyle {
		case AlignHTML:
			return fmt.Sprintf(`<h%d align="%s">%s</h%d>`+"\n\n", level, alignment, content, level), nil
		case AlignPandoc:
			return fmt.Sprintf("%s {style=\"text-align: %s;\"}\n\n", heading, alignment), nil
		}
	}

	return heading + "\n\n", nil // Newline after heading + blank line after
}

// convertBlockquote converts a blockquote node to markdown
func (s *state) convertBlockquote(node Node) (string, error) {
	// Handle empty blockquote
	if len(node.Content) == 0 {
		return "", nil
	}

	// Process child content recursively
	sbStr, err := s.convertChildren(node.Content)
	if err != nil {
		return "", err
	}

	// Use blockquoteContent helper to apply formatting
	// We pass empty prefix since standard blockquotes don't have special prefixes like panels
	content := s.blockquoteContent(sbStr, "")

	return content + "\n\n", nil
}

// convertRule converts a horizontal rule node to markdown
func (s *state) convertRule() (string, error) {
	return "---\n\n", nil
}

// convertHardBreak converts a hard line break to markdown (backslash + newline)
func (s *state) convertHardBreak() (string, error) {
	if s.config.HardBreakStyle == HardBreakHTML {
		return "<br>", nil
	}
	return "\\\n", nil
}

func (s *state) getNodeAlignment(node Node) string {
	alignment := node.GetStringAttr("align", "")
	if alignment == "" {
		alignment = node.GetStringAttr("layout", "")
	}

	switch alignment {
	case "left", "center", "right":
		return alignment
	default:
		return ""
	}
}

// blockquoteContent converts content to blockquoted format with optional first-line prefix
func (s *state) blockquoteContent(content, firstLinePrefix string) string {
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

// extractTextFromContent extracts raw text from a list of nodes (shallow, mainly for code blocks)
func (s *state) extractTextFromContent(content []Node) string {
	var sb strings.Builder
	for _, child := range content {
		if child.Type == "text" {
			sb.WriteString(child.Text)
		}
	}
	return sb.String()
}

// convertCodeBlock converts a code block node to markdown
func (s *state) convertCodeBlock(node Node) (string, error) {
	// Check if code block is empty or contains only whitespace
	if len(node.Content) == 0 {
		return "", nil
	}

	content := s.extractTextFromContent(node.Content)

	// If only whitespace, ignore per Core Principle #5
	if strings.TrimSpace(content) == "" {
		return "", nil
	}

	// Extract language attribute
	language := node.GetStringAttr("language", "")
	if mapped, ok := s.config.LanguageMap[language]; ok {
		language = mapped
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
func (s *state) indent(content, marker string) string {
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
func (s *state) convertPanel(node Node) (string, error) {
	// Handle empty panel
	if len(node.Content) == 0 {
		return "", nil
	}

	// Check if panel has actual content or just whitespace
	fullContent, err := s.convertChildren(node.Content)
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(fullContent) == "" {
		return "", nil
	}

	panelType := strings.ToLower(node.GetStringAttr("panelType", ""))
	hasPanelType := panelType != ""
	panelTitle := node.GetStringAttr("title", "")
	panelUpper, panelTitleCase := panelTypeLabels(panelType)

	switch s.config.PanelStyle {
	case PanelNone:
		quoted := s.blockquoteContent(fullContent, "")
		if quoted == "" {
			return "", nil
		}
		return quoted + "\n\n", nil
	case PanelBold:
		prefix := ""
		if hasPanelType {
			prefix = fmt.Sprintf("**%s**: ", panelTitleCase)
		}
		quoted := s.blockquoteContent(fullContent, prefix)
		if quoted == "" {
			return "", nil
		}
		return quoted + "\n\n", nil
	case PanelTitle:
		if !hasPanelType {
			quoted := s.blockquoteContent(fullContent, "")
			if quoted == "" {
				return "", nil
			}
			return quoted + "\n\n", nil
		}
		callout := fmt.Sprintf("[!%s]", panelUpper)
		if panelTitle != "" {
			callout = fmt.Sprintf("[!%s: %s]", panelUpper, panelTitle)
		}
		quoted := s.blockquoteContent(fullContent, "")
		if quoted == "" {
			return "> " + callout + "\n\n", nil
		}
		return "> " + callout + "\n" + quoted + "\n\n", nil
	default: // PanelGitHub
		if !hasPanelType {
			quoted := s.blockquoteContent(fullContent, "")
			if quoted == "" {
				return "", nil
			}
			return quoted + "\n\n", nil
		}
		callout := fmt.Sprintf("[!%s]", panelUpper)
		quoted := s.blockquoteContent(fullContent, "")
		if quoted == "" {
			return "> " + callout + "\n\n", nil
		}
		return "> " + callout + "\n" + quoted + "\n\n", nil
	}
}

// convertDecisionList converts a decision list to a single continuous blockquote
func (s *state) convertDecisionList(node Node) (string, error) {
	if len(node.Content) == 0 {
		return "", nil
	}

	var items []string
	for _, child := range node.Content {
		if child.Type != "decisionItem" {
			continue
		}

		itemContent, err := s.convertDecisionItemContent(child)
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
func (s *state) convertDecisionItem(node Node) (string, error) {
	return s.convertDecisionItemContent(node)
}

// convertDecisionItemContent processes a decision item's content
func (s *state) convertDecisionItemContent(node Node) (string, error) {
	if len(node.Content) == 0 {
		return "", nil
	}

	// Get decision state
	state := node.GetStringAttr("state", "")

	// Map state to prefix
	prefix := ""
	switch state {
	case "DECIDED":
		if s.config.DecisionStyle == DecisionText {
			prefix = "**DECIDED**: "
		} else {
			prefix = "**✓ Decision**: "
		}
	case "UNDECIDED":
		if s.config.DecisionStyle == DecisionText {
			prefix = "**UNDECIDED**: "
		} else {
			prefix = "**? Decision**: "
		}
	default:
		if s.config.DecisionStyle == DecisionText {
			prefix = "**DECISION**: "
		} else {
			prefix = "**Decision**: "
		}
	}

	// Process content
	sbStr, err := s.convertChildren(node.Content)
	if err != nil {
		return "", err
	}

	quoted := s.blockquoteContent(sbStr, prefix)
	if quoted == "" {
		return "", nil
	}

	return quoted, nil
}

func panelTypeLabels(panelType string) (string, string) {
	switch panelType {
	case "info":
		return "INFO", "Info"
	case "note":
		return "NOTE", "Note"
	case "success":
		return "SUCCESS", "Success"
	case "warning":
		return "WARNING", "Warning"
	case "error":
		return "ERROR", "Error"
	default:
		upper := strings.ToUpper(panelType)
		if upper == "" {
			upper = "INFO"
		}
		titleCase := panelType
		if titleCase == "" {
			titleCase = "Info"
		} else {
			titleCase = strings.ToUpper(titleCase[:1]) + strings.ToLower(titleCase[1:])
		}
		return upper, titleCase
	}
}

// convertExpand converts expand and nestedExpand nodes
func (s *state) convertExpand(node Node) (string, error) {
	// Extract title
	title := node.GetStringAttr("title", "")

	// Process content
	content, err := s.convertChildren(node.Content)
	if err != nil {
		return "", err
	}

	if s.config.ExpandStyle == ExpandHTML {
		var htmlBuilder strings.Builder
		htmlBuilder.WriteString("<details><summary>")
		htmlBuilder.WriteString(html.EscapeString(title))
		htmlBuilder.WriteString("</summary>\n\n")
		htmlBuilder.WriteString(strings.TrimRight(content, "\n"))
		htmlBuilder.WriteString("\n\n</details>\n\n")
		return htmlBuilder.String(), nil
	}
	if s.config.ExpandStyle == ExpandPandoc {
		var pandocBuilder strings.Builder
		pandocBuilder.WriteString(":::{ .details")
		if title != "" {
			escapedTitle := strings.ReplaceAll(title, "\\", "\\\\")
			escapedTitle = strings.ReplaceAll(escapedTitle, "\"", "\\\"")
			pandocBuilder.WriteString(fmt.Sprintf(` summary="%s"`, escapedTitle))
		}
		pandocBuilder.WriteString(" }\n\n")
		pandocBuilder.WriteString(strings.TrimRight(content, "\n"))
		pandocBuilder.WriteString("\n\n:::\n\n")
		return pandocBuilder.String(), nil
	}

	// Text Mode: Blockquote with bold title
	var text strings.Builder
	if title != "" {
		text.WriteString("> **" + title + "**\n> \n")
	}

	// Handle empty content - still need trailing newlines for block separation
	if content == "" {
		return text.String() + "\n\n", nil
	}
	// Apply blockquote to content
	// If title exists, we've already written the header, so we prefix content with just "> "
	quotedContent := s.blockquoteContent(content, "")
	text.WriteString(quotedContent)

	return text.String() + "\n\n", nil
}

// convertLayoutSection converts a layout section node
func (s *state) convertLayoutSection(node Node) (string, error) {
	if len(node.Content) == 0 {
		return "", nil
	}

	content, err := s.convertChildren(node.Content)
	if err != nil {
		return "", err
	}

	if s.config.LayoutSectionStyle == LayoutSectionHTML {
		return "<div class=\"layout-section\">\n\n" + content + "</div>\n\n", nil
	}

	if s.config.LayoutSectionStyle == LayoutSectionPandoc {
		return "::::{ .layoutSection }\n" + content + "::::\n\n", nil
	}

	// Default Standard (Lossy) strategy
	return content, nil
}

// convertLayoutColumn converts a layout column node
func (s *state) convertLayoutColumn(node Node) (string, error) {
	if len(node.Content) == 0 {
		return "", nil
	}

	content, err := s.convertChildren(node.Content)
	if err != nil {
		return "", err
	}

	if s.config.LayoutSectionStyle == LayoutSectionHTML {
		width := node.GetFloat64Attr("width", 0)
		if width > 0 {
			formattedWidth := fmt.Sprintf("%f", width)
			formattedWidth = strings.TrimRight(strings.TrimRight(formattedWidth, "0"), ".")
			return fmt.Sprintf("<div class=\"layout-column\" style=\"width: %s%%;\">\n\n%s\n</div>\n\n", formattedWidth, strings.TrimRight(content, "\n")), nil
		}
		return "<div class=\"layout-column\">\n\n" + strings.TrimRight(content, "\n") + "\n</div>\n\n", nil
	}

	if s.config.LayoutSectionStyle == LayoutSectionPandoc {
		width := node.GetFloat64Attr("width", 0)
		if width > 0 {
			formattedWidth := fmt.Sprintf("%f", width)
			formattedWidth = strings.TrimRight(strings.TrimRight(formattedWidth, "0"), ".")
			return fmt.Sprintf(":::{ .layoutColumn width=\"%s%%\" }\n\n%s\n\n:::\n\n", formattedWidth, strings.TrimRight(content, "\n")), nil
		}
		return ":::{ .layoutColumn }\n\n" + strings.TrimRight(content, "\n") + "\n\n:::\n\n", nil
	}

	// Default Standard (Lossy) strategy
	return content, nil
}
