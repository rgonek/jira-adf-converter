# Phase 6: Configuration/Params System (Revised v2)

## Overview
Replace the minimal `Config{AllowHTML, Strict}` with a comprehensive, granular configuration system optimized for:
1. **AI-readable output** - minimal token overhead, clean Markdown
2. **Flexible conversion** - per-element control without HTML comments clutter
3. **Round-trip fidelity** - structured syntax preserves metadata natively
4. **Extension handling** - per-type rules with JSON code block output
5. **Structured results** - warnings and metadata returned alongside markdown

**Scope (Phase 6)**: converter library implementation plus CLI preset integration. Full CLI flag redesign, config-file layering, and frontmatter handling are out of scope.

**Key Decisions**:
- **No HTML comments** for metadata preservation (too much fluff)
- **Structured Markdown** for panels (`> [!INFO: Title]`)
- **JSON code blocks** for extensions (readable, parseable)
- **Auto table detection** - pipe vs HTML based on complexity
- **Granular inline styles** - control underline, subsup, colors independently
- **Immutable config** - passed once to `New()`, thread-safe
- **Flat config struct** - no nested grouping structs, Go-idiomatic
- **Structured Result** - `Convert()` returns `Result{Markdown, Warnings}` instead of just string
- **No library-level profiles** - presets live in CLI only; library exposes orthogonal knobs
- **No dialect versioning** - deferred to Phase 7 when reverse converter defines its parsing needs
- **No `log.Printf`** - library never logs; warnings returned in `Result.Warnings`
- **Config is JSON-serializable** - reverse converter receives same Config to know the markdown schema

This is a **breaking API change**: `New()` returns `(*Converter, error)`, `Convert()` returns `(Result, error)`, old boolean-only configuration is replaced with granular per-element configuration.

Frontmatter is **not** part of the converter lib — it will be the CLI tool's responsibility.

---

## Deliverables
1. New `converter/config.go` with granular options and validation
2. New `converter/result.go` with `Result` and `Warning` types
3. Updated converter logic to support all new configuration options
4. New `converter/extensions.go` for extension node handling
5. CLI preset support in `cmd/jac` (`--preset` + preset-to-config mapping)
6. Full test coverage for all new configuration options and preset mapping

---

## Step-by-Step Implementation Plan

### Subphase Execution Order (Actionable)

### Subphase 6.1: Library API & Config Foundation
1. Create/new golden fixtures for granular config behavior (**Task 1**).
2. Add structured conversion result types (**Task 2**).
3. Implement granular `Config` types/defaults/validation (**Task 3**).
4. Add config-focused unit tests (**Task 4**).
5. Migrate converter constructor/return types and warning collection (**Task 5**).

### Subphase 6.2: Library Rendering Rollout
6. Implement mark-level granularity (**Task 6**).
7. Implement inline node granularity (**Task 7**).
8. Implement block-level granularity (**Task 8**).
9. Implement table mode auto-detection and HTML fallback (**Task 9**).
10. Add extension handling with per-type strategies (**Task 10**).
11. Add media base URL behavior (**Task 11**).
12. Add list marker/ordering options (**Task 12**).

### Subphase 6.3: CLI Preset Integration
13. Add CLI preset selection and mapping to library config (**Task 13**).
14. Add/update tests including preset mapping and precedence checks (**Task 14**).

### Subphase 6.4: Final Verification
15. Run full validation and regression checks (**Task 15**).

---

### Task 1: Create Phase 6 Test Data
**Goal**: Create Golden Files for all new config features before implementation (TDD).

**Directories**:
- Formatting: `testdata/marks/`
- Blocks/Panels: `testdata/blocks/`
- Lists: `testdata/lists/`
- Extensions: `testdata/extensions/`
- Media: `testdata/media/`
- Mentions/Links: `testdata/inline/`

**Golden File Naming Convention**:
Config suffix in filename maps to the non-default config used for that test:
- `_underline_bold` → `UnderlineStyle: "bold"`
- `_underline_html` → `UnderlineStyle: "html"`
- `_subsup_latex` → `SubSupStyle: "latex"`
- `_subsup_html` → `SubSupStyle: "html"`
- `_color_html` → `TextColorStyle: "html"`
- `_color_ignore` → `TextColorStyle: "ignore"`
- `_bgcolor_html` → `BackgroundColorStyle: "html"`
- `_panel_bold` → `PanelStyle: "bold"`
- `_panel_github` → `PanelStyle: "github"`
- `_panel_title` → `PanelStyle: "title"`
- `_bullet_star` → `BulletMarker: '*'`
- `_ext_json` → Extension rendered as JSON code block
- `_ext_strip` → Extension stripped
- `_ext_text` → Extension rendered as fallback text
- `_table_auto` → `TableMode: "auto"` (detects complexity)
- `_table_html` → `TableMode: "html"` (force HTML)
- `_mention_text` → `MentionStyle: "text"` (lossy, plain text)
- `_mention_link` → `MentionStyle: "link"` (preserve ID)
- `_mention_html` → `MentionStyle: "html"` (preserve ID and text)
- `_align_html` → `AlignmentStyle: "html"`
- `_expand_html` → `ExpandStyle: "html"`
- `_emoji_unicode` → `EmojiStyle: "unicode"`
- `_status_text` → `StatusStyle: "text"`
- `_date_iso` → `DateFormat` (default ISO 8601)
- `_heading_offset1` → `HeadingOffset: 1`

