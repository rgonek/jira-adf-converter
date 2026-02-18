package mdconverter

import (
	"regexp"
	"strings"

	"github.com/rgonek/jira-adf-converter/converter"
)

var (
	spanStyleColorRe     = regexp.MustCompile(`(?i)\bcolor\s*:\s*([^;"]+)`)
	spanStyleBgColorRe   = regexp.MustCompile(`(?i)\bbackground-color\s*:\s*([^;"]+)`)
	spanMentionAttrIDRe  = regexp.MustCompile(`(?i)\bdata-mention-id\s*=\s*"([^"]+)"`)
	openingSpanTagPrefix = "<span"
	closingSpanTagPrefix = "</span"
	openingUnderlineTag  = "<u>"
	closingUnderlineTag  = "</u>"
	openingSubTag        = "<sub>"
	closingSubTag        = "</sub>"
	openingSupTag        = "<sup>"
	closingSupTag        = "</sup>"
	hardBreakTag1        = "<br>"
	hardBreakTag2        = "<br/>"
	hardBreakTag3        = "<br />"
)

func (s *state) convertRawHTML(rawHTML string, stack *markStack) []converter.Node {
	trimmed := strings.TrimSpace(rawHTML)
	lower := strings.ToLower(trimmed)

	switch lower {
	case openingUnderlineTag:
		stack.push(converter.Mark{Type: "underline"})
		return nil
	case closingUnderlineTag:
		stack.popByType("underline")
		return nil

	case openingSubTag:
		stack.push(converter.Mark{
			Type: "subsup",
			Attrs: map[string]interface{}{
				"type": "sub",
			},
		})
		return nil
	case closingSubTag:
		stack.popByType("subsup")
		return nil

	case openingSupTag:
		stack.push(converter.Mark{
			Type: "subsup",
			Attrs: map[string]interface{}{
				"type": "sup",
			},
		})
		return nil
	case closingSupTag:
		stack.popByType("subsup")
		return nil

	case hardBreakTag1, hardBreakTag2, hardBreakTag3:
		return []converter.Node{{Type: "hardBreak"}}
	}

	if strings.HasPrefix(lower, openingSpanTagPrefix) {
		if mentionID, ok := extractSpanMentionID(trimmed); ok && s.shouldDetectMentionHTML() {
			s.pushHTMLMentionID(mentionID)
			return nil
		}

		if color, ok := extractSpanStyleColor(trimmed, true); ok {
			stack.push(converter.Mark{
				Type: "backgroundColor",
				Attrs: map[string]interface{}{
					"color": color,
				},
			})
			return nil
		}
		if color, ok := extractSpanStyleColor(trimmed, false); ok {
			stack.push(converter.Mark{
				Type: "textColor",
				Attrs: map[string]interface{}{
					"color": color,
				},
			})
			return nil
		}

		return nil
	}

	if strings.HasPrefix(lower, closingSpanTagPrefix) {
		if stack.popByType("backgroundColor") || stack.popByType("textColor") {
			return nil
		}
		s.popHTMLMentionID()
		return nil
	}

	return nil
}

func extractSpanStyleColor(tag string, background bool) (string, bool) {
	var match []string
	if background {
		match = spanStyleBgColorRe.FindStringSubmatch(tag)
	} else {
		match = spanStyleColorRe.FindStringSubmatch(tag)
	}
	if len(match) != 2 {
		return "", false
	}

	value := strings.TrimSpace(match[1])
	if value == "" {
		return "", false
	}

	return value, true
}

func extractSpanMentionID(tag string) (string, bool) {
	match := spanMentionAttrIDRe.FindStringSubmatch(tag)
	if len(match) != 2 {
		return "", false
	}

	id := strings.TrimSpace(match[1])
	if id == "" {
		return "", false
	}

	return id, true
}

func (s *state) pushHTMLMentionID(id string) {
	s.htmlMentionStack = append(s.htmlMentionStack, id)
}

func (s *state) popHTMLMentionID() {
	if len(s.htmlMentionStack) == 0 {
		return
	}
	s.htmlMentionStack = s.htmlMentionStack[:len(s.htmlMentionStack)-1]
}

func (s *state) currentHTMLMentionID() (string, bool) {
	if len(s.htmlMentionStack) == 0 {
		return "", false
	}
	return s.htmlMentionStack[len(s.htmlMentionStack)-1], true
}
