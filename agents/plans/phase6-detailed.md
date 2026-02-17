# Phase 6: Configuration/Params System (Revised)

## Overview
Replace the minimal `Config{AllowHTML, Strict}` with a comprehensive, granular configuration system optimized for:
1. **AI-readable output** - minimal token overhead, clean Markdown
2. **Flexible conversion** - per-element control without HTML comments clutter
3. **Round-trip fidelity** - structured syntax preserves metadata natively
4. **Extension handling** - per-type rules with JSON code block output
5. **Dual-mode fidelity** - default reversible output plus explicit opt-in lossy behavior

**Key Decisions**:
- ❌ **No HTML comments** for metadata preservation (too much fluff)
- ✅ **Structured Markdown** for panels (`> [!INFO: Title]`)
- ✅ **JSON code blocks** for extensions (readable, parseable)
- ✅ **Auto table detection** - pipe vs HTML based on complexity
- ✅ **Granular inline styles** - control underline, subsup, colors independently
- ✅ **Immutable config** - passed once to `New()`, thread-safe
- ✅ **Link & Mention Strategies** - Support for lossy (readable) vs lossless (round-trippable) rendering
- ✅ **Versioned Markdown Dialect** - stable, parseable contract for Markdown → ADF
- ✅ **Fidelity Profiles** - `balanced-lossless` (default), `strict-lossless`, `readable-lossy`
- ✅ **Explicit Loss Controls** - loss is opt-in and tracked by feature

This is a **breaking API change**: `New()` returns `(*Converter, error)`, old boolean-only configuration is replaced with profile + policy based configuration.

Frontmatter is **not** part of the converter lib — it will be the CLI tool's responsibility.

---

## Deliverables
1. New `converter/config.go` with granular options (Inline/Block/Extension strategies)
2. Updated converter logic to support custom bullets, panel styles, and extension rendering
3. Extended CLI with granular flags
4. Full test coverage for all new configuration options
5. **New Fidelity Presets** (`NewBalancedLossless`, `NewStrictLossless`, `NewReadableLossy`)
6. **Versioned Dialect Contract** (`MarkdownDialectVersion`) for reverse conversion stability

---

## Step-by-Step Implementation Plan

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
- `_underline_bold` → `UnderlineStyle: "bold"`
- `_underline_html` → `UnderlineStyle: "html"`
- `_subsup_latex` → `SubSupStyle: "latex"`
- `_subsup_html` → `SubSupStyle: "html"`
- `_color_html` → `ColorStyle: "html_span"`
- `_color_ignore` → `ColorStyle: "ignore"`
- `_panel_bold` → `PanelStrategy: "bold"`
- `_panel_github` → `PanelStrategy: "github"`
- `_panel_title` → `PanelStrategy: "title"`
- `_bullet_star` → `BulletChar: "*"`
- `_ext_json` → Extension rendered as JSON code block
- `_ext_strip` → Extension stripped
- `_table_auto` → TableFormat: "auto" (detects complexity)
- `_table_html` → TableFormat: "html" (force HTML)
- `_mention_link` → `MentionStyle: "link"` (preserve ID)
- `_mention_html` → `MentionStyle: "html"` (preserve ID and text)
- `_align_html` → `AlignmentStyle: "html"`
- `_fidelity_balanced` → `FidelityProfile: "balanced-lossless"` (default reversible readable output)
- `_fidelity_lossy` → `FidelityProfile: "readable-lossy"` with `Loss.Allowed` feature list

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
   * Input: Text with textColor attribute
   * Expected: `<span style="color: #ff0000">colored text</span>`

6. **`testdata/marks/color_ignore.json/.md`**
   * Input: Text with textColor attribute
   * Expected: `colored text` (color stripped)

7. **`testdata/blocks/panel_bold.json/.md`**
   * Input: Info Panel with title "Hello" and text "Content"
   * Expected: `> **Info**\n>\n> Content` (title in bold)

8. **`testdata/blocks/panel_github.json/.md`**
   * Input: Warning Panel with text "Watch out"
   * Expected: `> [!WARNING]\n> Watch out`

