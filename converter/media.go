package converter

import (
	"fmt"
	"strings"
)

// convertMediaSingle converts a mediaSingle node
func (s *state) convertMediaSingle(node Node) (string, error) {
	if len(node.Content) == 0 {
		return "", nil
	}

	// Pass through to children
	content, err := s.convertChildren(node.Content)
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(content) == "" {
		return "", nil
	}
	return content + "\n\n", nil
}

// convertMediaGroup converts a mediaGroup node
func (s *state) convertMediaGroup(node Node) (string, error) {
	if len(node.Content) == 0 {
		return "", nil
	}

	var items []string
	for _, child := range node.Content {
		result, err := s.convertNode(child)
		if err != nil {
			return "", err
		}
		items = append(items, result)
	}
	return strings.Join(items, "\n") + "\n\n", nil
}

// convertMedia converts a media node
func (s *state) convertMedia(node Node) (string, error) {
	mediaType := node.GetStringAttr("type", "")
	id := node.GetStringAttr("id", "")
	alt := node.GetStringAttr("alt", "")
	url := node.GetStringAttr("url", "")

	hookOutput, handled, err := s.applyMediaRenderHook(
		node.Type,
		MediaRenderInput{
			SourcePath: s.options.SourcePath,
			MediaType:  mediaType,
			ID:         id,
			URL:        url,
			Alt:        alt,
			Meta:       mediaMetadataFromAttrs(node.Attrs, id, url),
			Attrs:      cloneAnyMap(node.Attrs),
		},
	)
	if err != nil {
		return "", err
	}
	if handled {
		return hookOutput.Markdown, nil
	}

	// External image
	if mediaType == "image" && url != "" {
		if alt == "" {
			alt = "Image"
		}
		return fmt.Sprintf("![%s](%s)", alt, url), nil
	}

	// Internal media resolved via configured base URL.
	if url == "" && id != "" && s.config.MediaBaseURL != "" {
		if alt == "" {
			alt = "Image"
		}
		base := s.config.MediaBaseURL
		if !strings.HasSuffix(base, "/") {
			base += "/"
		}
		return fmt.Sprintf("![%s](%s%s)", alt, base, id), nil
	}

	// Internal image
	if mediaType == "image" {
		if id == "" {
			if s.config.UnknownNodes == UnknownError {
				return "", fmt.Errorf("media node of type image missing id")
			}
			s.addWarning(WarningMissingAttribute, node.Type, "media image missing id")
			return "[Image: (no id)]", nil
		}
		return fmt.Sprintf("[Image: %s]", id), nil
	}

	// File
	if mediaType == "file" {
		if id == "" {
			if s.config.UnknownNodes == UnknownError {
				return "", fmt.Errorf("media node of type file missing id")
			}
			s.addWarning(WarningMissingAttribute, node.Type, "media file missing id")
			return "[File: (no id)]", nil
		}
		return fmt.Sprintf("[File: %s]", id), nil
	}

	// Fallback/Unknown
	if id == "" {
		if s.config.UnknownNodes == UnknownError {
			return "", fmt.Errorf("media node missing id")
		}
		s.addWarning(WarningMissingAttribute, node.Type, "media node missing id")
		return "[Media: (no id)]", nil
	}
	return fmt.Sprintf("[Media: %s]", id), nil
}
