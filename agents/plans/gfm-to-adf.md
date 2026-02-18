# Plan: GFM to ADF Reverse Converter

## Goal
Build a Go library to convert GitHub Flavored Markdown (GFM) back to Jira Atlassian Document Format (ADF) JSON — the reverse of the existing `converter` package.

## Core Principles
1. **Full Parity**: Support all node types and marks that the forward (ADF→GFM) converter handles.
2. **Config-Driven Ambiguity**: Use configuration hints to resolve patterns that could map to multiple ADF nodes (e.g., `> text` → blockquote vs panel, `@Name` → plain text vs mention).
3. **Standard Parser**: Use `github.com/yuin/goldmark` with GFM extensions — no custom parser.
4. **Mirror Existing API**: `New(config)` + `Convert(markdown)` pattern, matching the forward converter.
5. **Roundtrip Fidelity**: Where the forward converter preserves enough information, converting ADF → MD → ADF should yield the original ADF structure.
6. **Graceful Degradation**: Unrecognized markdown constructs produce warnings, not errors (configurable).

---

## Architecture

### New Package: `mdconverter/`

A separate package at the same level as `converter/`. It imports `converter` for shared `Doc`, `Node`, `Mark` types but adds no reverse dependency.

```
mdconverter/
  mdconverter.go       # Public API: Converter struct, New(), Convert()
  config.go            # ReverseConfig struct, defaults, validation
  walker.go            # goldmark AST walker → ADF builder core
  blocks.go            # Block handlers: paragraph, heading, blockquote, codeBlock, rule, panel, expand, decision
  marks.go             # Mark stack: strong, em, strike, code, link, underline, subsup, colors
  inline.go            # Inline handlers: emoji, mention, status, date, inlineCard, media
  lists.go             # Lists: bullet, ordered, task (with nesting)
  tables.go            # GFM pipe tables + HTML tables → ADF table nodes
  html_parser.go       # Inline/block HTML → ADF (<u>, <sub>, <sup>, <span>, <details>, <table>)
  patterns.go          # Regex-based pattern detectors (emoji, status, dates, mention @Name)
  result.go            # ReverseResult type { ADF []byte, Warnings []converter.Warning }
  *_test.go            # Tests mirroring the forward converter's test structure
```

### Public API

```go
import "github.com/rgonek/jira-adf-converter/mdconverter"

conv, err := mdconverter.New(config)
result, err := conv.Convert(markdownString)
// result.ADF      []byte            — ADF JSON
// result.Warnings []converter.Warning
```

### Dependencies to Add to `go.mod`
- `github.com/yuin/goldmark` — Markdown parser with GFM extension support
- `golang.org/x/net/html` — HTML tokenizer for block-level HTML parsing

---

## Configuration: `ReverseConfig`

Each ambiguous pattern has a dedicated detection mode. Defaults are chosen to match the forward converter's **default output format**.

| Field | Type | Default | Purpose |
|---|---|---|---|
| `MentionDetection` | `MentionDetection` | `link` | Detect `[text](mention:id)` links as mention nodes |
| `EmojiDetection` | `EmojiDetection` | `shortcode` | Detect `:shortcode:` as emoji nodes |
| `StatusDetection` | `StatusDetection` | `bracket` | Detect `[Status: TEXT]` as status nodes |
| `DateDetection` | `DateDetection` | `iso` | Detect `YYYY-MM-DD` as date nodes |
| `PanelDetection` | `PanelDetection` | `github` | Detect `> [!NOTE]` and `> **Info**:` as panel nodes |
| `ExpandDetection` | `ExpandDetection` | `html` | Detect `<details>` as expand nodes |
| `DecisionDetection` | `DecisionDetection` | `emoji` | Detect `> **✓ Decision**:` as decision nodes |
| `DateFormat` | `string` | `2006-01-02` | Go time format for parsing dates to timestamps |
| `HeadingOffset` | `int` | `0` | Reverse heading level adjustment |
| `LanguageMap` | `map[string]string` | nil | Reverse code language aliases (`"cpp"` → `"c++"`) |
| `MediaBaseURL` | `string` | `""` | Strip base URL to extract media IDs |
| `MentionRegistry` | `map[string]string` | nil | Display name → account ID (for `@Name` detection) |
| `EmojiRegistry` | `map[string]string` | nil | Shortcode → emoji ID (for custom Jira emojis) |

