# Plan: Bidirectional Link and Media Hooks

## Goal
Add runtime hooks to both converter directions so tools embedding this library can:
1. rewrite Confluence page links to relative Markdown references during ADF -> Markdown
2. download attachments/images and emit local filesystem paths in Markdown
3. map relative Markdown links and local media paths back to Confluence URLs/IDs during Markdown -> ADF
4. support cancellation/timeouts and predictable unresolved-reference behavior for different workflows

## Scope
- `converter/` (ADF -> Markdown)
- `mdconverter/` (Markdown -> ADF)
- library API surface, tests, and docs
- context-aware conversion and hook invocation plumbing

## Non-Goals
- No built-in downloader/uploader/network client in this library.
- No automatic filesystem writes by the converter itself.
- No behavior changes when hooks are unset.

---

## Design Principles
1. Backward compatibility: existing output stays identical when hooks are nil.
2. Runtime hooks only: callback fields use `json:"-"` and are excluded from config serialization.
3. Explicit fallback: hook can return `Handled=false` to keep current behavior.
4. Predictable unresolved behavior: hook can return sentinel `ErrUnresolved`, handled by resolution mode.
5. Concurrency contract: concurrent `Convert` calls are safe only if converter code has no shared mutable state and caller-provided hooks are thread-safe.

---

## Resolution Modes

Add shared resolver mode semantics to both directions:

```go
var ErrUnresolved = errors.New("unresolved link or media reference")

type ResolutionMode string

const (
	ResolutionBestEffort ResolutionMode = "best_effort" // continue with fallback + warning
	ResolutionStrict     ResolutionMode = "strict"      // fail conversion on unresolved reference
)
```

`Config`/`ReverseConfig` additions:

```go
ResolutionMode ResolutionMode `json:"resolutionMode,omitempty"`
```

Recommended usage:
- `best_effort`: preview/diff flows
- `strict`: validate/push/publish flows

---

## Proposed API Surface

### Shared typed metadata

Typed optional metadata appears alongside raw attrs so hook implementations do not need to parse `map[string]any` in common cases.

```go
type LinkMetadata struct {
	PageID       string
	SpaceKey     string
	AttachmentID string
	Filename     string
	Anchor       string
}

type MediaMetadata struct {
	PageID       string
	SpaceKey     string
	AttachmentID string
	Filename     string
	Anchor       string
}
```

### Conversion options and context plumbing

Keep current `Convert(...)` methods as wrappers and add context-aware variants:

```go
type ConvertOptions struct {
	SourcePath string // path of source markdown/adf doc in caller workspace
}

ConvertWithContext(ctx context.Context, input []byte, opts ConvertOptions) (Result, error)      // converter
ConvertWithContext(ctx context.Context, markdown string, opts ConvertOptions) (Result, error) // mdconverter
```

Existing `Convert(...)` calls use `context.Background()` and empty options.

### `converter` (ADF -> Markdown)

```go
type LinkRenderHook func(ctx context.Context, in LinkRenderInput) (LinkRenderOutput, error)
type MediaRenderHook func(ctx context.Context, in MediaRenderInput) (MediaRenderOutput, error)

type LinkRenderInput struct {
	Source     string // "mark" | "inlineCard" | "blockCard"
	SourcePath string
	Href       string
	Title      string
	Text       string // link/inline-card/block-card text when available
	Meta       LinkMetadata
	Attrs      map[string]any // raw node/mark attrs for advanced cases
}

type LinkRenderOutput struct {
	Href     string
	Title    string
	TextOnly bool // true => render only text, no markdown link wrapper
	Handled  bool // false => fallback to existing behavior
}

type MediaRenderInput struct {
	SourcePath string
	MediaType  string // "image", "file", ...
	ID         string
	URL        string
	Alt        string
	Meta       MediaMetadata
	Attrs      map[string]any // raw node attrs for advanced cases
}

type MediaRenderOutput struct {
	Markdown string // final markdown snippet, e.g. ![alt](./assets/file.png)
	Handled  bool
}
```

`Config` additions:

