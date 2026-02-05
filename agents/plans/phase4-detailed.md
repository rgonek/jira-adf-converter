# Phase 4: Complex Layouts (Tables, Panels & Decisions)

## Overview
This phase implements ADF's table structure, panel nodes, and decision lists. Tables are mapped to GFM tables with support for both header and data cells. Panels are converted to blockquotes with semantic type prefixes. Decision lists are rendered as blockquotes with decision state indicators. All nodes support full nested content including paragraphs, lists, marks, and other inline elements.

**Note**: This implementation follows GitHub Flavored Markdown (GFM) specification as closely as possible.

## Deliverables
1. Table support with strict `tableHeader` vs `tableCell` detection (forcing empty header with column count if first row is data)
2. Panel support with semantic type labels (info, note, success, warning, error)
4. Decision list support with state indicators (DECIDED/UNDECIDED) as single continuous blockquote
5. Full preservation of block-level content in table cells (lists, code blocks, panels, etc.)
6. Comprehensive test coverage for all node combinations
7. Empty node handling (ignore empty panels/tables per Core Principle #5)
8. Bug fixes: Escape pipe characters in table cells and preserve nested list indentation

---

## Step-by-Step Implementation Plan

### Task 1: Create Phase 4 Test Data (Completed)
**Goal**: Create Golden Files for Phase 4 features before implementation (TDD).

**Directories**: 
- Tables: `testdata/tables/`
- Panels: `testdata/panels/`
- Decisions: `testdata/decisions/`
- Complex: `testdata/complex/` (integration tests)

**Test Cases to Create**:

#### Tables (`testdata/tables/`)

1. **table_simple.json/md** - 3x3 table with all `tableCell` nodes
   - Expected: GFM table with first row as header
   ```markdown
   | a | b | c |
   | --- | --- | --- |
   | 1 | 2 | 3 |
   | 4 | 5 | 6 |
   ```

2. **table_with_headers.json/md** - Table with `tableHeader` in first row
   - Expected: Proper header/data distinction
   ```markdown
   | Name | Age | City |
   | --- | --- | --- |
   | Alice | 30 | NYC |
   | Bob | 25 | LA |
   ```

3. **table_formatted_cells.json/md** - Cells with marks (strong, em, code, strike)
   ```markdown
   | **Bold** | *Italic* | `Code` |
   | --- | --- | --- |
   | ~~Strike~~ | Plain | **_Both_** |
   ```

4. **table_empty_cells.json/md** - Table with empty cells
   ```markdown
   | A |  | C |
   | --- | --- | --- |
   |  | B |  |
   ```

5. **table_multiline_cells.json/md** - Cells with multiple paragraphs
   ```markdown
   | Cell 1 | Cell 2 |
   | --- | --- |
   | Line 1<br>Line 2 | Single line |
   ```

6. **table_complex_content.json/md** - Cells with lists, code blocks, links
   ```markdown
   | Header | Content |
   | --- | --- |
   | List | - Item 1<br>- Item 2 |
   | Link | [Example](https://example.com) |
   ```

7. **table_single_column.json/md** - Single column table
   ```markdown
   | Column |
   | --- |
   | Row 1 |
   ```

8. **table_single_row.json/md** - Single row (headers only)
   ```markdown
   | A | B | C |
   | --- | --- | --- |
   ```

10. **table_no_headers.json/md** - Table with only `tableCell` in first row
    - Expected: Always check if first row contains `tableHeader` nodes, and if not, insert an empty header row with column count matching the data
    ```markdown
    |  |  |
    | --- | --- |
    | Data 1 | Data 2 |
    | Data 3 | Data 4 |
    ```



#### Panels (`testdata/panels/`)

1. **panel_info.json/md** - Info panel
   ```markdown
   > **Info**: Test info panel
   ```

2. **panel_note.json/md** - Note panel
   ```markdown
   > **Note**: Test note panel
   ```

3. **panel_success.json/md** - Success panel
   ```markdown
   > **Success**: Test success panel
   ```

4. **panel_warning.json/md** - Warning panel
   ```markdown
   > **Warning**: Test warning panel
   ```

5. **panel_error.json/md** - Error panel
   ```markdown
   > **Error**: Test error panel
   ```

6. **panel_multiline.json/md** - Multiple paragraphs
   ```markdown
   > **Info**: First paragraph
   > 
   > Second paragraph
   ```

7. **panel_nested_content.json/md** - Lists, code blocks, marks
   ```markdown
   > **Note**: Text with **bold** and *italic*
   > 
   > - List item 1
   > - List item 2
   > 
   > ```
   > code block
   > ```
   ```

8. **panel_no_type.json/md** - Panel without `panelType` → plain blockquote

9. **panel_empty.json/md** - Empty panel → empty string

#### Decisions (`testdata/decisions/`)

1. **decision_decided.json/md** - DECIDED state
   ```markdown
   > **✓ Decision**: test decision
   ```

2. **decision_undecided.json/md** - UNDECIDED state
   ```markdown
   > **? Decision**: test undecided
   ```

3. **decision_list_multiple.json/md** - Multiple items (single continuous blockquote)
    ```markdown
    > **✓ Decision**: First decision
    > 
    > **? Decision**: Second decision
    ```

4. **decision_formatted.json/md** - With marks
   ```markdown
   > **✓ Decision**: This is a **bold** decision
   ```

5. **decision_multiline.json/md** - Multiple paragraphs
   ```markdown
   > **✓ Decision**: First paragraph
   > 
   > Second paragraph
   ```

6. **decision_no_state.json/md** - Missing state → generic prefix

7. **decision_empty.json/md** - Empty list → empty string

---

### Task 2: Implement Table Conversion (Completed)
**Goal**: Convert ADF table nodes to GFM tables.

**File**: `converter/tables.go` (created)

**Implementation Details**:

- **`table` Node**: Process all `tableRow` nodes, generate GFM table format
- **`tableRow` Node**: Contains `tableHeader` or `tableCell` nodes
- **`tableHeader`/`tableCell` Nodes**: Process nested content recursively
- **Cell Content**: Preserve full block-level content (paragraphs with `<br>`, full multi-line lists, code blocks, etc.)
- **Empty cells**: Render as empty string between pipes
- **Empty tables**: Return empty string
- **No Headers**: If first row contains only `tableCell` nodes, insert empty header row with column count

**Table Assembly**:
1. Process first row for column count
2. Build header row: `| col1 | col2 |`
3. Build separator: `| --- | --- |`
4. Build data rows
5. Join with `\n`, add trailing `\n\n`

**Acceptance Criteria**:
- Tables with `tableHeader` vs `tableCell` render correctly
- Tables with no headers get empty header row with correct column count
- Multi-paragraph cells use `<br>` separator
- Nested lists preserved as full multi-line lists in cells
- Code blocks, panels, and other block content preserved in cells
- Empty tables output empty string
- All table test cases pass

---

### Task 3: Implement Panel Conversion (Completed)
**Goal**: Convert ADF panel nodes to blockquotes with type prefixes.

**File**: `converter/blocks.go`

**Implementation Details**:

- **`panel` Node**: Extract `panelType` from attrs
- **Type Mapping**:
  - `info` → `**Info**: `
  - `note` → `**Note**: `
  - `success` → `**Success**: `
  - `warning` → `**Warning**: `
  - `error` → `**Error**: `
  - Missing → no prefix
- **Blockquote Assembly**: Process content, prefix all lines with `> `, add type label to first line
- **Empty panels**: Return empty string

**Acceptance Criteria**:
- All panel types render with correct prefix
- Multi-paragraph panels work correctly
- Nested content properly blockquoted
- Empty panels output empty string
- Missing `panelType` renders as plain blockquote

---

### Task 4: Implement Decision List Conversion (Completed)
**Goal**: Convert decision lists and items to a single continuous blockquote.

**File**: `converter/blocks.go`

**Implementation Details**:

- **`decisionList` Node**: Process `decisionItem` nodes, join them into a single continuous blockquote.
- **`decisionItem` Node**: Extract `state` from attrs.
- **State Mapping**:
  - `DECIDED` → `**✓ Decision**: `
  - `UNDECIDED` → `**? Decision**: `
  - Missing → `**Decision**: `
- **Blockquote Assembly**: All items in a single continuous blockquote. Multiple items separated by quoted blank line (`> `).
- **Empty items**: Skip
- **Empty lists**: Return empty string

**Acceptance Criteria**:
- State indicators applied correctly (✓, ?)
- All items rendered in single continuous blockquote with blank quoted lines between items
- Multi-paragraph decisions work
- Empty items skipped
- Empty lists output empty string

---

### Task 5: Update Node Dispatcher (Completed)
**Goal**: Register new node types in `convertNode` switch.

**File**: `converter/converter.go`

Add cases for: `table`, `tableRow`, `tableHeader`, `tableCell`, `panel`, `decisionList`, `decisionItem`

**Acceptance Criteria**:
- All new node types registered
- Unknown node handling still works
- No regressions

---

### Task 6: Handle Complex Table Cell Content (Completed)
**Goal**: Properly preserve block-level content inside table cells.

**File**: `converter/tables.go`

**Implementation**:
- Create `convertCellContent()` method that preserves full block structure
- Paragraphs: Join with `<br>` separator within cell
- Lists: Preserve as full multi-line lists (bullet/ordered/task lists)
- Code blocks: Preserve with backticks (or `<pre><code>` if `AllowHTML` is true)
- Panels: Preserve as blockquotes
- Blockquotes: Preserve as nested blockquotes
- Decisions: Preserve with state indicators
- Marks: Always preserve
- **Note**: GFM tables support block-level content in cells, so preserve everything

**Acceptance Criteria**:
- Cells with paragraphs render correctly (joined with `<br>`)
- Lists preserved as full multi-line lists in cells
- Code blocks, panels, blockquotes preserved
- No content is flattened to inline format
- Proper line breaks between block elements in cells

---

### Task 7: Final Verification (Completed)
**Goal**: Ensure all tests pass and no regressions.

**Steps**:
1. Run `make test`
2. Run `make lint`
3. Review edge cases
4. Verify phase 1-3 tests still pass

**Acceptance Criteria**:
- All Phase 4 tests pass
- No regressions in previous phases
- Linting passes
- Empty nodes output empty string

---

### Task 8: Table Fixes (Completed)
**Goal**: Fix issues with pipe characters and list indentation in tables.

**File**: `converter/tables.go`

**Implementation**:
- Escape `|` as `\|` in `convertCellContent`.
- Replace `strings.TrimSpace` with `strings.TrimRight` in list processing within cells.

**Bug Fixes Applied** (Post-Implementation Code Review):

1. **Fixed: Nested list indentation lost in table cells** (`tables.go:148`)
   - **Issue**: `TrimRight(line, " \t\r\n")` removed all trailing whitespace including horizontal indentation, destroying nested list structure when `AllowHTML=false`.
   - **Fix**: Changed to `TrimRight(line, "\n")` to preserve horizontal whitespace (indentation) while removing newlines.
   - **Impact**: Nested list items in table cells now correctly retain their indentation in both HTML and non-HTML modes.

2. **Documented: Dead `isHeader` parameter** (`tables.go:112-115`)
   - **Issue**: `convertTableCell` accepted `isHeader bool` parameter but never used it.
   - **Fix**: Added documentation explaining why the parameter exists (API consistency, future extensibility) and noting that GFM tables don't require different header/data cell processing.

3. **Documented: Pipe escaping constraints** (`tables.go:236`)
   - **Issue**: Potential for double-escaping if child converters pre-escaped pipes.
   - **Fix**: Added comment warning that child converters must NOT pre-escape pipes.

**Acceptance Criteria**:
- Pipe characters in cells do not break table structure.
- Nested lists in cells retain their indentation.
- All regression tests pass.

---

## Success Criteria for Phase 4

- [x] All table nodes implemented (`table`, `tableRow`, `tableHeader`, `tableCell`)
- [x] Tables distinguish headers from data cells
- [x] Tables with no headers get empty header row with correct column count
- [x] Table cells support full nested block content
- [x] Multi-paragraph cells use `<br>` (HTML mode) or space (Default mode)
- [x] Lists in cells preserved as full multi-line lists with correct indentation (HTML mode)
- [x] Code blocks and other block content preserved in cells
- [x] Pipe characters in cells are escaped (`\|`)
- [x] Empty tables output empty string
- [x] All panel nodes implemented with type prefixes
- [x] All five panel types supported
- [x] Panels support nested content
- [x] Empty panels output empty string
- [x] Decision lists and items implemented
- [x] State indicators (✓, ?) correctly applied
- [x] Decision items rendered in single continuous blockquote
- [x] Empty decision lists output empty string
- [x] All 20+ test cases pass
- [x] `make test` passes
- [x] `make lint` passes
- [x] No regressions in Phase 1-3

---

## Next Phase Preview

Phase 5 will add:
- `expand` / `nestedExpand` (collapsible sections)
- `emoji` (unicode conversion)
- `mention` (user references)
- `status` (status badges)
- `media` (images and media embeds)
- Other rich media and interactive elements
