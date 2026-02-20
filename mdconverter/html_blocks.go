package mdconverter

import (
	"fmt"
	stdhtml "html"
	"regexp"
	"strconv"
	"strings"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	xhtml "golang.org/x/net/html"
)

var (
	detailsOpenPattern       = regexp.MustCompile(`(?is)^<details>\s*<summary>(.*?)</summary>\s*$`)
	detailsClosePattern      = regexp.MustCompile(`(?is)^</details>\s*$`)
	alignedDivPattern        = regexp.MustCompile(`(?is)^<div\s+align="(left|center|right)"\s*>(.*?)</div>\s*$`)
	alignedHeadingPattern    = regexp.MustCompile(`(?is)^<h([1-6])\s+align="(left|center|right)"\s*>(.*?)</h[1-6]>\s*$`)
	layoutSectionOpenPattern = regexp.MustCompile(`(?is)^<div\s+class="layout-section"\s*>\s*$`)
	layoutColumnOpenPattern  = regexp.MustCompile(`(?is)^<div\s+class="layout-column"(?:\s+style="width:\s*([0-9.]+)%;")?\s*>\s*$`)
	divClosePattern          = regexp.MustCompile(`(?is)^</div>\s*$`)
)

func parseDetailsOpenTagFromHTMLBlock(node *ast.HTMLBlock, source []byte) (string, bool) {
	return parseDetailsOpenTag(strings.TrimSpace(string(node.Text(source))))
}

func parseDetailsOpenTag(raw string) (string, bool) {
	match := detailsOpenPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if len(match) != 2 {
		return "", false
	}

	title := strings.TrimSpace(stdhtml.UnescapeString(match[1]))
	return title, true
}

func isDetailsCloseHTMLBlock(node *ast.HTMLBlock, source []byte) bool {
	return isDetailsCloseHTML(strings.TrimSpace(string(node.Text(source))))
}

func isDetailsCloseHTML(raw string) bool {
	return detailsClosePattern.MatchString(strings.TrimSpace(raw))
}

func parseLayoutSectionOpenTagFromHTMLBlock(node *ast.HTMLBlock, source []byte) bool {
	return parseLayoutSectionOpenTag(strings.TrimSpace(string(node.Text(source))))
}

func parseLayoutSectionOpenTag(raw string) bool {
	return layoutSectionOpenPattern.MatchString(strings.TrimSpace(raw))
}

func parseLayoutColumnOpenTagFromHTMLBlock(node *ast.HTMLBlock, source []byte) (float64, bool) {
	return parseLayoutColumnOpenTag(strings.TrimSpace(string(node.Text(source))))
}

func parseLayoutColumnOpenTag(raw string) (float64, bool) {
	match := layoutColumnOpenPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if len(match) == 0 {
		return 0, false
	}
	if match[1] != "" {
		if width, err := strconv.ParseFloat(match[1], 64); err == nil {
			return width, true
		}
	}
	return 0, true
}

func isDivCloseHTMLBlock(node *ast.HTMLBlock, source []byte) bool {
	return isDivCloseHTML(strings.TrimSpace(string(node.Text(source))))
}

func isDivCloseHTML(raw string) bool {
	return divClosePattern.MatchString(strings.TrimSpace(raw))
}

func (s *state) convertHTMLBlockNode(node *ast.HTMLBlock) (converter.Node, bool, error) {
	raw := strings.TrimSpace(string(node.Text(s.source)))
	if raw == "" {
		return converter.Node{}, false, nil
	}

	if headingNode, ok, err := s.parseAlignedHeading(raw); ok || err != nil {
		return headingNode, ok, err
	}
	if paragraphNode, ok, err := s.parseAlignedParagraph(raw); ok || err != nil {
		return paragraphNode, ok, err
	}
	if tableNode, ok, err := s.parseHTMLTable(raw); ok || err != nil {
		return tableNode, ok, err
	}

	s.addWarning(
		converter.WarningUnknownNode,
		node.Kind().String(),
		"unsupported html block converted to text",
	)
	return converter.Node{
		Type: "paragraph",
		Content: []converter.Node{
			{
				Type: "text",
				Text: raw,
			},
		},
	}, true, nil
}

func (s *state) parseAlignedParagraph(raw string) (converter.Node, bool, error) {
	if !s.shouldDetectAlignHTML() {
		return converter.Node{}, false, nil
	}

	match := alignedDivPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if len(match) != 3 {
		return converter.Node{}, false, nil
	}

	inlineContent, err := s.convertInlineFragment(match[2])
	if err != nil {
		return converter.Node{}, false, err
	}

	return converter.Node{
		Type: "paragraph",
		Attrs: map[string]interface{}{
			"layout": strings.ToLower(match[1]),
		},
		Content: inlineContent,
	}, true, nil
}