9. **`testdata/blocks/panel_title.json/.md`**
   * Input: Custom Panel with title "My Panel" and text "Content"
   * Expected: `> [!INFO: My Panel]\n> Content`

10. **`testdata/lists/bullet_star.json/.md`**
    * Input: Bullet list
    * Expected: `* Item 1` (instead of `-`)

11. **`testdata/extensions/ext_json.json/.md`**
    * Input: `bodiedExtension` (Jira Macro) with attrs `{"macro": "code", "language": "go"}`
    * Expected: Code block with raw JSON attributes
    ```markdown
    ```adf:macro
    {"type": "bodiedExtension", "attrs": {"macro": "code", "language": "go"}, "content": [...]}
    ```
    ```

12. **`testdata/extensions/ext_strip.json/.md`**
    * Input: `inlineExtension` (Mention)
    * Expected: Nothing (stripped)

13. **`testdata/extensions/ext_text.json/.md`**
    * Input: `inlineExtension` with fallback text "@username"
    * Expected: `@username`

14. **`testdata/blocks/heading_shift1.json/.md`**
    * Input: H1, H2
    * Expected: ##, ###

15. **`testdata/tables/table_auto_simple.json/.md`**
    * Input: Simple table (text cells only)
    * Expected: Pipe table format

16. **`testdata/tables/table_auto_complex.json/.md`**
    * Input: Table with colspan or nested blocks
    * Expected: HTML table format

17. **`testdata/media/media_baseurl.json/.md`**
    * Input: Internal media node
    * Expected: `![Image](https://example.com/media/{id})`

18. **`testdata/inline/mention_link.json/.md`**
    * Input: Mention with id "12345" and text "User Name"
    * Expected: `[@User Name](mention:12345)`

19. **`testdata/inline/mention_html.json/.md`**
    * Input: Mention with id "12345" and text "User Name"
    * Expected: `<span data-mention-id="12345">@User Name</span>`

20. **`testdata/blocks/align_html.json/.md`**
    * Input: Paragraph with `align: "center"`
    * Expected: `<div align="center">Centered text</div>`

21. **`testdata/fidelity/balanced_lossless.json/.md`**
    * Input: Mixed document (mentions, extensions, alignment, panel titles)
    * Expected: Human-readable markdown that preserves reversible metadata markers (default profile)

22. **`testdata/fidelity/readable_lossy_color_align.json/.md`**
    * Input: Same mixed document with `Loss.Allowed={"textColor","alignment"}`
    * Expected: Clean markdown with only allowed losses; other metadata remains reversible

---

### Task 2: Create `converter/config.go`
**Goal**: Define all configuration types, enums, and validation.

**File**: `converter/config.go`

**Types and Constants**:

