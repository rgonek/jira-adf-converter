# Phase 2: Structural Elements & Configuration

## Overview
This phase expands the converter's capabilities to support structural document elements (headings, blockquotes, horizontal rules, hard breaks) and advanced text formatting (links, subscript/superscript, underline). It also activates the `AllowHTML` configuration flag to control output format for features not natively supported in GFM.

**Note**: This implementation follows GitHub Flavored Markdown (GFM) specification as closely as possible. Where GFM lacks native support for certain features (e.g., subscript/superscript, underline), we provide fallback representations that prioritize readability and semantic preservation.

## Deliverables
1. Support for structural nodes: `heading`, `blockquote`, `rule`, `hardBreak`
2. Support for advanced marks: `link`, `subsup`, `underline`
3. Implementation of `AllowHTML` logic for `subsup` and `underline`
4. Expanded test suite with Phase 2 test cases in appropriate `testdata/` subdirectories
5. Updated `converter/` package with new node and mark handlers

---

## Step-by-Step Implementation Plan

### Task 1: Create Phase 2 Test Data
**Goal**: Create Golden Files for Phase 2 features before implementation (TDD).

**Directories**: 
- Structural nodes: `testdata/nodes/`
- Marks: `testdata/marks/`

**Test Cases to Create**:

1.  **heading.json/md**
    *   Input: `heading` nodes with levels 1 through 6.
    *   Expected: `# H1`, `## H2`, etc.

2.  **heading_with_marks.json/md**
    *   Input: `heading` containing text with marks (bold, italic, code).
    *   Expected: `## **Bold** heading with *italic*`.

3.  **heading_empty.json/md**
    *   Input: `heading` node with no text content or empty content array.
    *   Expected: Empty string (ignored).

4.  **blockquote.json/md**
    *   Input: `blockquote` containing a `paragraph`.
    *   Expected: `> Text content` (ensure proper spacing).

4.  **blockquote_multiline.json/md**
    *   Input: `blockquote` containing multiple paragraphs.
    *   Expected: Each line prefixed with `> `, blank lines within blockquote also prefixed.

5.  **blockquote_empty.json/md**
    *   Input: Empty `blockquote` node.
    *   Expected: Empty string (ignored).

6.  **nested_blockquote.json/md**
    *   Input: `blockquote` inside `blockquote`.
    *   Expected: `>> Text` (no space between `>` characters, per standard GFM).

7.  **rule.json/md**
    *   Input: `rule` node between paragraphs.
    *   Expected: `---`

