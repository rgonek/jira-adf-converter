package mdconverter

import (
	"encoding/json"
	"testing"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertEmptyDocument(t *testing.T) {
	conv, err := New(ReverseConfig{})
	require.NoError(t, err)

	result, err := conv.Convert("")
	require.NoError(t, err)
	assert.Empty(t, result.Warnings)

	var doc converter.Doc
	require.NoError(t, json.Unmarshal(result.ADF, &doc))

	assert.Equal(t, converter.Doc{
		Version: 1,
		Type:    "doc",
	}, doc)
}

func TestConvertSingleTextParagraph(t *testing.T) {
	conv, err := New(ReverseConfig{})
	require.NoError(t, err)

	result, err := conv.Convert("hello world")
	require.NoError(t, err)
	assert.Empty(t, result.Warnings)

	var doc converter.Doc
	require.NoError(t, json.Unmarshal(result.ADF, &doc))

	assert.Equal(t, converter.Doc{
		Version: 1,
		Type:    "doc",
		Content: []converter.Node{
			{
				Type: "paragraph",
				Content: []converter.Node{
					{
						Type: "text",
						Text: "hello world",
					},
				},
			},
		},
	}, doc)
}
