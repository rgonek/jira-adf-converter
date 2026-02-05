# Phase 5: Rich Media & Interactive Elements

## Overview
Add support for rich media (images, files) and interactive elements (expanders, status, mentions, emojis, dates, inline cards) to the Jira ADF to GFM converter. This enhances the semantic richness of the converted documents, ensuring vital context like user mentions, task statuses, dates, and attached media is preserved.

**Note**: This implementation follows GitHub Flavored Markdown (GFM) specification as closely as possible.

## Deliverables
1. **Interactive Containers**: Support for `expand` and `nestedExpand` nodes (using HTML details/summary or blockquote fallback).
2. **Inline Metadata**: Support for `emoji`, `mention`, `status`, and `date` nodes.
3. **Media**: Support for `media`, `mediaSingle`, and `mediaGroup` nodes (images and placeholders).
4. **Smart Links**: Support for `inlineCard` nodes (Atlassian smart links).

---

## Step-by-Step Implementation Plan

### Task 1: Create Test Data
**Goal**: Create Golden Files for Phase 5 features before implementation (TDD).

**Directories** (semantic subdirectories, NOT a "phase5" directory):
- Interactive: `testdata/expanders/`
- Media: `testdata/media/`
- Inline: `testdata/inline/`

**Test Cases to Create**:

#### Expander Tests (testdata/expanders/)

1.  **expand.json/md**
    *   Input: `expand` node with title "Click to see more" and paragraph content.
    *   Expected (Text Mode): `> **Click to see more**` followed by blank line, then blockquoted content.
    *   Expected (HTML Mode): `<details><summary>Click to see more</summary>...content...</details>`

2.  **expand_nested.json/md**
    *   Input: Nested expanders (expand inside expand).
    *   Expected: Correctly rendered nested structure in both HTML and text modes.

3.  **expand_in_list.json/md**
    *   Input: `nestedExpand` inside a list item.
    *   Expected: Similar to `expand` but verifying placement behavior inside lists.

4.  **expand_empty.json/md** *(Edge Case)*
    *   Input: `expand` node with title but no content.
    *   Expected (Text Mode): `> **Title**\n>\n` (empty blockquote)
    *   Expected (HTML Mode): `<details><summary>Title</summary>\n\n</details>`

5.  **expand_no_title.json/md** *(Edge Case)*
    *   Input: `expand` node with content but no title attribute.
    *   Expected (Text Mode): Blockquoted content without title line.
    *   Expected (HTML Mode): `<details><summary></summary>...content...</details>`

#### Inline Metadata Tests (testdata/inline/)

6.  **emoji.json/md**
    *   Input: Paragraph containing `emoji` node with `shortName` (e.g., ":smile:").
    *   Expected: Shortcode output `:smile:`

7.  **emoji_fallback.json/md** *(Edge Case)*
    *   Input: `emoji` node with no `shortName` but has `fallback` attribute.
    *   Expected: Fallback text output (e.g., "😊")

8.  **emoji_missing_both.json/md** *(Edge Case)*
    *   Input: `emoji` node with neither `shortName` nor `fallback`.
    *   Expected: Empty string or `[emoji]` placeholder (strict mode error).

9.  **mention.json/md**
    *   Input: `mention` node with `text` (display name) and `id` (account ID).
    *   Expected: `User Name (accountId:12345)`

10. **mention_no_text.json/md** *(Edge Case)*
    *   Input: `mention` node with `id` but no `text`.
    *   Expected: `Unknown User (accountId:12345)`

11. **mention_no_id.json/md** *(Edge Case)*
    *   Input: `mention` node with `text` but no `id`.
    *   Expected: `User Name (accountId:unknown)`

12. **status.json/md**
    *   Input: `status` node with `text` attribute (e.g., "In Progress").
    *   Expected: `[Status: In Progress]`

13. **status_with_color.json/md**
    *   Input: `status` node with `text` and `color` attributes.
    *   Expected: `[Status: In Progress]` (color ignored, per design philosophy)

