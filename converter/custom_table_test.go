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
		cfg := Config{AllowHTML: true}
		conv := New(cfg)
		output, err := conv.Convert(input)
		require.NoError(t, err)
		// Expected: | <code>fmt.Println(&quot;Hello&quot;)<br>return</code> |
		// Note: The surrounding table structure adds pipes and spaces
		assert.Contains(t, output, "<code>fmt.Println(&quot;Hello&quot;)<br>return</code>")
	})

	t.Run("AllowHTML=false", func(t *testing.T) {
		cfg := Config{AllowHTML: false}
		conv := New(cfg)
		output, err := conv.Convert(input)
		require.NoError(t, err)
		// Expected: | `fmt.Println("Hello") return` |
		assert.Contains(t, output, "`fmt.Println(\"Hello\") return`")
	})
}