**Test Cases to Create**:

1. **`testdata/marks/underline_bold.json/.md`**
   * Input: Text with underline mark
   * Expected: `**underlined text**` (underline → bold)

2. **`testdata/marks/underline_html.json/.md`**
   * Input: Text with underline mark
   * Expected: `<u>underlined text</u>`

3. **`testdata/marks/subsup_latex.json/.md`**
   * Input: Text with `subsup` marks
   * Expected: `$_{sub}$` and `$^{sup}$` syntax

4. **`testdata/marks/subsup_html.json/.md`**
   * Input: Text with `subsup` marks
   * Expected: `<sub>sub</sub>` and `<sup>sup</sup>`

5. **`testdata/marks/color_html.json/.md`**
   * Input: Text with textColor mark
   * Expected: `<span style="color: #ff0000">colored text</span>`

6. **`testdata/marks/color_ignore.json/.md`**
   * Input: Text with textColor mark
   * Expected: `colored text` (color stripped)

7. **`testdata/marks/bgcolor_html.json/.md`**
   * Input: Text with backgroundColor mark
   * Expected: `<span style="background-color: #ffff00">highlighted text</span>`

8. **`testdata/blocks/panel_bold.json/.md`**
   * Input: Info Panel with title "Hello" and text "Content"
   * Expected: `> **Info**: Content` (type in bold prefix)

9. **`testdata/blocks/panel_github.json/.md`**
   * Input: Warning Panel with text "Watch out"
   * Expected: `> [!WARNING]\n> Watch out`

10. **`testdata/blocks/panel_title.json/.md`**
    * Input: Info Panel with title "My Panel" and text "Content"
    * Expected: `> [!INFO: My Panel]\n> Content`

11. **`testdata/lists/bullet_star.json/.md`**
    * Input: Bullet list
    * Expected: `* Item 1` (instead of `-`)

12. **`testdata/extensions/ext_json.json/.md`**
    * Input: `bodiedExtension` (Jira Macro) with attrs `{"macro": "code", "language": "go"}`
    * Expected: JSON code block with `adf:extension` language tag

13. **`testdata/extensions/ext_strip.json/.md`**
    * Input: `inlineExtension`
    * Expected: Nothing (stripped)

14. **`testdata/extensions/ext_text.json/.md`**
    * Input: `inlineExtension` with fallback text "@username"
    * Expected: `@username`

15. **`testdata/blocks/heading_offset1.json/.md`**
    * Input: H1, H2
    * Expected: ##, ###

16. **`testdata/tables/table_auto_simple.json/.md`**
    * Input: Simple table (text cells only)
    * Expected: Pipe table format

17. **`testdata/tables/table_auto_complex.json/.md`**
    * Input: Table with colspan or nested blocks
    * Expected: HTML table format

18. **`testdata/media/media_baseurl.json/.md`**
    * Input: Internal media node with config `MediaBaseURL: "https://example.com/media/"`
    * Expected: `![Image](https://example.com/media/{id})`

19. **`testdata/inline/mention_text.json/.md`**
    * Input: Mention with id "12345" and text "User Name"
    * Expected: `@User Name`

20. **`testdata/inline/mention_link.json/.md`**
    * Input: Mention with id "12345" and text "User Name"
    * Expected: `[@User Name](mention:12345)`

21. **`testdata/inline/mention_html.json/.md`**
    * Input: Mention with id "12345" and text "User Name"
    * Expected: `<span data-mention-id="12345">@User Name</span>`

22. **`testdata/blocks/align_html.json/.md`**
    * Input: Paragraph with `align: "center"`
    * Expected: `<div align="center">Centered text</div>`

23. **`testdata/inline/emoji_unicode.json/.md`**
    * Input: Emoji with shortName `:smile:` and fallback unicode `😄`
    * Expected: `😄` (unicode mode)

24. **`testdata/inline/status_text.json/.md`**
    * Input: Status with text "IN PROGRESS"
    * Expected: `IN PROGRESS` (text mode, no brackets)

25. **`testdata/inline/inlinecard_embed.json/.md`**
    * Input: inlineCard with JSONLD data
    * Expected: JSON code block preserving full data

26. **`testdata/blocks/expand_html.json/.md`**
    * Input: Expand node with title and content
    * Expected: `<details><summary>title</summary>content</details>`

27. **`testdata/blocks/decision_text.json/.md`**
    * Input: Decision list with decided/undecided items
    * Expected: Configurable prefix rendering

28. **`testdata/lists/ordered_lazy.json/.md`**
    * Input: Ordered list with 3 items
    * Expected: `1. Item\n1. Item\n1. Item` (lazy numbering)

---

### Task 2: Create `converter/result.go`
**Goal**: Define the structured conversion result.

**File**: `converter/result.go`

