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

	case "paragraph":
		// Process paragraph content with mark continuity
		return c.convertParagraphContent(node.Content)

	case "text":
		// Text nodes should be processed within paragraph context
		// This case handles standalone text (shouldn't normally occur)
		return node.Text, nil

	default:
		if c.config.Strict {
			return "", fmt.Errorf("unknown node type: %s", node.Type)
		}
		return fmt.Sprintf("[Unknown node: %s]", node.Type), nil
	}
}

// convertParagraphContent processes all content nodes in a paragraph
// while maintaining mark continuity across adjacent text nodes
func (c *Converter) convertParagraphContent(content []Node) (string, error) {
	var sb strings.Builder
	var activeMarks []string // Track currently active marks

	// Check if any text node has both strong and em anywhere in the paragraph
	useUnderscoreForEm := c.hasStrongAndEm(content)

	for _, node := range content {
		if node.Type != "text" {
			// For non-text nodes, close all active marks, process node, reopen marks
			for i := len(activeMarks) - 1; i >= 0; i-- {
				closing, err := c.getClosingDelimiter(activeMarks[i], useUnderscoreForEm)
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

		// Get sorted marks for this text node
		currentMarks := c.getSortedMarkTypes(node.Marks)

		// Find marks to close and open
		marksToClose := c.getMarksToClose(activeMarks, currentMarks)
		marksToOpen := c.getMarksToOpen(activeMarks, currentMarks)

		// Close marks (in reverse order)
		for i := len(marksToClose) - 1; i >= 0; i-- {
			closing, err := c.getClosingDelimiter(marksToClose[i], useUnderscoreForEm)
			if err != nil {
				return "", err
			}
			sb.WriteString(closing)
		}

		// Open new marks (in priority order)
		for _, mark := range marksToOpen {
			opening, err := c.getOpeningDelimiter(mark, useUnderscoreForEm)
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
		closing, err := c.getClosingDelimiter(activeMarks[i], useUnderscoreForEm)
		if err != nil {
			return "", err
		}
		sb.WriteString(closing)
	}

	// Standard paragraph has two newlines to separate from next block
	return sb.String() + "\n\n", nil
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

// getSortedMarkTypes returns mark types in JSON order
// The order in the JSON represents nesting: first = outermost
func (c *Converter) getSortedMarkTypes(marks []Mark) []string {
	if len(marks) == 0 {
		return nil
	}

	types := make([]string, len(marks))
	for i, mark := range marks {
		types[i] = mark.Type
	}

	return types
}

// getMarksToClose returns marks that need to be closed (in activeMarks but not in currentMarks)
func (c *Converter) getMarksToClose(activeMarks, currentMarks []string) []string {
	var toClose []string

	// Find the first mark that differs or is missing in currentMarks
	// We need to close from that point onward
	closeFromIndex := -1
	for i, activeMark := range activeMarks {
		if i >= len(currentMarks) || activeMark != currentMarks[i] {
			closeFromIndex = i
			break
		}
	}

	if closeFromIndex >= 0 {
		toClose = activeMarks[closeFromIndex:]
	}

	return toClose
}

// getMarksToOpen returns marks that need to be opened (in currentMarks but not in activeMarks)
func (c *Converter) getMarksToOpen(activeMarks, currentMarks []string) []string {
	// Find common prefix length
	commonLen := 0
	for i := 0; i < len(activeMarks) && i < len(currentMarks); i++ {
		if activeMarks[i] == currentMarks[i] {
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

// isKnownMark checks if a mark type is supported
func (c *Converter) isKnownMark(markType string) bool {
	switch markType {
	case "strong", "em", "strike", "code", "underline":
		return true
	default:
		return false
	}
}

// getOpeningDelimiter returns the opening delimiter for a mark type
func (c *Converter) getOpeningDelimiter(markType string, useUnderscoreForEm bool) (string, error) {
	prefix, _, err := c.convertMark(Mark{Type: markType}, useUnderscoreForEm)
	return prefix, err
}

// getClosingDelimiter returns the closing delimiter for a mark type
func (c *Converter) getClosingDelimiter(markType string, useUnderscoreForEm bool) (string, error) {
	_, suffix, err := c.convertMark(Mark{Type: markType}, useUnderscoreForEm)
	return suffix, err
}

// convertMark returns opening delimiter, closing delimiter, and error for a mark type
// In strict mode, unknown marks return an error
// In non-strict mode, unknown marks are silently ignored (text content preserved, formatting lost)
// Special case: underline uses HTML <u> tag when AllowHTML is enabled
func (c *Converter) convertMark(mark Mark, useUnderscoreForEm bool) (string, string, error) {
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
	default:
		if c.config.Strict {
			return "", "", fmt.Errorf("unknown mark type: %s", mark.Type)
		}
		// In non-strict mode, ignore unknown marks (preserve text, lose formatting)
		// This is acceptable for minor semantic marks like colors, etc.
		return "", "", nil
	}
}
