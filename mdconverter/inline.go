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

	return content, nil
}

func (s *state) convertInlineNode(node ast.Node, stack *markStack) ([]converter.Node, error) {
	switch typed := node.(type) {
	case *ast.Text:
		var content []converter.Node
		textValue := string(typed.Value(s.source))
		if textValue != "" {
			content = append(content, newTextNode(textValue, stack.current()))
		}

		if typed.HardLineBreak() {
			content = append(content, converter.Node{Type: "hardBreak"})
		} else if typed.SoftLineBreak() {
			content = append(content, newTextNode(" ", stack.current()))
		}

		return content, nil

	case *ast.String:
		return []converter.Node{
			newTextNode(string(typed.Value), stack.current()),
		}, nil

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

		mark := converter.Mark{
			Type: "link",
			Attrs: map[string]interface{}{
				"href": href,
			},
		}
		if title := strings.TrimSpace(string(typed.Title)); title != "" {
			mark.Attrs["title"] = title
		}

		stack.push(mark)
		content, err := s.convertInlineChildren(typed, stack)
		stack.popByType("link")
		return content, err

	case *ast.Image:
		alt := strings.TrimSpace(string(typed.Text(s.source)))
		if alt == "" {
			alt = "Image"
		}
		s.addWarning(
			converter.WarningDroppedFeature,
			typed.Kind().String(),
			"image node parsing to media is not implemented yet",
		)
		return []converter.Node{
			newTextNode(alt, stack.current()),
		}, nil

	default:
		if node.HasChildren() {
			return s.convertInlineChildren(node, stack)
		}
		return s.warnUnknownInline(node, stack), nil
	}
}
