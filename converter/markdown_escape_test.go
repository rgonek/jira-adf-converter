package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEscapeMarkdownTextLiteralEscapesControlCharacters(t *testing.T) {
	input := "Literal * _ [ ] ( ) ` \\"
	expected := "Literal \\* \\_ \\[ \\] \\( \\) \\` \\\\"
	assert.Equal(t, expected, escapeMarkdownTextLiteral(input))
}

func TestEscapeMarkdownTextLiteralEscapesLineStartMarkers(t *testing.T) {
	input := "# heading\n  > quote\ntext # inline > inline"
	expected := "\\# heading\n  \\> quote\ntext # inline > inline"
	assert.Equal(t, expected, escapeMarkdownTextLiteral(input))
}

func TestEscapeMarkdownLinkDestination(t *testing.T) {
	assert.Equal(t, "https://example.com/a\\(b\\)", escapeMarkdownLinkDestination("https://example.com/a(b)"))
	assert.Equal(t, "<https://example.com/a b>", escapeMarkdownLinkDestination("https://example.com/a b"))
	assert.Equal(t, "mention:abc", escapeMarkdownLinkDestination(" mention:abc "))
}

func TestEscapeMarkdownLinkTitle(t *testing.T) {
	input := "He said \"hello\"\\next\nline"
	expected := "He said \\\"hello\\\"\\\\next line"
	assert.Equal(t, expected, escapeMarkdownLinkTitle(input))
}

func TestConvertInlineContentSkipsEscapingForCodeMark(t *testing.T) {
	conv, err := New(Config{})
	require.NoError(t, err)

	result, err := conv.Convert([]byte(`{"version":1,"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"*literal*","marks":[{"type":"code"}]}]}]}`))
	require.NoError(t, err)
	assert.Equal(t, "`*literal*`\n", result.Markdown)
}
