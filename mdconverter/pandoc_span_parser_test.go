package mdconverter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePandocAttributes(t *testing.T) {
	classes, attrs := parsePandocAttributes(`.underline .mention mention-id="abc" url='https://example.com'`)
	assert.Equal(t, []string{"underline", "mention"}, classes)
	assert.Equal(t, "abc", attrs["mention-id"])
	assert.Equal(t, "https://example.com", attrs["url"])
}

func TestPandocSpanUnderlineParsing(t *testing.T) {
	conv, err := New(ReverseConfig{
		UnderlineDetection: UnderlineDetectPandoc,
	})
	require.NoError(t, err)

	result, err := conv.Convert("[word]{.underline}")
	require.NoError(t, err)

	doc := decodeReverseDoc(t, result.ADF)
	require.Len(t, doc.Content, 1)
	require.Len(t, doc.Content[0].Content, 1)
	assert.Equal(t, "word", doc.Content[0].Content[0].Text)
	require.Len(t, doc.Content[0].Content[0].Marks, 1)
	assert.Equal(t, "underline", doc.Content[0].Content[0].Marks[0].Type)
}

func TestPandocSpanDoesNotOverrideLinkParsing(t *testing.T) {
	conv, err := New(ReverseConfig{
		UnderlineDetection: UnderlineDetectPandoc,
	})
	require.NoError(t, err)

	result, err := conv.Convert("[link](https://example.com) and [span]{.underline}")
	require.NoError(t, err)

	doc := decodeReverseDoc(t, result.ADF)
	require.Len(t, doc.Content, 1)
	require.Len(t, doc.Content[0].Content, 3)

	linkText := doc.Content[0].Content[0]
	assert.Equal(t, "link", linkText.Text)
	require.Len(t, linkText.Marks, 1)
	assert.Equal(t, "link", linkText.Marks[0].Type)

	spanText := doc.Content[0].Content[2]
	assert.Equal(t, "span", spanText.Text)
	require.Len(t, spanText.Marks, 1)
	assert.Equal(t, "underline", spanText.Marks[0].Type)
}

func TestPandocSpanDisabledFallsBackToLiteral(t *testing.T) {
	conv, err := New(ReverseConfig{
		MentionDetection:   MentionDetectPandoc,
		UnderlineDetection: UnderlineDetectNone,
	})
	require.NoError(t, err)

	result, err := conv.Convert("[word]{.underline}")
	require.NoError(t, err)

	doc := decodeReverseDoc(t, result.ADF)
	require.Len(t, doc.Content, 1)
	require.Len(t, doc.Content[0].Content, 1)
	assert.Equal(t, "[word]{.underline}", doc.Content[0].Content[0].Text)
}

func TestPandocSpanUnknownClassAddsWarning(t *testing.T) {
	conv, err := New(ReverseConfig{
		UnderlineDetection: UnderlineDetectPandoc,
	})
	require.NoError(t, err)

	result, err := conv.Convert("[word]{.custom}")
	require.NoError(t, err)
	require.NotEmpty(t, result.Warnings)
	assert.Equal(t, "[word]{.custom}", decodeReverseDoc(t, result.ADF).Content[0].Content[0].Text)
}