```go
LinkHook        LinkRenderHook `json:"-"`
MediaHook       MediaRenderHook `json:"-"`
ResolutionMode  ResolutionMode `json:"resolutionMode,omitempty"`
```

### `mdconverter` (Markdown -> ADF)

```go
type LinkParseHook func(ctx context.Context, in LinkParseInput) (LinkParseOutput, error)
type MediaParseHook func(ctx context.Context, in MediaParseInput) (MediaParseOutput, error)

type LinkParseInput struct {
	SourcePath  string // required for resolving ../relative.md consistently
	Destination string
	Title       string
	Text        string
	Meta        LinkMetadata
	Raw         map[string]any // parsed/raw extras when available
}

type LinkParseOutput struct {
	Destination string
	Title       string
	ForceLink   bool // bypass inlineCard/blockCard auto-detection
	ForceCard   bool // force card output; currently emits inlineCard
	Handled     bool
}

type MediaParseInput struct {
	SourcePath  string // required for resolving relative filesystem paths
	Destination string
	Alt         string
	Meta        MediaMetadata
	Raw         map[string]any // parsed/raw extras when available
}

type MediaParseOutput struct {
	MediaType string // "image" or "file"
	ID        string
	URL       string
	Alt       string
	Handled   bool
}
```

`ReverseConfig` additions:

```go
LinkHook       LinkParseHook `json:"-"`
MediaHook      MediaParseHook `json:"-"`
ResolutionMode ResolutionMode `json:"resolutionMode,omitempty"`
```

### Current card support (reverse path)

- Reverse conversion currently supports `inlineCard` output only.
- `LinkParseOutput.ForceCard=true` forces `inlineCard` for non-mention links.
- This MUST NOT error solely because block-card output is not implemented.
- If block-card support is added later, behavior can expand while preserving backward compatibility.

---

## Hook Invocation Order and Coverage

### ADF -> Markdown (`converter`)
1. Link marks (`marks.go`): run `LinkHook` before rendering `[text](href "title")`.
2. Inline cards (`inline.go`): run `LinkHook` with `Source="inlineCard"` before `InlineCardStyle` formatting.
3. Block cards (if supported): run the same `LinkHook` with `Source="blockCard"`.
4. Media nodes (`media.go`): run `MediaHook` at the start of `convertMedia()`.

### Markdown -> ADF (`mdconverter`)
1. `ast.Link` in `inline.go`:
   - keep `mention:` detection first
   - run `LinkHook` for non-mention links
   - if `ForceLink=true`, emit a normal link mark and bypass card heuristics
   - if `ForceCard=true`, emit `inlineCard` and bypass normal card heuristics
   - otherwise apply existing inline-card heuristics
2. Block card surfaces (if supported): apply `LinkHook` for those parser paths as well.
3. `ast.Image` in `inline.go`: run `MediaHook` before `MediaBaseURL` stripping.

### Unresolved handling
- Hook returns `ErrUnresolved`:
  - `ResolutionStrict`: fail `Convert`.
  - `ResolutionBestEffort`: add warning and continue with fallback behavior.

---

## Output Validation Rules

Validate hook outputs explicitly and return clear conversion errors for invalid payloads:

1. Link parse conflict: `ForceLink && ForceCard` is invalid.
2. Handled link parse result requires non-empty `Destination`.
3. Handled link render result requires non-empty `Href` unless `TextOnly=true`.
4. Handled media render result requires non-empty `Markdown`.
5. Handled media parse result requires valid payload:
   - `MediaType` must be supported
   - at least one of `ID` or `URL` must be set
   - reject structurally conflicting payloads
6. Hook returns `Handled=false` with populated output fields: ignored safely (and optionally warn in debug tests).
7. `Handled=true` with `ForceCard=true` requires non-empty `Destination`; current output mapping is `inlineCard`.

---

## Implementation Plan

### Task 1: Add hook types, context plumbing, and mode controls
Files:
- `converter/config.go`
- `mdconverter/config.go`
- `converter/converter.go`
- `mdconverter/mdconverter.go`