```go
package converter

// Result holds the output of a conversion.
type Result struct {
    Markdown string    `json:"markdown"`
    Warnings []Warning `json:"warnings,omitempty"`
}

// WarningType categorizes conversion warnings.
type WarningType string

const (
    WarningUnknownNode      WarningType = "unknown_node"
    WarningUnknownMark      WarningType = "unknown_mark"
    WarningDroppedFeature   WarningType = "dropped_feature"
    WarningExtensionFallback WarningType = "extension_fallback"
    WarningMissingAttribute WarningType = "missing_attribute"
)

// Warning represents a non-fatal issue encountered during conversion.
type Warning struct {
    Type     WarningType `json:"type"`
    NodeType string      `json:"nodeType,omitempty"`
    Message  string      `json:"message"`
}
```

---

### Task 3: Create `converter/config.go`
**Goal**: Define all configuration types, enums, and validation.

**File**: `converter/config.go`

**Types and Constants**:

```go
package converter

// --- Inline Mark Styles ---

// UnderlineStyle controls how underline marks are rendered.
type UnderlineStyle string
const (
    UnderlineIgnore UnderlineStyle = "ignore" // Strip underline
    UnderlineBold   UnderlineStyle = "bold"   // Render as **text**
    UnderlineHTML   UnderlineStyle = "html"   // Render as <u>text</u>
)

// SubSupStyle controls how subscript/superscript marks are rendered.
type SubSupStyle string
const (
    SubSupIgnore SubSupStyle = "ignore" // Strip sub/sup
    SubSupHTML   SubSupStyle = "html"   // Render as <sub>/<sup>
    SubSupLaTeX  SubSupStyle = "latex"  // Render as $_{sub}$ / $^{sup}$
)

// ColorStyle controls how text/background colors are rendered.
type ColorStyle string
const (
    ColorIgnore ColorStyle = "ignore" // Strip color info
    ColorHTML   ColorStyle = "html"   // Render as <span style="color: ...">
)

// MentionStyle controls how user mentions are rendered.
type MentionStyle string
const (
    MentionText MentionStyle = "text" // Render as @User Name (lossy)
    MentionLink MentionStyle = "link" // Render as [@User Name](mention:12345) (round-trip)
    MentionHTML MentionStyle = "html" // Render as <span data-mention-id="...">@User Name</span>
)

// EmojiStyle controls how emoji nodes are rendered.
type EmojiStyle string
const (
    EmojiShortcode EmojiStyle = "shortcode" // :smile: (default)
    EmojiUnicode   EmojiStyle = "unicode"   // Actual unicode character from fallback
)

// --- Block Styles ---

// PanelStyle controls how Info/Note/Warning panels are rendered.
type PanelStyle string
const (
    PanelNone   PanelStyle = "none"   // Just a blockquote (>)
    PanelBold   PanelStyle = "bold"   // > **Info**: ... (type in bold prefix)
    PanelGitHub PanelStyle = "github" // > [!INFO] (GFM alerts)
    PanelTitle  PanelStyle = "title"  // > [!INFO: Custom Title] (with title)
)

// AlignmentStyle controls how block alignment is rendered.
type AlignmentStyle string
const (
    AlignIgnore AlignmentStyle = "ignore" // Ignore alignment
    AlignHTML   AlignmentStyle = "html"   // Wrap in <div align="...">
)

// HardBreakStyle controls how hard line breaks are rendered.
type HardBreakStyle string
const (
    HardBreakBackslash HardBreakStyle = "backslash" // \<newline>
    HardBreakHTML      HardBreakStyle = "html"      // <br>
)

// ExpandStyle controls how expand/collapse sections are rendered.
type ExpandStyle string
const (
    ExpandBlockquote ExpandStyle = "blockquote" // > **title** (blockquote with bold title)
    ExpandHTML       ExpandStyle = "html"       // <details><summary>title</summary>...</details>
)

// StatusStyle controls how status badges are rendered.
type StatusStyle string
const (
    StatusBracket StatusStyle = "bracket" // [Status: IN PROGRESS]
    StatusText    StatusStyle = "text"    // IN PROGRESS (just the text)
)

// InlineCardStyle controls how smart links / inline cards are rendered.
type InlineCardStyle string
const (
    InlineCardLink  InlineCardStyle = "link"  // [title](url) (default)
    InlineCardURL   InlineCardStyle = "url"   // bare URL
    InlineCardEmbed InlineCardStyle = "embed" // JSON code block (preserves JSONLD data)
)

// DecisionStyle controls the prefix for decision items.
type DecisionStyle string
const (
    DecisionEmoji DecisionStyle = "emoji" // ✓ Decision / ? Decision (default)
    DecisionText  DecisionStyle = "text"  // DECIDED / UNDECIDED
)

// OrderedListStyle controls ordered list numbering.
type OrderedListStyle string
const (
    OrderedIncremental OrderedListStyle = "incremental" // 1. 2. 3. (default)
    OrderedLazy        OrderedListStyle = "lazy"        // 1. 1. 1. (diff-friendly)
)

// --- Tables ---

// TableMode controls how tables are rendered.
type TableMode string
const (
    TableAuto TableMode = "auto" // Detect complexity, use pipe or HTML
    TablePipe TableMode = "pipe" // Force pipe tables
    TableHTML TableMode = "html" // Force HTML tables
)

// --- Extensions ---

// ExtensionMode controls how extension nodes are handled.
type ExtensionMode string
const (
    ExtensionJSON  ExtensionMode = "json"  // Render as JSON code block (default)
    ExtensionText  ExtensionMode = "text"  // Render fallback text only
    ExtensionStrip ExtensionMode = "strip" // Remove entirely
)

// ExtensionRules allows per-extension-type configuration.
type ExtensionRules struct {
    Default ExtensionMode            `json:"default"`           // Default for all extensions
    ByType  map[string]ExtensionMode `json:"byType,omitempty"`  // Per-type overrides
}

// --- Unknown Handling ---

// UnknownPolicy controls behavior for unrecognized ADF elements.
type UnknownPolicy string
const (
    UnknownError       UnknownPolicy = "error"       // Return error
    UnknownSkip        UnknownPolicy = "skip"         // Silently skip
    UnknownPlaceholder UnknownPolicy = "placeholder"  // Render [Unknown node: type]
)

// --- Config ---

// Config holds all converter configuration options.
// All fields are JSON-serializable so the config can be stored alongside
// markdown output for the reverse converter to use.
type Config struct {
    // --- Inline marks ---
    UnderlineStyle       UnderlineStyle  `json:"underlineStyle,omitempty"`
    SubSupStyle          SubSupStyle     `json:"subSupStyle,omitempty"`
    TextColorStyle       ColorStyle      `json:"textColorStyle,omitempty"`
    BackgroundColorStyle ColorStyle      `json:"backgroundColorStyle,omitempty"`
    MentionStyle         MentionStyle    `json:"mentionStyle,omitempty"`
    EmojiStyle           EmojiStyle      `json:"emojiStyle,omitempty"`

    // --- Blocks ---
    PanelStyle     PanelStyle     `json:"panelStyle,omitempty"`
    HeadingOffset  int            `json:"headingOffset,omitempty"`
    HardBreakStyle HardBreakStyle `json:"hardBreakStyle,omitempty"`
    AlignmentStyle AlignmentStyle `json:"alignmentStyle,omitempty"`
    ExpandStyle    ExpandStyle    `json:"expandStyle,omitempty"`

    // --- Inline nodes ---
    StatusStyle    StatusStyle     `json:"statusStyle,omitempty"`
    InlineCardStyle InlineCardStyle `json:"inlineCardStyle,omitempty"`
    DecisionStyle  DecisionStyle   `json:"decisionStyle,omitempty"`
    DateFormat     string          `json:"dateFormat,omitempty"` // Go time format string

    // --- Tables ---
    TableMode TableMode `json:"tableMode,omitempty"`

    // --- Lists ---
    BulletMarker     rune             `json:"bulletMarker,omitempty"`
    OrderedListStyle OrderedListStyle `json:"orderedListStyle,omitempty"`

    // --- Extensions ---
    Extensions ExtensionRules `json:"extensions,omitempty"`

    // --- Media ---
    MediaBaseURL string `json:"mediaBaseURL,omitempty"`

    // --- Code blocks ---
    LanguageMap map[string]string `json:"languageMap,omitempty"` // e.g. "c++" → "cpp"

    // --- Error handling ---
    UnknownNodes UnknownPolicy `json:"unknownNodes,omitempty"`
    UnknownMarks UnknownPolicy `json:"unknownMarks,omitempty"`
}
```

