# Plan: Registry-based Extension Hook System

## Context

ADF extension nodes (e.g., PlantUML macros) currently render as raw JSON inside ` ```adf:extension ` code blocks. Users want to render specific extensions as clean, human-readable Markdown and reconstruct the original ADF nodes on the way back. This requires a hook system where library consumers register custom handlers per extension type. No built-in handlers are shipped — only the framework/interface.

## Design

Handlers are keyed by **`extensionKey`** in both directions (e.g., `"plantumlcloud"`).

**Forward (ADF → Markdown):** The handler produces the inner content (`Markdown`) plus optional string metadata (`Metadata`). The framework wraps them in a pandoc fenced div:

```
:::{ .adf-extension key="plantumlcloud" <handler metadata as attrs> }
<handler Markdown>
:::
```

**Reverse (Markdown → ADF):** The existing pandoc div parser detects `.adf-extension` divs. The walker routes them to the registered handler, passing the div body (`Body`) and parsed attributes (`Metadata`). The handler reconstructs the ADF node.

Handlers are responsible for serializing/deserializing all metadata values to/from strings. The framework round-trips strings verbatim as div attributes.

## Interface & Types

### NEW: `converter/extension_handler.go`

```go
type ExtensionRenderInput struct {
    SourcePath string
    Node       Node
}

type ExtensionRenderOutput struct {
    Markdown string
    Metadata map[string]string // handler serializes values; framework stores as div attrs
    Handled  bool
}

type ExtensionParseInput struct {
    SourcePath   string
    ExtensionKey string
    Body         string            // raw markdown content inside the .adf-extension div
    Metadata     map[string]string // div attrs (minus key and .adf-extension class)
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

Uses interface (not function type) because each handler binds both directions. `Handled bool` lets handlers decline and fall through to default behavior.

## Files to Create/Modify

### 1. NEW: `converter/extension_handler.go` — Interface & Types

Define the types and interface above.

### 2. MODIFY: `converter/config.go`

- Add `ExtensionHandlers map[string]ExtensionHandler` field to `Config` (json:`"-"`), keyed by `extensionKey`
- Add `cloneExtensionHandlerMap` helper (shallow copy; interfaces are reference-semantic)
- Update `clone()` to copy the map

### 3. MODIFY: `converter/extensions.go`

In `convertExtension()`, before the strategy switch:
1. Look up handler by `extensionKey` from `s.config.ExtensionHandlers`
2. If found, call `handler.ToMarkdown(s.ctx, input)`
3. If `Handled: true`:
   - Build `:::{ .adf-extension key="<extensionKey>" <metadata attrs> }\n<Markdown>\n:::`
   - Return the wrapped markdown
4. If `Handled: false`, fall through to existing strategy logic

### 4. MODIFY: `mdconverter/config.go`

- Add `ExtensionHandlers map[string]converter.ExtensionHandler` field to `ReverseConfig` (json:`"-"`), keyed by `extensionKey`
- Add `cloneExtensionHandlerMap` helper
- Update `clone()` to copy the map
- Update `needsPandocBlockExtension()` to also return `true` when `len(cfg.ExtensionHandlers) > 0`
  (ensures the pandoc div parser is registered even if no other pandoc block options are set)

### 5. MODIFY: `mdconverter/pandoc_div_convert.go`

Add handling for `.adf-extension` class divs before (or alongside) the existing `.details` / `align` cases:

1. If div has class `.adf-extension`:
   a. Extract `key` attr → look up handler in `s.config.ExtensionHandlers`
   b. If handler found: call `handler.FromMarkdown(s.ctx, input)` with `Body` = div body text, `Metadata` = remaining attrs
   c. If `Handled: true`: append the returned `Node`; return
   d. If `Handled: false` or no handler: fall through to unknown-class / blockquote warning behavior

No changes needed to `mdconverter/extensions.go` (`parseExtensionFence`) or `mdconverter/walker.go` — the existing pandoc div path handles everything.

## Wrapper Format

```
:::{ .adf-extension key="<extensionKey>" [<attr>="<value>" ...] }
<handler-produced markdown content>
:::
```

Examples:
```
:::{ .adf-extension key="plantumlcloud" extension-type="com.atlassian.confluence.macro.core" local-id="abc" layout="default" }
@startuml
Alice -> Bob
@enduml
:::
```
```
:::{ .adf-extension key="drawio" }
<diagram content>
:::
```

The framework emits and parses the wrapper. Handlers only produce/consume inner content and string metadata.

## Tests

### NEW: `converter/extension_handler_test.go`
- Handler called and output used; wrapper div emitted with correct attrs
- Handler returns `Handled: false` → falls back to `ExtensionMode` strategy (no wrapper emitted)
- Handler returns error → propagates
- No handler registered → existing behavior unchanged

### NEW: `mdconverter/extension_handler_test.go`
- Handler called for matching `extensionKey` in `.adf-extension` div
- Metadata attrs passed correctly to handler
- Div with no matching handler → blockquote warning (existing fallback)
- Handler returns `Handled: false` → blockquote warning fallback
- Handler returns error → propagates

### Existing test files (additions)
- `converter/config_test.go`: clone preserves `ExtensionHandlers`
- `mdconverter/config_test.go`: clone preserves `ExtensionHandlers`; `needsPandocBlockExtension()` returns true when handlers registered

## Verification

1. `go test ./...` — all existing tests pass (no regressions)
2. `go test -race ./...` — no data races
3. `go vet ./...` — no issues
4. New handler tests pass in both directions
