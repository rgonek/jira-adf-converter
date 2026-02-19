# Pandoc Markdown Strategy Support

## Overview

The library already uses HTML as a metadata-preserving strategy for lossless ADF↔Markdown round-trips (e.g. `<u>text</u>` for underline, `<details>` for expand). This plan adds Pandoc-flavored Markdown as a parallel strategy family. Every existing `*HTML` constant gets a matching `*Pandoc` constant; HTML strategies are never removed.

The implementation is fully bidirectional: new forward strategies emit Pandoc syntax, new reverse detection options parse it back to ADF. The default values for all new config fields are backward-compatible, so existing integrations require no changes.

**Pandoc syntax chosen:**

| ADF Feature | Pandoc Syntax |
|---|---|
| Underline | `[text]{.underline}` |
| Subscript | `~text~` |
| Superscript | `^text^` |
| Text color | `[text]{style="color: #rrggbb;"}` |
| Background color | `[text]{style="background-color: #rrggbb;"}` |
| Mention | `[Display Name]{.mention mention-id="accountId"}` |
| Alignment (Block) | `:::{ style="text-align: center;" }\n\ncontent\n\n:::` |
| Alignment (Heading) | `## Heading {style="text-align: center;"}` |
| Expand | `:::{ .details summary="Title" }\n\ncontent\n\n:::` |
| InlineCard | `[title]{.inline-card url="https://..."}` |
| Simple grid table | `+---+---+` with `+===+===+` header separator |

Note: `HardBreakBackslash` already produces valid Pandoc output — no new strategy is needed.

## Documentation References

