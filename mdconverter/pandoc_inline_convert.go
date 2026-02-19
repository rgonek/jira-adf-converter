package mdconverter

import (
	"strings"

	"github.com/rgonek/jira-adf-converter/converter"
)

func (s *state) convertPandocSubscriptNode(node *PandocSubscriptNode, stack *markStack) ([]converter.Node, error) {
	return s.convertPandocSubSupNode(node.Content, "sub", "~", stack)
}

func (s *state) convertPandocSuperscriptNode(node *PandocSuperscriptNode, stack *markStack) ([]converter.Node, error) {
	return s.convertPandocSubSupNode(node.Content, "sup", "^", stack)
}

func (s *state) convertPandocSubSupNode(content, kind, delimiter string, stack *markStack) ([]converter.Node, error) {
	if !s.shouldDetectSubSupPandoc() {
		return []converter.Node{newTextNode(delimiter+content+delimiter, stack.current())}, nil
	}

	inlineContent, err := s.convertInlineFragment(content)
	if err != nil {
		return nil, err
	}

	marked := applyMarkToInlineNodes(inlineContent, converter.Mark{
		Type: "subsup",
		Attrs: map[string]interface{}{
			"type": kind,
		},
	})
	return applyOuterMarksToInlineNodes(marked, stack.current()), nil
}

func (s *state) convertPandocSpanNode(node *PandocSpanNode, stack *markStack) ([]converter.Node, error) {
	literal := renderPandocSpanLiteral(node)
	if hasUnknownPandocSpanClass(node.Classes) || hasUnknownPandocSpanAttr(node.Attrs) {
		s.addWarning(converter.WarningDroppedFeature, "pandocSpan", "unsupported pandoc span class or attribute; preserved as text")
		return []converter.Node{newTextNode(literal, stack.current())}, nil
	}

	if hasPandocClass(node.Classes, "mention") {
		if !s.shouldDetectMentionPandoc() {
			return []converter.Node{newTextNode(literal, stack.current())}, nil
		}
		mentionID := strings.TrimSpace(node.Attrs["mention-id"])
		if mentionID == "" {
			s.addWarning(converter.WarningMissingAttribute, "pandocSpan", "pandoc mention span missing mention-id")
			return []converter.Node{newTextNode(literal, stack.current())}, nil
		}

		mentionText := strings.TrimSpace(node.Content)
		mentionText = strings.TrimPrefix(mentionText, "@")

		mention := converter.Node{
			Type: "mention",
			Attrs: map[string]interface{}{
				"id": mentionID,
			},
		}
		if mentionText != "" {
			mention.Attrs["text"] = mentionText
		}
		return []converter.Node{mention}, nil
	}

	if hasPandocClass(node.Classes, "inline-card") {
		if !s.shouldDetectInlineCardPandoc() {
			return []converter.Node{newTextNode(literal, stack.current())}, nil
		}
		url := strings.TrimSpace(node.Attrs["url"])
		if url == "" {
			s.addWarning(converter.WarningMissingAttribute, "pandocSpan", "pandoc inline-card span missing url")
			return []converter.Node{newTextNode(literal, stack.current())}, nil
		}

		displayTitle := strings.TrimSpace(node.Content)
		inlineCardAttrs := map[string]interface{}{
			"url": url,
		}
		if displayTitle != "" && displayTitle != url {
			inlineCardAttrs["data"] = map[string]interface{}{
				"name": displayTitle,
				"url":  url,
			}
		}
		return []converter.Node{
			{
				Type:  "inlineCard",
				Attrs: inlineCardAttrs,
			},
		}, nil
	}

	if hasPandocClass(node.Classes, "underline") && !s.shouldDetectUnderlinePandoc() {
		return []converter.Node{newTextNode(literal, stack.current())}, nil
	}
	if (strings.TrimSpace(node.Attrs["color"]) != "" || strings.TrimSpace(node.Attrs["background-color"]) != "" || strings.TrimSpace(node.Attrs["style"]) != "") && !s.shouldDetectColorPandoc() {
		return []converter.Node{newTextNode(literal, stack.current())}, nil
	}

	inlineContent, err := s.convertInlineFragment(node.Content)
	if err != nil {
		return nil, err
	}
	applied := false

	if hasPandocClass(node.Classes, "underline") {
		inlineContent = applyMarkToInlineNodes(inlineContent, converter.Mark{Type: "underline"})
		applied = true
	}

	style := node.Attrs["style"]

	color := strings.TrimSpace(node.Attrs["color"])
	if color == "" && style != "" {
		color = extractStyleColor(style, "color")
	}
	if color != "" {
		inlineContent = applyMarkToInlineNodes(inlineContent, converter.Mark{
			Type: "textColor",
			Attrs: map[string]interface{}{
				"color": color,
			},
		})
		applied = true
	}

	bgColor := strings.TrimSpace(node.Attrs["background-color"])
	if bgColor == "" && style != "" {
		bgColor = extractStyleColor(style, "background-color")
	}
	if bgColor != "" {
		inlineContent = applyMarkToInlineNodes(inlineContent, converter.Mark{
			Type: "backgroundColor",
			Attrs: map[string]interface{}{
				"color": bgColor,
			},
		})
		applied = true
	}

	if applied {
		return applyOuterMarksToInlineNodes(inlineContent, stack.current()), nil
	}

	s.addWarning(converter.WarningDroppedFeature, "pandocSpan", "pandoc span attributes were not mapped; preserved as text")
	return []converter.Node{newTextNode(literal, stack.current())}, nil
}

