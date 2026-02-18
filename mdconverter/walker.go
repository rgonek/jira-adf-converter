package mdconverter

import (
	"fmt"
	"strings"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/yuin/goldmark/ast"
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
		return s.convertParagraph(typed), true, nil
	default:
		nodeKind := typed.Kind().String()
		s.addWarning(
			converter.WarningUnknownNode,
			nodeKind,
			fmt.Sprintf("unsupported markdown block node: %s", nodeKind),
		)
		return converter.Node{}, false, nil
	}
}

func (s *state) convertParagraph(node *ast.Paragraph) converter.Node {
	paragraph := converter.Node{
		Type: "paragraph",
	}

	for inline := node.FirstChild(); inline != nil; inline = inline.NextSibling() {
		for _, converted := range s.convertInlineNode(inline) {
			if converted.Type == "text" && len(paragraph.Content) > 0 {
				last := &paragraph.Content[len(paragraph.Content)-1]
				if last.Type == "text" && len(last.Marks) == 0 && len(converted.Marks) == 0 {
					last.Text += converted.Text
					continue
				}
			}
			paragraph.Content = append(paragraph.Content, converted)
		}
	}

	return paragraph
}

func (s *state) convertInlineNode(node ast.Node) []converter.Node {
	switch typed := node.(type) {
	case *ast.Text:
		textValue := string(typed.Value(s.source))
		content := []converter.Node{}

		if textValue != "" {
			content = append(content, converter.Node{
				Type: "text",
				Text: textValue,
			})
		}

		if typed.HardLineBreak() {
			content = append(content, converter.Node{Type: "hardBreak"})
		} else if typed.SoftLineBreak() {
			content = append(content, converter.Node{
				Type: "text",
				Text: "\n",
			})
		}

		return content
	case *ast.String:
		if len(typed.Value) == 0 {
			return nil
		}
		return []converter.Node{
			{
				Type: "text",
				Text: string(typed.Value),
			},
		}
	default:
		nodeKind := typed.Kind().String()
		if strings.TrimSpace(string(node.Text(s.source))) == "" {
			return nil
		}
		s.addWarning(
			converter.WarningUnknownNode,
			nodeKind,
			fmt.Sprintf("unsupported markdown inline node: %s", nodeKind),
		)
		return []converter.Node{
			{
				Type: "text",
				Text: string(node.Text(s.source)),
			},
		}
	}
}
