package mdconverter

import (
	"fmt"
	"strings"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
)

func (s *state) convertDocument(root ast.Node) (converter.Doc, error) {
	doc := converter.Doc{
		Version: 1,
		Type:    "doc",
	}

	for child := root.FirstChild(); child != nil; child = child.NextSibling() {
		block, ok, err := s.convertBlockNode(child)
		if err != nil {
			return converter.Doc{}, err
		}
		if ok {
			doc.Content = append(doc.Content, block)
		}
	}

	return doc, nil
}

func (s *state) convertBlockNode(node ast.Node) (converter.Node, bool, error) {
	switch typed := node.(type) {
	case *ast.Paragraph:
		return s.convertParagraphNode(typed)
	case *ast.TextBlock:
		return s.convertTextBlockNode(typed)
	case *ast.Heading:
		return s.convertHeadingNode(typed)
	case *ast.Blockquote:
		return s.convertBlockquoteNode(typed)
	case *ast.ThematicBreak:
		return converter.Node{Type: "rule"}, true, nil
	case *ast.FencedCodeBlock:
		return s.convertFencedCodeBlockNode(typed)
	case *ast.CodeBlock:
		return s.convertCodeBlockNode(typed)
	case *ast.List:
		return s.convertListNode(typed)
	case *extast.Table:
		s.addWarning(
			converter.WarningDroppedFeature,
			typed.Kind().String(),
			"table parsing is not implemented yet",
		)
		return converter.Node{}, false, nil
	default:
		nodeKind := typed.Kind().String()
		textValue := strings.TrimSpace(string(node.Text(s.source)))
		if textValue == "" {
			return converter.Node{}, false, nil
		}
		s.addWarning(
			converter.WarningUnknownNode,
			nodeKind,
			fmt.Sprintf("unsupported markdown block node: %s", nodeKind),
		)
		return converter.Node{
			Type: "paragraph",
			Content: []converter.Node{
				{
					Type: "text",
					Text: textValue,
				},
			},
		}, true, nil
	}
}

func (s *state) convertBlockChildren(parent ast.Node) ([]converter.Node, error) {
	var content []converter.Node
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		converted, ok, err := s.convertBlockNode(child)
		if err != nil {
			return nil, err
		}
		if ok {
			content = append(content, converted)
		}
	}
	return content, nil
}

func (s *state) warnUnknownInline(node ast.Node, stack *markStack) []converter.Node {
	textValue := strings.TrimSpace(string(node.Text(s.source)))
	if textValue == "" {
		return nil
	}

	nodeKind := node.Kind().String()
	s.addWarning(
		converter.WarningUnknownNode,
		nodeKind,
		fmt.Sprintf("unsupported markdown inline node: %s", nodeKind),
	)

	return []converter.Node{
		newTextNode(textValue, stack.current()),
	}
}
