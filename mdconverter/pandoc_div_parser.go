package mdconverter

import (
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type PandocDivParser struct{}

func NewPandocDivParser() parser.BlockParser {
	return &PandocDivParser{}
}

func (p *PandocDivParser) Trigger() []byte {
	return []byte{':'}
}

func (p *PandocDivParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	trimmed := strings.TrimLeft(trimLineEnding(string(line)), " \t")
	if !strings.HasPrefix(trimmed, ":::") {
		return nil, parser.NoChildren
	}

	fenceLength := countLeadingChar(trimmed, ':')
	if fenceLength < 3 {
		return nil, parser.NoChildren
	}

	rest := strings.TrimSpace(trimmed[fenceLength:])
	if !strings.HasPrefix(rest, "{") {
		return nil, parser.NoChildren
	}

	rawAttrs, endPos, ok := readPandocAttrBlock([]byte(rest), 0)
	if !ok || strings.TrimSpace(rest[endPos:]) != "" {
		return nil, parser.NoChildren
	}

	classes, attrs := parsePandocAttributes(rawAttrs)
	node := NewPandocDivNode(fenceLength, rawAttrs, classes, attrs)
	reader.AdvanceLine()

	for {
		nextLine, _ := reader.PeekLine()
		if len(nextLine) == 0 {
			break
		}

		rawLine := trimLineEnding(string(nextLine))
		leftTrimmed := strings.TrimLeft(rawLine, " \t")
		if isPandocDivOpeningFence(leftTrimmed) {
			node.openDepth++
			node.appendBodyLine(rawLine)
			reader.AdvanceLine()
			continue
		}
		if isPandocDivClosingFence(leftTrimmed, node.FenceLength) {
			if node.openDepth == 1 {
				reader.AdvanceLine()
				break
			}
			node.openDepth--
			node.appendBodyLine(rawLine)
			reader.AdvanceLine()
			continue
		}

		node.appendBodyLine(rawLine)
		reader.AdvanceLine()
	}

	return node, parser.NoChildren
}

func (p *PandocDivParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	return parser.Close
}

func (p *PandocDivParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}

func (p *PandocDivParser) CanInterruptParagraph() bool {
	return true
}

func (p *PandocDivParser) CanAcceptIndentedLine() bool {
	return false
}

func isPandocDivClosingFence(line string, openingFenceLength int) bool {
	if !strings.HasPrefix(line, ":::") {
		return false
	}
	fenceLength := countLeadingChar(line, ':')
	if fenceLength < 3 {
		return false
	}
	if strings.TrimSpace(line[fenceLength:]) != "" {
		return false
	}
	return fenceLength == openingFenceLength
}

func isPandocDivOpeningFence(line string) bool {
	if !strings.HasPrefix(line, ":::") {
		return false
	}
	fenceLength := countLeadingChar(line, ':')
	if fenceLength < 3 {
		return false
	}
	rest := strings.TrimSpace(line[fenceLength:])
	if !strings.HasPrefix(rest, "{") {
		return false
	}
	_, endPos, ok := readPandocAttrBlock([]byte(rest), 0)
	return ok && strings.TrimSpace(rest[endPos:]) == ""
}

func countLeadingChar(value string, target byte) int {
	count := 0
	for count < len(value) && value[count] == target {
		count++
	}
	return count
}
