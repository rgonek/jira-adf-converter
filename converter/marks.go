package converter

import (
	"fmt"
	"strings"
)

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
				// Wait, I found a bug in existing code? sup uses </u>?
				// Looking at old code: sup used </sup>, sub used </sub>.
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

// intersectMarks returns the intersection of two mark slices, preserving the order of the first slice.
// This is used to maintain mark continuity across whitespace-only nodes without opening new marks.
func (c *Converter) intersectMarks(activeMarks, currentMarks []Mark) []Mark {
	var res []Mark
	for _, am := range activeMarks {
		for _, cm := range currentMarks {
			if c.marksEqual(am, cm) {
				res = append(res, am)
				break
			}
		}
	}
	return res
}
