# Supported Features

This document describes supported behavior for both conversion directions:

- ADF JSON -> Markdown (`converter` package)
- Markdown -> ADF JSON (`mdconverter` package)

## Conversion APIs

| Direction | Constructor | Convert API | Result |
|---|---|---|---|
| ADF -> Markdown | `converter.New(config)` | `Convert([]byte)` / `ConvertWithContext(ctx, []byte, opts)` | `converter.Result{Markdown, Warnings}` |
| Markdown -> ADF | `mdconverter.New(config)` | `Convert(string)` / `ConvertWithContext(ctx, string, opts)` | `mdconverter.Result{ADF, Warnings}` |

Both packages validate config at `New(...)` time and keep config immutable afterward.

## ADF -> Markdown (`converter`)

### Node Support Matrix

| ADF Node | Default Markdown Output | Notes / Config |
|---|---|---|
| `doc` | Root container | Ensures trailing newline for non-empty output. |
| `paragraph` | Text block separated by blank lines | Supports inline marks and inline nodes. |
| `text` | Plain text | Marks applied via mark stack continuity. |
| `heading` | `#` through `######` | `HeadingOffset` with clamping; optional HTML or Pandoc alignment. |
| `blockquote` | `>` blockquote | Nested content supported. |
| `rule` | `---` | Standard thematic break. |
| `hardBreak` | `\\` + newline | `HardBreakStyle`: `backslash` or `html` (`<br>`). |
| `codeBlock` | Fenced code block | Language aliasing via `LanguageMap`. |
| `bulletList` | `- item` | Marker configurable via `BulletMarker` (`-`, `*`, `+`). |
| `orderedList` | `1.`, `2.`, ... | `OrderedListStyle`: `incremental` or `lazy` (`1.` for every item). |
| `taskList` / `taskItem` | `- [ ]` / `- [x]` | Nested task structures supported. |
| `table` | Pipe, Grid (Pandoc) or HTML table | `TableMode`: `auto`, `pipe`, `pandoc`, `autopandoc`, `html`; auto-detects complex cells/spans. |
| `panel` | GitHub-style callout blockquote | `PanelStyle`: `none`, `bold`, `github`, `title`. |
| `decisionList` / `decisionItem` | Blockquote with decision prefix | `DecisionStyle`: `emoji` (`✓/? Decision`) or `text` (`DECIDED/UNDECIDED`). |
| `expand` / `nestedExpand` | `<details><summary>...</summary>` | `ExpandStyle`: `html` (default), `blockquote`, or `pandoc` (`:::{ .details }`). |
| `emoji` | `:shortcode:` | `EmojiStyle`: `shortcode` or `unicode` fallback. |
| `mention` | `[@Name](mention:id)` | `MentionStyle`: `text`, `link`, `html`, `pandoc`. |
| `status` | `[Status: TEXT]` | `StatusStyle`: `bracket` or `text`. |
| `date` | Formatted timestamp | Uses configurable `DateFormat`. |
| `inlineCard` | `[title](url)` | `InlineCardStyle`: `link`, `url`, `embed` (`adf:inlineCard` fenced JSON), `pandoc`. |
| `media` (+ `mediaSingle`/`mediaGroup`) | Image markdown or placeholders | External: `![alt](url)`; internal: `[Image: id]` / `[File: id]`; optional `MediaBaseURL` expansion. |
| `extension` / `inlineExtension` / `bodiedExtension` | Fenced JSON by default | `Extensions.Default`: `json`, `text`, `strip`; per-type override via `Extensions.ByType`. |

Unknown handling is policy driven:

- `UnknownNodes`: `placeholder`, `skip`, or `error`
- `UnknownMarks`: `skip`, `placeholder`, or `error`

### Mark Support

| Mark | Default Output | Alternatives |
|---|---|---|
| `strong` | `**text**` | - |
| `em` | `*text*` (or `_text_` in mixed emphasis scenarios) | - |
| `strike` | `~~text~~` | - |
| `code` | `` `text` `` | - |
| `link` | `[text](href "title")` | Can be rewritten by runtime `LinkHook`. |
| `underline` | `**text**` | `ignore`, `bold`, `html` (`<u>`), `pandoc` (`[text]{.underline}`). |
| `subsup` | HTML by default (`<sub>`, `<sup>`) | `ignore`, `html`, `latex`, `pandoc` (`~text~`, `^text^`). |
| `textColor` | dropped by default | `ignore`, `html` (`<span style="color: ...">`), `pandoc` (`[text]{color="..."}`). |
| `backgroundColor` | dropped by default | `ignore`, `html` (`<span style="background-color: ...">`), `pandoc` (`[text]{background-color="..."}`). |