- **Underline**: [Pandoc Manual - Bracketed Spans](https://pandoc.org/MANUAL.html#extension-bracketed_spans) (standard approach via classes)
- **Subscript/Superscript**: [Pandoc Manual - Superscripts and Subscripts](https://pandoc.org/MANUAL.html#superscripts-and-subscripts)
- **Attributes (Color, Alignment, etc.)**: [Pandoc Manual - Extension: attributes](https://pandoc.org/MANUAL.html#extension-attributes)
- **Fenced Divs**: [Pandoc Manual - Extension: fenced_divs](https://pandoc.org/MANUAL.html#extension-fenced_divs)
- **Bracketed Spans**: [Pandoc Manual - Extension: bracketed_spans](https://pandoc.org/MANUAL.html#extension-bracketed_spans)
- **Grid Tables**: [Pandoc Manual - Extension: grid_tables](https://pandoc.org/MANUAL.html#extension-grid_tables)

## Deliverables

1. New `*Pandoc` forward strategy constants in `converter/config.go` with full validation
2. Forward rendering of all Pandoc syntax variants
3. New `*Pandoc` reverse detection types in `mdconverter/config.go` with full validation
4. Four new goldmark extensions: Pandoc subscript/superscript, span `[text]{attrs}`, fenced div `:::`, and grid table `+---+`
5. Reverse ADF conversion of all Pandoc AST nodes
6. `pandoc` CLI preset in `cmd/jac/main.go`
7. Comprehensive golden test coverage for all new features (both directions)

---

## Testing Approach

All new features follow the established golden-file TDD pattern:

- **Forward tests**: Create `.json` input + `.md` expected output in `testdata/`; the test runner in `converter/converter_test.go` picks them up automatically via `goldenConfigForPath()`.
- **Reverse tests**: Create `.md` input + `.json` expected output; picked up by `mdconverter/golden_test.go` via `reverseGoldenConfigForPath()`.
- **Round-trip tests**: A dedicated test verifies that forward→reverse recovers semantically equivalent ADF for each Pandoc feature.
- **Unit tests**: goldmark extension internals (parser edge cases, attribute parsing) use standard Go table-driven unit tests alongside the golden files.

Test names must be descriptive and reflect the feature being tested, not an implementation detail or phase number. Examples of acceptable names: `underline_pandoc`, `mention_with_account_id_pandoc`, `expand_with_title_pandoc`, `table_with_header_pandoc`. Examples of unacceptable names: `test_step3`, `pandoc_feature_7`, `new_thing`.

`goldenConfigForPath()` and `reverseGoldenConfigForPath()` should recognize the `_pandoc` filename suffix and apply the corresponding Pandoc config.

---

## Step-by-Step Implementation Plan

---

### Task 1: Forward Config Constants

**Goal**: Add all new Pandoc strategy constants and update validation. Converter compiles; no behaviour changes yet.

**File**: `converter/config.go`

**New constants** (append to existing types):

```go
UnderlinePandoc  UnderlineStyle  = "pandoc"    // [text]{.underline}
SubSupPandoc     SubSupStyle     = "pandoc"    // ~text~ or ^text^
ColorPandoc      ColorStyle      = "pandoc"    // [text]{color="..."} / {background-color="..."}
MentionPandoc    MentionStyle    = "pandoc"    // [Name]{.mention mention-id="..."}
AlignPandoc      AlignmentStyle  = "pandoc"    // :::{ align="..." }...:::
ExpandPandoc     ExpandStyle     = "pandoc"    // :::{ .details summary="..." }...:::
InlineCardPandoc InlineCardStyle = "pandoc"    // [title]{.inline-card url="..."}
TablePandoc      TableMode       = "pandoc"    // always Pandoc grid table
TableAutoPandoc  TableMode       = "autopandoc"// pipe for simple, grid for complex
```

**`Validate()` changes**: extend each validation block to accept the new constant alongside existing ones. No default changes — all Pandoc options are opt-in.

**Acceptance Criteria**:
- `go build ./...` succeeds
- `Validate()` returns nil for each new constant
- `Validate()` returns an error for an invalid string (regression check)

**Test cases to create** (unit tests in `converter/config_test.go` or existing test file):
- `valid pandoc underline style` — `Config{UnderlineStyle: UnderlinePandoc}` validates without error
- `valid pandoc table modes` — `TablePandoc` and `TableAutoPandoc` both validate without error
- `invalid style string rejected` — `Config{UnderlineStyle: "invalid"}` returns a validation error

---

### Task 2: Forward Rendering — Inline Marks

**Goal**: Emit Pandoc span syntax for underline, subscript, superscript, and color marks.

**File**: `converter/marks.go` — `convertMarkFull()`

**Implementation Details**:

- **`UnderlinePandoc`**: return prefix `"["`, suffix `"]{.underline}"`, nil. This fits the existing `(prefix, suffix, error)` return signature.
- **`SubSupPandoc`**: return `"~"`, `"~"` for sub; `"^"`, `"^"` for sup.
- **`ColorPandoc` (textColor)**: return `"["`, `"]{style=\"color: " + color + ";\"}"`, nil.
- **`ColorPandoc` (backgroundColor)**: return `"["`, `"]{style=\"background-color: " + color + ";\"}"`, nil.

**Acceptance Criteria**:
- All Pandoc mark cases produce correct output
- Existing HTML mark cases are unaffected (no regressions)
- Mixed marks work correctly (e.g. underline+bold: `**[text]{.underline}**`)

**Test golden files to create** in `testdata/marks/`:

1. **`underline_pandoc`** — single underlined word
   - Input: paragraph with text node carrying `underline` mark
   - Expected: `[underlined text]{.underline}`

2. **`underline_and_bold_pandoc`** — underline combined with bold
   - Input: text with `strong` + `underline` marks
   - Expected: `**[bold underline]{.underline}**`

3. **`subscript_pandoc`** — subscript text
   - Input: text node with `subsup` mark, type `sub`
   - Expected: `~H2O~` (subscript characters)

4. **`superscript_pandoc`** — superscript text
   - Input: text node with `subsup` mark, type `sup`
   - Expected: `x^2^`

5. **`text_color_pandoc`** — text with hex color
   - Input: text with `textColor` mark, color `#ff0000`
   - Expected: `[red text]{color="#ff0000"}`

6. **`background_color_pandoc`** — text with background color
   - Input: text with `backgroundColor` mark, color `#ffff00`
   - Expected: `[highlighted]{background-color="#ffff00"}`

7. **`mixed_marks_pandoc`** — underline + color combined
   - Input: text with both `underline` and `textColor` marks
   - Expected: `[[colored underline]{.underline}]{color="#0000ff"}`

---

### Task 3: Forward Rendering — Inline Nodes

**Goal**: Emit Pandoc span syntax for mention and inline card nodes.

**File**: `converter/inline.go`

**Implementation Details**:

- **`MentionPandoc`** in `convertMention()`:
  ```go
  return fmt.Sprintf(`[%s]{.mention mention-id="%s"}`, mentionText, html.EscapeString(id)), nil
  ```
  If `id` is empty, fall back with a warning (same pattern as `MentionHTML`).

- **`InlineCardPandoc`** in `convertInlineCard()`:
  ```go
  displayTitle := title
  if displayTitle == "" { displayTitle = url }
  return fmt.Sprintf(`[%s]{.inline-card url="%s"}`, displayTitle, url), nil
  ```
  If `url` is empty, emit a warning and fall back to text.

**Acceptance Criteria**:
- Mention node with an account ID produces a valid Pandoc span
- Inline card with a URL and display title produces a valid Pandoc span
- Empty ID/URL cases emit a warning and produce reasonable fallback output
- Existing mention/card strategies are unaffected

**Test golden files to create** in `testdata/inline/`:

1. **`mention_with_account_id_pandoc`** — user mention with full metadata
   - Input: `mention` node with `id` = `"abc123"` and `text` = `"Alice"`
   - Expected: `[Alice]{.mention mention-id="abc123"}`

2. **`mention_display_text_only_pandoc`** — mention with `@` prefix preserved
   - Input: `mention` node where the display text includes `@`
   - Expected: `[@Bob]{.mention mention-id="def456"}`

3. **`inline_card_with_title_pandoc`** — smart link with display title
   - Input: `inlineCard` node with `url` and a title attribute
   - Expected: `[My Page]{.inline-card url="https://example.atlassian.net/wiki/..."}`

4. **`inline_card_url_only_pandoc`** — smart link with no display title
   - Input: `inlineCard` node with only a URL
   - Expected: `[https://example.com]{.inline-card url="https://example.com"}`

---

### Task 4: Forward Rendering — Block Nodes

**Goal**: Emit Pandoc fenced-div syntax for alignment and expand nodes.

**File**: `converter/blocks.go`

**Implementation Details**:

- **`AlignPandoc`** in `convertParagraph()`:
  ```go
  return fmt.Sprintf(":::{ style=\"text-align: %s;\" }\n\n%s\n\n:::\n\n", alignment, trimmedContent), nil
  ```
- **`AlignPandoc`** in `convertHeading()`:
  ```go
  return fmt.Sprintf("%s {style=\"text-align: %s;\"}\n\n", heading, alignment), nil
  ```
  Only emitted when an alignment attribute is present; falls through to plain rendering when no alignment is set.

- **`ExpandPandoc`** in `convertExpand()`:
  ```go
  var sb strings.Builder
  sb.WriteString(":::{ .details")
  if title != "" {
      sb.WriteString(fmt.Sprintf(` summary="%s"`, strings.ReplaceAll(title, `"`, `\"`)))
  }
  sb.WriteString(" }\n\n")
  sb.WriteString(strings.TrimRight(content, "\n"))
  sb.WriteString("\n\n:::\n\n")
  return sb.String(), nil
  ```

**Acceptance Criteria**:
- Center-aligned paragraph wraps content in `:::{ align="center" }...:::`
- Expand node with a title produces `:::{ .details summary="Title" }...:::`
- Expand node without a title produces `:::{ .details }...:::`
- `nestedExpand` inside another expand uses the same syntax
- Existing HTML and blockquote strategies are unaffected

**Test golden files to create** in `testdata/` (use existing subdirectory structure):

1. **`paragraph_aligned_center_pandoc`** (in `testdata/alignment/` or `testdata/nodes/`)
   - Input: paragraph with `layout: center` attribute
   - Expected: `:::{ align="center" }\n\nCentered text\n\n:::`

2. **`heading_aligned_right_pandoc`**
   - Input: heading H2 with `layout: right` attribute
   - Expected: `:::{ align="right" }\n\n## Heading\n\n:::`

3. **`expand_with_title_pandoc`** (in `testdata/expanders/`)
   - Input: `expand` node, title `"Click to expand"`, body with one paragraph
   - Expected: `:::{ .details summary="Click to expand" }\n\nbody\n\n:::`

4. **`expand_without_title_pandoc`**
   - Input: `expand` node with no title attribute
   - Expected: `:::{ .details }\n\nbody\n\n:::`

5. **`nested_expand_pandoc`**
   - Input: outer `expand` containing inner `nestedExpand`
   - Expected: nested `:::` blocks

6. **`expand_title_with_quotes_pandoc`**
   - Input: `expand` with a title containing double-quote characters
   - Expected: title attribute with escaped quotes

---

### Task 5: Forward Rendering — Tables

**Goal**: Add `TablePandoc` (always grid) and `TableAutoPandoc` (pipe for simple, grid for complex) table modes.

**File**: `converter/tables.go`

**Implementation Details**:

Extend `convertTable()`:
```go
if mode == TableAutoPandoc {
    if s.isComplexTable(node) { mode = TablePandoc } else { mode = TablePipe }
}
```

Add new `renderTableGrid(node Node) (string, error)` function:
- Reuse `extractTableRows()` to get `[][]string`
- Compute column widths: `max(len(header[i]), max over data rows len(row[i]))`
- Render rows with `+---+---+` borders; use `+===+===+` separator after the header row
- Pad cell content with trailing spaces to the column width
- For tables detected as complex (colspan/rowspan): emit a `Warning` and fall back to `renderTableHTML()`

Grid table format example:
```
+------+--------+
| Name | Value  |
+======+========+
| foo  | 1      |
+------+--------+
| bar  | 2      |
+------+--------+
```

**Acceptance Criteria**:
- `TablePandoc` always produces grid format, including simple tables
- `TableAutoPandoc` produces pipe format for simple tables (no regression vs `TableAuto`)
- `TableAutoPandoc` produces grid format for tables with complex block content in cells
- Complex tables with colspan/rowspan fall back to HTML with a warning
- Existing `TableAuto`, `TablePipe`, and `TableHTML` modes are unaffected
- Column widths accommodate the widest cell in each column
- Header separator uses `=` characters

**Test golden files to create** in `testdata/tables/`:

1. **`simple_table_pandoc`** — basic table, `TablePandoc` mode
   - Input: 2×2 table with header row, simple text cells
   - Expected: grid table with `+===+` separator

2. **`simple_table_autopandoc`** — simple table, `TableAutoPandoc` mode
   - Input: same as above
   - Expected: pipe table (same as `TablePipe` output — no regression)

3. **`table_with_wide_cells_pandoc`** — columns sized to widest cell
   - Input: table where data cells are wider than header cells
   - Expected: columns wide enough to fit data rows

4. **`complex_table_autopandoc_fallback`** — complex table falls back to HTML
   - Input: table with colspan attribute
   - Expected: HTML table output (same as `TableHTML`) plus a warning

---

### Task 6: Reverse Config & Detection Helpers

**Goal**: Add all new Pandoc detection types to `mdconverter/config.go`, add them to `ReverseConfig`, and implement `should*()` helpers. No behaviour changes yet — defaults match current implicit HTML parsing.

**Files**: `mdconverter/config.go`, `mdconverter/patterns.go`

**New detection types** to add in `mdconverter/config.go`:

```go
type UnderlineDetection string
const (
    UnderlineDetectNone   UnderlineDetection = "none"
    UnderlineDetectHTML   UnderlineDetection = "html"
    UnderlineDetectPandoc UnderlineDetection = "pandoc"
    UnderlineDetectAll    UnderlineDetection = "all"
)

type SubSupDetection string
const (
    SubSupDetectNone   SubSupDetection = "none"
    SubSupDetectHTML   SubSupDetection = "html"
    SubSupDetectPandoc SubSupDetection = "pandoc"
    SubSupDetectAll    SubSupDetection = "all"
)

type ColorDetection string
const (
    ColorDetectNone   ColorDetection = "none"
    ColorDetectHTML   ColorDetection = "html"
    ColorDetectPandoc ColorDetection = "pandoc"
    ColorDetectAll    ColorDetection = "all"
)

type AlignmentDetection string
const (
    AlignDetectNone   AlignmentDetection = "none"
    AlignDetectHTML   AlignmentDetection = "html"
    AlignDetectPandoc AlignmentDetection = "pandoc"
    AlignDetectAll    AlignmentDetection = "all"
)

type InlineCardDetection string
const (
    InlineCardDetectNone   InlineCardDetection = "none"
    InlineCardDetectLink   InlineCardDetection = "link"
    InlineCardDetectPandoc InlineCardDetection = "pandoc"
    InlineCardDetectAll    InlineCardDetection = "all"
)
```

Add `pandoc` values to existing types:
```go
MentionDetectPandoc MentionDetection = "pandoc"
ExpandDetectPandoc  ExpandDetection  = "pandoc"
```

Update the semantics of `MentionDetectAll` and `ExpandDetectAll` to include the new pandoc option.

**`ReverseConfig` struct additions**:
```go
UnderlineDetection  UnderlineDetection  `json:"underlineDetection,omitempty"`
SubSupDetection     SubSupDetection     `json:"subSupDetection,omitempty"`
ColorDetection      ColorDetection      `json:"colorDetection,omitempty"`
AlignmentDetection  AlignmentDetection  `json:"alignmentDetection,omitempty"`
InlineCardDetection InlineCardDetection `json:"inlineCardDetection,omitempty"`
```

**`applyDefaults()` additions** — backward-compatible (matching current implicit behaviour):
```
UnderlineDetection  → "html"
SubSupDetection     → "html"
ColorDetection      → "html"
AlignmentDetection  → "html"
InlineCardDetection → "none"   (pandoc inline-card is new; default off)
```

**`should*()` helpers** in `mdconverter/patterns.go`:
```go
func (s *state) shouldDetectUnderlineHTML()   bool
func (s *state) shouldDetectUnderlinePandoc() bool
func (s *state) shouldDetectSubSupHTML()      bool
func (s *state) shouldDetectSubSupPandoc()    bool
func (s *state) shouldDetectColorHTML()       bool
func (s *state) shouldDetectColorPandoc()     bool
func (s *state) shouldDetectAlignHTML()       bool
func (s *state) shouldDetectAlignPandoc()     bool
func (s *state) shouldDetectMentionPandoc()   bool
func (s *state) shouldDetectExpandPandoc()    bool
func (s *state) shouldDetectInlineCardPandoc() bool
```

**Guard `mdconverter/html_parser.go`** with the new detection flags:
- Wrap `<u>` detection with `shouldDetectUnderlineHTML()`
- Wrap `<sub>`/`<sup>` detection with `shouldDetectSubSupHTML()`
- Wrap `<span style="color:...">` detection with `shouldDetectColorHTML()`
- Wrap `<span data-mention-id="...">` detection with `shouldDetectMentionHTML()` (already guarded)
- `<br>` remains unconditional
- Wrap `<div align="...">` detection with `shouldDetectAlignHTML()`

Since defaults are `"html"`, all existing golden tests pass unchanged after this step.

**Acceptance Criteria**:
- `go build ./...` succeeds
- All existing golden tests pass without modification
- New detection types validate correctly
- Invalid detection values return validation errors
- `should*HTML()` helpers return true under the default config

**Test cases**:
- Unit tests for each new detection type's `Validate()` (valid values accepted, invalid rejected)
- Existing golden test suite passes unchanged (backward compatibility check)

---

### Task 7: Goldmark Extensions — Subscript and Superscript

**Goal**: Parse `~text~` and `^text^` back to ADF `subsup` marks.

**New file**: `mdconverter/pandoc_subsup_parser.go`

**Implementation Details**:

Two `goldmark/parser.InlineParser` implementations:

**SubscriptParser** (trigger: `~`):
- `Trigger()` returns `[]byte{'~'}`
- In `Parse()`: check that the opener `~` is NOT followed by another `~` (to avoid conflicting with GFM `~~strikethrough~~`). If the next byte is `~`, return false
- Scan forward until the closing `~` (single); produce a `SubscriptNode` wrapping the inline content

**SuperscriptParser** (trigger: `^`):
- `Trigger()` returns `[]byte{'^'}`
- In `Parse()`: scan forward until closing `^`; produce a `SuperscriptNode` wrapping the inline content

Both produce custom AST nodes (`SubscriptNode`, `SuperscriptNode`) extending `ast.BaseInline` with child nodes.

Registration in `mdconverter/mdconverter.go` `New()`:
```go
if cfg.needsPandocInlineExtension() {
    // Register SubscriptParser and SuperscriptParser via parser.WithInlineParsers(...)
}
```

Where `needsPandocInlineExtension()` returns true when any of `SubSupDetection`, `UnderlineDetection`, `ColorDetection`, `MentionDetection`, or `InlineCardDetection` is set to `pandoc` or `all`.

Conversion in `mdconverter/inline.go` (or a new `mdconverter/pandoc_subsup_convert.go`):
```go
case *pandocparser.SubscriptNode:
    if s.shouldDetectSubSupPandoc() {
        stack.push(converter.Mark{Type: "subsup", Attrs: map[string]interface{}{"type": "sub"}})
        // convert children
        stack.popByType("subsup")
    }
    // ...
```

**Acceptance Criteria**:
- `~H2O~` is parsed to a subscript ADF node
- `^2^` is parsed to a superscript ADF node
- `~~strikethrough~~` is still parsed as GFM strikethrough (not a subscript conflict)
- When `SubSupDetection = "none"`, `~text~` passes through as plain text
- When `SubSupDetection = "html"`, HTML `<sub>` still works and `~text~` passes through
- When `SubSupDetection = "all"`, both HTML and Pandoc syntax are detected

**Test golden files to create** in `testdata/marks/` (reverse direction):

1. **`subscript_pandoc`** (reverse) — `~H₂O~` → subscript mark
2. **`superscript_pandoc`** (reverse) — `x^2^` → superscript mark
3. **`subsup_not_strikethrough_pandoc`** — `~~strike~~` → strikethrough, not subsup
4. **`subsup_pandoc_disabled`** — with `SubSupDetection = "none"`, tildes pass through as text

Unit tests (in `mdconverter/pandoc_subsup_parser_test.go`):
- Tilde at start of word, mid-word, end-of-line
- Unmatched `~` (no closing tilde) — must not panic, falls through as text
- Nested marks: `~**bold sub**~` → subscript wrapping bold

---

### Task 8: Goldmark Extension — Pandoc Span `[text]{attrs}`

**Goal**: Parse Pandoc span syntax to recover underline, color, mention, and inline-card ADF nodes.

**New file**: `mdconverter/pandoc_span_parser.go`

**Implementation Details**:

One `InlineParser` (trigger: `[`):
- Priority: lower than goldmark's link parser (priority 79; link parser uses priority 80)
- In `Parse()`:
  1. Scan from `[` through matching `]` (balanced bracket counting)
  2. After `]`, check if next character is `{`; if next is `(` or `[`, return false (let link parser handle it)
  3. If `{`, parse the attribute block `{...}` until `}`
  4. Produce a `PandocSpanNode` with `Classes []string`, `Attrs map[string]string`, and inline children re-parsed from the bracketed content

**Attribute block grammar** (subset sufficient for our use cases):
- `.classname` → add to Classes
- `key="value"` → add to Attrs
- `key='value'` → add to Attrs (single quotes)
- Multiple entries separated by whitespace

**New file**: `mdconverter/pandoc_span_convert.go`

`convertPandocSpanNode()` dispatches by class and attributes:

```
.underline                → underline mark (if shouldDetectUnderlinePandoc())
.mention + mention-id=... → mention node (if shouldDetectMentionPandoc())
.inline-card + url=...    → inlineCard node (if shouldDetectInlineCardPandoc())
color=...                 → textColor mark (if shouldDetectColorPandoc())
background-color=...      → backgroundColor mark (if shouldDetectColorPandoc())
unknown class/attrs       → plain text with warning (best effort)
```

**Acceptance Criteria**:
- `[text]{.underline}` → underline mark
- `[Name]{.mention mention-id="abc123"}` → mention node
- `[Title]{.inline-card url="https://..."}` → inlineCard node
- `[text]{color="#ff0000"}` → textColor mark
- `[text]{background-color="#ff0000"}` → backgroundColor mark
- `[link text](url)` is still parsed as a link (no conflict)
- `[ref link][id]` is still parsed as a reference link (no conflict)
- Unknown span classes produce a warning and plain text
- When detection is disabled for a type, the span falls through as plain text

**Test golden files to create** (reverse direction):

1. **`underline_from_pandoc_span`** — `[word]{.underline}` → underline mark
2. **`mention_from_pandoc_span`** — `[Alice]{.mention mention-id="abc"}` → mention node
3. **`inline_card_from_pandoc_span`** — `[My Page]{.inline-card url="https://..."}` → inlineCard node
4. **`text_color_from_pandoc_span`** — `[red]{color="#ff0000"}` → textColor mark
5. **`background_color_from_pandoc_span`** — `[highlight]{background-color="#ffff00"}` → backgroundColor mark
6. **`pandoc_span_adjacent_to_link`** — `[link](url) and [span]{.underline}` → no conflict between link and span
7. **`pandoc_span_detection_disabled`** — span syntax passes through as literal text when detection is off

Unit tests (in `mdconverter/pandoc_span_parser_test.go`):
- Attribute parsing: `.class`, `key="val"`, `key='val'`, multiple attributes
- Balanced bracket handling inside content: `[a [b] c]{.underline}`
- Empty content: `[]{.underline}`
- Unclosed brace: `[text]{.underline` → falls through as text (no panic)
- Priority test: `[text](url)` triggers link parser, not span parser

---

### Task 9: Goldmark Extension — Pandoc Fenced Div `:::{ }`

**Goal**: Parse `:::{ .details summary="..." }...:::` and `:::{ align="center" }...:::` blocks back to ADF expand and alignment nodes.

**New file**: `mdconverter/pandoc_div_parser.go`

**Implementation Details**:

One `BlockParser` (trigger: `:`):
- `Trigger()` returns `[]byte{':'}`
- `Open()`:
  1. Check line starts with `:::` (3+ colons)
  2. After the colons, skip optional whitespace; if next char is `{`, parse the attribute block
  3. Return `parser.HasChildren` to let goldmark feed block children to this node
- `Continue()`: consume child blocks; stop when a closing `:::` line is encountered
- `Close()`: finalize the `PandocDivNode` with `RawAttrs string` and `FenceLength int`

**New file**: `mdconverter/pandoc_div_convert.go`

`convertPandocDivNode()` dispatches by parsed attributes:

```go
func parsePandocAttrString(raw string) (classes []string, attrs map[string]string)
// Parses "{ .details summary=\"Title\" align=\"center\" }" into (["details"], {"summary":"Title"})

func (s *state) convertPandocDivNode(node *pandocparser.DivNode) (converter.Node, bool, error)
```

Dispatch logic:
- Class `.details` present and `shouldDetectExpandPandoc()` → expand or nestedExpand node with `summary` attr as title
  - Use `expand` at top level, `nestedExpand` when inside another expand (check parent context)
- Attribute `align=...` present and `shouldDetectAlignPandoc()` → apply `layout` attribute to each immediate child block node
  - If multiple children, apply alignment to each individually (ADF alignment is per-block)
- Unknown div class → blockquote fallback with a warning

**Acceptance Criteria**:
- `:::{ .details summary="Title" }...:::` → `expand` node with title
- `:::{ .details }...:::` → `expand` node with no title
- Nested `:::` → `nestedExpand` inside `expand`
- `:::{ align="center" }...:::` → paragraph with `layout: center`
- `:::{ align="center" }` containing a heading → heading with `layout: center`
- Unknown div class → blockquote with warning
- When `ExpandDetection = "html"`, fenced divs are ignored
- Closing `:::` with more colons than the opening is rejected (mismatched fence)
- Closing `:::` with fewer or equal colons terminates the div correctly

**Test golden files to create** (reverse direction):

1. **`expand_from_pandoc_div`** — `:::{ .details summary="Title" }` → expand node
2. **`expand_no_title_from_pandoc_div`** — `:::{ .details }` → expand node without title
3. **`nested_expand_from_pandoc_divs`** — nested `:::` blocks → expand containing nestedExpand
4. **`center_alignment_from_pandoc_div`** — `:::{ align="center" }` → aligned paragraph
5. **`right_alignment_from_pandoc_div`** — `:::{ align="right" }` → aligned heading
6. **`unknown_div_class_fallback`** — `:::{ .custom-class }` → blockquote with warning
7. **`pandoc_div_detection_disabled`** — with `ExpandDetection = "none"`, div passes through as text

Unit tests (in `mdconverter/pandoc_div_parser_test.go`):
- Attribute block parsing (all relevant attr formats)
- Div containing multiple block types (paragraphs, lists, code blocks)
- Empty div body
- Mismatched fence lengths

---

### Task 10: Goldmark Extension — Pandoc Grid Table

**Goal**: Parse `+---+---+` grid tables back to ADF table nodes.

**New file**: `mdconverter/pandoc_table_parser.go`

**Implementation Details**:

One `BlockParser` (trigger: `+`):
- `Trigger()` returns `[]byte{'+'}`
- `Open()`:
  1. Check line matches grid border pattern: `+[-=]+[+[-=]+]*+`
  2. Parse column widths from the positions of `+` separators
  3. Record border style (`-` = data separator, `=` = header separator)
- `Continue()`:
  - Lines starting with `|` are cell content rows; accumulate per-column cell strings (multi-line cells: append within same cell)
  - Lines starting with `+` are separators; if `+=====+` encountered, mark following rows as data rows
  - Stop at a line not starting with `|` or `+` after at least one row

  Produce a `PandocGridTableNode` with:
  - `Columns int`
  - `HeaderRows [][]string` — cell content as raw string (may contain inline markdown)
  - `DataRows [][]string`

Conversion (`convertPandocGridTableNode()`):
- For each cell string, re-parse inline content using the inline parser chain
- Map header rows to `tableRow` nodes containing `tableHeader` cells
- Map data rows to `tableRow` nodes containing `tableCell` cells
- Colspan/rowspan detection: if a cell in the source contains a blank region matching a multi-column span pattern, emit a warning and produce a flat cell instead

**Enabling the grid table parser**: Registered when a new `ReverseConfig` field `TableGridDetection bool` is `true`, or when `TableAutoPandoc`-related strategy is detected. Simplest approach: add `TableGridDetection bool` to `ReverseConfig` (default `false` for backward compatibility).

**Acceptance Criteria**:
- Simple `+---+---+` grid table with one header row and two data rows → ADF table
- `+===+===+` separator correctly distinguishes header from data rows
- Multi-line cell content (two text lines in same cell) is concatenated with a space
- Empty cell produces an empty paragraph
- Grid table adjacent to normal text does not affect surrounding content
- Non-grid-table lines starting with `+` (e.g. arithmetic) are not consumed
- When `TableGridDetection = false`, grid syntax passes through as plain text

**Test golden files to create** (reverse direction):

1. **`grid_table_with_header`** — standard grid table with header row
2. **`grid_table_no_header_separator`** — grid table using only `-` separators (all rows are data rows)
3. **`grid_table_multiline_cell`** — cell with two lines of content
4. **`grid_table_round_trip`** — ADF table → `TablePandoc` forward → grid table → reverse → original ADF

Unit tests (in `mdconverter/pandoc_table_parser_test.go`):
- Column width extraction from border row
- `=` vs `-` separator discrimination
- Multi-line cell assembly
- Misaligned `|` characters

---

### Task 11: Extension Registration in `mdconverter/mdconverter.go`

**Goal**: Conditionally add Pandoc goldmark extensions to the parser at `New()` time.

**File**: `mdconverter/mdconverter.go`

**Implementation Details**:

Add helper methods to `ReverseConfig`:

```go
func (c ReverseConfig) needsPandocInlineExtension() bool {
    // true when any of UnderlineDetection, SubSupDetection, ColorDetection,
    // MentionDetection, or InlineCardDetection is set to pandoc or all
}

func (c ReverseConfig) needsPandocBlockExtension() bool {
    // true when ExpandDetection or AlignmentDetection is set to pandoc or all
}
```

Modify `New()`:

```go
options := []goldmark.Option{goldmark.WithExtensions(extension.GFM)}

if cfg.needsPandocInlineExtension() {
    options = append(options, goldmark.WithParserOptions(
        parser.WithInlineParsers(
            util.Prioritized(NewSubscriptParser(), 79),
            util.Prioritized(NewSuperscriptParser(), 79),
            util.Prioritized(NewPandocSpanParser(), 79),
        ),
    ))
}
if cfg.needsPandocBlockExtension() {
    options = append(options, goldmark.WithParserOptions(
        parser.WithBlockParsers(
            util.Prioritized(NewPandocDivParser(), 500),
        ),
    ))
}
if cfg.TableGridDetection {
    options = append(options, goldmark.WithParserOptions(
        parser.WithBlockParsers(
            util.Prioritized(NewPandocGridTableParser(), 501),
        ),
    ))
}

return &Converter{config: cfg, parser: goldmark.New(options...)}, nil
```

**Acceptance Criteria**:
- Default config (`needsPandocInlineExtension() = false`) creates a parser identical to the current one
- Pandoc extensions are not registered when not needed (no performance overhead)
- Enabling Pandoc inline extension does not break existing link parsing
- Enabling multiple Pandoc extensions simultaneously works correctly
- Adding `go.sum` / `go.mod` entries if new goldmark sub-packages are needed

---

### Task 12: CLI Preset

**Goal**: Add a `pandoc` preset that configures both forward and reverse converters with all Pandoc strategies.

**File**: `cmd/jac/main.go`

**Implementation Details**:

Add `"pandoc"` as a valid preset value. Forward config:
```go
converter.Config{
    UnderlineStyle:       converter.UnderlinePandoc,
    SubSupStyle:          converter.SubSupPandoc,
    TextColorStyle:       converter.ColorPandoc,
    BackgroundColorStyle: converter.ColorPandoc,
    MentionStyle:         converter.MentionPandoc,
    AlignmentStyle:       converter.AlignPandoc,
    ExpandStyle:          converter.ExpandPandoc,
    InlineCardStyle:      converter.InlineCardPandoc,
    TableMode:            converter.TableAutoPandoc,
}
```

Reverse config:
```go
mdconverter.ReverseConfig{
    UnderlineDetection:  mdconverter.UnderlineDetectPandoc,
    SubSupDetection:     mdconverter.SubSupDetectPandoc,
    ColorDetection:      mdconverter.ColorDetectPandoc,
    AlignmentDetection:  mdconverter.AlignDetectPandoc,
    MentionDetection:    mdconverter.MentionDetectPandoc,
    ExpandDetection:     mdconverter.ExpandDetectPandoc,
    InlineCardDetection: mdconverter.InlineCardDetectPandoc,
    TableGridDetection:  true,
}
```

Update help text and preset validation to include `"pandoc"`.

**Acceptance Criteria**:
- `jac --preset=pandoc input.adf.json` produces Pandoc-flavored markdown
- `jac --reverse --preset=pandoc input.md` parses Pandoc syntax back to ADF
- `jac --preset=invalid` returns an error listing `pandoc` as a valid option
- Existing presets (`balanced`, `strict`, `readable`, `lossy`) are unaffected

---

### Task 13: Round-trip Tests

**Goal**: Verify that each Pandoc feature survives a full forward→reverse round-trip without data loss.

**New file**: `converter/pandoc_roundtrip_test.go` (or add to existing `converter_test.go`)

**Test cases** (one per Pandoc feature):
1. Underline mark round-trip
2. Subscript and superscript marks round-trip
3. Text color and background color marks round-trip
4. Mention node round-trip (account ID preserved)
5. Inline card node round-trip (URL and title preserved)
6. Paragraph alignment round-trip
7. Expand with title round-trip
8. Expand without title round-trip
9. Nested expand round-trip
10. Simple table round-trip via `TablePandoc`
11. Complex table round-trip (expect warning + HTML fallback)
12. Combined: paragraph with underline + color + mention in same block

**Round-trip test helper pattern**:
```go
func testPandocRoundtrip(t *testing.T, adfJSON string, fwdConfig converter.Config, revConfig mdconverter.ReverseConfig) {
    // 1. Forward convert
    // 2. Reverse convert
    // 3. Normalize both ADF JSONs (strip localId, generated IDs, etc.)
    // 4. Assert deep equality
}
```

**Acceptance Criteria**:
- All round-trip tests pass
- No data loss for any supported Pandoc feature
- Warnings produced by unsupported features are stable and documented

---

### Task 14: Final Verification

**Goal**: Ensure all tests pass, no regressions, and the codebase compiles cleanly.

**Steps**:
1. Run `go build ./...` — zero errors
2. Run `go test ./...` — all tests pass
3. Run `go vet ./...` — zero issues
4. Manually run: `jac --preset=pandoc testdata/marks/underline_html.json` and verify Pandoc output
5. Manually run round-trip smoke test

**Acceptance Criteria**:
- All existing golden tests pass unchanged (backward compatibility)
- All new golden tests pass
- All round-trip tests pass
- `go vet` passes
- CLI `pandoc` preset works end-to-end

---

## Success Criteria

The implementation is complete when:
- [ ] All new `*Pandoc` strategy constants exist and validate correctly
- [ ] Forward converter produces correct Pandoc syntax for all 9 feature types
- [ ] Four goldmark extensions are implemented and registered conditionally
- [ ] Reverse converter reconstructs all Pandoc features back to ADF
- [ ] All new detection types are backward-compatible (defaults match current behaviour)
- [ ] `pandoc` CLI preset works for both forward and reverse conversion
- [ ] Round-trip tests pass for all Pandoc features
- [ ] All existing golden tests pass without modification
- [ ] `go build ./... && go test ./... && go vet ./...` all succeed

---

## Implementation Notes and Potential Challenges

### Pandoc span `[` vs GFM link `[`
The Pandoc span parser must run at lower goldmark priority than the link parser. The discriminator is what follows `]`: if `(` → link, if `[` → reference link, if `{` → Pandoc span. Test extensively with mixed content.

### Subscript `~` vs GFM strikethrough `~~`
Subscript parser must check that the opener `~` is not immediately followed by another `~`. The GFM strikethrough extension uses `~~`, so single `~` is unambiguous.

### Pandoc div `:::` nesting
Divs can be nested. The closer `:::` must belong to the innermost open div. Goldmark's block child model naturally handles this if `HasChildren` is returned in `Open()`.

### Grid table column width with inline markup
Column width is calculated from the rendered cell string length. Cells containing inline markdown (bold, links) may have different visual vs byte lengths. For the initial implementation, use byte-length — the output is valid Pandoc but may not be visually aligned. This is acceptable.

### Alignment applied to multiple blocks in a single div
`:::{ align="center" }` may contain multiple block nodes. ADF alignment is per-block. Apply `layout` attr to each child block individually. If a block type does not support alignment (e.g. a list item), emit a warning and skip.