**Methods**:

- `applyDefaults() Config` — fills zero-value fields with defaults:

| Field | Default | Rationale |
|-------|---------|-----------|
| `UnderlineStyle` | `"bold"` | Preserves emphasis semantically |
| `SubSupStyle` | `"html"` | Preserves sub/sup info for round-trip |
| `TextColorStyle` | `"ignore"` | Colors rarely meaningful for AI/sync |
| `BackgroundColorStyle` | `"ignore"` | Background color rarely meaningful |
| `MentionStyle` | `"link"` | Preserves user ID for round-trip |
| `EmojiStyle` | `"shortcode"` | `:smile:` is parseable and universal |
| `PanelStyle` | `"github"` | `> [!INFO]` is GFM-standard and parseable |
| `HeadingOffset` | `0` | No shift by default |
| `HardBreakStyle` | `"backslash"` | Visible and reliable |
| `AlignmentStyle` | `"ignore"` | Alignment rarely needed in markdown |
| `ExpandStyle` | `"html"` | `<details>` is standard and round-trips |
| `StatusStyle` | `"bracket"` | `[Status: X]` is parseable |
| `InlineCardStyle` | `"link"` | Standard markdown link |
| `DecisionStyle` | `"emoji"` | ✓/? prefix is compact and clear |
| `DateFormat` | `"2006-01-02"` | ISO 8601 |
| `OrderedListStyle` | `"incremental"` | Standard behavior |
| `TableMode` | `"auto"` | Smart detection |
| `BulletMarker` | `'-'` | Standard markdown |
| `Extensions.Default` | `"json"` | Preserves all data for sync/round-trip |
| `UnknownNodes` | `"placeholder"` | Visible but non-fatal |
| `UnknownMarks` | `"skip"` | Unknown marks silently dropped |