```go
// UnderlineStyle controls how underline marks are rendered
type UnderlineStyle string
const (
    UnderlineIgnore UnderlineStyle = "ignore" // Strip underline
    UnderlineBold   UnderlineStyle = "bold"   // Render as bold (**text**)
    UnderlineHTML   UnderlineStyle = "html"   // Render as <u>text</u>
)

// SubSupStyle controls how subscript/superscript marks are rendered
type SubSupStyle string
const (
    SubSupIgnore SubSupStyle = "ignore" // Strip sub/sup
    SubSupHTML   SubSupStyle = "html"   // Render as <sub>/<sup>
    SubSupLaTeX  SubSupStyle = "latex"  // Render as $_{sub}$ / $^{sup}$
)

// ColorStyle controls how text/background colors are rendered
type ColorStyle string
const (
    ColorIgnore    ColorStyle = "ignore"    // Strip color info
    ColorHTML      ColorStyle = "html"      // Render as <span style="color: ...">
    ColorHTMLShort ColorStyle = "html-short" // Render as <font color="..."> (if needed)
)

// MentionStyle controls how user mentions are rendered
type MentionStyle string
const (
    MentionText MentionStyle = "text" // Render as "@User Name" (lossy)
    MentionLink MentionStyle = "link" // Render as "[@User Name](mention:12345)" (round-trip friendly)
    MentionHTML MentionStyle = "html" // Render as <span data-mention-id="...">@User Name</span>
)

// FidelityProfile controls high-level conversion behavior
type FidelityProfile string
const (
    FidelityBalancedLossless FidelityProfile = "balanced-lossless" // Default: readable + reversible
    FidelityStrictLossless   FidelityProfile = "strict-lossless"   // Preserve everything possible
    FidelityReadableLossy    FidelityProfile = "readable-lossy"    // Prioritize clean readability
)

// MarkdownDialect versions the reversible markdown contract
type MarkdownDialect struct {
    Version string // e.g. "adf-md-v1"
}

// AlignmentStyle controls how block alignment is rendered
type AlignmentStyle string
const (
    AlignIgnore AlignmentStyle = "ignore" // Ignore alignment (default GFM)
    AlignHTML   AlignmentStyle = "html"   // Wrap in <div align="...">
)

// InlineStyles groups all inline formatting options
type InlineStyles struct {
    Underline       UnderlineStyle
    SubSup          SubSupStyle
    TextColor       ColorStyle
    BackgroundColor ColorStyle
    Mention         MentionStyle
}

// TableFormat controls how tables are rendered
type TableFormat string
const (
    TableAuto TableFormat = "auto"   // Detect complexity, use pipe or HTML
    TablePipe TableFormat = "pipe"   // Force pipe tables (ASCII art)
    TableHTML TableFormat = "html"   // Force HTML tables
)

// PanelStrategy controls how Info/Note/Warning panels are rendered
type PanelStrategy string
const (
    PanelNone   PanelStrategy = "none"   // Just a blockquote (>)
    PanelBold   PanelStrategy = "bold"   // > **Info** (type in bold)
    PanelGitHub PanelStrategy = "github" // > [!INFO] (GFM alerts)
    PanelTitle  PanelStrategy = "title"  // > [!INFO: Custom Title] (with title)
)

// ExtensionHandling controls how Jira/Confluence macros are handled
type ExtensionHandling string
const (
    ExtensionWarn  ExtensionHandling = "warn"  // Log warning, render text if available
    ExtensionStrip ExtensionHandling = "strip" // Remove entirely
    ExtensionText  ExtensionHandling = "text"  // Render fallback text only
    ExtensionJSON  ExtensionHandling = "json"  // Render as JSON code block
)

// ExtensionRules allows per-extension-type configuration
type ExtensionRules struct {
    Default ExtensionHandling              // Default for unknown extensions
    ByType  map[string]ExtensionHandling  // Per-type rules (e.g., "jira.issue", "code")
}

type UnknownPolicy string
const (
    UnknownError       UnknownPolicy = "error"
    UnknownWarn        UnknownPolicy = "warn"
    UnknownPlaceholder UnknownPolicy = "placeholder"
    UnknownStrip       UnknownPolicy = "strip"
)

type UnknownHandling struct {
    Node UnknownPolicy
    Mark UnknownPolicy
    Attr UnknownPolicy
}

type LossControl struct {
    Allowed map[string]bool // Explicitly allowed losses: textColor, backgroundColor, alignment, extensionPayload, etc.
}

type HardBreakStrategy string
const (
    HardBreakBackslash   HardBreakStrategy = "backslash"    // \\n
    HardBreakDoubleSpace HardBreakStrategy = "double-space" // "  \n"
    HardBreakHTML        HardBreakStrategy = "html"         // <br>
)

type MediaConfig struct {
    BaseURL        string
    DownloadImages bool
    AltTextPolicy  string // filename, occurrenceId, none
}

type LinkConfig struct {
    BaseURL        string            // Prefix for relative links (e.g. /wiki/...)
    LinkMapping    map[string]string // Static map for specific URLs
}

type MarkdownOptions struct {
    BulletChar rune // '-', '*', '+'
    LanguageMap map[string]string // Map "c++" -> "cpp"
}

type Config struct {
    // High-level fidelity and dialect
    Profile FidelityProfile
    Dialect MarkdownDialect

    // Granular inline control
    Inline InlineStyles
    
    // Block-level control
    TableFormat       TableFormat
    PanelStrategy     PanelStrategy
    HeadingShift      int
    HardBreakStrategy HardBreakStrategy
    Alignment         AlignmentStyle
    
    // Extension handling
    Extensions ExtensionRules
    
    // Markdown options
    Markdown MarkdownOptions
    
    // Media & Links
    Media MediaConfig
    Links LinkConfig

    // Explicitly allowed losses + unknown handling policies
    Loss    LossControl
    Unknown UnknownHandling
}
```