14. **date.json/md**
    *   Input: `date` node with `timestamp` attribute (Unix timestamp "1582152559").
    *   Expected: `2020-02-19` (ISO 8601 date format YYYY-MM-DD)

15. **date_invalid.json/md** *(Edge Case)*
    *   Input: `date` node with invalid or missing timestamp.
    *   Expected: `[Date: invalid]` or strict mode error.

16. **inline_combined.json/md**
    *   Input: Paragraph with mixed emoji, mention, status, and date nodes.
    *   Expected: `Here is a :smile: for User Name (accountId:12345) who is [Status: IN PROGRESS] as of 2020-02-19`.

17. **inline_card.json/md**
    *   Input: `inlineCard` node with `url` attribute.
    *   Expected: `[https://example.com]` (linked URL in brackets)

18. **inline_card_with_data.json/md**
    *   Input: `inlineCard` node with `data` (JSONLD) instead of `url`.
    *   Expected: Extract title/URL from JSONLD or fallback to `[Smart Link]`.

19. **inline_card_empty.json/md** *(Edge Case)*
    *   Input: `inlineCard` node with neither `url` nor `data`.
    *   Expected: `[Smart Link]` or strict mode error.

#### Media Tests (testdata/media/)

20. **media_image_url.json/md**
    *   Input: `mediaSingle` containing a `media` node with `type="image"`, `alt`, and `url`.
    *   Expected: `![Alt Text](http://example.com/image.png)`

21. **media_image_no_alt.json/md** *(Edge Case)*
    *   Input: `media` node with `type="image"` and `url` but no `alt`.
    *   Expected: `![Image](http://example.com/image.png)` (default alt text)

22. **media_image_id.json/md**
    *   Input: `media` node with `type="image"`, `id`, but no `url`.
    *   Expected: `[Image: id-123]`

23. **media_file.json/md**
    *   Input: `media` node with `type="file"`, `id`, and optional `collection`.
    *   Expected: `[File: id-456]`

24. **media_single.json/md**
    *   Input: `mediaSingle` node wrapping a media node.
    *   Expected: Pass through to media child (same as unwrapped media).

25. **media_group.json/md**
    *   Input: `mediaGroup` node with multiple `media` children.
    *   Expected: Each media on its own line (newline-separated).

26. **media_group_empty.json/md** *(Edge Case)*
    *   Input: `mediaGroup` node with no children.
    *   Expected: Empty output (structural node with no content).

27. **media_unknown_type.json/md** *(Edge Case)*
    *   Input: `media` node with unknown `type` (e.g., "video").
    *   Expected: `[Media: id-789]` (generic fallback).

28. **media_in_table.json/md**
    *   Input: Table cell containing `mediaSingle` with image.
    *   Expected: Image markdown inside table cell (test nesting).

### Task 2: Implement Interactive Containers (`expand`, `nestedExpand`)
**Goal**: Map expanders to GFM compatible representations.

**File**: `converter/blocks.go`

**Implementation Details**:
Add to `convertNode` switch in `converter/converter.go`:
```go
case "expand", "nestedExpand":
    return c.convertExpand(node)
```

**`convertExpand` function**:
*   Extract `title` attribute from `Attrs`.
*   **Logic**:
    *   If `config.AllowHTML`: 
        *   Render `<details><summary>{Title}</summary>\n{Content}\n</details>`.
        *   If title is empty, use empty `<summary></summary>`.
    *   If `!config.AllowHTML` (Text Mode):
        *   If `title` is present: Render `> **{Title}**\n>\n` (with blank line after title).
        *   Render content, ensuring it is prefixed with `> ` (blockquoted).
        *   Support for nesting: recursively handle nested expanders.
        *   If no content, output empty blockquote structure.

