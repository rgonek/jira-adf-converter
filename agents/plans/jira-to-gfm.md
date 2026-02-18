# Plan: Jira ADF to GFM Converter

## Goal
Build a Go library to convert Jira Atlassian Document Format (ADF) to GitHub Flavored Markdown (GFM), optimized for AI agent readability.

**Note**: This converter follows the [GitHub Flavored Markdown (GFM) specification](https://github.github.com/gfm/) as closely as possible. Where GFM lacks native support for certain ADF features (e.g., subscript/superscript, underline, panels), we provide fallback representations that prioritize readability and semantic preservation.

## Core Principles
1.  **Granularity**: Start with simple text, add complex nodes layer by layer.
2.  **Automated Testing**: Use data-driven tests (Golden Files) from the start.
3.  **No Data Loss**: Preserve all semantic information.
4.  **Configurable Output**:
    *   **Default**: Pure Markdown (no HTML tags). Use text formatting or symbolic representation for unsupported features.
    *   **Flag (`AllowHTML`)**: If enabled, use raw HTML (e.g., `<u>`, `<details>`) for features GFM doesn't support natively.
    *   **Flag (`Strict`)**:
        *   **Default (false)**: Render a placeholder (e.g., `[Unknown node: type]`) for unimplemented nodes.
        *   **True**: Return an error if an unknown node is encountered.
5.  **Ignore Empty Blocks**: Structural nodes (headings, blockquotes, paragraphs, panels) that contain no text or only whitespace should be ignored and output an empty string. This maintains a clean, readable document for AI agents and avoids noise.


## Development Phases

### Phase 1: Infrastructure & Basic Text
*   Initialize Go module (package name: `github.com/rgonek/jira-adf-converter`).
*   **Architecture**:
    *   Create `converter/` package for the library (AST structs and logic).
    *   Create `cmd/jac/` for the CLI tool (reads input file, outputs converted markdown to stdout).
*   Define core AST structs for ADF `Doc`, `Node`, `Mark` in `converter/ast.go`.
*   Implement `Converter` struct with configuration (e.g., `Converter{AllowHTML: bool, Strict: bool}`) in `converter/converter.go`.
*   **CLI Flags**:
    *   `--allow-html`: Enable HTML output for unsupported GFM features.
    *   `--strict`: Return error on unknown nodes (instead of placeholder).
*   **Test Harness**: 
    *   Use `testify/assert` for assertions and cleaner test output.
    *   Create a test runner that reads `testdata/*.json`, converts it, and compares it to `testdata/*.md`.
    *   Support two modes:
        *   **Normal mode** (`go test`): Fails on mismatch, requires manual `.md` file updates.
        *   **Update mode** (`go test -update`): Automatically overwrites `.md` files with actual output for review.
*   **Implementation**:
    *   `doc` (root node)
    *   `paragraph`
    *   `text`
    *   **Marks**: `strong` (**bold**), `em` (*italic*), `strike` (~~strike~~), `code` (`code`).

### Phase 2: Structural Elements & Configuration
*   **Nodes**:
    *   `heading` (# H1-H6)
    *   `blockquote` (> Text)
    *   `rule` (---)
    *   `hardBreak` (\  \n)
*   **Marks (with Config)**:
    *   `link` ([text](url))
    *   `subsup`:
        *   Default: Just the text (or `^text` for sup).
        *   HTML: `<sub>text</sub>` / `<sup>text</sup>`.
    *   `underline`:
        *   Default: Just the text.
        *   HTML: `<u>text</u>`.

### Phase 3: Lists (including Tasks) & Code Blocks
*   **Nodes**:
    *   `codeBlock` (```lang)
    *   `bulletList` (- item)
    *   `orderedList` (1. item)
    *   `listItem` (Support nested content)
    *   `taskList` (Container for tasks)
    *   `taskItem`:
        *   State `TODO`: `- [ ] Item`
        *   State `DONE`: `- [x] Item`

### Phase 4: Complex Layouts (Tables & Panels) (Completed)
*   **Nodes**:
    *   `table`: Map to GFM tables.
        *   **Requirement**: Escape pipe characters (`|`) to `\|` within cells.
        *   **Requirement**: Preserve indentation for nested lists when flattening to `<br>`.
    *   `panel`: Map to Blockquote with semantic label (e.g., `> **Info**: ...`).
    *   `decisionList` / `decisionItem`: Map to Blockquote with state indicators.
*   **Post-Implementation Fixes**:
    *   Removed unused `convertTableRow` function and its dispatcher registration (dead code elimination).
    *   Extracted common blockquoting logic into `blockquoteContent()` helper to eliminate duplication in `convertPanel` and `convertDecisionItemContent`.
    *   Fixed nested list indentation loss in table cells when `AllowHTML=false` (`tables.go:148` - changed `TrimRight(line, " \t\r\n")` to `TrimRight(line, "\n")`).
    *   Documented unused `isHeader` parameter in `convertTableCell` (API consistency, future extensibility).
    *   Added comment warning about pipe escaping constraints to prevent double-escaping.


### Phase 5: Rich Media & Interactive Elements (Completed)
*   **Nodes**:
    *   `expand` / `nestedExpand`:
        *   Default: `> **Expand: {Title}** \n > {Content}`
        *   HTML: `<details><summary>{Title}</summary>{Content}</details>`
    *   `emoji`: Convert to unicode or shortcode.
    *   `mention`: Convert to `[Name - @id]`.
    *   `status`: Convert to `[Status: TEXT]`.
    *   `media`: `![alt](url)` or `[Media: type]`.

### Phase 6: Configuration/Params System
*   **Breaking change**: `New()` returns `(*Converter, error)`, old `AllowHTML`/`Strict` removed entirely
*   **New config options**:
    *   `FormattingStrategy`: `simple` (default) | `html` | `latex`
    *   `UnrecognizedNodeHandling`: `warn` | `strip` | `stringify` | `error`
    *   `HeadingShift`: integer (0+), shifts heading levels
    *   `HardBreakStrategy`: `backslash` | `double-space` | `html`
    *   `MediaConfig`: `BaseURL`, `AltTextPolicy`, `DownloadImages`, `MediaResolver` interface
*   **New marks**: `textColor` (HTML: `<span>`, simple: stripped)
*   **Frontmatter**: Not in converter lib — CLI's responsibility (will include Confluence metadata)
*   See `agents/plans/phase6-detailed.md` for full implementation plan.

## Testing Strategy
*   **Framework**: Use `github.com/stretchr/testify/assert` for test assertions.
*   **Location**: `testdata/` directory.
*   **Format**: Pairs of files `*.json` and `*.md`.
*   **Feature**: Add specific test cases for `_html` variants to test the `AllowHTML` flag.
*   **Golden File Workflow**:
    *   `go test`: Fails on mismatch between actual output and `.md` files.
    *   `go test -update`: Auto-updates `.md` files with actual output (requires git diff review).

## Configuration
*   **Library**: Configuration via `Config` struct with strategy enums and option objects.
*   **CLI**: Configuration via command-line flags (strategy flags replacing old `--allow-html`, `--strict`).
*   See Phase 6 for the comprehensive configuration system.

## Next Step
*   Phase 6 Complete.
*   See `agents/plans/phase6-detailed.md` for details.
