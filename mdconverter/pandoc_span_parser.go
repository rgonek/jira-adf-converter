package mdconverter

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type PandocSpanParser struct{}

func NewPandocSpanParser() parser.InlineParser {
	return &PandocSpanParser{}
}

func (p *PandocSpanParser) Trigger() []byte {
	return []byte{'['}
}

func (p *PandocSpanParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) < 4 || line[0] != '[' {
		return nil
	}

	closing := findBalancedClosingBracket(line)
	if closing <= 0 || closing+1 >= len(line) {
		return nil
	}

	next := line[closing+1]
	if next == '(' || next == '[' || next != '{' {
		return nil
	}

	rawAttrs, endPos, ok := readPandocAttrBlock(line, closing+1)
	if !ok {
		return nil
	}

	classes, attrs := parsePandocAttributes(rawAttrs)
	block.Advance(endPos)
	return NewPandocSpanNode(string(line[1:closing]), rawAttrs, classes, attrs)
}

func findBalancedClosingBracket(line []byte) int {
	depth := 0
	for idx := 0; idx < len(line); idx++ {
		ch := line[idx]
		if ch == '\n' || ch == '\r' {
			break
		}
		if ch == '\\' {
			if idx+1 < len(line) {
				idx++
			}
			continue
		}
		switch ch {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return idx
			}
			if depth < 0 {
				return -1
			}
		}
	}
	return -1
}