**Acceptance Criteria**:
*   HTML mode produces valid `<details>` tags.
*   Text mode produces clean blockquotes with bold titles and blank line separator.
*   Content inside expanders is correctly rendered.
*   Nested expanders work correctly in both modes.
*   Empty expanders don't cause errors.

### Task 3: Implement Inline Metadata (`emoji`, `mention`, `status`, `date`)
**Goal**: Convert specific inline nodes to text representations.

**File**: `converter/inline.go`

**Implementation Details**:
Add to `convertNode` switch in `converter/converter.go`:
```go
case "emoji":
    return c.convertEmoji(node)
case "mention":
    return c.convertMention(node)
case "status":
    return c.convertStatus(node)
case "date":
    return c.convertDate(node)
```

**Implementation Details**:
*   **`emoji`**:
    *   Read `shortName` (e.g., ":smile:") attribute.
    *   If `shortName` missing, read `fallback` attribute.
    *   If both missing: return empty string (non-strict) or error (strict).
    *   Output the shortcode/fallback directly.
*   **`mention`**:
    *   Read `text` (display name) and `id` (account ID) from `Attrs`.
    *   **Format**: `{text} (accountId:{id})`
    *   If `text` is missing: use "Unknown User".
    *   If `id` is missing: use "unknown".
    *   **Rationale**: This format avoids confusion with file references (`@file`) while preserving both human-readable name and machine-readable ID.
*   **`status`**:
    *   Read `text` attribute (e.g., "In Progress") from `Attrs`.
    *   Ignore `color` and `localId` attributes (per design philosophy).
    *   Output format: `[Status: {text}]`.
    *   If `text` is missing: `[Status: Unknown]`.
*   **`date`**:
    *   Read `timestamp` attribute (Unix timestamp as string).
    *   Convert to ISO 8601 date format: `YYYY-MM-DD`.
    *   If timestamp is invalid/missing: return `[Date: invalid]` (non-strict) or error (strict).

**Acceptance Criteria**:
*   All inline nodes flow naturally within paragraph text.
*   Key information (who, what status, when) is preserved.
*   Fallback handling works for missing attributes.
*   Dates are human-readable and machine-parseable (ISO 8601).
*   No confusion with file references (no bare `@` prefix).

### Task 4: Implement Smart Links (`inlineCard`)
**Goal**: Handle Atlassian smart link cards.

**File**: `converter/inline.go`

**Implementation Details**:
Add to `convertNode` switch in `converter/converter.go`:
```go
case "inlineCard":
    return c.convertInlineCard(node)
```

**Implementation Details**:
*   **`inlineCard`**:
    *   Check `attrs.url` first (simple case).
    *   If `url` exists: Render `[{url}]` (URL in brackets).
    *   If `attrs.data` exists (JSONLD):
        *   Attempt to extract `name` or `url` from JSONLD.
        *   Render as `[{title}]({url})` if both available, otherwise `[{url}]`.
    *   If neither exists: return `[Smart Link]` (non-strict) or error (strict).

**Acceptance Criteria**:
*   URL-based cards render as clickable links.
*   JSONLD-based cards extract meaningful information when possible.
*   Missing data degrades gracefully with placeholder.

### Task 5: Implement Media (`media`, `mediaSingle`, `mediaGroup`)
**Goal**: Handle images and file placeholders.

**File**: `converter/media.go`

**Implementation Details**:
Add to `convertNode` switch in `converter/converter.go`:
```go
case "mediaSingle":
    return c.convertMediaSingle(node)
case "mediaGroup":
    return c.convertMediaGroup(node)
case "media":
    return c.convertMedia(node)
```

**Implementation Details**:
*   **`mediaSingle`**:
    *   Container for single media item.
    *   Pass through to children (render content).
*   **`mediaGroup`**:
    *   Container for multiple media items.
    *   Render each child `media` node on its own line (newline-separated).
    *   If empty, return empty string.