## Markdown -> ADF (`mdconverter`)

### Syntax Support Matrix

| Markdown / HTML Input | ADF Output | Notes / Config |
|---|---|---|
| Paragraph text | `paragraph` + `text` | Supports mark stack traversal. |
| `#`..`######` | `heading` | Applies reverse `HeadingOffset` with clamping. |
| `>` blockquote | `blockquote` or specialized node | Disambiguates to panel/decision/expand based on configured detection modes. |
| `---` | `rule` | - |
| Hard line break | `hardBreak` | Supports markdown hard breaks and `<br>`. |
| Fenced/indented code | `codeBlock` | Reverse language mapping via `LanguageMap`. |
| Bullet/ordered lists | `bulletList` / `orderedList` | Preserves ordered `start` when present. |
| Task lists (`- [ ]`, `- [x]`) | `taskList` / `taskItem` | State mapped to `TODO`/`DONE`. |
| GFM pipe tables | `table` nodes | Header/data cells reconstructed. |
| Grid tables (`+---+`) | `table` nodes | Reconstructs Pandoc grid tables into ADF tables. |
| HTML tables (`<table>`) | `table` nodes | Supports `colspan` / `rowspan` and nested markdown parsing in cells. |
| `[text](mention:id)` | `mention` | Controlled by `MentionDetection` (`link` / `all`). |
| `[Name]{.mention mention-id="..."}` | `mention` | Controlled by `MentionDetection` (`pandoc` / `all`). |
| `![alt](dest)` | `mediaSingle` + `media` | Hook runs first; fallback strips `MediaBaseURL` to `id` when configured. |
| `[Image: id]`, `[File: id]` | `mediaSingle` + `media` | Parsed from text patterns. |
| `:shortcode:` | `emoji` | Controlled by `EmojiDetection`. |
| `[Status: TEXT]` | `status` | Controlled by `StatusDetection`. |
| `YYYY-MM-DD` | `date` | Controlled by `DateDetection` + `DateFormat`. |
| `@Name` | `mention` | Requires `MentionRegistry`; controlled by `MentionDetection` (`at` / `all`). |
| `<u>`, `<sub>`, `<sup>` | `underline` / `subsup` marks | Parsed from inline HTML tags. |
| `~text~`, `^text^` | `subsup` marks | Pandoc subscript and superscript. |
| `[text]{.underline}` | `underline` mark | Pandoc underline span. |
| `<span style="color:...">` | `textColor` mark | Inline HTML parsing. |
| `[text]{color="..."}` | `textColor` mark | Pandoc color span. |
| `<span style="background-color:...">` | `backgroundColor` mark | Inline HTML parsing. |
| `[text]{background-color="..."}` | `backgroundColor` mark | Pandoc background color span. |
| `<span data-mention-id="...">` | `mention` node | Controlled by `MentionDetection` (`html` / `all`). |
| `<details><summary>...</summary>...</details>` | `expand` / `nestedExpand` | Controlled by `ExpandDetection` (`html` / `all`). |
| `:::{ .details summary="..." }...:::` | `expand` / `nestedExpand` | Controlled by `ExpandDetection` (`pandoc` / `all`). |
| `<div align="...">` | aligned `paragraph` | Alignment attr restored in ADF attrs. |
| `:::{ align="..." }` | aligned `paragraph`/`heading` | Pandoc fenced div with alignment attribute. |
| `<h1 align="...">...` | aligned `heading` | Alignment attr + heading level restoration. |
| `[title]{.inline-card url="..."}` | `inlineCard` | Controlled by `InlineCardDetection` (`pandoc` / `all`). |
| ```` ```adf:extension ```` | extension node | Reconstructs extension payload from JSON body. |
| `:::{ .adf-extension key="..." }` | extension node | Reconstructs handled extension from custom handler metadata/content. |
| ```` ```adf:inlineCard ```` | `inlineCard` | Reconstructs inline card attrs from JSON body. |

Unsupported markdown constructs are downgraded to text with warnings when possible instead of failing by default.

### Blockquote Disambiguation Order

When panel/decision/expand detection is enabled, blockquotes are checked in this order:

1. GitHub/title panel callouts (for example `> [!NOTE]`, `> [!INFO: Title]`)
2. Bold-prefix panels (for example `> **Info**: ...`)
3. Decision prefixes (for example `> **✓ Decision**: ...`, `> **DECIDED**: ...`)
4. Expand patterns (blockquote title style)
5. Fallback to plain `blockquote`

