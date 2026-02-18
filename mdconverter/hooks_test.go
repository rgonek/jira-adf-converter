package mdconverter

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type hookContextKey string

const traceContextKey hookContextKey = "trace"

func newHookReverseConverter(t testing.TB, cfg ReverseConfig) *Converter {
	t.Helper()
	conv, err := New(cfg)
	require.NoError(t, err)
	return conv
}

func decodeADFDoc(t testing.TB, payload []byte) converter.Doc {
	t.Helper()
	var doc converter.Doc
	require.NoError(t, json.Unmarshal(payload, &doc))
	return doc
}

func TestLinkParseHookRewritesDestination(t *testing.T) {
	conv := newHookReverseConverter(t, ReverseConfig{
		LinkHook: func(ctx context.Context, in LinkParseInput) (LinkParseOutput, error) {
			assert.Equal(t, "hook-test", ctx.Value(traceContextKey))
			assert.Equal(t, "docs/page.md", in.SourcePath)
			assert.Equal(t, "../page.md", in.Destination)
			assert.Equal(t, "Page", in.Text)
			assert.Equal(t, "page.md", in.Meta.Filename)
			return LinkParseOutput{
				Destination: "https://confluence.example/wiki/pages/123",
				Handled:     true,
			}, nil
		},
	})

	ctx := context.WithValue(context.Background(), traceContextKey, "hook-test")
	result, err := conv.ConvertWithContext(ctx, `[Page](../page.md)`, ConvertOptions{SourcePath: "docs/page.md"})
	require.NoError(t, err)

	doc := decodeADFDoc(t, result.ADF)
	require.Len(t, doc.Content, 1)
	require.Equal(t, "paragraph", doc.Content[0].Type)
	require.Len(t, doc.Content[0].Content, 1)
	textNode := doc.Content[0].Content[0]
	require.Equal(t, "text", textNode.Type)
	require.Equal(t, "Page", textNode.Text)
	require.Len(t, textNode.Marks, 1)
	assert.Equal(t, "link", textNode.Marks[0].Type)
	assert.Equal(t, "https://confluence.example/wiki/pages/123", textNode.Marks[0].Attrs["href"])
}

func TestLinkParseHookForceLinkBypassesCardHeuristics(t *testing.T) {
	conv := newHookReverseConverter(t, ReverseConfig{
		LinkHook: func(_ context.Context, in LinkParseInput) (LinkParseOutput, error) {
			return LinkParseOutput{
				Destination: in.Destination,
				ForceLink:   true,
				Handled:     true,
			}, nil
		},
	})

	result, err := conv.Convert(`[https://example.com](https://example.com)`)
	require.NoError(t, err)

	doc := decodeADFDoc(t, result.ADF)
	require.Len(t, doc.Content, 1)
	paragraph := doc.Content[0]
	require.Equal(t, "paragraph", paragraph.Type)
	require.Len(t, paragraph.Content, 1)
	textNode := paragraph.Content[0]
	require.Equal(t, "text", textNode.Type)
	require.Len(t, textNode.Marks, 1)
	assert.Equal(t, "link", textNode.Marks[0].Type)
	assert.Equal(t, "https://example.com", textNode.Marks[0].Attrs["href"])
}

func TestLinkParseHookForceCardEmitsInlineCard(t *testing.T) {
	conv := newHookReverseConverter(t, ReverseConfig{
		LinkHook: func(_ context.Context, _ LinkParseInput) (LinkParseOutput, error) {
			return LinkParseOutput{
				Destination: "https://confluence.example/wiki/pages/10",
				ForceCard:   true,
				Handled:     true,
			}, nil
		},
	})

	result, err := conv.Convert(`[Docs](../docs.md)`)
	require.NoError(t, err)

	doc := decodeADFDoc(t, result.ADF)
	require.Len(t, doc.Content, 1)
	paragraph := doc.Content[0]
	require.Equal(t, "paragraph", paragraph.Type)
	require.Len(t, paragraph.Content, 1)
	require.Equal(t, "inlineCard", paragraph.Content[0].Type)
	assert.Equal(t, "https://confluence.example/wiki/pages/10", paragraph.Content[0].Attrs["url"])
}

