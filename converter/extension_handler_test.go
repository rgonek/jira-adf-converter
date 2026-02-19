package converter

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

type mockExtensionHandler struct {
	toMarkdown func(ctx context.Context, in ExtensionRenderInput) (ExtensionRenderOutput, error)
}

func (m *mockExtensionHandler) ToMarkdown(ctx context.Context, in ExtensionRenderInput) (ExtensionRenderOutput, error) {
	if m.toMarkdown != nil {
		return m.toMarkdown(ctx, in)
	}
	return ExtensionRenderOutput{}, nil
}

func (m *mockExtensionHandler) FromMarkdown(ctx context.Context, in ExtensionParseInput) (ExtensionParseOutput, error) {
	return ExtensionParseOutput{}, nil
}

func TestExtensionHandler_ToMarkdown(t *testing.T) {
	t.Run("handler called and output used", func(t *testing.T) {
		handler := &mockExtensionHandler{
			toMarkdown: func(ctx context.Context, in ExtensionRenderInput) (ExtensionRenderOutput, error) {
				return ExtensionRenderOutput{
					Markdown: "CLEAN_MARKDOWN",
					Metadata: map[string]string{
						"foo": "bar",
					},
					Handled: true,
				}, nil
			},
		}

		cfg := Config{
			ExtensionHandlers: map[string]ExtensionHandler{
				"my-key": handler,
			},
		}
		conv, _ := New(cfg)

		adf := `{
			"version": 1,
			"type": "doc",
			"content": [
				{
					"type": "extension",
					"attrs": {
						"extensionKey": "my-key"
					}
				}
			]
		}`

		res, err := conv.Convert([]byte(adf))
		if err != nil {
			t.Fatalf("Convert failed: %v", err)
		}

		expected := "::: { .adf-extension key=\"my-key\" foo=\"bar\" }\nCLEAN_MARKDOWN\n:::\n"
		if res.Markdown != expected {
			t.Errorf("expected %q, got %q", expected, res.Markdown)
		}
	})

	t.Run("handler returns Handled: false", func(t *testing.T) {
		handler := &mockExtensionHandler{
			toMarkdown: func(ctx context.Context, in ExtensionRenderInput) (ExtensionRenderOutput, error) {
				return ExtensionRenderOutput{Handled: false}, nil
			},
		}

		cfg := Config{
			ExtensionHandlers: map[string]ExtensionHandler{
				"my-key": handler,
			},
			Extensions: ExtensionRules{
				Default: ExtensionJSON,
			},
		}
		conv, _ := New(cfg)

		adf := `{
			"version": 1,
			"type": "doc",
			"content": [
				{
					"type": "extension",
					"attrs": {
						"extensionKey": "my-key"
					}
				}
			]
		}`

		res, err := conv.Convert([]byte(adf))
		if err != nil {
			t.Fatalf("Convert failed: %v", err)
		}

		if !strings.Contains(res.Markdown, "```adf:extension") {
			t.Errorf("expected fallback to JSON, got %q", res.Markdown)
		}
		if strings.Contains(res.Markdown, "::: { .adf-extension") {
			t.Errorf("did not expect wrapper div")
		}
	})

	t.Run("handler returns error", func(t *testing.T) {
		handler := &mockExtensionHandler{
			toMarkdown: func(ctx context.Context, in ExtensionRenderInput) (ExtensionRenderOutput, error) {
				return ExtensionRenderOutput{}, fmt.Errorf("boom")
			},
		}

		cfg := Config{
			ExtensionHandlers: map[string]ExtensionHandler{
				"my-key": handler,
			},
		}
		conv, _ := New(cfg)

		adf := `{
			"version": 1,
			"type": "doc",
			"content": [
				{
					"type": "extension",
					"attrs": {
						"extensionKey": "my-key"
					}
				}
			]
		}`

		_, err := conv.Convert([]byte(adf))
		if err == nil || !strings.Contains(err.Error(), "boom") {
			t.Errorf("expected error 'boom', got %v", err)
		}
	})

	t.Run("no handler registered", func(t *testing.T) {
		cfg := Config{
			Extensions: ExtensionRules{
				Default: ExtensionJSON,
			},
		}
		conv, _ := New(cfg)

		adf := `{
			"version": 1,
			"type": "doc",
			"content": [
				{
					"type": "extension",
					"attrs": {
						"extensionKey": "my-key"
					}
				}
			]
		}`

		res, err := conv.Convert([]byte(adf))
		if err != nil {
			t.Fatalf("Convert failed: %v", err)
		}

		if !strings.Contains(res.Markdown, "```adf:extension") {
			t.Errorf("expected fallback to JSON, got %q", res.Markdown)
		}
	})
}