8.  **hard_break.json/md**
    *   Input: `paragraph` with `text`, `hardBreak`, `text`.
    *   Expected: `Line 1\` (with actual newline following) then `Line 2` (backslash at end of line + newline for explicit break, per GFM spec).

9.  **link.json/md**
    *   Input: `text` with `link` mark (`href` attribute).
    *   Expected: `[text](url)`

10. **link_with_title.json/md**
    *   Input: `text` with `link` mark containing both `href` and `title` attributes.
    *   Expected: `[text](url "title")`

11. **link_empty_text.json/md**
    *   Input: `link` mark with empty text content.
    *   Expected: `[](url)` or handle gracefully.

12. **link_missing_href.json/md**
    *   Input: `link` mark with no `href` attribute.
    *   Expected: Plain text output (link formatting dropped).

13. **link_title_with_quotes.json/md**
    *   Input: `link` mark with title containing double quotes (e.g., `title="He said \"hello\""`).
    *   Expected: `[text](url "He said \"hello\"")` (escape inner quotes with backslash).

14. **formatting_html.json/md** (Run with `-allow-html`)
    *   Input: `subsup` (sub and sup types), `underline`.
    *   Expected: `<sub>sub</sub>`, `<sup>sup</sup>`, `<u>underline</u>`.

15. **formatting_plain.json/md** (Run with default config)
    *   Input: Same as above.
    *   Expected: 
        *   Superscript: `^text` (carat prefix)
        *   Subscript: plain text only (no special indicator - semantic information is lost, but this preserves readability without conflicting with GFM syntax)
        *   Underline: plain text only

---

### Task 2: Implement Structural Nodes
**Goal**: Add handlers for `heading`, `rule`, `hardBreak`.

**File**: `converter/converter.go` (and potentially new files if refactoring is needed)

**Implementation Details**:
*   **Heading**:
    *   Extract `level` from attributes (default to 1 if missing/invalid).
    *   Output `strings.Repeat("#", level) + " " + content`.
    *   Ensure proper spacing around headings: blank line before AND after heading (except at document start/end).
    *   Handle nested marks correctly (bold, italic, etc. within heading text).
    *   **Empty Headings**: Ignore if content is empty (return empty string).
*   **Rule**:
    *   Output `---` surrounded by newlines (`\n\n---\n\n`).
*   **HardBreak**:
    *   Output `\` followed by `\n` (backslash at end of line, then actual newline character).
    *   This follows the GFM specification for hard line breaks.

**Acceptance Criteria**:
*   `make test` fails initially (due to missing implementation).
*   After implementation, heading, rule, and hard_break tests pass.

---

### Task 3: Implement Blockquote
**Goal**: Add handler for `blockquote` with correct indentation.

**Implementation Details**:
*   **Blockquote**:
    *   Process child content recursively.
    *   Prefix *every line* of the result with `> `.
    *   Ensure it handles multiline content correctly (blank lines within blockquote also get `> ` prefix).
    *   **Empty Blockquotes**: Ignore if content is empty (return empty string).
    *   Nested blockquotes should result in `>>` (no space between markers, per standard GFM).

**Acceptance Criteria**:
*   Blockquote tests pass.
*   Nested blockquotes render with proper spacing (`> > `).
*   Multiline blockquotes handle internal blank lines correctly.

---

### Task 4: Implement Link Mark
**Goal**: Support hyperlinks with optional titles.

**Implementation Details**:
*   **Link**:
    *   Extract `href` (required) and optional `title` from attributes.
    *   If `href` is missing or empty: output plain text (drop link formatting).
    *   Format: `[text](href "title")` (if title exists and is non-empty) or `[text](href)`.
    *   Handle edge cases:
        *   Empty link text: `[](url)` is valid GFM.
        *   Title with quotes: escape inner double quotes with backslash: `"He said \"hello\""`.
        *   Title with backslash: escape backslashes: `"path\\to\\file"`.

**Acceptance Criteria**:
*   Link tests pass.
*   Links with titles render correctly.
*   Links without titles render correctly.
*   Edge cases handled: empty text, missing href, quotes in title.

---

### Task 5: Implement HTML Configuration & Formatting Marks
**Goal**: Support `subsup` and `underline` with `AllowHTML` flag logic.

**Implementation Details**:
*   **Converter Struct**: Ensure `config.AllowHTML` is accessible in mark handlers.
*   **Underline**:
    *   If `AllowHTML`: Return `<u>`, `</u>`.
    *   Else: Return empty strings (just text, no special formatting).
*   **SubSup**:
    *   Extract `type` ("sub" or "sup").
    *   If `AllowHTML`: Return `<sub>`/`</sub>` or `<sup>`/`</sup>`.
    *   Else (plain mode):
        *   If `sup`: Return `^`, empty string (e.g., `^text`).
        *   If `sub`: Return empty strings (just text, no special indicator).
    *   **Note on subscript**: In plain mode, subscript semantic information is intentionally lost to avoid conflicts with GFM syntax (`~` is strikethrough, `_` can trigger italic/bold). This trades semantic precision for format compatibility and readability.

**Test Organization**:
*   Place tests in `testdata/marks/` with naming convention `*_html.json` for HTML config tests.

**Acceptance Criteria**:
*   `formatting_html` tests pass.
*   `formatting_plain` tests pass.
*   Subscript in plain mode outputs text only (no indicator).
*   Superscript in plain mode outputs with `^` prefix.

---

### Task 6: Refactor
**Goal**: Improve code organization before final verification. Split large conversion functions into dedicated methods for better maintainability and readability.

**Why before verification**: Refactoring earlier prevents accumulating technical debt and ensures the codebase is clean before adding more complexity in future phases.

**Steps**:
1.  Split `convertNode` into dedicated methods for each node type:
    *   `convertHeading`
    *   `convertBlockquote`
    *   `convertRule`
    *   `convertHardBreak`
    *   `convertParagraph`
    *   `convertText`
    *   etc.
2.  Consider splitting `convertMark` similarly if it becomes complex:
    *   `convertLinkMark`
    *   `convertSubSupMark`
    *   etc.
3.  Ensure proper method signatures and error handling.
4.  Run `make test` to verify all tests still pass after refactoring.
5.  Run `make lint` to verify code quality.

**Acceptance Criteria**:
*   Each node type has its own dedicated conversion method.
*   Methods have clear signatures and responsibility boundaries.
*   All tests pass after refactoring.
*   Code passes linting.
*   No behavioral changes (output remains identical).

---

### Task 7: Final Verification
**Goal**: Ensure all tests pass and no regressions exist.

**Steps**:
1.  Run `make test` to ensure all Phase 1 and Phase 2 tests pass.
2.  Run `make lint`.
3.  Review any test failures and fix issues.
4.  Manually review a few test outputs to ensure quality.

**Acceptance Criteria**:
*   All Phase 1 tests continue to pass (no regression).
*   All Phase 2 tests pass (all 16 test cases).
*   Linting passes with no errors.
*   Code is well-organized with dedicated methods per node type.

---

## Success Criteria for Phase 2
The phase is complete when:
- [x] All Phase 2 structural nodes (`heading`, `blockquote`, `rule`, `hardBreak`) convert correctly.
- [x] Headings support nested inline marks (bold, italic, code).
- [x] Headings have proper spacing (blank line before and after, except at document boundaries).
- [x] Empty headings are ignored (return empty string).
- [x] Blockquotes handle multiline content and nested blockquotes with proper GFM formatting (`>>`).
- [x] Empty blockquotes are ignored (return empty string).
- [x] Hard breaks use backslash + newline format (per GFM spec).
- [x] Links are rendered as standard Markdown with optional title support.
- [x] Link edge cases handled: empty text, missing href, quotes in title.
- [x] `AllowHTML` flag correctly toggles between HTML tags and plain text for `subsup` and `underline`.
- [x] Superscript in plain mode uses `^` prefix; subscript uses plain text only.
- [x] All new test cases pass (16 test cases total).
- [x] No regression in Phase 1 features.
- [x] Code passes linting.
- [x] Code is refactored with dedicated methods for each node type (refactoring done before final verification).
- [x] `marksEqual()` function compares both `Type` and attributes (href/title for links, type for subsup) to handle adjacent text nodes with different mark configurations.

---

## Implementation Notes

### Mark Comparison
The `marksEqual()` function must compare not just the mark `Type`, but also relevant attributes:

- **link**: Compare `href` and `title` attributes
- **subsup**: Compare `type` attribute ("sub" vs "sup")

This ensures that adjacent text nodes with different URLs (e.g., `[text1](url1)[text2](url2)`) are handled correctly, rather than incorrectly treating them as one continuous link.

Example scenario:
```json
// Input: two links with different URLs adjacent to each other
[
  {"type": "text", "text": "Google", "marks": [{"type": "link", "attrs": {"href": "https://google.com"}}]},
  {"type": "text", "text": "Bing", "marks": [{"type": "link", "attrs": {"href": "https://bing.com"}}]}
]

// Expected output: separate links
[Google](https://google.com)[Bing](https://bing.com)

// Without proper attribute comparison, would incorrectly output:
// [GoogleBing](https://google.com) - wrong!
```
