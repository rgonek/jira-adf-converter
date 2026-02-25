package mdconverter

import (
	"encoding/json"
	"testing"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnescapeMarkdownLiteralText(t *testing.T) {
	input := `\* \_ \[ \] \( \) \# \> \\`
	expected := "* _ [ ] ( ) # > \\"

	assert.Equal(t, expected, unescapeMarkdownLiteralText(input))
	assert.Equal(t, `\z`, unescapeMarkdownLiteralText(`\z`))
}

func TestConvertKeepsBackslashInCodeSpan(t *testing.T) {
	conv, err := New(ReverseConfig{})
	require.NoError(t, err)

	result, err := conv.Convert("`\\*`")
	require.NoError(t, err)

	var doc converter.Doc
	require.NoError(t, json.Unmarshal(result.ADF, &doc))
	require.Len(t, doc.Content, 1)
	require.Len(t, doc.Content[0].Content, 1)
	textNode := doc.Content[0].Content[0]
	assert.Equal(t, "\\*", textNode.Text)
	require.Len(t, textNode.Marks, 1)
	assert.Equal(t, "code", textNode.Marks[0].Type)
}
