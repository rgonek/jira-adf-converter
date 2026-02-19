package mdconverter

import (
	"strings"

	"github.com/rgonek/jira-adf-converter/converter"
)

func (s *state) convertPandocDivNode(node *PandocDivNode) (converter.Node, bool, error) {
	literalFallback := pandocLiteralParagraph(node.Literal())

	if hasPandocClass(node.Classes, "details") {
		if !s.shouldDetectExpandPandoc() {
			return literalFallback, true, nil
		}

		expandType := "expand"
		if s.pandocExpandDepth > 0 {
			expandType = "nestedExpand"
		}

		s.pandocExpandDepth++
		content, err := s.convertBlockFragment(node.Body())
		s.pandocExpandDepth--
		if err != nil {
			return converter.Node{}, false, err
		}

		expand := converter.Node{
			Type:    expandType,
			Content: content,
		}
		if title := strings.TrimSpace(node.Attrs["summary"]); title != "" {
			expand.Attrs = map[string]interface{}{
				"title": title,
			}
		}
		return expand, true, nil
	}

	if alignValue, hasAlign := node.Attrs["align"]; hasAlign {
		if !s.shouldDetectAlignPandoc() {
			return literalFallback, true, nil
		}

		alignment := normalizePandocAlignment(alignValue)
		if alignment == "" {
			s.addWarning(converter.WarningDroppedFeature, "pandocDiv", "invalid pandoc div alignment; preserved as text")
			return literalFallback, true, nil
		}

		content, err := s.convertBlockFragment(node.Body())
		if err != nil {
			return converter.Node{}, false, err
		}
		aligned := s.applyPandocAlignment(content, alignment)
		if len(aligned) == 0 {
			return converter.Node{}, false, nil
		}
		if len(aligned) == 1 {
			return aligned[0], true, nil
		}
		return converter.Node{
			Type:    "layoutSection",
			Content: aligned,
		}, true, nil
	}

	if hasUnknownPandocDivClass(node.Classes) {
		s.addWarning(converter.WarningDroppedFeature, "pandocDiv", "unknown pandoc div class converted to blockquote")
		content, err := s.convertBlockFragment(node.Body())
		if err != nil {
			return converter.Node{}, false, err
		}
		if len(content) == 0 {
			content = []converter.Node{pandocLiteralParagraph(node.Body())}
		}
		return converter.Node{
			Type:    "blockquote",
			Content: content,
		}, true, nil
	}

	return literalFallback, true, nil
}

func (s *state) applyPandocAlignment(content []converter.Node, alignment string) []converter.Node {
	if len(content) == 0 {
		return nil
	}

	out := make([]converter.Node, 0, len(content))
	for _, node := range content {
		next := node
		switch next.Type {
		case "paragraph", "heading":
			next.Attrs = cloneNodeAttrs(next.Attrs)
			if next.Attrs == nil {
				next.Attrs = map[string]interface{}{}
			}
			next.Attrs["layout"] = alignment
		default:
			s.addWarning(converter.WarningDroppedFeature, next.Type, "alignment skipped for unsupported block in pandoc div")
		}
		out = append(out, next)
	}
	return out
}

func normalizePandocAlignment(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "left", "center", "right":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func hasUnknownPandocDivClass(classes []string) bool {
	for _, className := range classes {
		if className != "details" {
			return true
		}
	}
	return false
}

func pandocLiteralParagraph(textValue string) converter.Node {
	return converter.Node{
		Type: "paragraph",
		Content: []converter.Node{
			{
				Type: "text",
				Text: textValue,
			},
		},
	}
}

func cloneNodeAttrs(attrs map[string]interface{}) map[string]interface{} {
	if attrs == nil {
		return nil
	}
	cloned := make(map[string]interface{}, len(attrs))
	for key, value := range attrs {
		cloned[key] = value
	}
	return cloned
}
