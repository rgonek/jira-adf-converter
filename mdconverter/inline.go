package mdconverter

import (
	"strings"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
)

func (s *state) convertInlineChildren(parent ast.Node, stack *markStack) ([]converter.Node, error) {
	var content []converter.Node

	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		converted, err := s.convertInlineNode(child, stack)
		if err != nil {
			return nil, err
		}
		for _, node := range converted {
			content = appendInlineNode(content, node)
		}
	}

	return s.applyInlinePatterns(content), nil
}

func (s *state) convertInlineNode(node ast.Node, stack *markStack) ([]converter.Node, error) {
	switch typed := node.(type) {
	case *ast.Text:
		var content []converter.Node
		textValue := string(typed.Value(s.source))
		if textValue != "" {
			content = append(content, s.convertInlineText(textValue, stack)...)
		}

		if typed.HardLineBreak() {
			content = append(content, converter.Node{Type: "hardBreak"})
		} else if typed.SoftLineBreak() {
			content = append(content, s.convertInlineText(" ", stack)...)
		}

		return content, nil

	case *ast.String:
		return s.convertInlineText(string(typed.Value), stack), nil

	case *ast.Emphasis:
		markType := "em"
		if typed.Level >= 2 {
			markType = "strong"
		}
		stack.push(converter.Mark{Type: markType})
		content, err := s.convertInlineChildren(typed, stack)
		stack.popByType(markType)
		return content, err

	case *extast.Strikethrough:
		stack.push(converter.Mark{Type: "strike"})
		content, err := s.convertInlineChildren(typed, stack)
		stack.popByType("strike")
		return content, err

	case *ast.CodeSpan:
		stack.push(converter.Mark{Type: "code"})
		content, err := s.convertInlineChildren(typed, stack)
		stack.popByType("code")
		return content, err

	case *ast.Link:
		href := strings.TrimSpace(string(typed.Destination))
		if href == "" {
			return s.convertInlineChildren(typed, stack)
		}
		linkText := strings.TrimSpace(string(typed.Text(s.source)))
		title := strings.TrimSpace(string(typed.Title))
		const mentionScheme = "mention:"

		if s.shouldDetectMentionLink() && strings.HasPrefix(strings.ToLower(href), mentionScheme) {
			id := strings.TrimSpace(href[len(mentionScheme):])
			if id != "" {
				mentionText := strings.TrimPrefix(linkText, "@")
				attrs := map[string]interface{}{
					"id": id,
				}
				if mentionText != "" {
					attrs["text"] = mentionText
				}
				return []converter.Node{
					{
						Type:  "mention",
						Attrs: attrs,
					},
				}, nil
			}
		}

		if title == "" && linkText != "" && linkText == href {
			return []converter.Node{
				{
					Type: "inlineCard",
					Attrs: map[string]interface{}{
						"url": href,
					},
				},
			}, nil
		}

		mark := converter.Mark{
			Type: "link",
			Attrs: map[string]interface{}{
				"href": href,
			},
		}
		if title != "" {
			mark.Attrs["title"] = title
		}

		stack.push(mark)
		content, err := s.convertInlineChildren(typed, stack)
		stack.popByType("link")
		return content, err

	case *ast.RawHTML:
		return s.convertRawHTML(string(typed.Text(s.source)), stack), nil

	case *ast.Image:
		alt := strings.TrimSpace(string(typed.Text(s.source)))
		if alt == "" {
			alt = "Image"
		}
		href := strings.TrimSpace(string(typed.Destination))
		mediaAttrs := map[string]interface{}{
			"type": "image",
		}
		if href != "" {
			mediaID := href
			strippedToID := false
			if s.config.MediaBaseURL != "" && strings.HasPrefix(href, s.config.MediaBaseURL) {
				candidateID := strings.TrimPrefix(href, s.config.MediaBaseURL)
				if candidateID != "" {
					mediaID = candidateID
					strippedToID = true
				}
			}

			lowerHref := strings.ToLower(href)
			if strippedToID {
				mediaAttrs["id"] = mediaID
				if alt != "" {
					mediaAttrs["alt"] = alt
				}
			} else if strings.HasPrefix(lowerHref, "http://") || strings.HasPrefix(lowerHref, "https://") {
				mediaAttrs["url"] = href
				mediaAttrs["alt"] = alt
			} else {
				mediaAttrs["id"] = mediaID
				if alt != "" {
					mediaAttrs["alt"] = alt
				}
			}
		}

		return []converter.Node{
			{
				Type: "mediaSingle",
				Content: []converter.Node{
					{
						Type:  "media",
						Attrs: mediaAttrs,
					},
				},
			},
		}, nil

	default:
		if node.HasChildren() {
			return s.convertInlineChildren(node, stack)
		}
		return s.warnUnknownInline(node, stack), nil
	}
}

func (s *state) convertInlineText(textValue string, stack *markStack) []converter.Node {
	if textValue == "" {
		return nil
	}

	if mentionID, ok := s.currentHTMLMentionID(); ok {
		if strings.TrimSpace(textValue) == "" {
			return []converter.Node{newTextNode(textValue, stack.current())}
		}
		mentionText := strings.TrimPrefix(strings.TrimSpace(textValue), "@")
		return []converter.Node{
			{
				Type: "mention",
				Attrs: map[string]interface{}{
					"id":   mentionID,
					"text": mentionText,
				},
			},
		}
	}

	return []converter.Node{newTextNode(textValue, stack.current())}
}

func (s *state) applyInlinePatterns(content []converter.Node) []converter.Node {
	var expanded []converter.Node

	for _, node := range content {
		if node.Type == "text" && len(node.Marks) == 0 {
			patternNodes := s.expandTextPatterns(node.Text, nil)
			for _, patternNode := range patternNodes {
				expanded = appendInlineNode(expanded, patternNode)
			}
			continue
		}
		expanded = appendInlineNode(expanded, node)
	}

	return expanded
}
