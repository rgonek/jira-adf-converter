package mdconverter

import (
	"strings"

	"github.com/yuin/goldmark/ast"
)

var (
	KindPandocSubscript   = ast.NewNodeKind("PandocSubscript")
	KindPandocSuperscript = ast.NewNodeKind("PandocSuperscript")
	KindPandocSpan        = ast.NewNodeKind("PandocSpan")
)

type PandocSubscriptNode struct {
	ast.BaseInline
	Content string
}

func NewPandocSubscriptNode(content string) *PandocSubscriptNode {
	return &PandocSubscriptNode{Content: content}
}

func (n *PandocSubscriptNode) Kind() ast.NodeKind {
	return KindPandocSubscript
}

func (n *PandocSubscriptNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Content": n.Content,
	}, nil)
}

type PandocSuperscriptNode struct {
	ast.BaseInline
	Content string
}

func NewPandocSuperscriptNode(content string) *PandocSuperscriptNode {
	return &PandocSuperscriptNode{Content: content}
}

func (n *PandocSuperscriptNode) Kind() ast.NodeKind {
	return KindPandocSuperscript
}

func (n *PandocSuperscriptNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Content": n.Content,
	}, nil)
}

type PandocSpanNode struct {
	ast.BaseInline
	Content  string
	RawAttrs string
	Classes  []string
	Attrs    map[string]string
}

func NewPandocSpanNode(content, rawAttrs string, classes []string, attrs map[string]string) *PandocSpanNode {
	return &PandocSpanNode{
		Content:  content,
		RawAttrs: rawAttrs,
		Classes:  classes,
		Attrs:    attrs,
	}
}

func (n *PandocSpanNode) Kind() ast.NodeKind {
	return KindPandocSpan
}

func (n *PandocSpanNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Content":  n.Content,
		"RawAttrs": n.RawAttrs,
		"Classes":  strings.Join(n.Classes, ","),
	}, nil)
}