func applyMarkToInlineNodes(content []converter.Node, mark converter.Mark) []converter.Node {
	if len(content) == 0 {
		return nil
	}

	out := make([]converter.Node, 0, len(content))
	for _, node := range content {
		next := node
		if next.Type == "text" {
			marks := make([]converter.Mark, 0, len(next.Marks)+1)
			marks = append(marks, cloneMark(mark))
			for _, existing := range next.Marks {
				marks = append(marks, cloneMark(existing))
			}
			next.Marks = marks
			out = appendInlineNode(out, next)
			continue
		}
		if len(next.Content) > 0 {
			next.Content = applyMarkToInlineNodes(next.Content, mark)
		}
		out = append(out, next)
	}

	return out
}

func applyOuterMarksToInlineNodes(content []converter.Node, outerMarks []converter.Mark) []converter.Node {
	if len(content) == 0 || len(outerMarks) == 0 {
		return content
	}

	out := make([]converter.Node, 0, len(content))
	for _, node := range content {
		next := node
		if next.Type == "text" {
			marks := make([]converter.Mark, 0, len(outerMarks)+len(next.Marks))
			for _, mark := range outerMarks {
				marks = append(marks, cloneMark(mark))
			}
			for _, existing := range next.Marks {
				marks = append(marks, cloneMark(existing))
			}
			next.Marks = marks
			out = appendInlineNode(out, next)
			continue
		}
		if len(next.Content) > 0 {
			next.Content = applyOuterMarksToInlineNodes(next.Content, outerMarks)
		}
		out = append(out, next)
	}
	return out
}

func renderPandocSpanLiteral(node *PandocSpanNode) string {
	var builder strings.Builder
	builder.WriteString("[")
	builder.WriteString(node.Content)
	builder.WriteString("]{")
	builder.WriteString(node.RawAttrs)
	builder.WriteString("}")
	return builder.String()
}

func hasPandocClass(classes []string, target string) bool {
	for _, className := range classes {
		if className == target {
			return true
		}
	}
	return false
}

func hasUnknownPandocSpanClass(classes []string) bool {
	for _, className := range classes {
		switch className {
		case "underline", "mention", "inline-card":
			continue
		default:
			return true
		}
	}
	return false
}

func hasUnknownPandocSpanAttr(attrs map[string]string) bool {
	for key := range attrs {
		switch key {
		case "mention-id", "url", "color", "background-color", "style":
			continue
		default:
			return true
		}
	}
	return false
}