**Methods**:
- `applyDefaults() Config` — fills zero-value fields with defaults:
  - `Profile`: `"balanced-lossless"` (**default**)
  - `Dialect.Version`: `"adf-md-v1"`
  - `UnderlineStyle`: `"bold"`
  - `SubSupStyle`: `"ignore"`
  - `TextColor`: `"ignore"`
  - `BackgroundColor`: `"ignore"`
  - `Mention`: `"text"`
  - `TableFormat`: `"auto"`
  - `PanelStrategy`: `"github"`
  - `ExtensionHandling.Default`: `"warn"`
  - `HardBreakStrategy`: `"backslash"`
  - `Alignment`: `"ignore"`
  - `Markdown.BulletChar`: `'-'`
  - `Loss.Allowed`: empty map (no intentional information loss by default)
  - `Unknown.Node`: `"placeholder"`
  - `Unknown.Mark`: `"warn"`
  - `Unknown.Attr`: `"warn"`
- `Validate() error` — check enums are valid, ranges (HeadingShift 0-5), bullet char is one of '-*+', dialect version is supported, and `Loss.Allowed` contains known feature keys only

**Presets**:
- `NewBalancedLossless()`: **default** profile; optimized for AI/human readability and reverse conversion with minimal visual noise.
- `NewStrictLossless()`: maximum fidelity; preserve all reversible metadata and fail fast on unknowns.
- `NewReadableLossy()`: readability-first; only drops information listed in `Loss.Allowed`.

---

### Task 3: Create `converter/config_test.go`
**Test Cases**:

1. **TestApplyDefaults**: Verify all zero values get sensible defaults
2. **TestValidateValid**: Test valid configurations pass
3. **TestValidateInvalid**: Test invalid enum values, out-of-range heading shift, invalid bullet char fail
4. **TestExtensionRules**: Test that ByType lookup falls back to Default
5. **TestPresets**: Verify `NewBalancedLossless()`, `NewStrictLossless()`, and `NewReadableLossy()` set expected values
6. **TestLossControl**: Verify only explicitly allowed losses are dropped in lossy profile

---

### Task 4: Modify `converter/converter.go`
**Goal**: Update core converter for new config system.

**Implementation Details**:
* Update `New(config Config)` to:
  1. Call `applyDefaults()`
  2. Call `Validate()`
  3. Return `(*Converter, error)` on validation failure
  4. Store immutable config
* Update `convertNode` default case to use `ExtensionRules`:
  1. Check if node type is extension (`extension`, `inlineExtension`, `bodiedExtension`)
  2. Look up extension type in `config.Extensions.ByType`
  3. Fall back to `config.Extensions.Default`
  4. Handle according to strategy:
     - `ExtensionWarn`: log warning, render fallback text if available
     - `ExtensionStrip`: return empty string
     - `ExtensionText`: render only fallback text
     - `ExtensionJSON`: marshal node.Attrs and Content to JSON, wrap in code block with language `adf:macro`
* Unknown node/mark/attr handling is driven by `config.Unknown` policies (`error|warn|placeholder|strip`)
* `--strict` can be retained as a CLI compatibility alias that maps unknown policies to `error`

---

### Task 5: Modify `converter/marks.go` & `converter/inline.go`
**Goal**: Implement granular inline styling.

**Implementation Details**:
* Replace old `FormattingStrategy` with individual style checks
* **Underline**:
  - `UnderlineIgnore`: skip the mark
  - `UnderlineBold`: render as `**text**`
  - `UnderlineHTML`: render as `<u>text</u>`
* **SubSup**:
  - `SubSupIgnore`: skip the mark
  - `SubSupHTML`: render as `<sub>text</sub>` or `<sup>text</sup>`
  - `SubSupLaTeX`: render as `$_{text}$` or `$^{text}$`
* **TextColor**:
  - `ColorIgnore`: skip the mark
  - `ColorHTML`: render as `<span style="color: #RRGGBB">text</span>`
