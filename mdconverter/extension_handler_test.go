package mdconverter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/rgonek/jira-adf-converter/converter"
)

type mockExtensionHandler struct {
	fromMarkdown func(ctx context.Context, in converter.ExtensionParseInput) (converter.ExtensionParseOutput, error)
}

func (m *mockExtensionHandler) ToMarkdown(ctx context.Context, in converter.ExtensionRenderInput) (converter.ExtensionRenderOutput, error) {
	return converter.ExtensionRenderOutput{}, nil
}

func (m *mockExtensionHandler) FromMarkdown(ctx context.Context, in converter.ExtensionParseInput) (converter.ExtensionParseOutput, error) {
	if m.fromMarkdown != nil {
		return m.fromMarkdown(ctx, in)
	}
	return converter.ExtensionParseOutput{}, nil
}

func TestExtensionHandler_FromMarkdown(t *testing.T) {
	t.Run("handler called for matching extensionKey", func(t *testing.T) {
		handler := &mockExtensionHandler{
			fromMarkdown: func(ctx context.Context, in converter.ExtensionParseInput) (converter.ExtensionParseOutput, error) {
				if in.ExtensionKey != "my-key" {
					return converter.ExtensionParseOutput{}, nil
				}
				if in.Metadata["foo"] != "bar" {
					return converter.ExtensionParseOutput{}, nil
				}
				return converter.ExtensionParseOutput{
					Node: converter.Node{
						Type: "extension",
						Attrs: map[string]interface{}{
							"extensionKey":  "my-key",
							"extensionType": "com.example",
							"body":          in.Body,
						},
					},
					Handled: true,
				}, nil
			},
		}

		cfg := ReverseConfig{
			ExtensionHandlers: map[string]converter.ExtensionHandler{
				"my-key": handler,
			},
		}
		conv, _ := New(cfg)

		markdown := "::: { .adf-extension key=\"my-key\" foo=\"bar\" }\nMY_BODY\n:::\n"

		res, err := conv.Convert(markdown)
		if err != nil {
			t.Fatalf("Convert failed: %v", err)
		}

		var doc converter.Doc
		if err := json.Unmarshal(res.ADF, &doc); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		// Verify result
		found := false
		for _, node := range doc.Content {
			if node.Type == "extension" && node.Attrs["extensionKey"] == "my-key" {
				if node.Attrs["body"] != "MY_BODY" {
					t.Errorf("expected body %q, got %v", "MY_BODY", node.Attrs["body"])
				}
				found = true
			}
		}
		if !found {
			t.Errorf("expected extension node not found in ADF")
		}
	})

	t.Run("handler returns Handled: false", func(t *testing.T) {
		handler := &mockExtensionHandler{
			fromMarkdown: func(ctx context.Context, in converter.ExtensionParseInput) (converter.ExtensionParseOutput, error) {
				return converter.ExtensionParseOutput{Handled: false}, nil
			},
		}

		cfg := ReverseConfig{
			ExtensionHandlers: map[string]converter.ExtensionHandler{
				"my-key": handler,
			},
		}
		conv, _ := New(cfg)

		markdown := "::: { .adf-extension key=\"my-key\" }\nMY_BODY\n:::\n"

		res, err := conv.Convert(markdown)
		if err != nil {
			t.Fatalf("Convert failed: %v", err)
		}

		var doc converter.Doc
		if err := json.Unmarshal(res.ADF, &doc); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		t.Logf("ADF: %s", string(res.ADF))

		// Should fall back to blockquote warning
		found := false
		for _, node := range doc.Content {
			if node.Type == "blockquote" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected blockquote fallback not found")
		}
	})

	t.Run("handler returns error", func(t *testing.T) {
		handler := &mockExtensionHandler{
			fromMarkdown: func(ctx context.Context, in converter.ExtensionParseInput) (converter.ExtensionParseOutput, error) {
				return converter.ExtensionParseOutput{}, fmt.Errorf("boom")
			},
		}

		cfg := ReverseConfig{
			ExtensionHandlers: map[string]converter.ExtensionHandler{
				"my-key": handler,
			},
		}
		conv, _ := New(cfg)

		markdown := "::: { .adf-extension key=\"my-key\" }\nMY_BODY\n:::\n"

		_, err := conv.Convert(markdown)
		if err == nil || !strings.Contains(err.Error(), "boom") {
			t.Errorf("expected error 'boom', got %v", err)
		}
	})

	t.Run("div with no matching handler", func(t *testing.T) {
		cfg := ReverseConfig{
			ExtensionHandlers: map[string]converter.ExtensionHandler{
				"some-other-key": &mockExtensionHandler{},
			},
		}
		conv, _ := New(cfg)

		markdown := "::: { .adf-extension key=\"unknown\" }\nMY_BODY\n:::\n"

		res, err := conv.Convert(markdown)
		if err != nil {
			t.Fatalf("Convert failed: %v", err)
		}

		var doc converter.Doc
		if err := json.Unmarshal(res.ADF, &doc); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		t.Logf("ADF: %s", string(res.ADF))

		// Should fall back to blockquote warning
		found := false
		for _, node := range doc.Content {
			if node.Type == "blockquote" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected blockquote fallback not found")
		}
	})
}