- `Validate() error` — validates:
  - All enum fields contain valid values
  - `HeadingOffset` is in range 0-5
  - `BulletMarker` is one of `-`, `*`, `+`
  - `DateFormat` is a valid Go time format (contains reference time components)
  - `Extensions.ByType` keys are non-empty strings
  - `LanguageMap` keys/values are non-empty strings

---

### Task 4: Create `converter/config_test.go`
**Test Cases**:

1. **TestApplyDefaults**: Verify all zero values get correct defaults
2. **TestValidateValid**: Test valid configurations pass
3. **TestValidateInvalidEnum**: Test invalid enum string values fail
4. **TestValidateInvalidRange**: Test out-of-range HeadingOffset, invalid BulletMarker fail
5. **TestExtensionRulesLookup**: Test that ByType lookup falls back to Default
6. **TestConfigSerialization**: Verify Config round-trips through JSON marshal/unmarshal
7. **TestZeroConfigUsable**: Verify `Config{}` with defaults applied produces valid config

---

### Task 5: Modify `converter/converter.go`
**Goal**: Update core converter for new config system and structured results.

**Breaking Changes**:
- `New(config Config)` → `New(config Config) (*Converter, error)`
- `Convert(input []byte) (string, error)` → `Convert(input []byte) (Result, error)`

**Implementation Details**:

1. **`New(config Config) (*Converter, error)`**:
   - Call `config.applyDefaults()`
   - Call `config.Validate()`
   - Return error on validation failure
   - Store immutable config
   - Initialize empty warnings slice

2. **`Convert()` returns `Result`**:
   - Collect warnings during conversion (unknown nodes, dropped features)
   - Return `Result{Markdown: ..., Warnings: collected}`

3. **Warning collection**:
   - Add `warnings []Warning` field to Converter (reset per Convert call)
   - Add `addWarning(warnType WarningType, nodeType, message string)` helper
   - Unknown node with `UnknownPlaceholder`: add warning + render placeholder
   - Unknown node with `UnknownSkip`: add warning + skip
   - Unknown mark with `UnknownSkip`: add warning + skip

4. **Extension routing in `convertNode` default case**:
   - Check if node type is `extension`, `inlineExtension`, or `bodiedExtension`
   - If yes, delegate to `convertExtension(node)`
   - Otherwise, apply `UnknownNodes` policy

5. **Library-level strictness primitives**:
   - `UnknownNodes` and `UnknownMarks` remain policy-driven in the library
   - CLI aliases (like `--strict`) can map to these policies in Task 13

---

### Task 6: Modify `converter/marks.go`
**Goal**: Implement granular inline mark styling.

**Implementation Details**:

Replace `c.config.AllowHTML` checks with individual style lookups:

* **Underline** (`convertMarkFull` case `"underline"`):
  - `UnderlineIgnore`: return `"", ""`
  - `UnderlineBold`: return `"**", "**"`
  - `UnderlineHTML`: return `"<u>", "</u>"`

* **SubSup** (`convertMarkFull` case `"subsup"`):
  - `SubSupIgnore`: return `"", ""`
  - `SubSupHTML`: return `"<sub>", "</sub>"` or `"<sup>", "</sup>"`
  - `SubSupLaTeX`: return `"$_{", "}$"` or `"$^{", "}$"`

* **TextColor** (new mark type `"textColor"`):
  - Add to `isKnownMark()`: `"textColor"` → true
  - `ColorIgnore`: return `"", ""`
  - `ColorHTML`: extract `color` attr, return `<span style="color: #RRGGBB">`, `</span>`

* **BackgroundColor** (new mark type `"backgroundColor"`):
  - Add to `isKnownMark()`: `"backgroundColor"` → true
  - `ColorIgnore`: return `"", ""`
  - `ColorHTML`: extract `color` attr, return `<span style="background-color: #RRGGBB">`, `</span>`

* **Unknown marks**: Use `config.UnknownMarks` policy instead of `config.Strict`

---

### Task 7: Modify `converter/inline.go`
**Goal**: Implement configurable mention, emoji, status, date, and inline card rendering.

**Implementation Details**:

* **`convertMention()`**:
  - `MentionText`: `@text` (current behavior)
  - `MentionLink`: `[@text](mention:id)` — extract `id` from attrs
  - `MentionHTML`: `<span data-mention-id="id">@text</span>`

* **`convertEmoji()`**:
  - `EmojiShortcode`: return `shortName` (current behavior)
  - `EmojiUnicode`: prefer `fallback` (unicode char), then `shortName`

* **`convertStatus()`**:
  - `StatusBracket`: `[Status: text]` (current behavior)
  - `StatusText`: just `text`

* **`convertDate()`**:
  - Use `config.DateFormat` (Go time format string) instead of hardcoded `"2006-01-02"`

