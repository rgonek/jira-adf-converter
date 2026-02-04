package converter

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Config holds converter configuration
type Config struct {
	AllowHTML bool // If true, use HTML for unsupported features
	Strict    bool // If true, return error on unknown nodes
}

// Converter converts ADF to GFM
type Converter struct {
	config Config
}

// New creates a new Converter with the given config
func New(config Config) *Converter {
	return &Converter{
		config: config,
	}
}

// Convert takes an ADF JSON document and returns GFM markdown
func (c *Converter) Convert(input []byte) (string, error) {
	var doc Doc
	if err := json.Unmarshal(input, &doc); err != nil {
		return "", fmt.Errorf("failed to parse ADF JSON: %w", err)
	}

	return c.convertNode(Node{Type: doc.Type, Content: doc.Content})
}

func (c *Converter) convertNode(node Node) (string, error) {
	switch node.Type {
	case "doc":
		return c.convertDoc(node)

	case "paragraph":
		return c.convertParagraph(node)

	case "heading":
		return c.convertHeading(node)

	case "blockquote":
		return c.convertBlockquote(node)

	case "rule":
		return c.convertRule()

	case "hardBreak":
		return c.convertHardBreak()

	case "text":
		return c.convertText(node)

	default:
		if c.config.Strict {
			return "", fmt.Errorf("unknown node type: %s", node.Type)
		}
		return fmt.Sprintf("[Unknown node: %s]", node.Type), nil
	}
}

// convertDoc converts the root document node
func (c *Converter) convertDoc(node Node) (string, error) {
	var sb strings.Builder
	for _, child := range node.Content {
		res, err := c.convertNode(child)
		if err != nil {
			return "", err
		}
		sb.WriteString(res)
	}
	// Trim right to avoid excessive newlines at the end of file, then ensure exactly one.
	return strings.TrimRight(sb.String(), "\n") + "\n", nil
}

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
	var sb strings.Builder
	var activeMarks []Mark // Track currently active marks (full Mark objects)

	// Check if any text node has both strong and em anywhere in the paragraph
	useUnderscoreForEm := c.hasStrongAndEm(content)

	for _, node := range content {
		if node.Type != "text" {
			// For non-text nodes, close all active marks, process node, reset marks
			for i := len(activeMarks) - 1; i >= 0; i-- {
				closing, err := c.getClosingDelimiterForMark(activeMarks[i], useUnderscoreForEm)
				if err != nil {
					return "", err
				}
				sb.WriteString(closing)
			}
			result, err := c.convertNode(node)
			if err != nil {
				return "", err
			}
			sb.WriteString(result)
			activeMarks = nil
			continue
		}

		// Validate marks in strict mode
		if c.config.Strict {
			for _, mark := range node.Marks {
				if !c.isKnownMark(mark.Type) {
					return "", fmt.Errorf("unknown mark type: %s", mark.Type)
				}
			}
		}

		// Get marks for this text node
		currentMarks := node.Marks

		// Find marks to close and open
		marksToClose := c.getMarksToCloseFull(activeMarks, currentMarks)
		marksToOpen := c.getMarksToOpenFull(activeMarks, currentMarks)

		// Close marks (in reverse order)
		for i := len(marksToClose) - 1; i >= 0; i-- {
			closing, err := c.getClosingDelimiterForMark(marksToClose[i], useUnderscoreForEm)
			if err != nil {
				return "", err
			}
			sb.WriteString(closing)
		}

		// Open new marks (in priority order)
		for _, mark := range marksToOpen {
			opening, err := c.getOpeningDelimiterForMark(mark, useUnderscoreForEm)
			if err != nil {
				return "", err
			}
			sb.WriteString(opening)
		}

		// Write text content
		sb.WriteString(node.Text)

		// Update active marks
		activeMarks = currentMarks
	}

	// Close any remaining marks at end of paragraph
	for i := len(activeMarks) - 1; i >= 0; i-- {
		closing, err := c.getClosingDelimiterForMark(activeMarks[i], useUnderscoreForEm)
		if err != nil {
			return "", err
		}
		sb.WriteString(closing)
	}

	// Standard paragraph has two newlines to separate from next block
	res := sb.String()
	if res == "" {
		return "", nil
	}
	return res + "\n\n", nil
}

// hasStrongAndEm checks if any text node in content has both strong and em marks
func (c *Converter) hasStrongAndEm(content []Node) bool {
	for _, node := range content {
		if node.Type != "text" {
			continue
		}
		hasStrong := false
		hasEm := false
		for _, mark := range node.Marks {
			if mark.Type == "strong" {
				hasStrong = true
			}
			if mark.Type == "em" {
				hasEm = true
			}
		}
		if hasStrong && hasEm {
			return true
		}
	}
	return false
}

// getMarksToCloseFull returns marks that need to be closed
func (c *Converter) getMarksToCloseFull(activeMarks, currentMarks []Mark) []Mark {
	// Find the first mark that differs or is missing in currentMarks
	closeFromIndex := -1
	for i, activeMark := range activeMarks {
		if i >= len(currentMarks) || !c.marksEqual(activeMark, currentMarks[i]) {
			closeFromIndex = i
			break
		}
	}

	if closeFromIndex >= 0 {
		return activeMarks[closeFromIndex:]
	}

	return nil
}

