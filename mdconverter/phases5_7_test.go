package mdconverter

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertTable(t *testing.T) {
	conv, err := New(ReverseConfig{})
	require.NoError(t, err)

	result, err := conv.Convert("| A | B |\n| --- | :---: |\n| 1 | 2 |\n")
	require.NoError(t, err)
	assert.Empty(t, result.Warnings)

	var doc converter.Doc
	require.NoError(t, json.Unmarshal(result.ADF, &doc))

	assert.Equal(t, converter.Doc{
		Version: 1,
		Type:    "doc",
		Content: []converter.Node{
			{
				Type: "table",
				Content: []converter.Node{
					{
						Type: "tableRow",
						Content: []converter.Node{
							{
								Type: "tableHeader",
								Content: []converter.Node{
									{
										Type: "paragraph",
										Content: []converter.Node{
											{Type: "text", Text: "A"},
										},
									},
								},
							},
							{
								Type: "tableHeader",
								Attrs: map[string]interface{}{
									"alignment": "center",
								},
								Content: []converter.Node{
									{
										Type: "paragraph",
										Content: []converter.Node{
											{Type: "text", Text: "B"},
										},
									},
								},
							},
						},
					},
					{
						Type: "tableRow",
						Content: []converter.Node{
							{
								Type: "tableCell",
								Content: []converter.Node{
									{
										Type: "paragraph",
										Content: []converter.Node{
											{Type: "text", Text: "1"},
										},
									},
								},
							},
							{
								Type: "tableCell",
								Attrs: map[string]interface{}{
									"alignment": "center",
								},
								Content: []converter.Node{
									{
										Type: "paragraph",
										Content: []converter.Node{
											{Type: "text", Text: "2"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}, doc)
}

func TestConvertInlinePatterns(t *testing.T) {
	conv, err := New(ReverseConfig{
		MentionDetection: MentionDetectAt,
		MentionRegistry: map[string]string{
			"User Name": "12345",
		},
	})
	require.NoError(t, err)

	result, err := conv.Convert("Hey @User Name :smile: [Status: In Progress] Due 2024-02-19")
	require.NoError(t, err)
	assert.Empty(t, result.Warnings)

	var doc converter.Doc
	require.NoError(t, json.Unmarshal(result.ADF, &doc))

	expectedDate, err := time.Parse("2006-01-02", "2024-02-19")
	require.NoError(t, err)

	assert.Equal(t, converter.Doc{
		Version: 1,
		Type:    "doc",
		Content: []converter.Node{
			{
				Type: "paragraph",
				Content: []converter.Node{
					{Type: "text", Text: "Hey "},
					{
						Type: "mention",
						Attrs: map[string]interface{}{
							"id":   "12345",
							"text": "User Name",
						},
					},
					{Type: "text", Text: " "},
					{
						Type: "emoji",
						Attrs: map[string]interface{}{
							"shortName": ":smile:",
						},
					},
					{Type: "text", Text: " "},
					{
						Type: "status",
						Attrs: map[string]interface{}{
							"text": "In Progress",
						},
					},
					{Type: "text", Text: " Due "},
					{
						Type: "date",
						Attrs: map[string]interface{}{
							"timestamp": strconv.FormatInt(expectedDate.Unix(), 10),
						},
					},
				},
			},
		},
	}, doc)
}

func TestConvertMentionLinkImageAndInlineCard(t *testing.T) {
	conv, err := New(ReverseConfig{})
	require.NoError(t, err)

	result, err := conv.Convert("[@User Name](mention:12345)\n\n![Alt Text](http://example.com/image.png)\n\n[https://example.com](https://example.com)")
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
						Type: "mention",
						Attrs: map[string]interface{}{
							"id":   "12345",
							"text": "User Name",
						},
					},
				},
			},
			{
				Type: "mediaSingle",
				Content: []converter.Node{
					{
						Type: "media",
						Attrs: map[string]interface{}{
							"type": "image",
							"url":  "http://example.com/image.png",
							"alt":  "Alt Text",
						},
					},
				},
			},
			{
				Type: "paragraph",
				Content: []converter.Node{
					{
						Type: "inlineCard",
						Attrs: map[string]interface{}{
							"url": "https://example.com",
						},
					},
				},
			},
		},
	}, doc)
}

func TestConvertInlineHTML(t *testing.T) {
	conv, err := New(ReverseConfig{
		MentionDetection: MentionDetectHTML,
	})
	require.NoError(t, err)

	result, err := conv.Convert("This is <sub>sub</sub> and <sup>sup</sup> and <u>under</u>\n\n<span style=\"color: #ff0000\">red</span>\n\n<span data-mention-id=\"42\">@Jane</span>")
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
					{Type: "text", Text: "This is "},
					{
						Type: "text",
						Text: "sub",
						Marks: []converter.Mark{
							{
								Type: "subsup",
								Attrs: map[string]interface{}{
									"type": "sub",
								},
							},
						},
					},
					{Type: "text", Text: " and "},
					{
						Type: "text",
						Text: "sup",
						Marks: []converter.Mark{
							{
								Type: "subsup",
								Attrs: map[string]interface{}{
									"type": "sup",
								},
							},
						},
					},
					{Type: "text", Text: " and "},
					{
						Type: "text",
						Text: "under",
						Marks: []converter.Mark{
							{Type: "underline"},
						},
					},
				},
			},
			{
				Type: "paragraph",
				Content: []converter.Node{
					{
						Type: "text",
						Text: "red",
						Marks: []converter.Mark{
							{
								Type: "textColor",
								Attrs: map[string]interface{}{
									"color": "#ff0000",
								},
							},
						},
					},
				},
			},
			{
				Type: "paragraph",
				Content: []converter.Node{
					{
						Type: "mention",
						Attrs: map[string]interface{}{
							"id":   "42",
							"text": "Jane",
						},
					},
				},
			},
		},
	}, doc)
}