* **BackgroundColor**:
  - `ColorIgnore`: skip the mark
  - `ColorHTML`: render as `<span style="background-color: #RRGGBB">text</span>`
* **Mentions**:
  - `MentionText`: Render `@text`
  - `MentionLink`: Render `[@text](mention:id)`
  - `MentionHTML`: Render `<span data-mention-id="id">@text</span>`

---

### Task 6: Modify `converter/blocks.go`
**Goal**: Implement `PanelStrategy`, `HeadingShift`, `HardBreakStrategy`, `Alignment`.

**Implementation Details**:

* **`convertPanel()`**:
  * Extract `panelType` and optional `title` from node attrs
  * Map panelType to display type: "info" → "INFO", "warning" → "WARNING", etc.
  * Check `PanelStrategy`:
    - `PanelNone`: Standard blockquote `> ...`
    - `PanelBold`: `> **Type**\n> ...` (Map panelType to Title Case)
    - `PanelGitHub`: `> [!TYPE]\n> ...` (Map panelType to UPPERCASE)
    - `PanelTitle`: `> [!TYPE: Title]\n> ...` (include custom title if present)

* **`convertParagraph/Heading()`**:
  * Check `attrs.align`
  * If `AlignmentStyle == AlignHTML` and align is present:
    - Wrap content in `<div align="center">...</div>` or `<p align="...">`

* **`convertExpand()`**:
  * If expand has complex content: Render as HTML `<details><summary>title</summary>...`
  * If simple content: Render title in bold, content as normal paragraph
  * Note: ADF expand node has `title` attr

* **`convertHeading()`**:
  * Calculate new level: `newLevel = min(6, max(1, originalLevel + HeadingShift))`
  * Render appropriate number of `#` characters

* **`convertHardBreak()`**:
  * `HardBreakBackslash`: append `\\n`
  * `HardBreakDoubleSpace`: append `  \n` (two spaces + newline)
  * `HardBreakHTML`: append `<br>`

---

### Task 7: Modify `converter/tables.go`
**Goal**: Implement `TableFormat` with auto-detection.

**Implementation Details**:

* **`isComplexTable(tableNode)` helper**:
  * Returns true if table has:
    - Any cell with `colspan` > 1
    - Any cell with `rowspan` > 1
    - Any cell containing block nodes (lists, code blocks, nested tables)
  * Returns false for simple text-only tables

* **`convertTable()`**:
  * If `TableFormat == TableHTML`: Use HTML tables
  * If `TableFormat == TablePipe`: Use pipe tables (strip complex content)
  * If `TableFormat == TableAuto`:
    - Call `isComplexTable()`
    - If complex → HTML
    - If simple → Pipe

* **HTML Table Rendering**:
  * Use `<table>`, `<thead>`, `<tbody>`, `<tr>`, `<th>`, `<td>` tags
  * Preserve `colspan` and `rowspan` attributes
  * Render cell content recursively

* **Pipe Table Rendering** (existing logic):
  * Strip newlines from cell content (replace with space)
  * Escape pipe characters
  * Generate header separator line with alignment markers

---

### Task 8: Modify `converter/extensions.go` (New File)
**Goal**: Handle extension nodes with per-type rules.

**File**: `converter/extensions.go`

**Implementation Details**:

```go
func (c *Converter) convertExtension(node *ADFNode) (string, error) {
    // Determine extension type/identifier
    extType := getExtensionType(node) // e.g., "jira.issue", "confluence.toc"
    
    // Look up handling strategy
    strategy := c.config.Extensions.Default
    if specific, ok := c.config.Extensions.ByType[extType]; ok {
        strategy = specific
    }
    
    switch strategy {
    case ExtensionWarn:
        log.Printf("Warning: unhandled extension %s", extType)
        return c.getFallbackText(node), nil
    case ExtensionStrip:
        return "", nil
    case ExtensionText:
        return c.getFallbackText(node), nil
    case ExtensionJSON:
        return c.renderExtensionJSON(node), nil
    default:
        return "", fmt.Errorf("unknown extension strategy: %s", strategy)
    }
}

func (c *Converter) renderExtensionJSON(node *ADFNode) string {
    // Marshal node to JSON
    data, _ := json.Marshal(node)
    // Wrap in code block with adf:macro language
    return fmt.Sprintf("```adf:macro\n%s\n```", string(data))
}
```