* **`convertInlineCard()`**:
  - `InlineCardLink`: `[title](url)` (current behavior)
  - `InlineCardURL`: bare URL
  - `InlineCardEmbed`: JSON code block with `adf:inlineCard` language tag

---

### Task 8: Modify `converter/blocks.go`
**Goal**: Implement `PanelStyle`, `HeadingOffset`, `HardBreakStyle`, `AlignmentStyle`, `ExpandStyle`, `DecisionStyle`.

**Implementation Details**:

* **`convertPanel()`**:
  * Extract `panelType` and optional `title` from node attrs
  * Map panelType to display: `"info"` → `"INFO"` / `"Info"` depending on style
  * Check `PanelStyle`:
    - `PanelNone`: Standard blockquote `> ...`
    - `PanelBold`: `> **Info**: ...` (type in Title Case, colon separator)
    - `PanelGitHub`: `> [!INFO]\n> ...` (type in UPPERCASE)
    - `PanelTitle`: `> [!INFO: Custom Title]\n> ...` (include title if present, fall back to `[!INFO]`)

* **`convertHeading()`**:
  * Calculate: `newLevel = min(6, max(1, originalLevel + HeadingOffset))`
  * Render appropriate number of `#` characters
  * Check `AlignmentStyle` + `attrs.align` for alignment wrapping

* **`convertParagraph()`**:
  * Check `attrs.align` + `AlignmentStyle`:
    - `AlignHTML`: wrap in `<div align="center">...</div>`
    - `AlignIgnore`: render normally

* **`convertHardBreak()`**:
  * `HardBreakBackslash`: `\<newline>` (current behavior)
  * `HardBreakHTML`: `<br>`

* **`convertExpand()`**:
  * `ExpandHTML`: `<details><summary>title</summary>...</details>` (current AllowHTML behavior)
  * `ExpandBlockquote`: `> **title**\n> ...` (current non-HTML behavior)

* **`convertDecisionItemContent()`**:
  * `DecisionEmoji`: `✓ Decision` / `? Decision` (current behavior)
  * `DecisionText`: `DECIDED` / `UNDECIDED`

---

### Task 9: Modify `converter/tables.go`
**Goal**: Implement `TableMode` with auto-detection.

**Implementation Details**:

* **`isComplexTable(node Node) bool` helper**:
  * Returns true if table has:
    - Any cell with `colspan` > 1
    - Any cell with `rowspan` > 1
    - Any cell containing block nodes (lists, code blocks, nested tables)
  * Returns false for simple text-only tables

* **`convertTable()`** routing:
  * `TableHTML`: always use HTML table rendering
  * `TablePipe`: always use pipe table rendering (current behavior)
  * `TableAuto`: call `isComplexTable()` → complex=HTML, simple=Pipe

* **HTML Table Rendering** (new):
  * `<table>`, `<thead>`, `<tbody>`, `<tr>`, `<th>`, `<td>`
  * Preserve `colspan` and `rowspan` attributes
  * Render cell content recursively

* **Pipe Table Rendering** (existing, with config):
  * Use `HardBreakStyle` for in-cell line breaks instead of hardcoded `<br>` / space
  * Escape pipe characters (existing)

---

### Task 10: Create `converter/extensions.go`
**Goal**: Handle extension nodes with per-type rules.

**File**: `converter/extensions.go` (new)

**Implementation Details**:

```go
func (c *Converter) convertExtension(node Node) (string, error) {
    // Determine extension type from attrs
    extType := node.GetStringAttr("extensionType", "")
    if extType == "" {
        extType = node.GetStringAttr("extensionKey", "")
    }
    if extType == "" {
        extType = node.Type // fallback to node type itself
    }

    // Look up handling strategy
    strategy := c.config.Extensions.Default
    if c.config.Extensions.ByType != nil {
        if specific, ok := c.config.Extensions.ByType[extType]; ok {
            strategy = specific
        }
    }

    switch strategy {
    case ExtensionStrip:
        c.addWarning(WarningDroppedFeature, node.Type,
            fmt.Sprintf("extension %q stripped", extType))
        return "", nil
    case ExtensionText:
        text := c.getExtensionFallbackText(node)
        if text == "" {
            c.addWarning(WarningExtensionFallback, node.Type,
                fmt.Sprintf("extension %q has no fallback text", extType))
        }
        return text, nil
    case ExtensionJSON:
        return c.renderExtensionJSON(node), nil
    default:
        return "", fmt.Errorf("unknown extension strategy: %s", strategy)
    }
}

func (c *Converter) renderExtensionJSON(node Node) string {
    data, _ := json.MarshalIndent(node, "", "  ")
    return fmt.Sprintf("```adf:extension\n%s\n```\n\n", string(data))
}

func (c *Converter) getExtensionFallbackText(node Node) string {
    // Try to extract text from content children
    if len(node.Content) > 0 {
        text, _ := c.convertChildren(node.Content)
        return strings.TrimSpace(text)
    }
    // Try attrs for fallback text
    return node.GetStringAttr("text", "")
}
```

---