---

## goldmark AST → ADF Node Mapping

| goldmark Node | ADF Output | Notes |
|---|---|---|
| `ast.Document` | `doc` | Root container |
| `ast.Paragraph` | `paragraph` | |
| `ast.Heading` | `heading` | `Level` applied with reverse `HeadingOffset` |
| `ast.Blockquote` | `blockquote` / `panel` / `expand` / `decisionList` | **Disambiguated by config + content** |
| `ast.FencedCodeBlock` | `codeBlock` | Language from info string, reverse `LanguageMap` |
| `ast.ThematicBreak` | `rule` | |
| `ast.List` (unordered) | `bulletList` | |
| `ast.List` (ordered) | `orderedList` | `start` attr from `list.Start` |
| `ast.List` + `TaskCheckBox` | `taskList` / `taskItem` | Detected via GFM extension |
| `ast.ListItem` | `listItem` | |
| `ext.Table` | `table` | |
| `ext.TableHeader/Row/Cell` | `tableRow` / `tableHeader` / `tableCell` | |
| `ast.Text` | `text` | Pattern-scanned for emoji, status, date, mention, media |
| `ast.Emphasis(Level=1)` | mark: `em` | |
| `ast.Emphasis(Level=2)` | mark: `strong` | |
| `ext.Strikethrough` | mark: `strike` | |
| `ast.CodeSpan` | mark: `code` | |
| `ast.Link` | mark: `link` / `mention` / `inlineCard` | Checks `mention:` scheme, inlineCard config |
| `ast.Image` | `mediaSingle` + `media` | |
| `ast.RawHTML` | Various marks/nodes | `<u>`, `<sub>`, `<sup>`, `<span>`, `<br>`, `<span data-mention-id>` |
| `ast.HTMLBlock` | Various | `<details>`, `<table>`, `<div align>` |
| `ast.HardLineBreak` | `hardBreak` | |

### Mark Handling Strategy
Marks are tracked as a **stack** during inline traversal. goldmark represents formatting as nested AST nodes; the walker pushes a mark when entering a node and pops it when leaving. Text nodes are emitted with a copy of the current mark stack.

```
Enter Emphasis(2) → push {strong}
  Enter Emphasis(1) → push {em}
    Text "bold italic" → emit text node with marks [{strong}, {em}]
  Leave Emphasis(1) → pop
Leave Emphasis(2) → pop
```

### Blockquote Disambiguation
The most complex ambiguity: the forward converter renders panels, decisions, and expands as blockquotes with special prefixes. Detection order (when `PanelDetection: all`):

1. `> [!NOTE]` or `> [!NOTE: title]` → panel (GitHub / title style)
2. `> **Info**: ...` → panel (bold style)
3. `> **✓ Decision**: ...` or `> **DECIDED**: ...` → decisionList
4. `> **Title**\n> content` (blockquote-style expand) → expand
5. Default → blockquote

---

## Implementation Phases

### Phase 1: Foundation ✅
**Files**: `mdconverter.go`, `config.go`, `walker.go`, `result.go`

**Status**: Implemented (package scaffold, config defaults/validation/clone, parser wiring, document/paragraph/text walker, and initial tests).

**Deliverables**:
- Package structure with public `New()` and `Convert()` API
- `ReverseConfig` with all fields, `applyDefaults()`, `Validate()`, and deep clone
- `ReverseResult` type
- goldmark initialized with `extension.GFM`
- Basic walker skeleton: converts `ast.Document` → `doc`, `ast.Paragraph` → `paragraph`, `ast.Text` → `text`
- `go.mod` updated with goldmark dependency

**Test**: Empty document, single plain-text paragraph

---

### Phase 2: Block Nodes ✅
**File**: `blocks.go`

