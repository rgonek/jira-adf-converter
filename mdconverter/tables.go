package mdconverter

import (
	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
)

func (s *state) convertTableNode(node *extast.Table) (converter.Node, bool, error) {
	table := converter.Node{
		Type: "table",
	}

	for row := node.FirstChild(); row != nil; row = row.NextSibling() {
		converted, ok, err := s.convertTableRowNode(row)
		if err != nil {
			return converter.Node{}, false, err
		}
		if ok {
			table.Content = append(table.Content, converted)
		}
	}

	if len(table.Content) == 0 {
		return converter.Node{}, false, nil
	}

	return table, true, nil
}

func (s *state) convertTableRowNode(node ast.Node) (converter.Node, bool, error) {
	row := converter.Node{
		Type: "tableRow",
	}

	isHeader := false
	switch typed := node.(type) {
	case *extast.TableHeader:
		isHeader = true
		for cell := typed.FirstChild(); cell != nil; cell = cell.NextSibling() {
			converted, ok, err := s.convertTableCellNode(cell, isHeader)
			if err != nil {
				return converter.Node{}, false, err
			}
			if ok {
				row.Content = append(row.Content, converted)
			}
		}
	case *extast.TableRow:
		for cell := typed.FirstChild(); cell != nil; cell = cell.NextSibling() {
			converted, ok, err := s.convertTableCellNode(cell, isHeader)
			if err != nil {
				return converter.Node{}, false, err
			}
			if ok {
				row.Content = append(row.Content, converted)
			}
		}
	default:
		return converter.Node{}, false, nil
	}

	if len(row.Content) == 0 {
		return converter.Node{}, false, nil
	}

	return row, true, nil
}

func (s *state) convertTableCellNode(node ast.Node, isHeader bool) (converter.Node, bool, error) {
	cell, ok := node.(*extast.TableCell)
	if !ok {
		return converter.Node{}, false, nil
	}

	cellType := "tableCell"
	if isHeader {
		cellType = "tableHeader"
	}

	inlineContent, err := s.convertInlineChildren(cell, newMarkStack())
	if err != nil {
		return converter.Node{}, false, err
	}

	cellNode := converter.Node{
		Type: cellType,
		Content: []converter.Node{
			{
				Type:    "paragraph",
				Content: inlineContent,
			},
		},
	}

	if cell.Alignment != extast.AlignNone {
		cellNode.Attrs = map[string]interface{}{
			"alignment": cell.Alignment.String(),
		}
	}

	return cellNode, true, nil
}
