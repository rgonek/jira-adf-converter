package converter

import (
	"fmt"
	"time"
)

// convertEmoji converts an emoji node to a shortcode or fallback
func (s *state) convertEmoji(node Node) (string, error) {
	shortName := node.GetStringAttr("shortName", "")
	fallback := node.GetStringAttr("fallback", "")

	if shortName != "" {
		return shortName, nil
	}
	if fallback != "" {
		return fallback, nil
	}

	// Fallback if neither exists
	if s.config.UnknownNodes == UnknownError {
		return "", fmt.Errorf("emoji node missing shortName and fallback")
	}
	s.addWarning(WarningMissingAttribute, node.Type, "emoji node missing shortName and fallback")
	return "", nil
}

// convertMention converts a mention node to text representation
func (s *state) convertMention(node Node) (string, error) {
	text := node.GetStringAttr("text", "Unknown User")
	return text, nil
}

// convertStatus converts a status node to text representation
func (s *state) convertStatus(node Node) (string, error) {
	text := node.GetStringAttr("text", "Unknown")
	return fmt.Sprintf("[Status: %s]", text), nil
}

// convertDate converts a date node to ISO 8601 format
func (s *state) convertDate(node Node) (string, error) {
	timestamp := node.GetStringAttr("timestamp", "")

	if timestamp == "" || timestamp == "invalid" {
		if s.config.UnknownNodes == UnknownError {
			return "", fmt.Errorf("date node missing or invalid timestamp")
		}
		s.addWarning(WarningMissingAttribute, node.Type, "date node missing or invalid timestamp")
		return "[Date: invalid]", nil
	}

	// Parse timestamp
	var ts int64
	_, err := fmt.Sscanf(timestamp, "%d", &ts)
	if err != nil {
		if s.config.UnknownNodes == UnknownError {
			return "", fmt.Errorf("date node has invalid timestamp format: %s", timestamp)
		}
		s.addWarning(WarningMissingAttribute, node.Type, fmt.Sprintf("date node has invalid timestamp format: %s", timestamp))
		return "[Date: invalid]", nil
	}

	// Heuristic to handle milliseconds vs seconds
	// Jira timestamps can be in either format. We use a cutoff to detect:
	// - Values > 10000000000 are treated as milliseconds (divided by 1000)
	// - Values <= 10000000000 are treated as seconds
	// This cutoff represents year 2286 in seconds, which is safe for typical Jira usage.
	// Note: This heuristic will fail for:
	// - Timestamps after year 2286 (will be incorrectly treated as milliseconds)
	// - Millisecond timestamps before Nov 20, 1970 (will be incorrectly treated as seconds)
	if ts > 10000000000 {
		ts = ts / 1000
	}

	t := time.Unix(ts, 0).UTC()
	return t.Format("2006-01-02"), nil
}

// convertInlineCard converts an inlineCard node
func (s *state) convertInlineCard(node Node) (string, error) {
	url := node.GetStringAttr("url", "")

	// if url is present, return [url](url)
	if url != "" {
		return fmt.Sprintf("[%s](%s)", url, url), nil
	}

	// check for data
	if node.Attrs != nil && node.Attrs["data"] != nil {
		data, ok := node.Attrs["data"].(map[string]interface{})
		if ok {
			name := ""
			dataUrl := ""
			if n, ok := data["name"].(string); ok {
				name = n
			}
			if u, ok := data["url"].(string); ok {
				dataUrl = u
			}

			if name != "" && dataUrl != "" {
				return fmt.Sprintf("[%s](%s)", name, dataUrl), nil
			}
			if dataUrl != "" {
				return fmt.Sprintf("[%s](%s)", dataUrl, dataUrl), nil
			}
			if name != "" {
				return name, nil
			}
		}
	}

	// Fallback
	if s.config.UnknownNodes == UnknownError {
		return "", fmt.Errorf("inlineCard missing url and valid data")
	}
	s.addWarning(WarningMissingAttribute, node.Type, "inlineCard missing url and valid data")
	return "[Smart Link]", nil
}
