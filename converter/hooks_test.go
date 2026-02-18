package converter

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type hookContextKey string

const traceContextKey hookContextKey = "trace"

func TestLinkHookRewritesLinkMark(t *testing.T) {
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Page","marks":[{"type":"link","attrs":{"href":"https://confluence.example/wiki/pages/123","title":"Original"}}]}]}]}`)

	var called atomic.Bool
	conv := newTestConverter(t, Config{
		LinkHook: func(ctx context.Context, in LinkRenderInput) (LinkRenderOutput, error) {
			called.Store(true)
			assert.Equal(t, "hook-test", ctx.Value(traceContextKey))
			assert.Equal(t, "mark", in.Source)
			assert.Equal(t, "docs/input.adf.json", in.SourcePath)
			assert.Equal(t, "https://confluence.example/wiki/pages/123", in.Href)
			assert.Equal(t, "Original", in.Title)
			return LinkRenderOutput{
				Href:    "../pages/123.md",
				Title:   "Updated",
				Handled: true,
			}, nil
		},
	})

	ctx := context.WithValue(context.Background(), traceContextKey, "hook-test")
	result, err := conv.ConvertWithContext(ctx, input, ConvertOptions{SourcePath: "docs/input.adf.json"})
	require.NoError(t, err)
	assert.True(t, called.Load())
	assert.Equal(t, "[Page](../pages/123.md \"Updated\")\n", result.Markdown)
}

func TestLinkHookRewritesInlineCard(t *testing.T) {
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"inlineCard","attrs":{"url":"https://confluence.example/wiki/pages/10"}}]}]}`)

	conv := newTestConverter(t, Config{
		LinkHook: func(_ context.Context, in LinkRenderInput) (LinkRenderOutput, error) {
			assert.Equal(t, "inlineCard", in.Source)
			assert.Equal(t, "https://confluence.example/wiki/pages/10", in.Href)
			return LinkRenderOutput{
				Href:    "../pages/10.md",
				Title:   "Page 10",
				Handled: true,
			}, nil
		},
	})

	result, err := conv.Convert(input)
	require.NoError(t, err)
	assert.Equal(t, "[Page 10](../pages/10.md)\n", result.Markdown)
}

func TestMediaHookOverridesMarkdown(t *testing.T) {
	input := []byte(`{"type":"doc","content":[{"type":"mediaSingle","content":[{"type":"media","attrs":{"type":"image","url":"https://cdn.example.com/a.png","alt":"Preview"}}]}]}`)

	conv := newTestConverter(t, Config{
		MediaHook: func(_ context.Context, in MediaRenderInput) (MediaRenderOutput, error) {
			assert.Equal(t, "image", in.MediaType)
			assert.Equal(t, "https://cdn.example.com/a.png", in.URL)
			assert.Equal(t, "Preview", in.Alt)
			return MediaRenderOutput{Markdown: "![Preview](./assets/a.png)", Handled: true}, nil
		},
	})

	result, err := conv.ConvertWithContext(context.Background(), input, ConvertOptions{SourcePath: "docs/input.adf.json"})
	require.NoError(t, err)
	assert.Equal(t, "![Preview](./assets/a.png)\n", result.Markdown)
}

func TestMediaHookOverridesFileMarkdown(t *testing.T) {
	input := []byte(`{"type":"doc","content":[{"type":"mediaSingle","content":[{"type":"media","attrs":{"type":"file","id":"att-9","alt":"Spec"}}]}]}`)

	conv := newTestConverter(t, Config{
		MediaHook: func(_ context.Context, in MediaRenderInput) (MediaRenderOutput, error) {
			assert.Equal(t, "file", in.MediaType)
			assert.Equal(t, "att-9", in.ID)
			return MediaRenderOutput{Markdown: "[Spec](./assets/spec.pdf)", Handled: true}, nil
		},
	})

	result, err := conv.Convert(input)
	require.NoError(t, err)
	assert.Equal(t, "[Spec](./assets/spec.pdf)\n", result.Markdown)
}

func TestUnhandledHooksFallbackToExistingBehavior(t *testing.T) {
	t.Run("link", func(t *testing.T) {
		input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Link","marks":[{"type":"link","attrs":{"href":"https://example.com"}}]}]}]}`)

		conv := newTestConverter(t, Config{
			LinkHook: func(_ context.Context, _ LinkRenderInput) (LinkRenderOutput, error) {
				return LinkRenderOutput{Href: "../ignored.md", Handled: false}, nil
			},
		})

		result, err := conv.Convert(input)
		require.NoError(t, err)
		assert.Equal(t, "[Link](https://example.com)\n", result.Markdown)
	})

	t.Run("media", func(t *testing.T) {
		input := []byte(`{"type":"doc","content":[{"type":"mediaSingle","content":[{"type":"media","attrs":{"type":"image","url":"https://example.com/a.png"}}]}]}`)

		conv := newTestConverter(t, Config{
			MediaHook: func(_ context.Context, _ MediaRenderInput) (MediaRenderOutput, error) {
				return MediaRenderOutput{Markdown: "![Ignored](./ignored.png)", Handled: false}, nil
			},
		})

		result, err := conv.Convert(input)
		require.NoError(t, err)
		assert.Equal(t, "![Image](https://example.com/a.png)\n", result.Markdown)
	})
}