func (s *state) parseAlignedHeading(raw string) (converter.Node, bool, error) {
	if !s.shouldDetectAlignHTML() {
		return converter.Node{}, false, nil
	}

	match := alignedHeadingPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if len(match) != 4 {
		return converter.Node{}, false, nil
	}

	level, err := strconv.Atoi(match[1])
	if err != nil {
		return converter.Node{}, false, fmt.Errorf("invalid heading level in html block: %w", err)
	}
	level += s.config.HeadingOffset
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}

	inlineContent, err := s.convertInlineFragment(match[3])
	if err != nil {
		return converter.Node{}, false, err
	}

	return converter.Node{
		Type: "heading",
		Attrs: map[string]interface{}{
			"level": level,
			"align": strings.ToLower(match[2]),
		},
		Content: inlineContent,
	}, true, nil
}

func (s *state) parseHTMLTable(raw string) (converter.Node, bool, error) {
	if !strings.Contains(strings.ToLower(raw), "<table") {
		return converter.Node{}, false, nil
	}

	document, err := xhtml.Parse(strings.NewReader(raw))
	if err != nil {
		return converter.Node{}, false, fmt.Errorf("failed to parse html table: %w", err)
	}

	tableElement := findHTMLElement(document, "table")
	if tableElement == nil {
		return converter.Node{}, false, nil
	}

	rows, err := s.convertHTMLTableRows(tableElement)
	if err != nil {
		return converter.Node{}, false, err
	}

	if len(rows) == 0 {
		return converter.Node{}, false, nil
	}

	return converter.Node{
		Type:    "table",
		Content: rows,
	}, true, nil
}

func findHTMLElement(node *xhtml.Node, tag string) *xhtml.Node {
	if node == nil {
		return nil
	}
	if node.Type == xhtml.ElementNode && strings.EqualFold(node.Data, tag) {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findHTMLElement(child, tag); found != nil {
			return found
		}
	}
	return nil
}

func (s *state) convertHTMLTableRows(table *xhtml.Node) ([]converter.Node, error) {
	rows := make([]converter.Node, 0, 4)

	for child := table.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != xhtml.ElementNode {
			continue
		}

		switch strings.ToLower(child.Data) {
		case "thead":
			sectionRows, err := s.convertHTMLTableSection(child, true)
			if err != nil {
				return nil, err
			}
			rows = append(rows, sectionRows...)
		case "tbody", "tfoot":
			sectionRows, err := s.convertHTMLTableSection(child, false)
			if err != nil {
				return nil, err
			}
			rows = append(rows, sectionRows...)
		case "tr":
			rowNode, ok, err := s.convertHTMLTableRow(child, false)
			if err != nil {
				return nil, err
			}
			if ok {
				rows = append(rows, rowNode)
			}
		}
	}

	return rows, nil
}

func (s *state) convertHTMLTableSection(section *xhtml.Node, headerSection bool) ([]converter.Node, error) {
	rows := make([]converter.Node, 0, 2)
	for child := section.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != xhtml.ElementNode || !strings.EqualFold(child.Data, "tr") {
			continue
		}
		rowNode, ok, err := s.convertHTMLTableRow(child, headerSection)
		if err != nil {
			return nil, err
		}
		if ok {
			rows = append(rows, rowNode)
		}
	}
	return rows, nil
}

func (s *state) convertHTMLTableRow(row *xhtml.Node, headerSection bool) (converter.Node, bool, error) {
	rowNode := converter.Node{
		Type: "tableRow",
	}

	for cell := row.FirstChild; cell != nil; cell = cell.NextSibling {
		if cell.Type != xhtml.ElementNode {
			continue
		}
		tag := strings.ToLower(cell.Data)
		if tag != "td" && tag != "th" {
			continue
		}

		cellNode, ok, err := s.convertHTMLTableCell(cell, headerSection || tag == "th")
		if err != nil {
			return converter.Node{}, false, err
		}
		if ok {
			rowNode.Content = append(rowNode.Content, cellNode)
		}
	}

	if len(rowNode.Content) == 0 {
		return converter.Node{}, false, nil
	}

	return rowNode, true, nil
}