*   **`media`**:
    *   Check attributes: `type` (file, image), `alt` (alt text), `url` (if available), `id` (if no URL).
    *   **Logic**:
        *   If `type="image"` and `url` exists: Render `![{alt}]({url})`.
        *   If `type="image"` and no `url`: Render `[Image: {id}]`.
        *   If `type="file"`: Render `[File: {id}]`.
        *   If `type` is unknown or missing: Render `[Media: {id}]`.
        *   If `alt` is missing for images with URL, default to "Image".

**Acceptance Criteria**:
*   External images render as standard Markdown images.
*   Internal media renders as distinct, identifiable placeholders (not broken images).
*   File type is preserved in placeholder text using format `[Image: id]` or `[File: id]`.
*   Media groups render multiple items correctly.
*   Empty media groups don't cause errors.

### Task 6: Final Verification
**Goal**: Ensure all tests pass and no regressions exist.

**Steps**:
1.  Run `make test`.
2.  Run `make lint`.

**Acceptance Criteria**:
*   All tests pass (including 28 new Phase 5 tests).
*   Linting passes.
*   No regressions in existing Phase 1-4 tests.

---

## Success Criteria for Phase 5
The phase is complete when:
- [ ] **Expanders**: `expand` and `nestedExpand` render as `<details>` (HTML) or blockquotes (Text).
- [ ] **Nested expanders**: Work correctly in both HTML and text modes.
- [ ] **Emojis**: Render as shortcodes (`:smile:`) with fallback to `fallback` attribute.
- [ ] **Mentions**: Render as `User Name (accountId:12345)` format (avoids `@` confusion).
- [ ] **Statuses**: Render as `[Status: text]` with color attributes ignored.
- [ ] **Dates**: Render as ISO 8601 dates (`YYYY-MM-DD`) from Unix timestamps.
- [ ] **Inline cards**: Render as `[url]` or extract data from JSONLD.
- [ ] **Media images**: Render as `![alt](url)` or `[Image: id]` placeholder.
- [ ] **Media files**: Render as `[File: id]` placeholder.
- [ ] **Media groups**: Render multiple media items on separate lines.
- [ ] **Edge cases**: All 28 test cases pass, including empty nodes, missing attributes, and nested structures.
- [ ] **No regressions**: All existing Phase 1-4 tests still pass.
- [ ] **Linting**: `make lint` passes with no issues.

---

## Complete Node Coverage Summary

After Phase 5 completion, the following nodes will be **fully implemented**:

### ✅ Implemented (Phase 1-5)
- **Root**: `doc`
- **Block**: `paragraph`, `heading`, `blockquote`, `rule`, `hardBreak`, `codeBlock`, `panel`, `expand`, `nestedExpand`
- **Lists**: `bulletList`, `orderedList`, `listItem`, `taskList`, `taskItem`, `decisionList`, `decisionItem`
- **Tables**: `table`, `tableRow`, `tableHeader`, `tableCell`
- **Media**: `media`, `mediaSingle`, `mediaGroup`
- **Inline**: `text`, `emoji`, `mention`, `status`, `date`, `inlineCard`
- **Marks**: `strong`, `em`, `strike`, `code`, `underline`, `link`, `subsup`

### ⚠️ Intentionally Not Supported (Out of Scope)
Per the project's design philosophy, the following are intentionally not supported:
- **`textColor`** mark: Color styling not preserved (text preserved, color lost).
- **`backgroundColor`** mark: Background color not preserved.
- **`alignment`** mark: Text alignment not preserved in Markdown.
- **`border`** mark: Border styling not preserved.
- **`multiBodiedExtension`**: Complex extension framework (out of scope for MVP).
- **`extensionFrame`**: Extension child node (out of scope for MVP).
- **`mediaInline`**: Inline media (rare, can be added in future phase if needed).

These nodes will be handled according to the **strict mode** setting:
- **Strict mode OFF** (default): Unknown nodes return `[Unknown node: type]`, unknown marks are silently ignored.
- **Strict mode ON**: Unknown nodes/marks return an error.
