package converter

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"time"
)

// convertEmoji converts an emoji node to a shortcode or fallback
func (s *state) convertEmoji(node Node) (string, error) {
	shortName := node.GetStringAttr("shortName", "")
	fallback := node.GetStringAttr("fallback", "")

	switch s.config.EmojiStyle {
	case EmojiUnicode:
		if fallback != "" {
			return fallback, nil
		}
		if shortName != "" {
			return shortName, nil
		}
	default:
		if shortName != "" {
			return shortName, nil
		}
		if fallback != "" {
			return fallback, nil
		}
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
	id := node.GetStringAttr("id", "")
	rawText := node.GetStringAttr("text", "")
	text := rawText
	if text == "" {
		text = "Unknown User"
	}
	mentionText := text
	if rawText != "" && !strings.HasPrefix(mentionText, "@") {
		mentionText = "@" + mentionText
	}

	switch s.config.MentionStyle {
	case MentionText:
		return mentionText, nil
	case MentionLink:
		if id == "" {
			s.addWarning(WarningMissingAttribute, node.Type, "mention node missing id")
			return mentionText, nil
		}
		return fmt.Sprintf("[%s](mention:%s)", mentionText, id), nil
	case MentionHTML:
		if id == "" {
			s.addWarning(WarningMissingAttribute, node.Type, "mention node missing id")
			return mentionText, nil
		}
		return fmt.Sprintf(`<span data-mention-id="%s">%s</span>`, html.EscapeString(id), html.EscapeString(mentionText)), nil
	default:
		return mentionText, nil
	}
}

// convertStatus converts a status node to text representation
func (s *state) convertStatus(node Node) (string, error) {
	text := node.GetStringAttr("text", "Unknown")
	if s.config.StatusStyle == StatusText {
		return text, nil
	}
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
	return t.Format(s.config.DateFormat), nil
}

// convertInlineCard converts an inlineCard node
func (s *state) convertInlineCard(node Node) (string, error) {
	title, url := s.getInlineCardLinkData(node)
	hookHandled := false

	hookOutput, handled, err := s.applyLinkRenderHook(
		node.Type,
		LinkRenderInput{
			Source:     "inlineCard",
			SourcePath: s.options.SourcePath,
			Href:       url,
			Title:      title,
			Text:       title,
			Meta:       linkMetadataFromAttrs(node.Attrs, url),
			Attrs:      cloneAnyMap(node.Attrs),
		},
	)
	if err != nil {
		return "", err
	}
	if handled {
		hookHandled = true
		if hookOutput.TextOnly {
			textValue := firstNonEmptyTrimmed(hookOutput.Title, title, url)
			if textValue != "" {
				return textValue, nil
			}
			if s.config.UnknownNodes == UnknownError {
				return "", fmt.Errorf("inlineCard missing url and valid data")
			}
			s.addWarning(WarningMissingAttribute, node.Type, "inlineCard missing url and valid data")
			return "[Smart Link]", nil
		}
		title = hookOutput.Title
		url = hookOutput.Href
	}

	switch s.config.InlineCardStyle {
	case InlineCardURL:
		if url != "" {
			return url, nil
		}
	case InlineCardEmbed:
		embedAttrs := node.Attrs
		if hookHandled {
			embedAttrs = rewriteInlineCardAttrs(node.Attrs, title, url)
		}
		if len(embedAttrs) > 0 {
			data, err := json.MarshalIndent(embedAttrs, "", "  ")
			if err != nil {
				return "", fmt.Errorf("failed to marshal inlineCard attrs: %w", err)
			}
			return fmt.Sprintf("```adf:inlineCard\n%s\n```\n\n", string(data)), nil
		}
	case InlineCardLink:
		if url != "" {
			if title == "" {
				title = url
			}
			return fmt.Sprintf("[%s](%s)", title, url), nil
		}
		if title != "" {
			return title, nil
		}
	}

	// Fallback
	if s.config.UnknownNodes == UnknownError {
		return "", fmt.Errorf("inlineCard missing url and valid data")
	}
	s.addWarning(WarningMissingAttribute, node.Type, "inlineCard missing url and valid data")
	return "[Smart Link]", nil
}

func rewriteInlineCardAttrs(attrs map[string]any, title, href string) map[string]any {
	rewritten := cloneAnyMap(attrs)
	if rewritten == nil {
		rewritten = map[string]any{}
	}

	href = strings.TrimSpace(href)
	title = strings.TrimSpace(title)

	if href != "" {
		rewritten["url"] = href
	}

	rawData, ok := rewritten["data"]
	if !ok {
		return rewritten
	}

	dataMap, ok := rawData.(map[string]any)
	if !ok {
		return rewritten
	}

	clonedData := cloneAnyMap(dataMap)
	if href != "" {
		clonedData["url"] = href
	}
	if title != "" {
		clonedData["name"] = title
	}
	rewritten["data"] = clonedData

	return rewritten
}

func firstNonEmptyTrimmed(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func (s *state) getInlineCardLinkData(node Node) (string, string) {
	url := node.GetStringAttr("url", "")
	title := ""
	if url != "" {
		title = url
	}

	if node.Attrs == nil || node.Attrs["data"] == nil {
		return title, url
	}

	data, ok := node.Attrs["data"].(map[string]interface{})
	if !ok {
		return title, url
	}

	if n, ok := data["name"].(string); ok && n != "" {
		title = n
	}
	if u, ok := data["url"].(string); ok && u != "" {
		url = u
		if title == "" {
			title = u
		}
	}

	return title, url
}
