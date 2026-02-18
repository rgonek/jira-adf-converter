package mdconverter

import (
	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
)

func (s *state) convertListNode(node *ast.List) (converter.Node, bool, error) {
	if s.isTaskList(node) {
		return s.convertTaskListNode(node)
	}

	listNode := converter.Node{
		Type: "bulletList",
	}
	if node.IsOrdered() {
		listNode.Type = "orderedList"
		if node.Start > 0 {
			listNode.Attrs = map[string]interface{}{
				"order": node.Start,
			}
		}
	}

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		itemNode, ok, err := s.convertListItemNode(child)
		if err != nil {
			return converter.Node{}, false, err
		}
		if ok {
			listNode.Content = append(listNode.Content, itemNode)
		}
	}

	if len(listNode.Content) == 0 {
		return converter.Node{}, false, nil
	}

	return listNode, true, nil
}

func (s *state) convertListItemNode(node ast.Node) (converter.Node, bool, error) {
	listItem, ok := node.(*ast.ListItem)
	if !ok {
		return converter.Node{}, false, nil
	}

	itemNode := converter.Node{
		Type: "listItem",
	}

	content, err := s.convertBlockChildren(listItem)
	if err != nil {
		return converter.Node{}, false, err
	}
	itemNode.Content = content

	return itemNode, true, nil
}

func (s *state) isTaskList(node *ast.List) bool {
	hasTaskItems := false

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		item, ok := child.(*ast.ListItem)
		if !ok {
			return false
		}

		container := item.FirstChild()
		if container == nil {
			return false
		}

		switch typed := container.(type) {
		case *ast.TextBlock:
			if _, ok := typed.FirstChild().(*extast.TaskCheckBox); ok {
				hasTaskItems = true
				continue
			}
			return false
		case *ast.Paragraph:
			if _, ok := typed.FirstChild().(*extast.TaskCheckBox); ok {
				hasTaskItems = true
				continue
			}
			return false
		default:
			return false
		}
	}

	return hasTaskItems
}

func (s *state) convertTaskListNode(node *ast.List) (converter.Node, bool, error) {
	taskList := converter.Node{
		Type: "taskList",
	}

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		item, ok := child.(*ast.ListItem)
		if !ok {
			continue
		}

		taskItem, nestedLists, hasTaskItem, err := s.convertTaskListItem(item)
		if err != nil {
			return converter.Node{}, false, err
		}
		if hasTaskItem {
			taskList.Content = append(taskList.Content, taskItem)
		}
		taskList.Content = append(taskList.Content, nestedLists...)
	}

	if len(taskList.Content) == 0 {
		return converter.Node{}, false, nil
	}

	return taskList, true, nil
}

func (s *state) convertTaskListItem(node *ast.ListItem) (converter.Node, []converter.Node, bool, error) {
	taskItem := converter.Node{
		Type: "taskItem",
		Attrs: map[string]interface{}{
			"state": "TODO",
		},
	}

	var nestedLists []converter.Node
	hasInlineContent := false

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch typed := child.(type) {
		case *ast.TextBlock, *ast.Paragraph:
			inlineContent, state, hasCheckbox, err := s.convertTaskInlineContent(typed)
			if err != nil {
				return converter.Node{}, nil, false, err
			}
			if hasCheckbox {
				taskItem.Attrs["state"] = state
				hasInlineContent = true
			}
			if len(inlineContent) == 0 {
				continue
			}
			if len(taskItem.Content) > 0 {
				taskItem.Content = append(taskItem.Content, converter.Node{Type: "hardBreak"})
			}
			for _, inlineNode := range inlineContent {
				taskItem.Content = appendInlineNode(taskItem.Content, inlineNode)
			}
			hasInlineContent = true
		case *ast.List:
			converted, ok, err := s.convertListNode(typed)
			if err != nil {
				return converter.Node{}, nil, false, err
			}
			if ok {
				nestedLists = append(nestedLists, converted)
			}
		default:
			converted, ok, err := s.convertBlockNode(typed)
			if err != nil {
				return converter.Node{}, nil, false, err
			}
			if ok {
				nestedLists = append(nestedLists, converted)
			}
		}
	}

	if !hasInlineContent {
		return converter.Node{}, nestedLists, false, nil
	}

	return taskItem, nestedLists, true, nil
}

func (s *state) convertTaskInlineContent(container ast.Node) ([]converter.Node, string, bool, error) {
	stack := newMarkStack()
	state := "TODO"
	hasCheckbox := false
	var content []converter.Node

	for child := container.FirstChild(); child != nil; child = child.NextSibling() {
		checkbox, isCheckbox := child.(*extast.TaskCheckBox)
		if isCheckbox {
			hasCheckbox = true
			if checkbox.IsChecked {
				state = "DONE"
			}
			continue
		}

		converted, err := s.convertInlineNode(child, stack)
		if err != nil {
			return nil, "", false, err
		}
		for _, node := range converted {
			if node.Type == "mediaSingle" || node.Type == "table" {
				s.addWarning(
					converter.WarningDroppedFeature,
					node.Type,
					"task item only supports inline content; embedded block converted to placeholder text",
				)
				node = newTextNode("[Embedded content]", nil)
			}
			content = appendInlineNode(content, node)
		}
	}

	return s.applyInlinePatterns(content), state, hasCheckbox, nil
}
