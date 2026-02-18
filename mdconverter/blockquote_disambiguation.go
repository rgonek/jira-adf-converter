package mdconverter

import (
	"regexp"
	"strings"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/yuin/goldmark/ast"
)

var panelCalloutPattern = regexp.MustCompile(`(?i)^\[!([a-z]+)(?::\s*([^\]]+))?\](?:\s*(.*))?$`)

func (s *state) tryConvertPanelBlockquote(content []converter.Node) (converter.Node, bool) {
	if len(content) == 0 || s.config.PanelDetection == PanelDetectNone {
		return converter.Node{}, false
	}

	firstParagraph, ok := asParagraphNode(content[0])
	if !ok {
		return converter.Node{}, false
	}

	if s.config.PanelDetection == PanelDetectGitHub ||
		s.config.PanelDetection == PanelDetectTitle ||
		s.config.PanelDetection == PanelDetectAll {
		if panelNode, matched := detectCalloutPanel(firstParagraph, content[1:]); matched {
			return panelNode, true
		}
	}

	if s.config.PanelDetection == PanelDetectBold || s.config.PanelDetection == PanelDetectAll {
		if panelNode, matched := detectBoldPanel(firstParagraph, content[1:]); matched {
			return panelNode, true
		}
	}

	return converter.Node{}, false
}

func detectCalloutPanel(firstParagraph converter.Node, remaining []converter.Node) (converter.Node, bool) {
	textValue := strings.TrimSpace(paragraphPlainText(firstParagraph))
	match := panelCalloutPattern.FindStringSubmatch(textValue)
	if len(match) != 4 {
		return converter.Node{}, false
	}

	panelType := normalizePanelType(match[1])
	if panelType == "" {
		return converter.Node{}, false
	}

	panel := converter.Node{
		Type: "panel",
		Attrs: map[string]interface{}{
			"panelType": panelType,
		},
		Content: []converter.Node{},
	}
	if title := strings.TrimSpace(match[2]); title != "" {
		panel.Attrs["title"] = title
	}
	if firstLineContent := strings.TrimSpace(match[3]); firstLineContent != "" {
		panel.Content = append(panel.Content, converter.Node{
			Type: "paragraph",
			Content: []converter.Node{
				{
					Type: "text",
					Text: firstLineContent,
				},
			},
		})
	}
	panel.Content = append(panel.Content, cloneNodes(remaining)...)

	return panel, true
}

func detectBoldPanel(firstParagraph converter.Node, remaining []converter.Node) (converter.Node, bool) {
	label, remainder, ok := extractLeadingStrongPrefix(firstParagraph.Content)
	if !ok {
		return converter.Node{}, false
	}

	panelType := normalizePanelType(label)
	if panelType == "" {
		return converter.Node{}, false
	}

	trimmedRemainder, hasPrefix := trimLeadingColon(remainder)
	if !hasPrefix {
		return converter.Node{}, false
	}

	panelContent := make([]converter.Node, 0, len(remaining)+1)
	if len(trimmedRemainder) > 0 {
		panelContent = append(panelContent, converter.Node{
			Type:    "paragraph",
			Content: trimmedRemainder,
		})
	}
	panelContent = append(panelContent, cloneNodes(remaining)...)

	return converter.Node{
		Type: "panel",
		Attrs: map[string]interface{}{
			"panelType": panelType,
		},
		Content: panelContent,
	}, true
}

func normalizePanelType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "info":
		return "info"
	case "note":
		return "note"
	case "success":
		return "success"
	case "warning":
		return "warning"
	case "error":
		return "error"
	default:
		return ""
	}
}

func (s *state) tryConvertDecisionBlockquote(content []converter.Node) (converter.Node, bool) {
	if len(content) == 0 || s.config.DecisionDetection == DecisionDetectNone {
		return converter.Node{}, false
	}

	items := make([]converter.Node, 0, 2)
	currentItemIndex := -1

	for _, block := range content {
		paragraph, isParagraph := asParagraphNode(block)
		if isParagraph {
			state, trimmedParagraph, matched := s.parseDecisionPrefix(paragraph)
			if matched {
				item := converter.Node{
					Type: "decisionItem",
				}
				if state != "" {
					item.Attrs = map[string]interface{}{
						"state": state,
					}
				}
				if len(trimmedParagraph.Content) > 0 {
					item.Content = append(item.Content, trimmedParagraph)
				}
				items = append(items, item)
				currentItemIndex = len(items) - 1
				continue
			}
		}

		if currentItemIndex == -1 {
			return converter.Node{}, false
		}

		items[currentItemIndex].Content = append(items[currentItemIndex].Content, cloneNode(block))
	}

	if len(items) == 0 {
		return converter.Node{}, false
	}

	return converter.Node{
		Type:    "decisionList",
		Content: items,
	}, true
}

func (s *state) parseDecisionPrefix(paragraph converter.Node) (string, converter.Node, bool) {
	label, remainder, ok := extractLeadingStrongPrefix(paragraph.Content)
	if !ok {
		return "", converter.Node{}, false
	}

	remainder, hasColon := trimLeadingColon(remainder)
	if !hasColon {
		return "", converter.Node{}, false
	}

	state, matched := s.matchDecisionLabel(label)
	if !matched {
		return "", converter.Node{}, false
	}

	return state, converter.Node{
		Type:    "paragraph",
		Content: remainder,
	}, true
}