// getMarksToOpenFull returns marks that need to be opened
func (c *Converter) getMarksToOpenFull(activeMarks, currentMarks []Mark) []Mark {
	// Find common prefix length
	commonLen := 0
	for i := 0; i < len(activeMarks) && i < len(currentMarks); i++ {
		if c.marksEqual(activeMarks[i], currentMarks[i]) {
			commonLen++
		} else {
			break
		}
	}

	// Return marks after common prefix
	if commonLen < len(currentMarks) {
		return currentMarks[commonLen:]
	}
	return nil
}

// marksEqual compares two marks for equality
// For marks with attributes (link, subsup), it also compares the attributes
func (c *Converter) marksEqual(m1, m2 Mark) bool {
	// Type must match
	if m1.Type != m2.Type {
		return false
	}

	// For marks with attributes, compare them as well
	switch m1.Type {
	case "link":
		return c.markAttrsEqual(m1.Attrs, m2.Attrs, []string{"href", "title"})
	case "subsup":
		return c.markAttrsEqual(m1.Attrs, m2.Attrs, []string{"type"})
	}

	return true
}

// markAttrsEqual compares specific attributes between two marks
func (c *Converter) markAttrsEqual(attrs1, attrs2 map[string]any, keys []string) bool {
	for _, key := range keys {
		val1, has1 := attrs1[key]
		val2, has2 := attrs2[key]
		if has1 != has2 {
			return false
		}
		if has1 && val1 != val2 {
			return false
		}
	}
	return true
}

// isKnownMark checks if a mark type is supported
func (c *Converter) isKnownMark(markType string) bool {
	switch markType {
	case "strong", "em", "strike", "code", "underline", "link", "subsup":
		return true
	default:
		return false
	}
}

// getOpeningDelimiterForMark returns the opening delimiter for a mark
func (c *Converter) getOpeningDelimiterForMark(mark Mark, useUnderscoreForEm bool) (string, error) {
	prefix, _, err := c.convertMarkFull(mark, useUnderscoreForEm)
	return prefix, err
}

// getClosingDelimiterForMark returns the closing delimiter for a mark
func (c *Converter) getClosingDelimiterForMark(mark Mark, useUnderscoreForEm bool) (string, error) {
	_, suffix, err := c.convertMarkFull(mark, useUnderscoreForEm)
	return suffix, err
}

// convertMarkFull returns opening delimiter, closing delimiter, and error for a mark
func (c *Converter) convertMarkFull(mark Mark, useUnderscoreForEm bool) (string, string, error) {
	switch mark.Type {
	case "strong":
		return "**", "**", nil
	case "em":
		if useUnderscoreForEm {
			return "_", "_", nil
		}
		return "*", "*", nil
	case "strike":
		return "~~", "~~", nil
	case "code":
		return "`", "`", nil
	case "underline":
		if c.config.AllowHTML {
			return "<u>", "</u>", nil
		}
		// In non-HTML mode, silently drop underline formatting
		return "", "", nil
	case "link":
		// Extract href and title from attrs
		if mark.Attrs == nil {
			// No attributes - just return plain text
			return "", "", nil
		}
		href, hasHref := mark.Attrs["href"].(string)
		if !hasHref || href == "" {
			// No href - just return plain text
			return "", "", nil
		}

		// Build link syntax: [text](href) or [text](href "title")
		opening := "["
		closing := "](" + href

		if title, hasTitle := mark.Attrs["title"].(string); hasTitle && title != "" {
			// Escape quotes in title
			escapedTitle := strings.ReplaceAll(title, "\\", "\\\\")
			escapedTitle = strings.ReplaceAll(escapedTitle, "\"", "\\\"")
			closing += " \"" + escapedTitle + "\""
		}
		closing += ")"

		return opening, closing, nil
	case "subsup":
		// Extract sub/sup type from attrs
		if mark.Attrs == nil {
			return "", "", nil
		}
		subSupType, ok := mark.Attrs["type"].(string)
		if !ok {
			return "", "", nil
		}

		if c.config.AllowHTML {
			if subSupType == "sub" {
				return "<sub>", "</sub>", nil
			} else if subSupType == "sup" {
				return "<sup>", "</sup>", nil
			}
		} else {
			// Plain mode
			if subSupType == "sup" {
				return "^", "", nil
			} else if subSupType == "sub" {
				// Subscript in plain mode: just plain text (no indicator)
				return "", "", nil
			}
		}
		return "", "", nil
	default:
		if c.config.Strict {
			return "", "", fmt.Errorf("unknown mark type: %s", mark.Type)
		}
		// In non-strict mode, ignore unknown marks (preserve text, lose formatting)
		// This is acceptable for minor semantic marks like colors, etc.
		return "", "", nil
	}
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
	content, err := c.convertParagraphContent(node.Content)
	if err != nil {
		return "", err
	}

	// Remove trailing newlines from paragraph content (we'll add our own spacing)
	content = strings.TrimRight(content, "\n")
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
