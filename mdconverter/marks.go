package mdconverter

import "github.com/rgonek/jira-adf-converter/converter"

type markStack struct {
	items []converter.Mark
}

func newMarkStack() *markStack {
	return &markStack{}
}

func (s *markStack) push(mark converter.Mark) {
	s.items = append(s.items, cloneMark(mark))
}

func (s *markStack) pop() (converter.Mark, bool) {
	if len(s.items) == 0 {
		return converter.Mark{}, false
	}

	last := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return cloneMark(last), true
}

func (s *markStack) popByType(markType string) bool {
	for i := len(s.items) - 1; i >= 0; i-- {
		if s.items[i].Type != markType {
			continue
		}
		s.items = append(s.items[:i], s.items[i+1:]...)
		return true
	}

	return false
}

func (s *markStack) current() []converter.Mark {
	if len(s.items) == 0 {
		return nil
	}

	marks := make([]converter.Mark, 0, len(s.items))
	for _, mark := range s.items {
		marks = append(marks, cloneMark(mark))
	}

	return marks
}

func cloneMark(mark converter.Mark) converter.Mark {
	cloned := mark
	if mark.Attrs != nil {
		cloned.Attrs = make(map[string]interface{}, len(mark.Attrs))
		for key, value := range mark.Attrs {
			cloned.Attrs[key] = value
		}
	}
	return cloned
}

func marksEqual(left, right []converter.Mark) bool {
	if len(left) != len(right) {
		return false
	}

	for idx := range left {
		if left[idx].Type != right[idx].Type {
			return false
		}
		if !attrsEqual(left[idx].Attrs, right[idx].Attrs) {
			return false
		}
	}

	return true
}

func attrsEqual(left, right map[string]interface{}) bool {
	if len(left) != len(right) {
		return false
	}
	for key, leftValue := range left {
		rightValue, ok := right[key]
		if !ok || leftValue != rightValue {
			return false
		}
	}
	return true
}

func newTextNode(textValue string, marks []converter.Mark) converter.Node {
	node := converter.Node{
		Type: "text",
		Text: textValue,
	}
	if len(marks) > 0 {
		node.Marks = marks
	}
	return node
}

func appendInlineNode(content []converter.Node, next converter.Node) []converter.Node {
	if next.Type == "text" && next.Text == "" {
		return content
	}

	if len(content) == 0 {
		return append(content, next)
	}

	last := &content[len(content)-1]
	if last.Type == "text" && next.Type == "text" && marksEqual(last.Marks, next.Marks) {
		last.Text += next.Text
		return content
	}

	return append(content, next)
}