func TestMediaParseHookMapsLocalImageToID(t *testing.T) {
	conv := newHookReverseConverter(t, ReverseConfig{
		MediaHook: func(_ context.Context, in MediaParseInput) (MediaParseOutput, error) {
			assert.Equal(t, "./assets/a.png", in.Destination)
			assert.Equal(t, "a.png", in.Meta.Filename)
			return MediaParseOutput{
				MediaType: "image",
				ID:        "att-1",
				Handled:   true,
			}, nil
		},
	})

	result, err := conv.ConvertWithContext(context.Background(), `![Screenshot](./assets/a.png)`, ConvertOptions{SourcePath: "docs/page.md"})
	require.NoError(t, err)

	doc := decodeADFDoc(t, result.ADF)
	require.Len(t, doc.Content, 1)
	mediaSingle := doc.Content[0]
	require.Equal(t, "mediaSingle", mediaSingle.Type)
	require.Len(t, mediaSingle.Content, 1)
	media := mediaSingle.Content[0]
	require.Equal(t, "media", media.Type)
	assert.Equal(t, "image", media.Attrs["type"])
	assert.Equal(t, "att-1", media.Attrs["id"])
	assert.Equal(t, "Screenshot", media.Attrs["alt"])
}

func TestMediaParseHookMapsLocalFileToURL(t *testing.T) {
	conv := newHookReverseConverter(t, ReverseConfig{
		MediaHook: func(_ context.Context, _ MediaParseInput) (MediaParseOutput, error) {
			return MediaParseOutput{
				MediaType: "file",
				URL:       "https://files.example/spec.pdf",
				Alt:       "Spec",
				Handled:   true,
			}, nil
		},
	})

	result, err := conv.Convert(`![Spec](./assets/spec.pdf)`)
	require.NoError(t, err)

	doc := decodeADFDoc(t, result.ADF)
	require.Len(t, doc.Content, 1)
	mediaSingle := doc.Content[0]
	require.Equal(t, "mediaSingle", mediaSingle.Type)
	require.Len(t, mediaSingle.Content, 1)
	media := mediaSingle.Content[0]
	require.Equal(t, "media", media.Type)
	assert.Equal(t, "file", media.Attrs["type"])
	assert.Equal(t, "https://files.example/spec.pdf", media.Attrs["url"])
	assert.Equal(t, "Spec", media.Attrs["alt"])
}

func TestUnhandledHooksFallbackToParserBehavior(t *testing.T) {
	t.Run("link", func(t *testing.T) {
		conv := newHookReverseConverter(t, ReverseConfig{
			LinkHook: func(_ context.Context, _ LinkParseInput) (LinkParseOutput, error) {
				return LinkParseOutput{Destination: "https://ignored.example", Handled: false}, nil
			},
		})

		result, err := conv.Convert(`[https://example.com](https://example.com)`)
		require.NoError(t, err)

		doc := decodeADFDoc(t, result.ADF)
		require.Len(t, doc.Content, 1)
		paragraph := doc.Content[0]
		require.Equal(t, "paragraph", paragraph.Type)
		require.Len(t, paragraph.Content, 1)
		assert.Equal(t, "inlineCard", paragraph.Content[0].Type)
		assert.Equal(t, "https://example.com", paragraph.Content[0].Attrs["url"])
	})

	t.Run("media", func(t *testing.T) {
		conv := newHookReverseConverter(t, ReverseConfig{
			MediaHook: func(_ context.Context, _ MediaParseInput) (MediaParseOutput, error) {
				return MediaParseOutput{MediaType: "image", ID: "ignored", Handled: false}, nil
			},
		})

		result, err := conv.Convert(`![Alt](./assets/a.png)`)
		require.NoError(t, err)

		doc := decodeADFDoc(t, result.ADF)
		require.Len(t, doc.Content, 1)
		mediaSingle := doc.Content[0]
		require.Equal(t, "mediaSingle", mediaSingle.Type)
		require.Len(t, mediaSingle.Content, 1)
		media := mediaSingle.Content[0]
		assert.Equal(t, "image", media.Attrs["type"])
		assert.Equal(t, "./assets/a.png", media.Attrs["id"])
	})
}