func TestHookErrUnresolvedBestEffortWarnsAndFallsBack(t *testing.T) {
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Page","marks":[{"type":"link","attrs":{"href":"https://confluence.example/wiki/pages/123"}}]}]}]}`)

	conv := newTestConverter(t, Config{
		LinkHook: func(_ context.Context, _ LinkRenderInput) (LinkRenderOutput, error) {
			return LinkRenderOutput{}, ErrUnresolved
		},
	})

	result, err := conv.Convert(input)
	require.NoError(t, err)
	assert.Equal(t, "[Page](https://confluence.example/wiki/pages/123)\n", result.Markdown)
	require.NotEmpty(t, result.Warnings)
	assert.Equal(t, WarningUnresolvedReference, result.Warnings[0].Type)
}

func TestHookErrUnresolvedStrictFailsConversion(t *testing.T) {
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Page","marks":[{"type":"link","attrs":{"href":"https://confluence.example/wiki/pages/123"}}]}]}]}`)

	conv := newTestConverter(t, Config{
		ResolutionMode: ResolutionStrict,
		LinkHook: func(_ context.Context, _ LinkRenderInput) (LinkRenderOutput, error) {
			return LinkRenderOutput{}, ErrUnresolved
		},
	})

	_, err := conv.Convert(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unresolved link reference")
}

func TestHookValidationErrors(t *testing.T) {
	t.Run("link output requires href", func(t *testing.T) {
		input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Page","marks":[{"type":"link","attrs":{"href":"https://example.com"}}]}]}]}`)

		conv := newTestConverter(t, Config{
			LinkHook: func(_ context.Context, _ LinkRenderInput) (LinkRenderOutput, error) {
				return LinkRenderOutput{Handled: true}, nil
			},
		})

		_, err := conv.Convert(input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "handled link render output requires non-empty href")
	})

	t.Run("media output requires markdown", func(t *testing.T) {
		input := []byte(`{"type":"doc","content":[{"type":"mediaSingle","content":[{"type":"media","attrs":{"type":"image","url":"https://example.com/a.png"}}]}]}`)

		conv := newTestConverter(t, Config{
			MediaHook: func(_ context.Context, _ MediaRenderInput) (MediaRenderOutput, error) {
				return MediaRenderOutput{Handled: true}, nil
			},
		})

		_, err := conv.Convert(input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "handled media render output requires non-empty markdown")
	})
}

func TestConvertWithContextCancellationPropagatesToHook(t *testing.T) {
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Page","marks":[{"type":"link","attrs":{"href":"https://example.com"}}]}]}]}`)

	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), traceContextKey, "context-cancel"))

	hookCalled := atomic.Bool{}
	conv := newTestConverter(t, Config{
		LinkHook: func(hookCtx context.Context, _ LinkRenderInput) (LinkRenderOutput, error) {
			hookCalled.Store(true)
			assert.Equal(t, "context-cancel", hookCtx.Value(traceContextKey))
			cancel()
			<-hookCtx.Done()
			return LinkRenderOutput{}, hookCtx.Err()
		},
	})

	_, err := conv.ConvertWithContext(ctx, input, ConvertOptions{})
	require.Error(t, err)
	assert.True(t, hookCalled.Load())
	assert.ErrorIs(t, err, context.Canceled)
}

func TestConvertWithContextCancellationAfterHandledHookReturnsCanceled(t *testing.T) {
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Page","marks":[{"type":"link","attrs":{"href":"https://example.com"}}]}]}]}`)

	ctx, cancel := context.WithCancel(context.Background())
	var hookCalled atomic.Bool

	conv := newTestConverter(t, Config{
		LinkHook: func(_ context.Context, in LinkRenderInput) (LinkRenderOutput, error) {
			hookCalled.Store(true)
			cancel()
			return LinkRenderOutput{
				Href:    in.Href,
				Title:   in.Title,
				Handled: true,
			}, nil
		},
	})

	_, err := conv.ConvertWithContext(ctx, input, ConvertOptions{})
	require.Error(t, err)
	assert.True(t, hookCalled.Load())
	assert.ErrorIs(t, err, context.Canceled)
}

func TestConcurrentConvertWithThreadSafeHook(t *testing.T) {
	input := []byte(`{"type":"doc","content":[{"type":"mediaSingle","content":[{"type":"media","attrs":{"type":"image","url":"https://example.com/a.png"}}]}]}`)

	var (
		mu    sync.Mutex
		calls int
	)

	conv := newTestConverter(t, Config{
		MediaHook: func(_ context.Context, _ MediaRenderInput) (MediaRenderOutput, error) {
			mu.Lock()
			calls++
			mu.Unlock()
			return MediaRenderOutput{Handled: false}, nil
		},
	})

	const workers = 8
	const iterations = 100

	errCh := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				result, err := conv.Convert(input)
				if err != nil {
					errCh <- err
					return
				}
				if result.Markdown != "![Image](https://example.com/a.png)\n" {
					errCh <- errors.New("unexpected markdown output")
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		require.NoError(t, err)
	}

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, workers*iterations, calls)
}
