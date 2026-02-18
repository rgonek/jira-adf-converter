package mdconverter

import (
	"strings"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/yuin/goldmark/ast"
)

func (s *state) convertParagraphNode(node *ast.Paragraph) (converter.Node, bool, error) {
	content, err := s.convertInlineChildren(node, newMarkStack())
	if err != nil {
		return converter.Node{}, false, err
	}
	content = s.normalizeParagraphInline(content)

	if len(content) == 1 && isParagraphBlockReplacement(content[0].Type) {
		return content[0], true, nil
	}

	paragraph := converter.Node{
		Type:    "paragraph",
		Content: content,
	}

	return paragraph, true, nil
}

func (s *state) convertTextBlockNode(node *ast.TextBlock) (converter.Node, bool, error) {
	content, err := s.convertInlineChildren(node, newMarkStack())
	if err != nil {
		return converter.Node{}, false, err
	}
	content = s.normalizeParagraphInline(content)

	if len(content) == 1 && isParagraphBlockReplacement(content[0].Type) {
		return content[0], true, nil
	}

	paragraph := converter.Node{
		Type:    "paragraph",
		Content: content,
	}

	return paragraph, true, nil
}

func (s *state) convertHeadingNode(node *ast.Heading) (converter.Node, bool, error) {
	content, err := s.convertInlineChildren(node, newMarkStack())
	if err != nil {
		return converter.Node{}, false, err
	}

	level := node.Level + s.config.HeadingOffset
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}

	heading := converter.Node{
		Type:    "heading",
		Content: content,
		Attrs: map[string]interface{}{
			"level": level,
		},
	}

	return heading, true, nil
}

func (s *state) convertBlockquoteNode(node *ast.Blockquote) (converter.Node, bool, error) {
	content, err := s.convertBlockChildren(node)
	if err != nil {
		return converter.Node{}, false, err
	}

	if panelNode, ok := s.tryConvertPanelBlockquote(content); ok {
		return panelNode, true, nil
	}
	if decisionNode, ok := s.tryConvertDecisionBlockquote(content); ok {
		return decisionNode, true, nil
	}
	if expandNode, ok := s.tryConvertExpandBlockquote(node, content); ok {
		return expandNode, true, nil
	}

	blockquote := converter.Node{
		Type:    "blockquote",
		Content: content,
	}

	return blockquote, true, nil
}

func (s *state) convertFencedCodeBlockNode(node *ast.FencedCodeBlock) (converter.Node, bool, error) {
	language := strings.TrimSpace(string(node.Language(s.source)))
	textValue := string(node.Text(s.source))
	textValue = strings.TrimRight(textValue, "\n")

	if extensionNode, ok, err := s.parseExtensionFence(language, textValue); ok || err != nil {
		return extensionNode, ok, err
	}

	if mapped, ok := s.config.LanguageMap[language]; ok {
		language = mapped
	}

	codeBlock := converter.Node{
		Type: "codeBlock",
	}
	if language != "" {
		codeBlock.Attrs = map[string]interface{}{
			"language": language,
		}
	}

	if textValue != "" {
		codeBlock.Content = []converter.Node{
			{
				Type: "text",
				Text: textValue,
			},
		}
	}

	return codeBlock, true, nil
}

func (s *state) convertCodeBlockNode(node *ast.CodeBlock) (converter.Node, bool, error) {
	textValue := string(node.Text(s.source))
	textValue = strings.TrimRight(textValue, "\n")

	codeBlock := converter.Node{
		Type: "codeBlock",
	}
	if textValue != "" {
		codeBlock.Content = []converter.Node{
			{
				Type: "text",
				Text: textValue,
			},
		}
	}

	return codeBlock, true, nil
}

func (s *state) normalizeParagraphInline(content []converter.Node) []converter.Node {
	if len(content) <= 1 {
		return content
	}

	normalized := make([]converter.Node, 0, len(content))
	for _, node := range content {
		if isParagraphBlockReplacement(node.Type) {
			s.addWarning(
				converter.WarningDroppedFeature,
				node.Type,
				"inline block-style node mixed with text; converted to placeholder text",
			)
			normalized = appendInlineNode(normalized, newTextNode("[Embedded content]", nil))
			continue
		}
		normalized = appendInlineNode(normalized, node)
	}

	return normalized
}

func isParagraphBlockReplacement(nodeType string) bool {
	switch nodeType {
	case "mediaSingle", "table":
		return true
	default:
		return false
	}
}
