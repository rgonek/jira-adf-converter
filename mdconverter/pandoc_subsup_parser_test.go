package mdconverter

import (
	"encoding/json"
	"testing"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPandocSubSupParsing(t *testing.T) {
	conv, err := New(ReverseConfig{
		SubSupDetection: SubSupDetectPandoc,
	})
	require.NoError(t, err)

	result, err := conv.Convert("~H2O~")
	require.NoError(t, err)

	doc := decodeReverseDoc(t, result.ADF)
	require.Len(t, doc.Content, 1)
	require.Len(t, doc.Content[0].Content, 1)

	textNode := doc.Content[0].Content[0]
	assert.Equal(t, "text", textNode.Type)
	assert.Equal(t, "H2O", textNode.Text)
	require.Len(t, textNode.Marks, 1)
	assert.Equal(t, "subsup", textNode.Marks[0].Type)
	assert.Equal(t, "sub", textNode.Marks[0].Attrs["type"])

	result, err = conv.Convert("x^2^")
	require.NoError(t, err)
	doc = decodeReverseDoc(t, result.ADF)
	require.Len(t, doc.Content[0].Content, 2)
	assert.Equal(t, "x", doc.Content[0].Content[0].Text)
	assert.Equal(t, "2", doc.Content[0].Content[1].Text)
	require.Len(t, doc.Content[0].Content[1].Marks, 1)
	assert.Equal(t, "subsup", doc.Content[0].Content[1].Marks[0].Type)
	assert.Equal(t, "sup", doc.Content[0].Content[1].Marks[0].Attrs["type"])
}

func TestPandocSubSupDoesNotConflictWithStrikethrough(t *testing.T) {
	conv, err := New(ReverseConfig{
		SubSupDetection: SubSupDetectPandoc,
	})
	require.NoError(t, err)

	result, err := conv.Convert("~~strike~~")
	require.NoError(t, err)

	doc := decodeReverseDoc(t, result.ADF)
	require.Len(t, doc.Content, 1)
	require.Len(t, doc.Content[0].Content, 1)

	textNode := doc.Content[0].Content[0]
	assert.Equal(t, "strike", textNode.Text)
	require.Len(t, textNode.Marks, 1)
	assert.Equal(t, "strike", textNode.Marks[0].Type)
}

func TestPandocSubSupDisabledPreservesLiteral(t *testing.T) {
	conv, err := New(ReverseConfig{
		SubSupDetection:    SubSupDetectNone,
		UnderlineDetection: UnderlineDetectPandoc,
	})
	require.NoError(t, err)

	result, err := conv.Convert("~text~ and ~unclosed")
	require.NoError(t, err)

	doc := decodeReverseDoc(t, result.ADF)
	require.Len(t, doc.Content, 1)
	require.Len(t, doc.Content[0].Content, 1)
	assert.Equal(t, "~text~ and ~unclosed", doc.Content[0].Content[0].Text)
}

func decodeReverseDoc(t *testing.T, raw []byte) converter.Doc {
	t.Helper()

	var doc converter.Doc
	require.NoError(t, json.Unmarshal(raw, &doc))
	return doc
}
