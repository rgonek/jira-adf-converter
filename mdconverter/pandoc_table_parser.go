package mdconverter

import (
	"regexp"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var pandocGridBorderRe = regexp.MustCompile(`^\+[=-]+(?:\+[=-]+)+\+$`)

type PandocGridTableParser struct{}

func NewPandocGridTableParser() parser.BlockParser {
	return &PandocGridTableParser{}
}

func (p *PandocGridTableParser) Trigger() []byte {
	return []byte{'+'}
}

func (p *PandocGridTableParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	trimmed := strings.TrimLeft(trimLineEnding(string(line)), " \t")
	if !pandocGridBorderRe.MatchString(trimmed) {
		return nil, parser.NoChildren
	}

	node := NewPandocGridTableNode()
	node.appendLine(trimmed)
	reader.AdvanceLine()

	for {
		nextLine, _ := reader.PeekLine()
		if len(nextLine) == 0 {
			break
		}
		rawLine := strings.TrimLeft(trimLineEnding(string(nextLine)), " \t")
		if strings.HasPrefix(rawLine, "|") || strings.HasPrefix(rawLine, "+") {
			node.appendLine(rawLine)
			reader.AdvanceLine()
			continue
		}
		break
	}

	return node, parser.NoChildren
}

func (p *PandocGridTableParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	return parser.Close
}

func (p *PandocGridTableParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}

func (p *PandocGridTableParser) CanInterruptParagraph() bool {
	return true
}

func (p *PandocGridTableParser) CanAcceptIndentedLine() bool {
	return false
}
