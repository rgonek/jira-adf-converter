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

	case "codeBlock":
		return c.convertCodeBlock(node)

	case "bulletList":
		return c.convertBulletList(node)

	case "orderedList":
		return c.convertOrderedList(node)

	case "taskList":
		return c.convertTaskList(node)

	case "taskItem":
		return c.convertTaskItem(node)

	case "listItem":
		return c.convertListItem(node)

	case "text":
		return c.convertText(node)

	case "emoji":
		return c.convertEmoji(node)

	case "mention":
		return c.convertMention(node)

	case "status":
		return c.convertStatus(node)

	case "date":
		return c.convertDate(node)

	case "inlineCard":
		return c.convertInlineCard(node)

	case "table":
		return c.convertTable(node)

	case "tableRow":
		// Table rows are processed within convertTable, not standalone
		return "", nil

	case "tableHeader":
		return c.convertTableCell(node, true)

	case "tableCell":
		return c.convertTableCell(node, false)

	case "panel":
		return c.convertPanel(node)

	case "expand", "nestedExpand":
		return c.convertExpand(node)

	case "mediaSingle":
		return c.convertMediaSingle(node)

	case "mediaGroup":
		return c.convertMediaGroup(node)

	case "media":
		return c.convertMedia(node)

	case "decisionList":
		return c.convertDecisionList(node)

	case "decisionItem":
		return c.convertDecisionItem(node)

	default:
		if c.config.Strict {
			return "", fmt.Errorf("unknown node type: %s", node.Type)
		}
		return fmt.Sprintf("[Unknown node: %s]", node.Type), nil
	}
}

// convertChildren processes a slice of nodes and concatenates their results
func (c *Converter) convertChildren(content []Node) (string, error) {
	var sb strings.Builder
	for _, child := range content {
		res, err := c.convertNode(child)
		if err != nil {
			return "", err
		}
		sb.WriteString(res)
	}
	return sb.String(), nil
}

// convertDoc converts the root document node
func (c *Converter) convertDoc(node Node) (string, error) {
	res, err := c.convertChildren(node.Content)
	if err != nil {
		return "", err
	}
	// Trim right to avoid excessive newlines at the end of file, then ensure exactly one.
	result := strings.TrimRight(res, "\n")
	if result == "" {
		return "", nil
	}
	return result + "\n", nil
}

// convertInlineContent processes a slice of nodes (typically text with marks)
// and returns the markdown string without trailing block-level newlines.
func (c *Converter) convertInlineContent(content []Node) (string, error) {
	var sb strings.Builder
	var activeMarks []Mark // Track currently active marks (full Mark objects)

	// Check if any text node has both strong and em anywhere in the paragraph
	useUnderscoreForEm := c.hasStrongAndEm(content)

	for _, node := range content {
		if node.Type != "text" {
			// For non-text nodes, close all active marks, process node, reset marks
			if err := c.closeMarks(activeMarks, useUnderscoreForEm, &sb); err != nil {
				return "", err
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

		// Special handling for whitespace-only nodes to avoid "stupid" markdown like ** **
		// or marks starting/ending on whitespace.
		effectiveMarks := currentMarks
		if strings.TrimSpace(node.Text) == "" {
			// For whitespace-only nodes, we don't want to open new marks.
			// We only keep marks that were already active.
			effectiveMarks = c.intersectMarks(activeMarks, currentMarks)
		}

		// Find marks to close and open
		marksToClose := c.getMarksToCloseFull(activeMarks, effectiveMarks)
		marksToOpen := c.getMarksToOpenFull(activeMarks, effectiveMarks)

		// Close marks
		if err := c.closeMarks(marksToClose, useUnderscoreForEm, &sb); err != nil {
			return "", err
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
		activeMarks = effectiveMarks
	}

	// Close any remaining marks at end of content
	if err := c.closeMarks(activeMarks, useUnderscoreForEm, &sb); err != nil {
		return "", err
	}

	return sb.String(), nil
}

// closeMarks closes the provided marks in reverse order
func (c *Converter) closeMarks(marks []Mark, useUnderscoreForEm bool, sb *strings.Builder) error {
	for i := len(marks) - 1; i >= 0; i-- {
		closing, err := c.getClosingDelimiterForMark(marks[i], useUnderscoreForEm)
		if err != nil {
			return err
		}
		sb.WriteString(closing)
	}
	return nil
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
