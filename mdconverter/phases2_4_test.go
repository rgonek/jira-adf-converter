package mdconverter

import (
	"encoding/json"
	"testing"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertBlockNodes(t *testing.T) {
	conv, err := New(ReverseConfig{})
	require.NoError(t, err)

	result, err := conv.Convert("# Heading\n\n> Quote\n\nBefore\n\n---\n\nAfter\n\nLine 1\\\nLine 2\n\n```go\nfmt.Println(\"Hello\")\n```")
	require.NoError(t, err)
	assert.Empty(t, result.Warnings)

	var doc converter.Doc
	require.NoError(t, json.Unmarshal(result.ADF, &doc))

	assert.Equal(t, converter.Doc{
		Version: 1,
		Type:    "doc",
		Content: []converter.Node{
			{
				Type: "heading",
				Attrs: map[string]interface{}{
					"level": float64(1),
				},
				Content: []converter.Node{
					{Type: "text", Text: "Heading"},
				},
			},
			{
				Type: "blockquote",
				Content: []converter.Node{
					{
						Type: "paragraph",
						Content: []converter.Node{
							{Type: "text", Text: "Quote"},
						},
					},
				},
			},
			{
				Type: "paragraph",
				Content: []converter.Node{
					{Type: "text", Text: "Before"},
				},
			},
			{
				Type: "rule",
			},
			{
				Type: "paragraph",
				Content: []converter.Node{
					{Type: "text", Text: "After"},
				},
			},
			{
				Type: "paragraph",
				Content: []converter.Node{
					{Type: "text", Text: "Line 1"},
					{Type: "hardBreak"},
					{Type: "text", Text: "Line 2"},
				},
			},
			{
				Type: "codeBlock",
				Attrs: map[string]interface{}{
					"language": "go",
				},
				Content: []converter.Node{
					{Type: "text", Text: "fmt.Println(\"Hello\")"},
				},
			},
		},
	}, doc)
}

func TestConvertBlockConfigOptions(t *testing.T) {
	conv, err := New(ReverseConfig{
		HeadingOffset: -1,
		LanguageMap: map[string]string{
			"cpp": "c++",
		},
	})
	require.NoError(t, err)

	result, err := conv.Convert("## Heading\n\n```cpp\nstd::cout << \"x\";\n```")
	require.NoError(t, err)
	assert.Empty(t, result.Warnings)

	var doc converter.Doc
	require.NoError(t, json.Unmarshal(result.ADF, &doc))

	assert.Equal(t, converter.Doc{
		Version: 1,
		Type:    "doc",
		Content: []converter.Node{
			{
				Type: "heading",
				Attrs: map[string]interface{}{
					"level": float64(1),
				},
				Content: []converter.Node{
					{Type: "text", Text: "Heading"},
				},
			},
			{
				Type: "codeBlock",
				Attrs: map[string]interface{}{
					"language": "c++",
				},
				Content: []converter.Node{
					{Type: "text", Text: "std::cout << \"x\";"},
				},
			},
		},
	}, doc)
}

func TestConvertMarks(t *testing.T) {
	conv, err := New(ReverseConfig{})
	require.NoError(t, err)

	result, err := conv.Convert("**bold _bold+italic_ end** ~~strike~~ [link](https://example.com \"Example\") `code`")
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
						Text: "bold ",
						Marks: []converter.Mark{
							{Type: "strong"},
						},
					},
					{
						Type: "text",
						Text: "bold+italic",
						Marks: []converter.Mark{
							{Type: "strong"},
							{Type: "em"},
						},
					},
					{
						Type: "text",
						Text: " end",
						Marks: []converter.Mark{
							{Type: "strong"},
						},
					},
					{Type: "text", Text: " "},
					{
						Type: "text",
						Text: "strike",
						Marks: []converter.Mark{
							{Type: "strike"},
						},
					},
					{Type: "text", Text: " "},
					{
						Type: "text",
						Text: "link",
						Marks: []converter.Mark{
							{
								Type: "link",
								Attrs: map[string]interface{}{
									"href":  "https://example.com",
									"title": "Example",
								},
							},
						},
					},
					{Type: "text", Text: " "},
					{
						Type: "text",
						Text: "code",
						Marks: []converter.Mark{
							{Type: "code"},
						},
					},
				},
			},
		},
	}, doc)
}

func TestConvertLists(t *testing.T) {
	conv, err := New(ReverseConfig{})
	require.NoError(t, err)

	result, err := conv.Convert("- First item\n- Second item\n  - Nested item\n\n1. Step one\n2. Step two\n\n- [ ] Todo\n- [x] Done")
	require.NoError(t, err)
	assert.Empty(t, result.Warnings)

	var doc converter.Doc
	require.NoError(t, json.Unmarshal(result.ADF, &doc))

	assert.Equal(t, converter.Doc{
		Version: 1,
		Type:    "doc",
		Content: []converter.Node{
			{
				Type: "bulletList",
				Content: []converter.Node{
					{
						Type: "listItem",
						Content: []converter.Node{
							{
								Type: "paragraph",
								Content: []converter.Node{
									{Type: "text", Text: "First item"},
								},
							},
						},
					},
					{
						Type: "listItem",
						Content: []converter.Node{
							{
								Type: "paragraph",
								Content: []converter.Node{
									{Type: "text", Text: "Second item"},
								},
							},
							{
								Type: "bulletList",
								Content: []converter.Node{
									{
										Type: "listItem",
										Content: []converter.Node{
											{
												Type: "paragraph",
												Content: []converter.Node{
													{Type: "text", Text: "Nested item"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Type: "orderedList",
				Attrs: map[string]interface{}{
					"order": float64(1),
				},
				Content: []converter.Node{
					{
						Type: "listItem",
						Content: []converter.Node{
							{
								Type: "paragraph",
								Content: []converter.Node{
									{Type: "text", Text: "Step one"},
								},
							},
						},
					},
					{
						Type: "listItem",
						Content: []converter.Node{
							{
								Type: "paragraph",
								Content: []converter.Node{
									{Type: "text", Text: "Step two"},
								},
							},
						},
					},
				},
			},
			{
				Type: "taskList",
				Content: []converter.Node{
					{
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state": "TODO",
						},
						Content: []converter.Node{
							{Type: "text", Text: "Todo"},
						},
					},
					{
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state": "DONE",
						},
						Content: []converter.Node{
							{Type: "text", Text: "Done"},
						},
					},
				},
			},
		},
	}, doc)
}