func (s *state) matchDecisionLabel(label string) (string, bool) {
	normalized := strings.ToUpper(strings.TrimSpace(label))

	switch normalized {
	case "✓ DECISION":
		return "DECIDED", s.config.DecisionDetection == DecisionDetectEmoji || s.config.DecisionDetection == DecisionDetectAll
	case "? DECISION":
		return "UNDECIDED", s.config.DecisionDetection == DecisionDetectEmoji || s.config.DecisionDetection == DecisionDetectAll
	case "DECIDED":
		return "DECIDED", s.config.DecisionDetection == DecisionDetectText || s.config.DecisionDetection == DecisionDetectAll
	case "UNDECIDED":
		return "UNDECIDED", s.config.DecisionDetection == DecisionDetectText || s.config.DecisionDetection == DecisionDetectAll
	case "DECISION":
		return "", s.config.DecisionDetection != DecisionDetectNone
	default:
		return "", false
	}
}

func (s *state) tryConvertExpandBlockquote(node *ast.Blockquote, content []converter.Node) (converter.Node, bool) {
	if s.config.ExpandDetection == ExpandDetectNone || s.config.ExpandDetection == ExpandDetectHTML {
		return converter.Node{}, false
	}

	expandType := "expand"
	if isNestedExpandContext(node.Parent()) {
		expandType = "nestedExpand"
	}

	title, remaining, hasTitle := extractExpandTitle(content)
	if hasTitle {
		expand := converter.Node{
			Type:    expandType,
			Content: remaining,
			Attrs: map[string]interface{}{
				"title": title,
			},
		}
		return expand, true
	}

	if s.config.ExpandDetection == ExpandDetectBlockquote {
		return converter.Node{
			Type:    expandType,
			Content: cloneNodes(content),
		}, true
	}

	return converter.Node{}, false
}

func extractExpandTitle(content []converter.Node) (string, []converter.Node, bool) {
	if len(content) == 0 {
		return "", nil, false
	}

	firstParagraph, ok := asParagraphNode(content[0])
	if !ok {
		return "", nil, false
	}

	title, remainder, ok := extractLeadingStrongPrefix(firstParagraph.Content)
	if !ok || strings.TrimSpace(paragraphInlineText(remainder)) != "" {
		return "", nil, false
	}

	return title, cloneNodes(content[1:]), true
}

func asParagraphNode(node converter.Node) (converter.Node, bool) {
	return node, node.Type == "paragraph"
}

func extractLeadingStrongPrefix(content []converter.Node) (string, []converter.Node, bool) {
	if len(content) == 0 {
		return "", nil, false
	}

	label := strings.Builder{}
	index := 0
	for ; index < len(content); index++ {
		node := content[index]
		if node.Type != "text" || !hasStrongMark(node.Marks) {
			break
		}
		label.WriteString(node.Text)
	}

	if index == 0 {
		return "", nil, false
	}

	return strings.TrimSpace(label.String()), cloneNodes(content[index:]), true
}

func hasStrongMark(marks []converter.Mark) bool {
	for _, mark := range marks {
		if mark.Type == "strong" {
			return true
		}
	}
	return false
}

func trimLeadingColon(content []converter.Node) ([]converter.Node, bool) {
	if len(content) == 0 {
		return nil, false
	}

	trimmed := cloneNodes(content)
	for index := 0; index < len(trimmed); index++ {
		node := trimmed[index]
		if node.Type != "text" {
			return nil, false
		}

		textValue := node.Text
		if strings.TrimSpace(textValue) == "" {
			continue
		}

		firstNonSpace := 0
		for firstNonSpace < len(textValue) && (textValue[firstNonSpace] == ' ' || textValue[firstNonSpace] == '\t') {
			firstNonSpace++
		}
		if firstNonSpace >= len(textValue) || textValue[firstNonSpace] != ':' {
			return nil, false
		}

		updated := strings.TrimLeft(textValue[firstNonSpace+1:], " \t")
		trimmed[index].Text = updated

		out := trimmed[index:]
		if len(out) > 0 && out[0].Type == "text" && out[0].Text == "" {
			out = out[1:]
		}
		return out, true
	}

	return nil, false
}

func paragraphPlainText(paragraph converter.Node) string {
	return paragraphInlineText(paragraph.Content)
}

func paragraphInlineText(content []converter.Node) string {
	var textBuilder strings.Builder
	for _, node := range content {
		switch node.Type {
		case "text":
			textBuilder.WriteString(node.Text)
		case "hardBreak":
			textBuilder.WriteString("\n")
		}
	}
	return textBuilder.String()
}

func cloneNodes(nodes []converter.Node) []converter.Node {
	if len(nodes) == 0 {
		return nil
	}
	cloned := make([]converter.Node, 0, len(nodes))
	for _, node := range nodes {
		cloned = append(cloned, cloneNode(node))
	}
	return cloned
}

func cloneNode(node converter.Node) converter.Node {
	cloned := node
	cloned.Content = cloneNodes(node.Content)

	if len(node.Marks) > 0 {
		cloned.Marks = make([]converter.Mark, 0, len(node.Marks))
		for _, mark := range node.Marks {
			clonedMark := mark
			if mark.Attrs != nil {
				clonedMark.Attrs = make(map[string]interface{}, len(mark.Attrs))
				for key, value := range mark.Attrs {
					clonedMark.Attrs[key] = value
				}
			}
			cloned.Marks = append(cloned.Marks, clonedMark)
		}
	}

	if node.Attrs != nil {
		cloned.Attrs = make(map[string]interface{}, len(node.Attrs))
		for key, value := range node.Attrs {
			cloned.Attrs[key] = value
		}
	}

	return cloned
}
