package mdconverter

import (
	"encoding/json"
	"testing"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertBlockquoteDisambiguation(t *testing.T) {
	t.Run("github panel", func(t *testing.T) {
		conv, err := New(ReverseConfig{})
		require.NoError(t, err)

		result, err := conv.Convert("> [!WARNING]\n> watch out")
		require.NoError(t, err)

		var doc converter.Doc
		require.NoError(t, json.Unmarshal(result.ADF, &doc))
		assert.Equal(t, converter.Doc{
			Version: 1,
			Type:    "doc",
			Content: []converter.Node{
				{
					Type: "panel",
					Attrs: map[string]interface{}{
						"panelType": "warning",
					},
					Content: []converter.Node{
						{
							Type: "paragraph",
							Content: []converter.Node{
								{Type: "text", Text: "watch out"},
							},
						},
					},
				},
			},
		}, doc)
	})

	t.Run("bold panel", func(t *testing.T) {
		conv, err := New(ReverseConfig{
			PanelDetection: PanelDetectBold,
		})
		require.NoError(t, err)

		result, err := conv.Convert("> **Info**: First paragraph\n>\n> Second paragraph")
		require.NoError(t, err)

		var doc converter.Doc
		require.NoError(t, json.Unmarshal(result.ADF, &doc))
		assert.Equal(t, converter.Doc{
			Version: 1,
			Type:    "doc",
			Content: []converter.Node{
				{
					Type: "panel",
					Attrs: map[string]interface{}{
						"panelType": "info",
					},
					Content: []converter.Node{
						{
							Type: "paragraph",
							Content: []converter.Node{
								{Type: "text", Text: "First paragraph"},
							},
						},
						{
							Type: "paragraph",
							Content: []converter.Node{
								{Type: "text", Text: "Second paragraph"},
							},
						},
					},
				},
			},
		}, doc)
	})

	t.Run("decision list", func(t *testing.T) {
		conv, err := New(ReverseConfig{})
		require.NoError(t, err)

		result, err := conv.Convert("> **✓ Decision**: First decision\n>\n> **? Decision**: Second decision")
		require.NoError(t, err)

		var doc converter.Doc
		require.NoError(t, json.Unmarshal(result.ADF, &doc))
		assert.Equal(t, converter.Doc{
			Version: 1,
			Type:    "doc",
			Content: []converter.Node{
				{
					Type: "decisionList",
					Content: []converter.Node{
						{
							Type: "decisionItem",
							Attrs: map[string]interface{}{
								"state": "DECIDED",
							},
							Content: []converter.Node{
								{
									Type: "paragraph",
									Content: []converter.Node{
										{Type: "text", Text: "First decision"},
									},
								},
							},
						},
						{
							Type: "decisionItem",
							Attrs: map[string]interface{}{
								"state": "UNDECIDED",
							},
							Content: []converter.Node{
								{
									Type: "paragraph",
									Content: []converter.Node{
										{Type: "text", Text: "Second decision"},
									},
								},
							},
						},
					},
				},
			},
		}, doc)
	})

	t.Run("expand blockquote", func(t *testing.T) {
		conv, err := New(ReverseConfig{
			ExpandDetection: ExpandDetectBlockquote,
		})
		require.NoError(t, err)

		result, err := conv.Convert("> **Title**\n>\n> Body")
		require.NoError(t, err)

		var doc converter.Doc
		require.NoError(t, json.Unmarshal(result.ADF, &doc))
		assert.Equal(t, converter.Doc{
			Version: 1,
			Type:    "doc",
			Content: []converter.Node{
				{
					Type: "expand",
					Attrs: map[string]interface{}{
						"title": "Title",
					},
					Content: []converter.Node{
						{
							Type: "paragraph",
							Content: []converter.Node{
								{Type: "text", Text: "Body"},
							},
						},
					},
				},
			},
		}, doc)
	})
}

func TestConvertBlockHTMLAndExtensions(t *testing.T) {
	t.Run("details html", func(t *testing.T) {
		conv, err := New(ReverseConfig{})
		require.NoError(t, err)

		result, err := conv.Convert("<details><summary>Click to see more</summary>\n\nHidden content here.\n\n</details>")
		require.NoError(t, err)

		var doc converter.Doc
		require.NoError(t, json.Unmarshal(result.ADF, &doc))
		assert.Equal(t, converter.Doc{
			Version: 1,
			Type:    "doc",
			Content: []converter.Node{
				{
					Type: "expand",
					Attrs: map[string]interface{}{
						"title": "Click to see more",
					},
					Content: []converter.Node{
						{
							Type: "paragraph",
							Content: []converter.Node{
								{Type: "text", Text: "Hidden content here."},
							},
						},
					},
				},
			},
		}, doc)
	})

	t.Run("aligned html blocks", func(t *testing.T) {
		conv, err := New(ReverseConfig{})
		require.NoError(t, err)

		result, err := conv.Convert("<div align=\"center\">Centered text</div>\n\n<h3 align=\"right\">**Bold Heading**</h3>")
		require.NoError(t, err)

		var doc converter.Doc
		require.NoError(t, json.Unmarshal(result.ADF, &doc))
		assert.Equal(t, converter.Doc{
			Version: 1,
			Type:    "doc",
			Content: []converter.Node{
				{
					Type: "paragraph",
					Attrs: map[string]interface{}{
						"layout": "center",
					},
					Content: []converter.Node{
						{Type: "text", Text: "Centered text"},
					},
				},
				{
					Type: "heading",
					Attrs: map[string]interface{}{
						"level": float64(3),
						"align": "right",
					},
					Content: []converter.Node{
						{
							Type: "text",
							Text: "Bold Heading",
							Marks: []converter.Mark{
								{Type: "strong"},
							},
						},
					},
				},
			},
		}, doc)
	})

	t.Run("html table and extension fences", func(t *testing.T) {
		conv, err := New(ReverseConfig{})
		require.NoError(t, err)

		markdown := "<table>\n  <tbody>\n    <tr>\n      <td colspan=\"2\">\n        complex cell\n      </td>\n    </tr>\n  </tbody>\n</table>\n\n```adf:extension\n{\n  \"type\": \"inlineExtension\",\n  \"attrs\": {\n    \"extensionKey\": \"demo\"\n  }\n}\n```\n\n```adf:inlineCard\n{\n  \"data\": {\n    \"name\": \"Example\",\n    \"url\": \"https://example.com\"\n  }\n}\n```"
		result, err := conv.Convert(markdown)
		require.NoError(t, err)

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
									Type: "tableCell",
									Attrs: map[string]interface{}{
										"colspan": float64(2),
									},
									Content: []converter.Node{
										{
											Type: "paragraph",
											Content: []converter.Node{
												{Type: "text", Text: "complex cell"},
											},
										},
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
							Type: "inlineExtension",
							Attrs: map[string]interface{}{
								"extensionKey": "demo",
							},
						},
						{
							Type: "inlineCard",
							Attrs: map[string]interface{}{
								"data": map[string]interface{}{
									"name": "Example",
									"url":  "https://example.com",
								},
							},
						},
					},
				},
			},
		}, doc)
	})

	t.Run("inline extension and inline card merge with text", func(t *testing.T) {
		conv, err := New(ReverseConfig{})
		require.NoError(t, err)

		markdown := "Before\n```adf:extension\n{\n  \"type\": \"inlineExtension\",\n  \"attrs\": {\n    \"extensionKey\": \"demo\"\n  }\n}\n```\n\nafter\n\nBefore2\n```adf:inlineCard\n{\n  \"data\": {\n    \"name\": \"Example\",\n    \"url\": \"https://example.com\"\n  }\n}\n```\n\nafter2"
		result, err := conv.Convert(markdown)
		require.NoError(t, err)

		var doc converter.Doc
		require.NoError(t, json.Unmarshal(result.ADF, &doc))
		assert.Equal(t, converter.Doc{
			Version: 1,
			Type:    "doc",
			Content: []converter.Node{
				{
					Type: "paragraph",
					Content: []converter.Node{
						{Type: "text", Text: "Before"},
						{
							Type: "inlineExtension",
							Attrs: map[string]interface{}{
								"extensionKey": "demo",
							},
						},
						{Type: "text", Text: "after"},
					},
				},
				{
					Type: "paragraph",
					Content: []converter.Node{
						{Type: "text", Text: "Before2"},
						{
							Type: "inlineCard",
							Attrs: map[string]interface{}{
								"data": map[string]interface{}{
									"name": "Example",
									"url":  "https://example.com",
								},
							},
						},
						{Type: "text", Text: "after2"},
					},
				},
			},
		}, doc)
	})
}
