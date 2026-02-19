# Plan: Registry-based Extension Hook System

## Context

ADF extension nodes (e.g., PlantUML macros) currently render as raw JSON inside ` ```adf:extension ` code blocks. Users want to render specific extensions as clean, human-readable Markdown (e.g., PlantUML code blocks with metadata in HTML comments) and reconstruct the original ADF nodes on the way back. This requires a hook system where library consumers register custom handlers per extension type. No built-in handlers are shipped — only the framework/interface.

## Approach

Add an `ExtensionHandler` interface following the existing `LinkRenderHook`/`MediaRenderHook` pattern (typed Input/Output structs, `context.Context`). Register handlers on both `Config` and `ReverseConfig`. Forward direction checks handlers before the existing `ExtensionMode` strategy. Reverse direction checks handlers in two places: the `convertBlockSlice` walker (for HTML comment + code block pairs) and `parseExtensionFence` (for standalone code blocks without metadata).

## Files to Create/Modify

### 1. NEW: `converter/extension_handler.go` — Interface & Types

```go
type ExtensionRenderInput struct {
    SourcePath string
    Node       Node
}

type ExtensionRenderOutput struct {
    Markdown string
    Handled  bool
}

type ExtensionParseInput struct {
    SourcePath string
    Language   string
    Body       string
    Metadata   map[string]any  // from HTML comment, nil if absent
}

type ExtensionParseOutput struct {
    Node    Node
    Handled bool
}

type ExtensionHandler interface {
    ToMarkdown(ctx context.Context, in ExtensionRenderInput) (ExtensionRenderOutput, error)
    FromMarkdown(ctx context.Context, in ExtensionParseInput) (ExtensionParseOutput, error)
}
```

Follows the typed Input/Output struct pattern from `converter/hooks.go`. Uses interface (not function type) because the handler binds both directions together. `Handled bool` allows handlers to decline and fall back to default behavior.

### 2. MODIFY: `converter/config.go`

- Add `ExtensionHandlers map[string]ExtensionHandler` field to `Config` (json:"-"), keyed by `extensionKey`
- Add `cloneExtensionHandlerMap` helper (shallow copy of map, interfaces are reference-semantic)
- Update `clone()` to copy the map

### 3. MODIFY: `converter/extensions.go`

In `convertExtension()`, before the strategy switch:
1. Look up handler by `extensionKey` from `s.config.ExtensionHandlers`
2. If found, call `handler.ToMarkdown(s.ctx, input)`
3. If `Handled: true`, return the markdown; if `Handled: false`, fall through to existing strategy logic

### 4. MODIFY: `mdconverter/config.go`

- Add `ExtensionHandlers map[string]converter.ExtensionHandler` field to `ReverseConfig` (json:"-"), keyed by code block language
- Add `cloneExtensionHandlerMap` helper
- Update `clone()` to copy the map

### 5. MODIFY: `mdconverter/extensions.go`

- In `parseExtensionFence()`, before the `switch` on language: check `s.config.ExtensionHandlers[language]`, call `handler.FromMarkdown()` with `metadata: nil`
- Add `parseExtensionComment(node, source)` — regex-based parser for `<!-- adf:extension <key> <json> -->`
- Add `applyExtensionParseHandler()` helper

### 6. MODIFY: `mdconverter/walker.go`

In `convertBlockSlice()`, add a peek-ahead block (following the `consumeDetailsBlock` pattern):
1. If current child is `*ast.HTMLBlock` matching `<!-- adf:extension <key> <json> -->`
2. AND next child is `*ast.FencedCodeBlock`
3. AND a handler is registered for the code block's language
4. Call `handler.FromMarkdown()` with the parsed metadata
5. If handled: consume both nodes (index += 2)
6. If not handled: fall through to normal processing

### HTML Comment Format

```
<!-- adf:extension <extensionKey> <optional JSON> -->
```

Examples:
```
<!-- adf:extension plantumlcloud {"extensionType":"com.atlassian.confluence.macro.core","layout":"default","localId":"abc"} -->
```
```
<!-- adf:extension drawio -->
```

The framework only **parses** these comments — handlers are responsible for **producing** them in their `ToMarkdown` output.

## Tests

### NEW: `converter/extension_handler_test.go`
- Handler called and output used
- Handler returns `Handled: false` → falls back to ExtensionMode strategy
- Handler returns error → propagates
- No handler registered → existing behavior unchanged

### NEW: `mdconverter/extension_handler_test.go`
- Handler called for matching code block language
- HTML comment metadata passed to handler
- No preceding comment → `metadata` is nil
- Handler returns `Handled: false` → normal code block processing
- Handler returns error → propagates

### Existing test files (additions)
- `converter/config_test.go`: clone preserves ExtensionHandlers
- `mdconverter/config_test.go`: clone preserves ExtensionHandlers

## Verification

1. `go test ./...` — all existing tests pass (no regressions)
2. `go test -race ./...` — no data races
3. `go vet ./...` — no issues
4. New handler tests pass in both directions