### Reverse Detection Defaults

| Field | Default |
|---|---|
| `MentionDetection` | `link` |
| `EmojiDetection` | `shortcode` |
| `StatusDetection` | `bracket` |
| `DateDetection` | `iso` |
| `PanelDetection` | `github` |
| `ExpandDetection` | `html` |
| `DecisionDetection` | `emoji` |

## Runtime Hooks (Link, Media, Extensions)

Both directions support optional runtime hooks. Hook fields are runtime-only (`json:"-"`) and are not serialized in config JSON.

### Hook Surfaces

| Direction | Link Hook | Media Hook | Extension Handler |
|---|---|---|---|
| ADF -> Markdown | `LinkHook(ctx, LinkRenderInput) (LinkRenderOutput, error)` | `MediaHook(ctx, MediaRenderInput) (MediaRenderOutput, error)` | `ExtensionHandlers[key].ToMarkdown(ctx, ...)` |
| Markdown -> ADF | `LinkHook(ctx, LinkParseInput) (LinkParseOutput, error)` | `MediaHook(ctx, MediaParseInput) (MediaParseOutput, error)` | `ExtensionHandlers[key].FromMarkdown(ctx, ...)` |

Typed metadata is available in both directions (`PageID`, `SpaceKey`, `AttachmentID`, `Filename`, `Anchor`) plus raw attrs payloads.

### Invocation Ordering

- ADF -> Markdown:
  1. Link marks
  2. `inlineCard`
  3. Media nodes
  4. Extensions (matches by `extensionKey`)
- Markdown -> ADF:
  1. Mention-link detection (`mention:`) first
  2. Link hook for non-mention links
  3. Card heuristics (`inlineCard`) unless forced by hook output
  4. Media hook before `MediaBaseURL` stripping
  5. Extension handler on `:::{ .adf-extension key="..." }`

### Unresolved and Validation Behavior

- `ErrUnresolved` + `ResolutionBestEffort`: warn and fallback.
- `ErrUnresolved` + `ResolutionStrict`: fail conversion.

Handled hook outputs are validated:

1. Forward link output requires non-empty `Href` unless `TextOnly=true`.
2. Forward media output requires non-empty `Markdown`.
3. Reverse link output requires non-empty `Destination`.
4. Reverse link output cannot set both `ForceLink` and `ForceCard`.
5. Reverse media output requires `MediaType` of `image` or `file`.
6. Reverse media output must set exactly one of `ID` or `URL`.
7. Extensions gracefully fall back to default behavior (e.g. fenced JSON block) if handler declines (`Handled: false`).

`ForceCard` currently maps to `inlineCard` output (block-card output is not currently emitted by reverse conversion).

### SourcePath and Context

Use context-aware entrypoints to pass deterministic source location and cancellation/timeouts:

- `converter.ConvertWithContext(ctx, input, converter.ConvertOptions{SourcePath: ...})`
- `mdconverter.ConvertWithContext(ctx, markdown, mdconverter.ConvertOptions{SourcePath: ...})`

## CLI Presets

`jac --preset=...` supports `balanced`, `strict`, `readable`, `lossy`, and `pandoc` in both directions.

| Preset | Forward Intent | Reverse Intent |
|---|---|---|
| `balanced` | library defaults | library defaults |
| `strict` | error on unknown nodes/marks; preserve IDs/extensions | conservative detection set matching round-trip formats |
| `readable` | human-focused markdown (text mentions, text extensions, blockquote expands) | readable pattern set (`@Name`, text status, bold panels, blockquote expands) |
| `lossy` | minimize metadata (`inlineCard` URLs, stripped extensions, text mentions) | disable most semantic detectors (`none`) |
| `pandoc` | Pandoc-flavored Markdown using span/div syntax and grid tables | detect and parse Pandoc syntax back to ADF metadata |

CLI compatibility flags are layered on top of preset output:

- `--allow-html` adjusts HTML-oriented style/detection overrides.
- `--strict` applies strict forward unknown policy and strict reverse detection overrides.

## Result and Warning Model

- Forward returns `converter.Result{Markdown, Warnings}`.
- Reverse returns `mdconverter.Result{ADF, Warnings}`.
- Warnings include categories such as unknown nodes/marks, dropped features, extension fallback, missing attributes, and unresolved references.

## Concurrency Contract

- Converter internals are safe for concurrent calls when using the same converter instance.
- Hook closures are caller-owned and must synchronize shared mutable state when reused across goroutines.
