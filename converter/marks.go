package converter

import (
	"fmt"
	"strings"
)

// getMarksToCloseFull returns marks that need to be closed
func (s *state) getMarksToCloseFull(activeMarks, currentMarks []Mark) []Mark {
	// Find the first mark that differs or is missing in currentMarks
	closeFromIndex := -1
	for i, activeMark := range activeMarks {
		if i >= len(currentMarks) || !s.marksEqual(activeMark, currentMarks[i]) {
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
func (s *state) getMarksToOpenFull(activeMarks, currentMarks []Mark) []Mark {
	// Find common prefix length
	commonLen := 0
	for i := 0; i < len(activeMarks) && i < len(currentMarks); i++ {
		if s.marksEqual(activeMarks[i], currentMarks[i]) {
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
func (s *state) marksEqual(m1, m2 Mark) bool {
	// Type must match
	if m1.Type != m2.Type {
		return false
	}

	// For marks with attributes, compare them as well
	switch m1.Type {
	case "link":
		return s.markAttrsEqual(m1.Attrs, m2.Attrs, []string{"href", "title"})
	case "subsup":
		return s.markAttrsEqual(m1.Attrs, m2.Attrs, []string{"type"})
	case "textColor", "backgroundColor":
		return s.markAttrsEqual(m1.Attrs, m2.Attrs, []string{"color"})
	}

	return true
}

// markAttrsEqual compares specific attributes between two marks
func (s *state) markAttrsEqual(attrs1, attrs2 map[string]any, keys []string) bool {
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
func (s *state) isKnownMark(markType string) bool {
	switch markType {
	case "strong", "em", "strike", "code", "underline", "link", "subsup", "textColor", "backgroundColor":
		return true
	default:
		return false
	}
}

// getOpeningDelimiterForMark returns the opening delimiter for a mark
func (s *state) getOpeningDelimiterForMark(mark Mark, useUnderscoreForEm bool) (string, error) {
	prefix, _, err := s.convertMarkFull(mark, useUnderscoreForEm)
	return prefix, err
}

// getClosingDelimiterForMark returns the closing delimiter for a mark
func (s *state) getClosingDelimiterForMark(mark Mark, useUnderscoreForEm bool) (string, error) {
	_, suffix, err := s.convertMarkFull(mark, useUnderscoreForEm)
	return suffix, err
}

// convertMarkFull returns opening delimiter, closing delimiter, and error for a mark
func (s *state) convertMarkFull(mark Mark, useUnderscoreForEm bool) (string, string, error) {
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
		switch s.config.UnderlineStyle {
		case UnderlineIgnore:
			return "", "", nil
		case UnderlineHTML:
			return "<u>", "</u>", nil
		case UnderlineBold:
			return "**", "**", nil
		default:
			return "", "", nil
		}
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

		switch s.config.SubSupStyle {
		case SubSupIgnore:
			return "", "", nil
		case SubSupHTML:
			if subSupType == "sub" {
				return "<sub>", "</sub>", nil
			}
			if subSupType == "sup" {
				return "<sup>", "</sup>", nil
			}
		case SubSupLaTeX:
			if subSupType == "sub" {
				return "$_{", "}$", nil
			}
			if subSupType == "sup" {
				return "$^{", "}$", nil
			}
		}

		return "", "", nil
	case "textColor":
		switch s.config.TextColorStyle {
		case ColorIgnore:
			return "", "", nil
		case ColorHTML:
			color := mark.GetStringAttr("color", "")
			if color == "" {
				return "", "", nil
			}
			return `<span style="color: ` + color + `">`, "</span>", nil
		default:
			return "", "", nil
		}
	case "backgroundColor":
		switch s.config.BackgroundColorStyle {
		case ColorIgnore:
			return "", "", nil
		case ColorHTML:
			color := mark.GetStringAttr("color", "")
			if color == "" {
				return "", "", nil
			}
			return `<span style="background-color: ` + color + `">`, "</span>", nil
		default:
			return "", "", nil
		}
	default:
		if s.config.UnknownMarks == UnknownError {
			return "", "", fmt.Errorf("unknown mark type: %s", mark.Type)
		}
		s.addWarning(WarningUnknownMark, mark.Type, fmt.Sprintf("unknown mark skipped: %s", mark.Type))
		return "", "", nil
	}
}

// intersectMarks returns the intersection of two mark slices, preserving the order of the first slice.
// This is used to maintain mark continuity across whitespace-only nodes without opening new marks.
func (s *state) intersectMarks(activeMarks, currentMarks []Mark) []Mark {
	var res []Mark
	for _, am := range activeMarks {
		for _, cm := range currentMarks {
			if s.marksEqual(am, cm) {
				res = append(res, am)
				break
			}
		}
	}
	return res
}