**Status**: Implemented (heading with offset, blockquote, rule, hardBreak, fenced/indented code blocks with reverse language mapping).

**Deliverables**:
- Paragraph, heading (with level + reverse HeadingOffset), plain blockquote, rule, hardBreak
- Fenced code blocks with language extraction and reverse LanguageMap

**Test**: Roundtrip `testdata/simple/`, `testdata/nodes/`, `testdata/codeblocks/`

---

### Phase 3: Text Marks ✅
**File**: `marks.go`

**Status**: Implemented (mark stack with push/pop/current/popByType and support for strong/em/strike/code/link marks).

**Deliverables**:
- Mark stack with `push`, `pop`, `current`, `popByType`
- `strong`, `em`, `strike`, `code` from goldmark AST
- `link` mark from `ast.Link` (href, optional title)

**Test**: Roundtrip `testdata/marks/` (basic mark tests)

---

### Phase 4: Lists ✅
**File**: `lists.go`

**Status**: Implemented (bullet and ordered lists, task list detection via `TaskCheckBox`, nested list handling including nested task lists).

**Deliverables**:
- Bullet lists, ordered lists (with `start` attribute)
- Task list detection: list items containing `ext.TaskCheckBox` → `taskList` + `taskItem`
- Nested list support

**Test**: Roundtrip `testdata/lists/`

---

### Phase 5: Tables ✅
**File**: `tables.go`

**Status**: Implemented (GFM table parsing with row/header/cell mapping and alignment attributes).

**Deliverables**:
- GFM pipe table → ADF `table`/`tableRow`/`tableHeader`/`tableCell`
- Cell alignment attributes
- Cell content as inline paragraph content

**Test**: Roundtrip `testdata/tables/` (simple GFM tables)

---

### Phase 6: Inline HTML Parsing ✅
**File**: `html_parser.go`

**Status**: Implemented (`<u>`, `<sub>`, `<sup>`, `<span style=...>`, mention spans, and `<br>` handling during inline traversal).

**Deliverables**:
- `<u>` → underline mark
- `<sub>` / `<sup>` → subsup mark
- `<span style="color:...">` → textColor mark
- `<span style="background-color:...">` → backgroundColor mark
- `<span data-mention-id="...">` → mention node (with HTML detection mode)
- `<br>` → hardBreak node

**Test**: Roundtrip `testdata/marks/*_html*`

---

### Phase 7: Inline Pattern Detection ✅
**Files**: `patterns.go`, `inline.go`

**Status**: Implemented (`:shortcode:`, `[Status: ...]`, ISO dates, `@Name` registry mentions, media placeholders, markdown images, and `mention:` links).

**Deliverables**:
- `:shortcode:` → emoji node (EmojiDetection)
- `[Status: TEXT]` → status node (StatusDetection)
- `YYYY-MM-DD` → date node with timestamp (DateDetection)
- `@Name` → mention node via MentionRegistry lookup (MentionDetection: at)
- `[Image: id]` / `[File: id]` → media node placeholders
- `![alt](url)` → mediaSingle + media (image)
- `[text](mention:id)` → mention node (MentionDetection: link)

**Test**: Roundtrip `testdata/inline/`, `testdata/media/`

---

### Phase 8: Blockquote Disambiguation ✅
**In**: `blocks.go`

**Status**: Implemented (panel, decision list, and blockquote-style expand disambiguation with config-gated detection and nested expand context handling).

**Deliverables**:
- Panel detection: github (`> [!NOTE]`), bold (`> **Info**: ...`), title (`> [!NOTE: title]`) styles
- Decision list detection: emoji (`> **✓ Decision**: ...`) and text (`> **DECIDED**: ...`) styles
- Expand detection from blockquote: `> **Title**\n> content`

**Test**: Roundtrip `testdata/panels/`, `testdata/decisions/`, `testdata/expanders/`

---

### Phase 9: Block-Level HTML ✅
**In**: `html_parser.go`

**Status**: Implemented (`<details>` block sequences, `<div align>`, `<h1-6 align>`, and HTML table parsing including colspan/rowspan and cell markdown parsing).