---

### Task 9: Modify `converter/media.go`
**Goal**: Support BaseURL and media resolution.

**Implementation Details**:
* Update `convertMedia()` to:
  * Construct URL using `config.Media.BaseURL` as prefix
  * Handle `AltTextPolicy`:
    - `filename`: use file name from attrs
    - `occurrenceId`: use occurrence ID
    - `none`: empty alt text
* Support both internal and external media nodes
* If `DownloadImages` is true, return placeholder (CLI will handle actual download)

---

### Task 10: Modify `converter/lists.go`
**Goal**: Implement custom bullets.

**Implementation Details**:
* Use `config.Markdown.BulletChar` in `convertBulletList`
* Default to `-`
* Validate bullet char is one of: `-`, `*`, `+`

---

### Task 11: Modify `cmd/jac/main.go`
**Goal**: Expose granular flags.

**New Flags**:
```
# Inline styles (granular)
--underline-style          ignore|bold|html
--subsup-style             ignore|html|latex
--text-color-style         ignore|html
--background-color-style   ignore|html
--mention-style            text|link|html
--alignment-style          ignore|html

# Block formatting
--table-format             auto|pipe|html
--panel-strategy           none|bold|github|title
--heading-shift            N (0-5)
--hard-break-strategy      backslash|double-space|html

# Extensions
--extension-default        warn|strip|text|json
--extension-rule           type:strategy (can be used multiple times)
                           e.g., --extension-rule jira.issue:json --extension-rule confluence.toc:strip

# Markdown options
--bullet-char              - | * | +

# Media
--media-baseurl            URL prefix for media
--media-alt-policy         filename|occurrenceId|none

# Links
--link-baseurl             URL prefix for relative links

# General
--fidelity                 balanced-lossless|strict-lossless|readable-lossy
--dialect-version          adf-md-v1
--loss-allow               feature (repeatable)
                           e.g., --loss-allow textColor --loss-allow alignment
--unknown-node-policy      error|warn|placeholder|strip
--unknown-mark-policy      error|warn|strip
--unknown-attr-policy      error|warn|strip
--strict                   bool (compatibility alias for strict-lossless)
```

**Old Flags to Remove**:
- `--formatting-strategy` (replaced by granular flags)
- `--allow-html` (replaced by individual HTML options)

---

### Task 12: Tests and Verification

**Update existing tests**:
1. `converter/converter_test.go` - Update to use new Config struct
2. `converter/custom_table_test.go` - Update to use new Config struct
3. All existing golden files remain valid (using default config)

**New test files**:
1. `converter/config_test.go` - Config validation and defaults
2. `converter/marks_test.go` - Granular mark rendering
3. `converter/extensions_test.go` - Extension handling strategies
4. `converter/tables_test.go` - Auto-detection logic
5. `converter/fidelity_test.go` - Profile behavior and loss-control matrix

**Test coverage requirements**:
- All new config options have at least one test case
- ExtensionRules lookup and fallback tested
- Table auto-detection logic tested with both simple and complex tables
- All PanelStrategy options tested
- All InlineStyles combinations tested
- Mention and Alignment rendering tested
- Fidelity profiles and `Loss.Allowed` behavior tested
- Unknown policy behavior (`error|warn|placeholder|strip`) tested

---

### Task 13: Final Verification
**Goal**: Ensure all tests pass and no regressions exist.

**Steps**:
1. Run `make test` to ensure all tests pass.
2. Run `make lint`.
3. Run `make test-update` to regenerate golden files, then review diffs.
4. Manual CLI test with various flag combinations:
   ```bash
   go run ./cmd/jac --panel-strategy=title testdata/blocks/panel_title.json
   go run ./cmd/jac --table-format=html testdata/tables/complex.json
   go run ./cmd/jac --underline-style=html --subsup-style=latex testdata/marks/mixed.json
   go run ./cmd/jac --fidelity=readable-lossy --loss-allow=textColor --loss-allow=alignment testdata/fidelity/balanced_lossless.json
   ```
