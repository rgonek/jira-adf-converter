package mdconverter

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type SubscriptParser struct{}

func NewSubscriptParser() parser.InlineParser {
	return &SubscriptParser{}
}

func (p *SubscriptParser) Trigger() []byte {
	return []byte{'~'}
}

func (p *SubscriptParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) < 3 || line[0] != '~' {
		return nil
	}
	if line[1] == '~' {
		return nil
	}

	closing := -1
	for idx := 1; idx < len(line); idx++ {
		if line[idx] == '\n' || line[idx] == '\r' {
			break
		}
		if line[idx] == '~' {
			closing = idx
			break
		}
	}
	if closing <= 1 {
		return nil
	}

	block.Advance(closing + 1)
	return NewPandocSubscriptNode(string(line[1:closing]))
}

type SuperscriptParser struct{}

func NewSuperscriptParser() parser.InlineParser {
	return &SuperscriptParser{}
}

func (p *SuperscriptParser) Trigger() []byte {
	return []byte{'^'}
}

func (p *SuperscriptParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) < 3 || line[0] != '^' {
		return nil
	}

	closing := -1
	for idx := 1; idx < len(line); idx++ {
		if line[idx] == '\n' || line[idx] == '\r' {
			break
		}
		if line[idx] == '^' {
			closing = idx
			break
		}
	}
	if closing <= 1 {
		return nil
	}

	block.Advance(closing + 1)
	return NewPandocSuperscriptNode(string(line[1:closing]))
}
