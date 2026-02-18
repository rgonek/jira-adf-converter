package mdconverter

import (
	"encoding/json"
	"testing"
)

func FuzzConvertMarkdown(f *testing.F) {
	seeds := []string{
		"",
		"Hello World",
		"**bold** _italic_ ~~strike~~",
		"> [!WARNING]\n> watch out",
		"```adf:extension\n{\"type\":\"inlineExtension\",\"attrs\":{\"extensionKey\":\"demo\"}}\n```",
		"<details><summary>Title</summary>\n\nBody\n\n</details>",
		"| A | B |\n| --- | --- |\n| 1 | 2 |",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	conv, err := New(ReverseConfig{})
	if err != nil {
		f.Fatalf("failed to create converter: %v", err)
	}

	f.Fuzz(func(t *testing.T, markdown string) {
		result, err := conv.Convert(markdown)
		if err != nil {
			t.Fatalf("convert returned error: %v", err)
		}

		var doc map[string]interface{}
		if err := json.Unmarshal(result.ADF, &doc); err != nil {
			t.Fatalf("invalid adf json: %v", err)
		}
	})
}
