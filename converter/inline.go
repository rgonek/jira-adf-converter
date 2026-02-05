package converter

import (
	"fmt"
	"time"
)

// convertEmoji converts an emoji node to a shortcode or fallback
func (c *Converter) convertEmoji(node Node) (string, error) {
	shortName := ""
	fallback := ""

	if node.Attrs != nil {
		if s, ok := node.Attrs["shortName"].(string); ok {
			shortName = s
		}
		if f, ok := node.Attrs["fallback"].(string); ok {
			fallback = f
		}
	}

	if shortName != "" {
		return shortName, nil
	}
	if fallback != "" {
		return fallback, nil
	}

	// Fallback if neither exists
	if c.config.Strict {
		return "", fmt.Errorf("emoji node missing shortName and fallback")
	}
	return "", nil
}

// convertMention converts a mention node to text representation
func (c *Converter) convertMention(node Node) (string, error) {
	text := "Unknown User"

	if node.Attrs != nil {
		if t, ok := node.Attrs["text"].(string); ok {
			text = t
		}
	}

	return text, nil
}

// convertStatus converts a status node to text representation
func (c *Converter) convertStatus(node Node) (string, error) {
	text := "Unknown"
	if node.Attrs != nil {
		if t, ok := node.Attrs["text"].(string); ok {
			text = t
		}
	}
	return fmt.Sprintf("[Status: %s]", text), nil
}

// convertDate converts a date node to ISO 8601 format
func (c *Converter) convertDate(node Node) (string, error) {
	timestamp := ""
	if node.Attrs != nil {
		if t, ok := node.Attrs["timestamp"].(string); ok {
			timestamp = t
		}
	}

	if timestamp == "" || timestamp == "invalid" {
		if c.config.Strict {
			return "", fmt.Errorf("date node missing or invalid timestamp")
		}
		return "[Date: invalid]", nil
	}

	// Parse timestamp
	var ts int64
	_, err := fmt.Sscanf(timestamp, "%d", &ts)
	if err != nil {
		if c.config.Strict {
			return "", fmt.Errorf("date node has invalid timestamp format: %s", timestamp)
		}
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
func (c *Converter) convertInlineCard(node Node) (string, error) {
	url := ""
	if node.Attrs != nil {
		if u, ok := node.Attrs["url"].(string); ok {
			url = u
		}
	}

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
	if c.config.Strict {
		return "", fmt.Errorf("inlineCard missing url and valid data")
	}
	return "[Smart Link]", nil
}
