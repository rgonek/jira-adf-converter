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

---

## Step-by-Step Implementation Plan

### Task 1: Create Phase 4 Test Data
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

### Task 2: Implement Table Conversion
**Goal**: Convert ADF table nodes to GFM tables.

**File**: `converter/converter.go`

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

### Task 3: Implement Panel Conversion
**Goal**: Convert ADF panel nodes to blockquotes with type prefixes.

**File**: `converter/converter.go`

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

### Task 4: Implement Decision List Conversion
**Goal**: Convert decision lists and items to a single continuous blockquote.

**File**: `converter/converter.go`

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

### Task 5: Update Node Dispatcher
**Goal**: Register new node types in `convertNode` switch.

**File**: `converter/converter.go`

Add cases for: `table`, `tableRow`, `tableHeader`, `tableCell`, `panel`, `decisionList`, `decisionItem`

**Acceptance Criteria**:
- All new node types registered
- Unknown node handling still works
- No regressions

---

### Task 6: Handle Complex Table Cell Content
**Goal**: Properly preserve block-level content inside table cells.

**File**: `converter/converter.go`

**Implementation**:
- Create `convertCellContent()` method that preserves full block structure
- Paragraphs: Join with `<br>` separator within cell
- Lists: Preserve as full multi-line lists (bullet/ordered/task lists)
- Code blocks: Preserve with backticks
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

### Task 7: Final Verification
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

## Success Criteria for Phase 4

- [ ] All table nodes implemented (`table`, `tableRow`, `tableHeader`, `tableCell`)
- [ ] Tables distinguish headers from data cells
- [ ] Tables with no headers get empty header row with correct column count
- [ ] Table cells support full nested block content
- [ ] Multi-paragraph cells use `<br>`
- [ ] Lists in cells preserved as full multi-line lists
- [ ] Code blocks and other block content preserved in cells
- [ ] Empty tables output empty string
- [ ] All panel nodes implemented with type prefixes
- [ ] All five panel types supported
- [ ] Panels support nested content
- [ ] Empty panels output empty string
- [ ] Decision lists and items implemented
- [ ] State indicators (✓, ?) correctly applied
- [ ] Decision items rendered in single continuous blockquote
- [ ] Empty decision lists output empty string
- [ ] All 20+ test cases pass
- [ ] `make test` passes
- [ ] `make lint` passes
- [ ] No regressions in Phase 1-3

---

## Next Phase Preview

Phase 5 will add:
- `expand` / `nestedExpand` (collapsible sections)
- `emoji` (unicode conversion)
- `mention` (user references)
- `status` (status badges)
- `media` (images and media embeds)
- Other rich media and interactive elements