### Task 11: Modify `converter/media.go`
**Goal**: Support MediaBaseURL for internal media resolution.

**Implementation Details**:
* Update `convertMedia()`:
  * If `config.MediaBaseURL` is set and media is internal (no URL):
    - Construct: `![alt](baseURL + id)`
  * If no BaseURL and internal: keep current `[Image: id]` placeholder
  * External media: unchanged (already has URL)

---

### Task 12: Modify `converter/lists.go`
**Goal**: Implement custom bullet marker and ordered list style.

**Implementation Details**:
* **`convertBulletList()`**: Use `config.BulletMarker` + " " as marker
* **`convertOrderedList()`**:
  - `OrderedIncremental`: `1. 2. 3.` (current behavior)
  - `OrderedLazy`: always use `1.` regardless of index

---

### Task 13: Modify `cmd/jac/main.go` (CLI Presets)
**Goal**: Expose preset-based configuration in CLI while keeping library config canonical.

**Implementation Details**:
1. Add `--preset` flag with allowed values:
   - `balanced` (default)
   - `strict`
   - `readable`
   - `lossy` (one-way optimization)

2. Implement preset mapping to `converter.Config`:
   - `balanced`: rely on library defaults (no explicit overrides)
   - `strict`: `UnknownNodes=error`, `UnknownMarks=error`, `MentionStyle=link`, `Extensions.Default=json`
   - `readable`: `MentionStyle=text`, `TextColorStyle=ignore`, `BackgroundColorStyle=ignore`, `AlignmentStyle=ignore`, `Extensions.Default=text`, `ExpandStyle=blockquote`
   - `lossy`: `MentionStyle=text`, `Extensions.Default=strip`, `TextColorStyle=ignore`, `BackgroundColorStyle=ignore`, `InlineCardStyle=url`

3. Preset precedence:
   - Apply preset first
   - Apply explicit CLI flags second (flags override preset)
   - Pass final config to `converter.New(config)` (library still applies defaults + validation)

4. Compatibility:
   - Retain `--strict` CLI alias by mapping it to `UnknownNodes=error` + `UnknownMarks=error`
   - Return clear error for unknown preset values listing allowed options

---

### Task 14: Tests and Verification

**Update existing tests**:
1. `converter/converter_test.go` - Update to use new Config struct and `Result` return type
2. `converter/custom_table_test.go` - Update to use new Config struct
3. `cmd/jac/main_test.go` (or equivalent) - Add preset parsing and config resolution tests
4. All existing golden files remain valid with default config

**Golden file test infrastructure update**:
- Detect config from filename suffix (e.g., `_mention_link` → set `MentionStyle: "link"`)
- Or use a companion `.config.json` file per test for complex configs
- Existing `_html` suffix tests need migration to specific config options

**New test files**:
1. `converter/config_test.go` - Config validation, defaults, serialization
2. `converter/result_test.go` - Warning collection and Result structure
3. `converter/marks_config_test.go` - Granular mark rendering (all style combinations)
4. `converter/extensions_test.go` - Extension handling strategies
5. `converter/tables_config_test.go` - Auto-detection logic, HTML table rendering
6. `converter/inline_config_test.go` - Mention, emoji, status, date, inline card styles
7. `cmd/jac/preset_test.go` - Preset mapping and precedence validation

**Test coverage requirements**:
- Every config option has at least 2 test cases (default + non-default)
- ExtensionRules ByType lookup and fallback tested
- Table auto-detection tested with simple and complex tables
- All PanelStyle options tested
- All mark style combinations tested
- Warning collection tested (unknown nodes produce warnings)
- `Result.Warnings` populated correctly
- Config JSON round-trip tested
- All CLI presets (`balanced`, `strict`, `readable`, `lossy`) map to expected config values
- Preset precedence tested (explicit flag overrides preset)
- Invalid preset value returns explicit CLI error

---

### Task 15: Final Verification
**Goal**: Ensure all tests pass and no regressions exist.

**Steps**:
1. Run `go test ./...` to ensure all tests pass
2. Run linter
3. Regenerate golden files, review diffs
4. Manual CLI smoke test with presets:
   ```bash
   go run ./cmd/jac --preset=balanced testdata/simple/hello.json
   go run ./cmd/jac --preset=readable testdata/inline/mention_text.json
   go run ./cmd/jac --preset=lossy testdata/extensions/ext_text.json
   ```
5. Verify zero `c.config.AllowHTML` references remain
6. Verify zero `c.config.Strict` references remain (replaced by UnknownNodes/UnknownMarks)
7. Verify zero `log.Printf` calls in converter package
8. Verify no HTML comment generation logic exists

**Acceptance Criteria**:
- All tests pass
- Linting passes
- No old config field references anywhere in converter package
- CLI `--preset` supports `balanced|strict|readable|lossy`
- Preset application order is deterministic (preset first, explicit flags second)
- Default config produces AI-readable, human-readable, reverse-convertible Markdown
- `Convert()` returns structured `Result` with markdown and warnings
- Config JSON-serializes and deserializes correctly

---

