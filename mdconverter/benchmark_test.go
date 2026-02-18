package mdconverter

import "testing"

func BenchmarkConvertMarkdown(b *testing.B) {
	conv, err := New(ReverseConfig{})
	if err != nil {
		b.Fatalf("failed to create converter: %v", err)
	}

	input := `# Heading

This is **bold** text with [link](https://example.com) and :smile:.

> [!WARNING]
> Warning text

- [ ] Task one
- [x] Task two

| Name | Value |
| --- | --- |
| A | 1 |
| B | 2 |
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := conv.Convert(input); err != nil {
			b.Fatalf("convert failed: %v", err)
		}
	}
}
