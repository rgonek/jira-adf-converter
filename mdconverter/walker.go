package mdconverter

import (
	"fmt"
	"strings"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
)

func (s *state) convertDocument(root ast.Node) (converter.Doc, error) {
	if err := s.checkContext(); err != nil {
		return converter.Doc{}, err
	}

	doc := converter.Doc{
		Version: 1,
		Type:    "doc",
	}

	content, err := s.convertNodeSequence(root)
	if err != nil {
		return converter.Doc{}, err
	}
	doc.Content = content

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
	case *ast.HTMLBlock:
		return s.convertHTMLBlockNode(typed)
	case *extast.Table:
		return s.convertTableNode(typed)
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

func (s *state) convertNodeSequence(parent ast.Node) ([]converter.Node, error) {
	children := make([]ast.Node, 0, parent.ChildCount())
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		children = append(children, child)
	}
	return s.convertBlockSlice(children, parent)
}

func (s *state) convertBlockSlice(children []ast.Node, parent ast.Node) ([]converter.Node, error) {
	var content []converter.Node
	mergeNextParagraph := false

	for index := 0; index < len(children); {
		if err := s.checkContext(); err != nil {
			return nil, err
		}

		if s.shouldDetectExpandHTML() {
			if opening, ok := children[index].(*ast.HTMLBlock); ok {
				if title, isOpen := parseDetailsOpenTagFromHTMLBlock(opening, s.source); isOpen {
					expandNode, consumed, consumedOK, err := s.consumeDetailsBlock(children, index, parent, title)
					if err != nil {
						return nil, err
					}
					if consumedOK {
						content = s.appendConvertedBlock(content, expandNode, &mergeNextParagraph)
						index += consumed
						continue
					}
				}
			}
		}

		converted, ok, err := s.convertBlockNode(children[index])
		if err != nil {
			return nil, err
		}
		if ok {
			content = s.appendConvertedBlock(content, converted, &mergeNextParagraph)
		} else {
			mergeNextParagraph = false
		}
		index++
	}

	return content, nil
}

func (s *state) consumeDetailsBlock(children []ast.Node, start int, parent ast.Node, title string) (converter.Node, int, bool, error) {
	end := -1
	depth := 1
	for idx := start + 1; idx < len(children); idx++ {
		htmlNode, ok := children[idx].(*ast.HTMLBlock)
		if !ok {
			continue
		}
		if _, isOpen := parseDetailsOpenTagFromHTMLBlock(htmlNode, s.source); isOpen {
			depth++
			continue
		}
		if isDetailsCloseHTMLBlock(htmlNode, s.source) {
			depth--
			if depth == 0 {
				end = idx
				break
			}
		}
	}
	if end == -1 {
		return converter.Node{}, 0, false, nil
	}

	innerParent := parent
	if opening, ok := children[start].(*ast.HTMLBlock); ok {
		innerParent = opening
	}

	innerContent, err := s.convertBlockSlice(children[start+1:end], innerParent)
	if err != nil {
		return converter.Node{}, 0, false, err
	}

	expandType := "expand"
	if isNestedExpandContext(parent) {
		expandType = "nestedExpand"
	}

	expandNode := converter.Node{
		Type:    expandType,
		Content: innerContent,
	}
	if title != "" {
		expandNode.Attrs = map[string]interface{}{
			"title": title,
		}
	}

	return expandNode, end - start + 1, true, nil
}

func isNestedExpandContext(parent ast.Node) bool {
	switch parent.(type) {
	case *ast.ListItem, *ast.Blockquote, *ast.HTMLBlock:
		return true
	default:
		return false
	}
}

func (s *state) shouldDetectExpandHTML() bool {
	return s.config.ExpandDetection == ExpandDetectHTML || s.config.ExpandDetection == ExpandDetectAll
}

func (s *state) appendConvertedBlock(content []converter.Node, next converter.Node, mergeNextParagraph *bool) []converter.Node {
	if isInlineBlockNodeType(next.Type) {
		if len(content) == 0 || content[len(content)-1].Type != "paragraph" {
			content = append(content, converter.Node{
				Type:    "paragraph",
				Content: []converter.Node{next},
			})
		} else {
			lastParagraph := &content[len(content)-1]
			lastParagraph.Content = append(lastParagraph.Content, next)
		}
		*mergeNextParagraph = true
		return content
	}

	if *mergeNextParagraph && next.Type == "paragraph" && len(content) > 0 && content[len(content)-1].Type == "paragraph" {
		lastParagraph := &content[len(content)-1]
		for _, inlineNode := range next.Content {
			lastParagraph.Content = appendInlineNode(lastParagraph.Content, inlineNode)
		}
		*mergeNextParagraph = false
		return content
	}

	*mergeNextParagraph = false
	return append(content, next)
}

func isInlineBlockNodeType(nodeType string) bool {
	switch nodeType {
	case "inlineCard", "inlineExtension", "mention", "emoji", "status", "date":
		return true
	default:
		return false
	}
}

func (s *state) convertBlockChildren(parent ast.Node) ([]converter.Node, error) {
	children := make([]ast.Node, 0, parent.ChildCount())
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		children = append(children, child)
	}
	return s.convertBlockSlice(children, parent)
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