## Success Criteria for Phase 6
The phase is complete when:
- [x] `converter/config.go` defines all granular types, flat Config, Validate(), applyDefaults()
- [x] `converter/result.go` defines Result and Warning types
- [x] `New()` returns `(*Converter, error)` with config validation
- [x] `Convert()` returns `(Result, error)` with warnings
- [x] No FidelityProfile, LossControl, or MarkdownDialect in the library
- [x] Granular inline controls work (underline, subsup, text color, background color independently)
- [x] Mention rendering supports text/link/html modes
- [x] Emoji rendering supports shortcode/unicode modes
- [x] Status rendering supports bracket/text modes
- [x] Date rendering uses configurable format string
- [x] InlineCard rendering supports link/url/embed modes
- [x] Decision rendering supports emoji/text prefix styles
- [x] TableMode "auto" detects complexity and chooses pipe vs HTML
- [x] PanelStyle supports none/bold/github/title with title preservation
- [x] ExtensionRules support per-type handling with JSON code block output (default: JSON)
- [x] HeadingOffset shifts heading levels with clamping
- [x] HardBreakStyle supports backslash/html
- [x] ExpandStyle supports blockquote/html
- [x] AlignmentStyle supports ignore/html
- [x] OrderedListStyle supports incremental/lazy
- [x] BulletMarker configurable (`-`, `*`, `+`)
- [x] MediaBaseURL constructs proper image URLs for internal media
- [x] LanguageMap maps code block languages
- [x] CLI `--preset` supports balanced/strict/readable/lossy
- [x] CLI preset mapping sets expected config values
- [x] Explicit CLI flags override preset values
- [x] All tests pass with new config API
- [x] All new features have dedicated test coverage
- [x] Zero references to old `AllowHTML` / `Strict` fields in converter package
- [x] Zero `log.Printf` calls in converter package
- [x] Unknown handling is policy-driven (UnknownNodes, UnknownMarks)
- [x] Config is JSON-serializable for future reverse converter use
- [x] No HTML comment generation code exists

---

## Files Modified Summary

| File | Action |
|------|--------|
| `converter/config.go` | **NEW** — flat Config, all types/enums, Validate, applyDefaults |
| `converter/result.go` | **NEW** — Result, Warning, WarningType |
| `converter/config_test.go` | **NEW** — config validation, defaults, serialization tests |
| `converter/result_test.go` | **NEW** — warning collection tests |
| `converter/extensions.go` | **NEW** — extension handling with per-type rules |
| `converter/extensions_test.go` | **NEW** — extension strategy tests |
| `converter/converter.go` | Remove old Config, change New() → error, Convert() → Result, warning collection |
| `converter/marks.go` | Granular inline styles (underline, subsup, textColor, backgroundColor) |
| `converter/blocks.go` | PanelStyle, HeadingOffset, HardBreakStyle, AlignmentStyle, ExpandStyle, DecisionStyle |
| `converter/tables.go` | TableMode with auto-detection, HTML table rendering |
| `converter/media.go` | MediaBaseURL support |
| `converter/lists.go` | BulletMarker, OrderedListStyle |
| `converter/inline.go` | MentionStyle, EmojiStyle, StatusStyle, DateFormat, InlineCardStyle |
| `cmd/jac/main.go` | Add `--preset`, preset-to-config mapping, strict alias compatibility |
| `cmd/jac/preset_test.go` | **NEW** — preset mapping and precedence tests |
| `converter/converter_test.go` | Migrate to new Config + Result |
| `converter/custom_table_test.go` | Migrate to new Config |
| `converter/marks_config_test.go` | **NEW** — mark style combination tests |
| `converter/tables_config_test.go` | **NEW** — table mode tests |
| `converter/inline_config_test.go` | **NEW** — inline node style tests |
| `testdata/**` | ~28 new golden file pairs |

---

## Design Decisions Summary

1. **Flat Config**: No nested structs; Go-idiomatic flat struct with clear field names
2. **Structured Result**: `Convert()` returns `Result{Markdown, Warnings}` — library never logs
3. **No Library Profiles**: Presets (balanced/strict/readable/lossy) live in CLI only
4. **No Dialect Versioning**: Deferred to Phase 7 when reverse converter defines needs
5. **JSON-Serializable Config**: Reverse converter receives same Config to parse markdown
6. **Granular Inline Control**: Each mark type configurable independently
7. **Auto Table Detection**: Chooses pipe or HTML based on table complexity
8. **JSON Extensions Default**: Extensions preserved as JSON by default for sync/round-trip
9. **Per-Type Extension Rules**: Different handling for different macro types
10. **Immutable Config**: Thread-safe, validated at construction

---

## Future Considerations (Not in Phase 6)

- **Frontmatter Integration**: CLI tool responsibility, converter could support metadata hooks
- **Reverse Conversion (MD → ADF)**: Phase 7 — receives Config to know markdown schema
- **Sync Tool**: Phase 8 — uses converter library with stored Config
- **Plugin System**: Allow users to register custom converters for specific node types
- **Streaming API**: Process large documents in chunks for memory efficiency
- **Additional config options**: link base URL, link mapping, media alt text policy (if needed)
