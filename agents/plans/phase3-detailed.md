# Phase 3: Lists & Code Blocks

## Overview
This phase focuses on implementing list structures (bullet, ordered, task) and code blocks. These are container nodes that require careful handling of indentation and nesting to produce valid GFM.

**Note**: This implementation follows GitHub Flavored Markdown (GFM) specification as closely as possible.

## Deliverables
1. Support for `codeBlock` with language syntax highlighting.
2. Support for `bulletList` and `orderedList` (nested and mixed).
3. Support for `taskList` and `taskItem` with state (`TODO`/`DONE`).
4. Robust indentation logic for nested content.
5. Expanded test suite in `testdata/codeblocks/` and `testdata/lists/`.

---

## Step-by-Step Implementation Plan

### Task 1: Create Phase 3 Test Data (Done)
**Goal**: Create Golden Files for Phase 3 features before implementation (TDD).

**Directories**: 
- Code Blocks: `testdata/codeblocks/`
- Lists: `testdata/lists/`

**Test Cases to Create**:

1.  **testdata/codeblocks/basic.json/md**
    *   Input: `codeBlock` with `language` attribute (e.g., "go") and text content.
    *   Expected: 
        ```markdown
        ```go
        content
        ```
        ```

2.  **testdata/codeblocks/empty.json/md**
    *   Input: `codeBlock` with no content or only whitespace.
    *   Expected: Empty string (Core Principle #5 - ignore empty blocks).
    *   Note: According to empty block rule, empty code blocks produce no output.

2.  **testdata/lists/bullet.json/md**
    *   Input: `bulletList` containing `listItem`s. Includes nested lists.
    *   Expected: List items starting with `- ` and properly indented nested items.

3.  **testdata/lists/ordered.json/md**
    *   Input: `orderedList` with `order` attribute.
    *   Expected: List items starting with `1. `, `2. `, etc.

4.  **testdata/lists/task.json/md**
    *   Input: `taskList` with `taskItem` nodes having `state` "TODO" or "DONE".
    *   Expected: Items starting with `- [ ] ` or `- [x] `.

5.  **testdata/lists/mixed.json/md**
    *   Input: Lists containing code blocks, paragraphs, and other lists nested deep.
    *   Expected: Correct indentation for all nested content relative to the parent item.
    *   **Progressive approach**: Start with simpler mixed cases (bullet list with nested ordered list), then add complexity (lists with code blocks, multiple paragraphs).

### TaskList Structure
Task lists in ADF have a specific structure with `taskList` as container and `taskItem` as children:

```json
{
  "type": "taskList",
  "attrs": { "localId": "..." },
  "content": [
    {
      "type": "taskItem",
      "attrs": {
        "localId": "...",
        "state": "TODO"  // or "DONE"
      },
      "content": [...]
    }
  ]
}
```

### Task 2: Implement Helper Functions (Done)
**Goal**: Add necessary utility methods for handling indentation.

**File**: `converter/converter.go`

**Implementation Details**:
*   **`indent(content, marker)`**: 
    *   Applies **uniform indentation** to all content within a list item.
    *   First line: prefixed with the list marker (e.g., `- `, `1. `, `- [ ] `).
    *   Subsequent lines: prefixed with spaces matching the marker's length:
        *   `- ` (2 chars) → indent with 2 spaces
        *   `1. ` (3 chars) → indent with 3 spaces
        *   `- [ ] ` (6 chars) → indent with 6 spaces
    *   **Note**: Even fenced code blocks are indented for consistent visual hierarchy and unambiguous parsing by both humans and AI.
    *   **Rationale**: Simple uniform indentation provides clear visual hierarchy for humans and consistent, predictable structure for AI parsing. Avoids ambiguity about whether content belongs to a list item.

**Acceptance Criteria**:
*   Helper function exists and correctly handles multiline strings.
*   All nested content (paragraphs, code blocks, sub-lists) receives uniform indentation based on parent marker width.

### Task 3: Implement Code Blocks (Done)
**Goal**: Add handler for `codeBlock`.

**File**: `converter/converter.go`

**Implementation Details**:
*   **`codeBlock`**:
    *   Extract `language` attribute.
    *   Output fenced code block syntax (```).
    *   Ensure a blank line follows the block.
    *   **When nested in lists**: Apply uniform indentation from the parent list item marker (see Task 2). This provides clear visual hierarchy and consistent structure for AI parsing.

**Acceptance Criteria**:
*   Code block tests pass.
*   Language attribute is respected.

### Task 4: Implement Lists (Done)
**Goal**: Add handlers for all list types and items.

**File**: `converter/converter.go`

**Implementation Details**:
*   **`bulletList`**:
    *   Iterate over `content`.
    *   Validate `listItem` child type in `Strict` mode.
    *   Apply marker `- ` using indentation logic.
    *   *Note*: `-` chosen over `*` for consistency with task lists.
*   **`orderedList`**:
    *   Iterate over `content`.
    *   Validate `listItem` child type in `Strict` mode.
    *   Determine starting index from `order` attribute (default 1).
    *   Apply marker `N. ` (incrementing) using indentation logic.
*   **`taskList`**:
    *   Validate `taskItem` child type in `Strict` mode.
    *   Treat similar to bullet list but delegates to `taskItem`.
*   **`taskItem`**:
    *   Check `state` attribute (`TODO` vs `DONE`).
    *   Convert inline content directly (avoiding paragraph spacing).
    *   Return content prefixed with `- [ ] ` or `- [x] `.
*   **`listItem`**:
    *   Convert children and join with blank lines (`\n\n`) to preserve paragraph separation within the list item.

**Acceptance Criteria**:
*   All list tests pass (bullet, ordered, task, mixed).
*   Nesting works correctly with proper indentation.

### Task 5: Final Verification (Done)
**Goal**: Ensure all tests pass and no regressions exist.

**Steps**:
1.  Run `make test` to ensure all tests pass.
2.  Run `make lint`.
3.  Review any test failures and fix issues.

**Acceptance Criteria**:
*   All tests pass.
*   Linting passes.

---

## Success Criteria for Phase 3
The phase is complete when:
- [x] `codeBlock` converts correctly with language support.
- [x] `codeBlock` with empty/whitespace content produces empty string.
- [x] `bulletList` uses `- ` marker and handles nesting.
- [x] `orderedList` uses incrementing numbers (`1.`, `2.`, etc.).
- [x] `taskList` renders checkbox syntax (`- [ ]`, `- [x]`).
- [x] Multiline content within list items is properly indented.
- [x] Mixed and complex nested structures render valid GFM (including progressive test cases from simple to complex).
- [x] All new tests pass.
- [x] No regressions in Phase 1 & 2 features.

---

## Next Phase Preview
Phase 4 will add:
- `table` (Map to GFM tables)
- `panel` (Map to Blockquote with semantic label)
