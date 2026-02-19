package mdconverter

import (
	"strconv"
	"strings"

	"github.com/yuin/goldmark/ast"
)

var (
	KindPandocDiv       = ast.NewNodeKind("PandocDiv")
	KindPandocGridTable = ast.NewNodeKind("PandocGridTable")
)

type PandocDivNode struct {
	ast.BaseBlock
	FenceLength int
	RawAttrs    string
	Classes     []string
	Attrs       map[string]string
	bodyLines   []string
	openDepth   int
}

func NewPandocDivNode(fenceLength int, rawAttrs string, classes []string, attrs map[string]string) *PandocDivNode {
	return &PandocDivNode{
		FenceLength: fenceLength,
		RawAttrs:    rawAttrs,
		Classes:     classes,
		Attrs:       attrs,
		openDepth:   1,
	}
}

func (n *PandocDivNode) Kind() ast.NodeKind {
	return KindPandocDiv
}

func (n *PandocDivNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"RawAttrs":    n.RawAttrs,
		"FenceLength": strconv.Itoa(n.FenceLength),
	}, nil)
}

func (n *PandocDivNode) appendBodyLine(line string) {
	n.bodyLines = append(n.bodyLines, line)
}

func (n *PandocDivNode) Body() string {
	return strings.Join(n.bodyLines, "\n")
}

func (n *PandocDivNode) Literal() string {
	var builder strings.Builder
	builder.WriteString(strings.Repeat(":", maxInt(n.FenceLength, 3)))
	builder.WriteString("{")
	builder.WriteString(n.RawAttrs)
	builder.WriteString("}\n")
	if len(n.bodyLines) > 0 {
		builder.WriteString(n.Body())
		builder.WriteString("\n")
	}
	builder.WriteString(strings.Repeat(":", 3))
	return builder.String()
}

type PandocGridTableNode struct {
	ast.BaseBlock
	lines []string
}

func NewPandocGridTableNode() *PandocGridTableNode {
	return &PandocGridTableNode{}
}

func (n *PandocGridTableNode) Kind() ast.NodeKind {
	return KindPandocGridTable
}

func (n *PandocGridTableNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"LineCount": strconv.Itoa(len(n.lines)),
	}, nil)
}

func (n *PandocGridTableNode) appendLine(line string) {
	n.lines = append(n.lines, line)
}

func (n *PandocGridTableNode) RawLines() []string {
	if len(n.lines) == 0 {
		return nil
	}
	out := make([]string, len(n.lines))
	copy(out, n.lines)
	return out
}

func (n *PandocGridTableNode) Literal() string {
	return strings.Join(n.lines, "\n")
}

func trimLineEnding(raw string) string {
	raw = strings.TrimSuffix(raw, "\n")
	return strings.TrimSuffix(raw, "\r")
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
