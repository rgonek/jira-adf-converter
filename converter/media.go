package converter

import (
	"fmt"
	"strings"
)

// convertMediaSingle converts a mediaSingle node
func (c *Converter) convertMediaSingle(node Node) (string, error) {
	if len(node.Content) == 0 {
		return "", nil
	}

	// Pass through to children
	var sb strings.Builder
	for _, child := range node.Content {
		result, err := c.convertNode(child)
		if err != nil {
			return "", err
		}
		sb.WriteString(result)
	}
	content := sb.String()
	if strings.TrimSpace(content) == "" {
		return "", nil
	}
	return content + "\n\n", nil
}

// convertMediaGroup converts a mediaGroup node
func (c *Converter) convertMediaGroup(node Node) (string, error) {
	if len(node.Content) == 0 {
		return "", nil
	}

	var items []string
	for _, child := range node.Content {
		result, err := c.convertNode(child)
		if err != nil {
			return "", err
		}
		items = append(items, result)
	}
	return strings.Join(items, "\n") + "\n\n", nil
}

// convertMedia converts a media node
func (c *Converter) convertMedia(node Node) (string, error) {
	mediaType := ""
	id := ""
	alt := ""
	url := ""

	if node.Attrs != nil {
		if t, ok := node.Attrs["type"].(string); ok {
			mediaType = t
		}
		if i, ok := node.Attrs["id"].(string); ok {
			id = i
		}
		if a, ok := node.Attrs["alt"].(string); ok {
			alt = a
		}
		if u, ok := node.Attrs["url"].(string); ok {
			url = u
		}
	}

	// External image
	if mediaType == "image" && url != "" {
		if alt == "" {
			alt = "Image"
		}
		return fmt.Sprintf("![%s](%s)", alt, url), nil
	}

	// Internal image
	if mediaType == "image" {
		if id == "" {
			if c.config.Strict {
				return "", fmt.Errorf("media node of type image missing id")
			}
			return "[Image: (no id)]", nil
		}
		return fmt.Sprintf("[Image: %s]", id), nil
	}

	// File
	if mediaType == "file" {
		if id == "" {
			if c.config.Strict {
				return "", fmt.Errorf("media node of type file missing id")
			}
			return "[File: (no id)]", nil
		}
		return fmt.Sprintf("[File: %s]", id), nil
	}

	// Fallback/Unknown
	if id == "" {
		if c.config.Strict {
			return "", fmt.Errorf("media node missing id")
		}
		return "[Media: (no id)]", nil
	}
	return fmt.Sprintf("[Media: %s]", id), nil
}