**Deliverables**:
- `<details><summary>Title</summary>content</details>` → expand node
- `<table>` with colspan/rowspan → ADF table with cell attributes
- `<div align="...">` → paragraph with alignment attr
- `<h1-6 align="...">` → heading with alignment

**Test**: Roundtrip HTML-mode golden files (`*_html*` variants for expanders, tables, alignment)

---

### Phase 10: Extensions ✅
**File**: `extensions.go` (in mdconverter)

**Status**: Implemented (`adf:extension` fenced JSON reconstruction plus `adf:inlineCard` fence parsing with inline merge into surrounding paragraph flow).

**Deliverables**:
- Code fences with `adf:extension` info string → reconstruct extension node from JSON body
- InlineCard embed code blocks → inlineCard node reconstruction

**Test**: Roundtrip `testdata/extensions/`

---

### Phase 11: CLI Integration ✅
**File**: `cmd/jac/main.go`

**Status**: Implemented (`--reverse` mode, reverse preset resolution with override precedence, and pretty-printed ADF JSON output).

**Deliverables**:
- `--reverse` flag to select MD→ADF direction
- Reverse config preset mapping (matching forward presets: balanced, strict, readable, lossy)
- Pretty-printed JSON output (`json.MarshalIndent`)

**Test**: CLI integration: `jac --reverse file.md` produces valid ADF JSON

---

### Phase 12: Polish & Edge Cases ✅
**Deliverables**:
- Full roundtrip golden file test suite with categorized lossless vs. lossy expectations
- `reverseGoldenConfigForPath()` mirroring `goldenConfigForPath()` from `converter/converter_test.go`
- Fuzz test: any markdown input → valid ADF JSON (no panics)
- Benchmark tests
- Edge cases: empty docs, whitespace-only, deeply nested content, unknown HTML tags

**Status**: Implemented (reverse golden harness with config mapping, fuzz + benchmark tests, and normalized comparisons for expected lossy metadata differences).

---

## Testing Strategy

| Test Type | What | How |
|---|---|---|
| **Roundtrip golden files** | Read `testdata/**/*.md`, convert to ADF, compare with `testdata/**/*.json` | Structural JSON comparison via `json.Unmarshal` + `assert.Equal` (not byte-for-byte) |
| **Reverse-specific fixtures** | `testdata_reverse/` — markdown not produced by forward converter | New JSON+MD fixture pairs |
| **Unit tests** | Regex patterns, HTML tag parsing, mark stack, config validation | Table-driven tests per file |
| **Fuzz tests** | Any markdown → valid ADF JSON, no panics | `testing.F` with golden file seeds |
| **Lossless vs lossy** | Not all golden files roundtrip perfectly (forward converter drops some info) | Categorize: skip/mark-expected-different for lossy tests |

---

## Critical Files

| File | Action |
|---|---|
| `converter/ast.go` | **Reference** — shared `Doc`, `Node`, `Mark` types imported by mdconverter |
| `converter/result.go` | **Reference** — reuse `Warning`, `WarningType` |
| `converter/inline.go` | **Reference** — exact output patterns the reverse converter must detect |
| `converter/converter_test.go:27` | **Reference** — `goldenConfigForPath` pattern to mirror as `reverseGoldenConfigForPath` |
| `cmd/jac/main.go` | **Modify** — add `--reverse` flag |
| `go.mod` | **Modify** — add goldmark, x/net/html |
| `mdconverter/*` | **Create** — all new package files |

---

## Success Criteria
- [x] `go build ./...` compiles with no errors
- [x] `go test ./mdconverter/...` all tests pass
- [x] `go test ./...` no regressions in forward converter tests
- [x] Golden file roundtrip tests pass for all lossless test cases
- [x] `jac --reverse input.md` produces valid, well-formed ADF JSON
- [x] Fuzz tests run without panics for 30 seconds
- [x] `make lint` passes

---

## Next Steps
See `phase7-detailed.md` for a step-by-step breakdown of Phase 1 (Foundation) implementation tasks.
