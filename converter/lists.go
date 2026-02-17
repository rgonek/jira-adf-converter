package converter

import (
	"fmt"
	"strings"
)

// convertListItems iterates over list items, converts them, and applies indentation with the provided marker.
func (s *state) convertListItems(content []Node, childType string, getMarker func(index int) string) (string, error) {
	var sb strings.Builder

	for i, item := range content {
		if item.Type != childType {
			if s.config.UnknownNodes == UnknownError {
				// We don't have the parent type here easily, so we give a generic error
				return "", fmt.Errorf("expected %s child, got %s", childType, item.Type)
			}
			s.addWarning(WarningUnknownNode, item.Type, fmt.Sprintf("unexpected list child %s, expected %s", item.Type, childType))
			continue
		}

		itemContent, err := s.convertListItemContent(item.Content)
		if err != nil {
			return "", err
		}

		marker := getMarker(i)
		indented := s.indent(itemContent, marker)
		sb.WriteString(indented)
		sb.WriteString("\n")
	}

	return sb.String() + "\n", nil
}

// convertBulletList converts a bullet list node to markdown
func (s *state) convertBulletList(node Node) (string, error) {
	return s.convertListItems(node.Content, "listItem", func(i int) string {
		return "- "
	})
}

// convertOrderedList converts an ordered list node to markdown
func (s *state) convertOrderedList(node Node) (string, error) {
	// Extract starting order from attributes (default to 1)
	order := node.GetIntAttr("order", 1)

	return s.convertListItems(node.Content, "listItem", func(i int) string {
		return fmt.Sprintf("%d. ", order+i)
	})
}

// convertTaskList converts a task list node to markdown
func (s *state) convertTaskList(node Node) (string, error) {
	var sb strings.Builder

	for _, item := range node.Content {
		if item.Type == "taskList" {
			res, err := s.convertTaskList(item)
			if err != nil {
				return "", err
			}
			// Indent nested task lists to preserve hierarchy
			// We use 2 spaces which is standard for nested lists
			indented := s.indent(res, "  ")
			sb.WriteString(indented)
			sb.WriteString("\n")
			continue
		}

		if item.Type != "taskItem" {
			if s.config.UnknownNodes == UnknownError {
				return "", fmt.Errorf("taskList expects taskItem child, got %s", item.Type)
			}
			s.addWarning(WarningUnknownNode, item.Type, fmt.Sprintf("unexpected task list child %s", item.Type))
			continue
		}

		itemContent, err := s.convertTaskItem(item)
		if err != nil {
			return "", err
		}

		sb.WriteString(itemContent)
	}

	return sb.String() + "\n", nil
}

// convertTaskItem converts a task item node to markdown
func (s *state) convertTaskItem(node Node) (string, error) {
	// Extract state from attributes
	state := node.GetStringAttr("state", "TODO")

	// Determine checkbox marker
	marker := "- [ ] "
	if state == "DONE" {
		marker = "- [x] "
	}

	// Convert content using inline content converter to support marks
	itemContent, err := s.convertInlineContent(node.Content)
	if err != nil {
		return "", err
	}

	indented := s.indent(itemContent, marker)

	return indented + "\n", nil
}

// convertListItem converts a list item node to markdown
func (s *state) convertListItem(node Node) (string, error) {
	return s.convertListItemContent(node.Content)
}

// convertListItemContent processes the content of a list item
func (s *state) convertListItemContent(content []Node) (string, error) {
	var sb strings.Builder

	for i, child := range content {
		result, err := s.convertNode(child)
		if err != nil {
			return "", err
		}

		// Remove trailing newlines from each child except the last
		result = strings.TrimRight(result, "\n")
		sb.WriteString(result)

		// Add a blank line between children to preserve block separation
		if i < len(content)-1 {
			sb.WriteString("\n\n")
		}
	}

	return sb.String(), nil
}