5. Verify no `c.config.AllowHTML` references remain, and unknown handling uses `config.Unknown` policies.
6. Verify no HTML comment generation logic exists.

**Acceptance Criteria**:
- All tests pass.
- Linting passes.
- No old config field references anywhere in the codebase.
- CLI help shows all new flags.
- Default behavior (`balanced-lossless`) produces AI-readable, human-readable, reverse-convertible Markdown.

---

## Success Criteria for Phase 6
The phase is complete when:
- [ ] `converter/config.go` defines all granular types, Config, Validate(), applyDefaults()
- [ ] `New()` returns `(*Converter, error)` with config validation
- [ ] `FidelityProfile` supports `balanced-lossless` (default), `strict-lossless`, `readable-lossy`
- [ ] `MarkdownDialectVersion` is set and validated for reverse parser compatibility
- [ ] Loss is explicit and opt-in via `Loss.Allowed`
- [ ] Granular inline controls work (underline, subsup, text color independently)
- [ ] TableFormat "auto" detects complexity and chooses pipe vs HTML
- [ ] PanelStrategy supports none/bold/github/title with title preservation
- [ ] ExtensionRules support per-type handling with JSON code block output
- [ ] HeadingShift shifts heading levels with clamping
- [ ] HardBreakStrategy supports backslash/double-space/html
- [ ] MediaConfig supports BaseURL and AltTextPolicy
- [ ] CLI has all new granular flags (old flags removed)
- [ ] All tests pass with new config API
- [ ] All new features have dedicated test coverage
- [ ] Zero references to old AllowHTML field
- [ ] Unknown handling is policy-driven (`Node`, `Mark`, `Attr`) instead of a single strict boolean
- [ ] No HTML comment generation code exists

---

## Files Modified Summary

| File | Action |
|------|--------|
| `converter/config.go` | **NEW** — granular types, Config, Validate, applyDefaults |
| `converter/config_test.go` | **NEW** — config validation and defaults tests |
| `converter/extensions.go` | **NEW** — extension handling with per-type rules |
| `converter/converter.go` | Remove old Config, change New() → error, apply fidelity/loss/unknown policies |
| `converter/marks.go` | Granular inline styles (underline, subsup, color) |
| `converter/blocks.go` | PanelStrategy, HeadingShift, HardBreakStrategy |
| `converter/tables.go` | TableFormat with auto-detection |
| `converter/media.go` | BaseURL, AltTextPolicy |
| `converter/lists.go` | Custom bullet char |
| `converter/inline.go` | Use new config structure |
| `cmd/jac/main.go` | New granular CLI flags |
| `converter/converter_test.go` | Migrate to new config |
| `converter/custom_table_test.go` | Migrate to new config |
| `testdata/**` | ~20-24 new golden file pairs (including fidelity profile fixtures) |

---

## Design Decisions Summary

1. **No HTML Comments**: Metadata preserved through structured Markdown syntax
2. **Granular Inline Control**: Each mark type (underline, subsup, color) configurable independently
3. **Auto Table Detection**: Automatically chooses pipe or HTML based on table complexity
4. **JSON Code Blocks for Extensions**: Readable, parseable, no token overhead from HTML comments
5. **Per-Type Extension Rules**: Different handling for different macro types
6. **Immutable Config**: Thread-safe, simple, passed once to constructor
7. **Balanced-Lossless Default**: Default output is AI-readable, human-readable, and reversible
8. **Explicit Loss Budget**: Lossy behavior is opt-in and limited to `Loss.Allowed`
9. **Versioned Dialect Contract**: `adf-md-v1` enables stable Markdown → ADF parsing

---

## Future Considerations (Not in Phase 6)

- **Frontmatter Integration**: CLI tool responsibility, but converter could support metadata hooks
- **Reverse Conversion (Markdown → ADF)**: Phase 7 - will parse versioned dialect (`adf-md-v1`) back to ADF
- **Sync Tool**: Phase 8 - will use this converter as library
- **Plugin System**: Allow users to register custom converters for specific node types
- **Streaming API**: Process large documents in chunks for memory efficiency
