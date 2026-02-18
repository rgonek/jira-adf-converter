package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTableCodeBlock(t *testing.T) {
	// JSON input representing a table with a code block in a cell
	input := []byte(`{
		"type": "doc",
		"content": [
			{
				"type": "table",
				"content": [
					{
						"type": "tableRow",
						"content": [
							{
								"type": "tableCell",
								"content": [
									{
										"type": "codeBlock",
										"content": [
											{
												"type": "text",
												"text": "fmt.Println(\"Hello\")\nreturn"
											}
										]
									}
								]
							}
						]
					}
				]
			}
		]
	}`)

	t.Run("AllowHTML=true", func(t *testing.T) {
		cfg := Config{HardBreakStyle: HardBreakHTML, TableMode: TablePipe}
		conv, err := New(cfg)
		require.NoError(t, err)
		result, err := conv.Convert(input)
		require.NoError(t, err)
		// Expected: | <code>fmt.Println(&quot;Hello&quot;)<br>return</code> |
		// Note: The surrounding table structure adds pipes and spaces
		assert.Contains(t, result.Markdown, "<code>fmt.Println(&quot;Hello&quot;)<br>return</code>")
	})

	t.Run("AllowHTML=false", func(t *testing.T) {
		cfg := Config{HardBreakStyle: HardBreakBackslash, TableMode: TablePipe}
		conv, err := New(cfg)
		require.NoError(t, err)
		result, err := conv.Convert(input)
		require.NoError(t, err)
		// Expected: | `fmt.Println("Hello") return` |
		assert.Contains(t, result.Markdown, "`fmt.Println(\"Hello\") return`")
	})

	t.Run("TableAuto renders HTML", func(t *testing.T) {
		cfg := Config{TableMode: TableAuto}
		conv, err := New(cfg)
		require.NoError(t, err)
		result, err := conv.Convert(input)
		require.NoError(t, err)
		assert.Contains(t, result.Markdown, "<table>")
		assert.Contains(t, result.Markdown, "```")
	})
}