Work:
- define hook input/output and typed metadata structs
- add `context.Context` to hook signatures
- add `ResolutionMode` + default (`best_effort`) + validation
- add `ConvertWithContext(..., ConvertOptions)` methods and wrap existing `Convert(...)`
- ensure `SourcePath` reaches hook inputs (especially reverse path resolution)
- keep hook fields `json:"-"`; copy function fields in `clone()`

### Task 2: Integrate forward link hooks across all link surfaces
Files:
- `converter/marks.go`
- `converter/inline.go`
- `converter/converter.go` (if `blockCard` dispatch support is needed)

Work:
- apply `LinkHook` for link mark rendering
- apply `LinkHook` for `inlineCard`
- apply `LinkHook` for `blockCard` where supported
- enforce validation and `ErrUnresolved` behavior based on `ResolutionMode`

### Task 3: Integrate forward media hooks
Files:
- `converter/media.go`

Work:
- apply `MediaHook` before existing media branching
- allow full markdown override for downloaded local paths
- enforce media output validation
- preserve current behavior when hook is nil/unhandled

### Task 4: Integrate reverse link hooks with source-path context
Files:
- `mdconverter/inline.go`
- `mdconverter/extensions.go` (if block-card-like parsing path exists)

Work:
- apply `LinkHook` after mention detection and before card heuristics
- include `SourcePath` in input so `../relative.md` is resolvable
- support `ForceLink`/`ForceCard` with conflict validation
- enforce `ErrUnresolved` mode behavior

### Task 5: Integrate reverse media hooks with source-path context
Files:
- `mdconverter/inline.go`

Work:
- apply `MediaHook` before base URL stripping
- include `SourcePath` and typed metadata
- validate handled media payload and map to consistent ADF attrs
- preserve current behavior when unhandled

### Task 6: Tests
Files:
- `converter/hooks_test.go` (new)
- `mdconverter/hooks_test.go` (new)
- optionally extend `converter/converter_test.go` and `mdconverter/golden_test.go`

Forward test cases:
1. link hook rewrites Confluence URL to `../page.md`
2. inlineCard URL is rewritten through the same link hook
3. blockCard path (if supported) is also rewritten through link hook
4. media hook returns local image path markdown
5. media hook returns local file link markdown
6. unhandled hook falls back to existing output
7. `ErrUnresolved` in best-effort warns + continues
8. `ErrUnresolved` in strict fails conversion
9. context cancellation propagates to hook and aborts conversion

Reverse test cases:
1. link hook rewrites `../page.md` to Confluence URL in link mark
2. link hook forces regular link (no card auto-conversion)
3. link hook forces card mode where supported
4. media hook maps `./assets/a.png` to media `id`
5. media hook maps local file path to media `url`
6. unhandled hook falls back to existing parser behavior
7. `ErrUnresolved` strict vs best-effort behavior
8. context cancellation propagation

Validation test cases:
1. `ForceLink && ForceCard` returns error
2. `Handled=true` with empty destination returns error
3. handled media with invalid payload returns error

Concurrency tests:
- verify converter internals remain free of shared mutable state across concurrent `Convert` calls
- document and test expectation that hook closures with mutable state must synchronize access

### Task 7: Documentation
Files:
- `README.md`
- `docs/features.md`

Work:
- document context-aware conversion entrypoints and hook signatures
- document strict vs best-effort usage guidance for preview vs publish flows
- document source-path requirement for reverse resolution
- add examples for:
  - Confluence URL <-> relative path mapping
  - attachment/image ID <-> local asset path mapping
- include explicit concurrency note in docs

---

## Success Criteria
- Hook signatures include `context.Context` in both directions.
- `ResolutionMode` and `ErrUnresolved` behavior are implemented and tested.
- Reverse hook inputs include `SourcePath` for deterministic relative-path resolution.
- Typed metadata (`page ID`, `space key`, `attachment ID`, `filename`, `anchor`) is available alongside raw attrs.
- Conflicting/invalid hook outputs are validated with explicit errors.
- Link hooks cover mark links, inline cards, and block cards where supported.
- Default behavior is unchanged when hooks are unset.
- Concurrency contract is clear in docs/tests: concurrent `Convert` is safe only when hooks are thread-safe and converter has no shared mutable state.
- New tests pass without regressions in existing golden suites.