func (s *state) convertHTMLTableCell(cell *xhtml.Node, isHeader bool) (converter.Node, bool, error) {
	cellType := "tableCell"
	if isHeader {
		cellType = "tableHeader"
	}

	content := normalizeHTMLCellText(extractHTMLNodeText(cell))
	blocks, err := s.convertBlockFragment(content)
	if err != nil {
		return converter.Node{}, false, err
	}
	if len(blocks) == 0 {
		blocks = []converter.Node{{Type: "paragraph"}}
	}

	cellNode := converter.Node{
		Type:    cellType,
		Content: blocks,
	}

	attrs := map[string]interface{}{}
	if colspan := getIntHTMLAttr(cell, "colspan"); colspan > 1 {
		attrs["colspan"] = colspan
	}
	if rowspan := getIntHTMLAttr(cell, "rowspan"); rowspan > 1 {
		attrs["rowspan"] = rowspan
	}
	if len(attrs) > 0 {
		cellNode.Attrs = attrs
	}

	return cellNode, true, nil
}

func getIntHTMLAttr(node *xhtml.Node, key string) int {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, key) {
			parsed, err := strconv.Atoi(strings.TrimSpace(attr.Val))
			if err == nil {
				return parsed
			}
			return 0
		}
	}
	return 0
}

func extractHTMLNodeText(node *xhtml.Node) string {
	var builder strings.Builder

	var walk func(current *xhtml.Node)
	walk = func(current *xhtml.Node) {
		switch current.Type {
		case xhtml.TextNode:
			builder.WriteString(current.Data)
		case xhtml.ElementNode:
			if strings.EqualFold(current.Data, "br") {
				builder.WriteString("\n")
				return
			}
			for child := current.FirstChild; child != nil; child = child.NextSibling {
				walk(child)
			}
			switch strings.ToLower(current.Data) {
			case "p", "div", "li":
				builder.WriteString("\n")
			}
		default:
			for child := current.FirstChild; child != nil; child = child.NextSibling {
				walk(child)
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		walk(child)
	}

	return builder.String()
}

func normalizeHTMLCellText(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	lines := strings.Split(value, "\n")

	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return ""
	}

	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := countLeadingWhitespace(line)
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent > 0 {
		for index, line := range lines {
			if len(line) >= minIndent {
				lines[index] = line[minIndent:]
			}
		}
	}

	for index, line := range lines {
		lines[index] = strings.TrimRight(line, " \t")
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func countLeadingWhitespace(value string) int {
	count := 0
	for count < len(value) {
		if value[count] != ' ' && value[count] != '\t' {
			break
		}
		count++
	}
	return count
}

func (s *state) convertInlineFragment(fragment string) ([]converter.Node, error) {
	if err := s.checkContext(); err != nil {
		return nil, err
	}

	trimmed := strings.TrimSpace(fragment)
	if trimmed == "" {
		return nil, nil
	}

	originalSource := s.source
	originalMentionStack := s.htmlMentionStack
	originalSpanStack := s.htmlSpanStack
	defer func() {
		s.source = originalSource
		s.htmlMentionStack = originalMentionStack
		s.htmlSpanStack = originalSpanStack
	}()

	s.source = []byte(trimmed)
	s.htmlMentionStack = nil
	s.htmlSpanStack = nil

	root := s.parser.Parser().Parse(text.NewReader(s.source))
	if err := s.checkContext(); err != nil {
		return nil, err
	}
	content := make([]converter.Node, 0, 4)
	for child := root.FirstChild(); child != nil; child = child.NextSibling() {
		switch typed := child.(type) {
		case *ast.Paragraph:
			inlineContent, err := s.convertInlineChildren(typed, newMarkStack())
			if err != nil {
				return nil, err
			}
			for _, inlineNode := range inlineContent {
				content = appendInlineNode(content, inlineNode)
			}
		case *ast.TextBlock:
			inlineContent, err := s.convertInlineChildren(typed, newMarkStack())
			if err != nil {
				return nil, err
			}
			for _, inlineNode := range inlineContent {
				content = appendInlineNode(content, inlineNode)
			}
		}
	}

	return content, nil
}

func (s *state) convertBlockFragment(fragment string) ([]converter.Node, error) {
	if err := s.checkContext(); err != nil {
		return nil, err
	}

	trimmed := strings.TrimSpace(fragment)
	if trimmed == "" {
		return nil, nil
	}

	originalSource := s.source
	originalMentionStack := s.htmlMentionStack
	originalSpanStack := s.htmlSpanStack
	defer func() {
		s.source = originalSource
		s.htmlMentionStack = originalMentionStack
		s.htmlSpanStack = originalSpanStack
	}()

	s.source = []byte(trimmed)
	s.htmlMentionStack = nil
	s.htmlSpanStack = nil

	root := s.parser.Parser().Parse(text.NewReader(s.source))
	if err := s.checkContext(); err != nil {
		return nil, err
	}
	return s.convertNodeSequence(root)
}

