package converter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Converter converts ADF to GFM
type Converter struct {
	config Config
}

// state holds the per-conversion state, making the converter thread-safe.
type state struct {
	config   Config
	ctx      context.Context
	options  ConvertOptions
	warnings []Warning
}

// New creates a new Converter with the given config
func New(config Config) (*Converter, error) {
	cfg := config.applyDefaults().clone()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Converter{
		config: cfg,
	}, nil
}

// Convert takes an ADF JSON document and returns GFM markdown
func (c *Converter) Convert(input []byte) (Result, error) {
	return c.ConvertWithContext(context.Background(), input, ConvertOptions{})
}

// ConvertWithContext takes an ADF JSON document and returns GFM markdown.
func (c *Converter) ConvertWithContext(ctx context.Context, input []byte, opts ConvertOptions) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	var doc Doc
	if err := json.Unmarshal(input, &doc); err != nil {
		return Result{}, fmt.Errorf("failed to parse ADF JSON: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	s := &state{
		config:  c.config,
		ctx:     ctx,
		options: opts,
	}

	markdown, err := s.convertNode(Node{Type: doc.Type, Content: doc.Content})
	if err != nil {
		return Result{}, err
	}
	if err := s.checkContext(); err != nil {
		return Result{}, err
	}

	return Result{Markdown: markdown, Warnings: s.warnings}, nil
}

func (s *state) convertNode(node Node) (string, error) {
	if err := s.checkContext(); err != nil {
		return "", err
	}

	switch node.Type {
	case "doc":
		return s.convertDoc(node)

	case "paragraph":
		return s.convertParagraph(node)

	case "heading":
		return s.convertHeading(node)

	case "blockquote":
		return s.convertBlockquote(node)

	case "rule":
		return s.convertRule()

	case "hardBreak":
		return s.convertHardBreak()

	case "codeBlock":
		return s.convertCodeBlock(node)

	case "bulletList":
		return s.convertBulletList(node)

	case "orderedList":
		return s.convertOrderedList(node)

	case "taskList":
		return s.convertTaskList(node)

	case "taskItem":
		return s.convertTaskItem(node)

	case "listItem":
		return s.convertListItem(node)

	case "text":
		return s.convertText(node)

	case "emoji":
		return s.convertEmoji(node)

	case "mention":
		return s.convertMention(node)

	case "status":
		return s.convertStatus(node)

	case "date":
		return s.convertDate(node)

	case "inlineCard":
		return s.convertInlineCard(node)

	case "table":
		return s.convertTable(node)

	case "tableRow":
		// Table rows are processed within convertTable, not standalone
		return "", nil

	case "tableHeader":
		return s.convertTableCell(node, true)

	case "tableCell":
		return s.convertTableCell(node, false)

	case "panel":
		return s.convertPanel(node)

	case "expand", "nestedExpand":
		return s.convertExpand(node)

	case "mediaSingle":
		return s.convertMediaSingle(node)

	case "mediaGroup":
		return s.convertMediaGroup(node)

	case "media":
		return s.convertMedia(node)

	case "decisionList":
		return s.convertDecisionList(node)

	case "decisionItem":
		return s.convertDecisionItem(node)

	default:
		if s.isExtensionNode(node.Type) {
			return s.convertExtension(node)
		}

		switch s.config.UnknownNodes {
		case UnknownError:
			return "", fmt.Errorf("unknown node type: %s", node.Type)
		case UnknownSkip:
			s.addWarning(WarningUnknownNode, node.Type, fmt.Sprintf("unknown node skipped: %s", node.Type))
			return "", nil
		default:
			s.addWarning(WarningUnknownNode, node.Type, fmt.Sprintf("unknown node rendered as placeholder: %s", node.Type))
			return fmt.Sprintf("[Unknown node: %s]", node.Type), nil
		}
	}
}

func (s *state) isExtensionNode(nodeType string) bool {
	switch nodeType {
	case "extension", "inlineExtension", "bodiedExtension":
		return true
	default:
		return false
	}
}

func (s *state) addWarning(warnType WarningType, nodeType, message string) {
	s.warnings = append(s.warnings, Warning{
		Type:     warnType,
		NodeType: nodeType,
		Message:  message,
	})
}

func (s *state) checkContext() error {
	if s.ctx == nil {
		return nil
	}

	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
		return nil
	}
}