func TestErrUnresolvedHandlingModes(t *testing.T) {
	t.Run("best effort", func(t *testing.T) {
		conv := newHookReverseConverter(t, ReverseConfig{
			LinkHook: func(_ context.Context, _ LinkParseInput) (LinkParseOutput, error) {
				return LinkParseOutput{}, ErrUnresolved
			},
		})

		result, err := conv.Convert(`[Page](../page.md)`)
		require.NoError(t, err)
		require.NotEmpty(t, result.Warnings)
		assert.Equal(t, converter.WarningUnresolvedReference, result.Warnings[0].Type)

		doc := decodeADFDoc(t, result.ADF)
		require.Len(t, doc.Content, 1)
		paragraph := doc.Content[0]
		require.Equal(t, "paragraph", paragraph.Type)
		require.Len(t, paragraph.Content, 1)
		textNode := paragraph.Content[0]
		require.Equal(t, "text", textNode.Type)
		require.Len(t, textNode.Marks, 1)
		assert.Equal(t, "../page.md", textNode.Marks[0].Attrs["href"])
	})

	t.Run("strict", func(t *testing.T) {
		conv := newHookReverseConverter(t, ReverseConfig{
			ResolutionMode: ResolutionStrict,
			LinkHook: func(_ context.Context, _ LinkParseInput) (LinkParseOutput, error) {
				return LinkParseOutput{}, ErrUnresolved
			},
		})

		_, err := conv.Convert(`[Page](../page.md)`)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unresolved link destination")
	})
}

func TestConvertWithContextCancellationPropagatesToHook(t *testing.T) {
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), traceContextKey, "cancel-check"))

	hookCalled := false
	conv := newHookReverseConverter(t, ReverseConfig{
		LinkHook: func(hookCtx context.Context, _ LinkParseInput) (LinkParseOutput, error) {
			hookCalled = true
			assert.Equal(t, "cancel-check", hookCtx.Value(traceContextKey))
			cancel()
			<-hookCtx.Done()
			return LinkParseOutput{}, hookCtx.Err()
		},
	})

	_, err := conv.ConvertWithContext(ctx, `[Page](../page.md)`, ConvertOptions{})
	require.Error(t, err)
	assert.True(t, hookCalled)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestConvertWithContextCancellationAfterHandledHookReturnsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	hookCalled := false
	conv := newHookReverseConverter(t, ReverseConfig{
		LinkHook: func(_ context.Context, in LinkParseInput) (LinkParseOutput, error) {
			hookCalled = true
			cancel()
			return LinkParseOutput{
				Destination: in.Destination,
				ForceCard:   true,
				Handled:     true,
			}, nil
		},
	})

	_, err := conv.ConvertWithContext(ctx, `[Docs](../docs.md)`, ConvertOptions{})
	require.Error(t, err)
	assert.True(t, hookCalled)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestLinkParseHookValidation(t *testing.T) {
	t.Run("forceLink and forceCard conflict", func(t *testing.T) {
		conv := newHookReverseConverter(t, ReverseConfig{
			LinkHook: func(_ context.Context, in LinkParseInput) (LinkParseOutput, error) {
				return LinkParseOutput{
					Destination: in.Destination,
					ForceLink:   true,
					ForceCard:   true,
					Handled:     true,
				}, nil
			},
		})

		_, err := conv.Convert(`[Page](../page.md)`)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "both forceLink and forceCard")
	})

	t.Run("handled requires destination", func(t *testing.T) {
		conv := newHookReverseConverter(t, ReverseConfig{
			LinkHook: func(_ context.Context, _ LinkParseInput) (LinkParseOutput, error) {
				return LinkParseOutput{Handled: true}, nil
			},
		})

		_, err := conv.Convert(`[Page](../page.md)`)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "handled link parse output requires non-empty destination")
	})
}

func TestMediaParseHookValidation(t *testing.T) {
	conv := newHookReverseConverter(t, ReverseConfig{
		MediaHook: func(_ context.Context, _ MediaParseInput) (MediaParseOutput, error) {
			return MediaParseOutput{
				MediaType: "video",
				Handled:   true,
			}, nil
		},
	})

	_, err := conv.Convert(`![Alt](./assets/a.png)`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "supported mediaType")
}
