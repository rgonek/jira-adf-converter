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
	content, err := s.convertChildrenWithParent(node.Content, node.Type)
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
	id := s.getMediaID(node)
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
		escapedURL := escapeMarkdownLinkDestination(url)
		if escapedURL != "" {
			return fmt.Sprintf("![%s](%s)", escapeMarkdownTextLiteral(alt), escapedURL), nil
		}
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
		escapedURL := escapeMarkdownLinkDestination(base + id)
		if escapedURL != "" {
			return fmt.Sprintf("![%s](%s)", escapeMarkdownTextLiteral(alt), escapedURL), nil
		}
	}

	// Internal image
	if mediaType == "image" {
		if id == "" {
			if s.config.UnknownNodes == UnknownError {
				return "", fmt.Errorf("media node of type image missing id")
			}
			s.addWarningWithContext(WarningMissingAttribute, node, "media image missing id")
			return "[Image: (no id)]", nil
		}
		return fmt.Sprintf("[Image: %s]", escapeMarkdownTextLiteral(id)), nil
	}

	// File
	if mediaType == "file" {
		if id == "" {
			if s.config.UnknownNodes == UnknownError {
				return "", fmt.Errorf("media node of type file missing id")
			}
			s.addWarningWithContext(WarningMissingAttribute, node, "media file missing id")
			return "[File: (no id)]", nil
		}
		return fmt.Sprintf("[File: %s]", escapeMarkdownTextLiteral(id)), nil
	}

	// Fallback/Unknown
	if id == "" {
		if s.config.UnknownNodes == UnknownError {
			return "", fmt.Errorf("media node missing id")
		}
		s.addWarningWithContext(WarningMissingAttribute, node, "media node missing id")
		return "[Media: (no id)]", nil
	}
	return fmt.Sprintf("[Media: %s]", escapeMarkdownTextLiteral(id)), nil
}

// convertMediaInline converts a mediaInline node
func (s *state) convertMediaInline(node Node) (string, error) {
	if s.config.MediaInlineStyle == MediaInlinePandoc {
		id := s.getMediaID(node)
		mediaType := node.GetStringAttr("type", "file")
		if id == "" {
			s.addWarningWithContext(WarningMissingAttribute, node, "mediaInline missing id")
			return "[Media: (no id)]", nil
		}
		var label string
		switch mediaType {
		case "image":
			label = fmt.Sprintf("Image: %s", id)
		case "file":
			label = fmt.Sprintf("File: %s", id)
		default:
			label = fmt.Sprintf("Media: %s", id)
		}
		span := fmt.Sprintf("[%s]{.media-inline", escapeMarkdownTextLiteral(label))
		span += fmt.Sprintf(` media-id=%q`, id)
		span += fmt.Sprintf(` media-type=%q`, mediaType)
		span += "}"
		return span, nil
	}
	return s.convertMedia(node)
}

// convertCaption converts a caption node
func (s *state) convertCaption(node Node) (string, error) {
	content, err := s.convertChildrenWithParent(node.Content, node.Type)
	if err != nil {
		return "", err
	}
	if content == "" {
		return "", nil
	}
	if s.config.CaptionStyle == CaptionPandoc {
		return fmt.Sprintf("[%s]{.media-caption}", content), nil
	}
	// For now, just render as text. In many systems, this follows an image.
	return content, nil
}

func (s *state) getMediaID(node Node) string {
	if id := node.GetStringAttr("id", ""); id != "" {
		return id
	}
	if id := node.GetStringAttr("attachmentId", ""); id != "" {
		return id
	}
	if id := node.GetStringAttr("fileId", ""); id != "" {
		return id
	}
	if id := node.GetStringAttr("attachment-id", ""); id != "" {
		return id
	}
	return ""
}