// convertChildren processes a slice of nodes and concatenates their results
func (s *state) convertChildren(content []Node) (string, error) {
	var sb strings.Builder
	for _, child := range content {
		if err := s.checkContext(); err != nil {
			return "", err
		}
		res, err := s.convertNode(child)
		if err != nil {
			return "", err
		}
		sb.WriteString(res)
	}
	return sb.String(), nil
}

// convertDoc converts the root document node
func (s *state) convertDoc(node Node) (string, error) {
	res, err := s.convertChildren(node.Content)
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
func (s *state) convertInlineContent(content []Node) (string, error) {
	var sb strings.Builder
	var activeMarks []Mark // Track currently active marks (full Mark objects)

	// Check if any text node has both strong and em anywhere in the paragraph
	useUnderscoreForEm := s.hasStrongAndEm(content)

	for _, node := range content {
		if err := s.checkContext(); err != nil {
			return "", err
		}

		if node.Type != "text" {
			// For non-text nodes, close all active marks, process node, reset marks
			if err := s.closeMarks(activeMarks, useUnderscoreForEm, &sb); err != nil {
				return "", err
			}

			result, err := s.convertNode(node)
			if err != nil {
				return "", err
			}
			if startsWithFence(result) {
				ensureFenceLineStart(&sb)
			}
			sb.WriteString(result)
			activeMarks = nil
			continue
		}

		// Filter marks according to unknown-mark policy.
		currentMarks := make([]Mark, 0, len(node.Marks))
		var unknownPlaceholder strings.Builder
		for _, mark := range node.Marks {
			if s.isKnownMark(mark.Type) {
				currentMarks = append(currentMarks, mark)
				continue
			}

			switch s.config.UnknownMarks {
			case UnknownError:
				return "", fmt.Errorf("unknown mark type: %s", mark.Type)
			case UnknownSkip:
				s.addWarning(WarningUnknownMark, mark.Type, fmt.Sprintf("unknown mark skipped: %s", mark.Type))
			case UnknownPlaceholder:
				s.addWarning(WarningUnknownMark, mark.Type, fmt.Sprintf("unknown mark rendered as placeholder: %s", mark.Type))
				unknownPlaceholder.WriteString(fmt.Sprintf("[Unknown mark: %s]", mark.Type))
			}
		}

		// Special handling for whitespace-only nodes to avoid "stupid" markdown like ** **
		// or marks starting/ending on whitespace.
		effectiveMarks := currentMarks
		if strings.TrimSpace(node.Text) == "" {
			// For whitespace-only nodes, we don't want to open new marks.
			// We only keep marks that were already active.
			effectiveMarks = s.intersectMarks(activeMarks, currentMarks)
		}

		// Find marks to close and open
		marksToClose := s.getMarksToCloseFull(activeMarks, effectiveMarks)
		marksToOpen := s.getMarksToOpenFull(activeMarks, effectiveMarks)

		// Close marks
		if err := s.closeMarks(marksToClose, useUnderscoreForEm, &sb); err != nil {
			return "", err
		}

		// Open new marks (in priority order)
		for _, mark := range marksToOpen {
			opening, err := s.getOpeningDelimiterForMark(mark, useUnderscoreForEm)
			if err != nil {
				return "", err
			}
			sb.WriteString(opening)
		}

		// Write text content (including placeholders for unknown marks).
		if unknownPlaceholder.Len() > 0 {
			sb.WriteString(unknownPlaceholder.String())
		}
		sb.WriteString(node.Text)

		// Update active marks
		activeMarks = effectiveMarks
	}

	// Close any remaining marks at end of content
	if err := s.closeMarks(activeMarks, useUnderscoreForEm, &sb); err != nil {
		return "", err
	}

	return sb.String(), nil
}

func startsWithFence(value string) bool {
	trimmed := strings.TrimLeft(value, "\n")
	return strings.HasPrefix(trimmed, "```")
}

func ensureFenceLineStart(sb *strings.Builder) {
	if sb.Len() == 0 {
		return
	}
	content := sb.String()
	if strings.HasSuffix(content, "\n") {
		return
	}
	sb.WriteString("\n")
}

// closeMarks closes the provided marks in reverse order
func (s *state) closeMarks(marks []Mark, useUnderscoreForEm bool, sb *strings.Builder) error {
	for i := len(marks) - 1; i >= 0; i-- {
		closing, err := s.getClosingDelimiterForMark(marks[i], useUnderscoreForEm)
		if err != nil {
			return err
		}
		sb.WriteString(closing)
	}
	return nil
}

// hasStrongAndEm checks if any text node in content has both strong and em marks
func (s *state) hasStrongAndEm(content []Node) bool {
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
