# Supported Features

This document details the mapping between Jira Atlassian Document Format (ADF) nodes and GitHub Flavored Markdown (GFM).

## Node Support Matrix

| ADF Node | GFM Representation | Notes |
|----------|-------------------|-------|
| `doc` | Root Document | - |
| `paragraph` | Text Block | Separated by blank lines. |
| `text` | Text | Preserves content. |
| `heading` | `h1` - `h6` (`#` - `######`) | Supports nested marks. |
| `blockquote` | Blockquote (`>`) | Supports nesting (`>>`) and multiline content. |
| `rule` | Horizontal Rule (`---`) | - |
| `hardBreak` | Line Break (`\` + `\n`) | Follows GFM spec for hard breaks. |
| `codeBlock` | Fenced Code Block (```) | Supports language syntax highlighting. |
| `bulletList` | Bullet List (`-`) | Supports nesting. |
| `orderedList` | Ordered List (`1.`) | Supports nesting. |
| `listItem` | List Item | Supports complex content (paragraphs, code blocks, etc.). |
| `taskList` | Task List | - |
| `taskItem` | Checkbox (`- [ ]` / `- [x]`) | Based on `state` attribute (`TODO`/`DONE`). |
| `table` | Table | - |
| `tableRow` | Table Row | - |
| `tableHeader`| Header Cell | - |
| `tableCell` | Data Cell | Supports multiline content (joined by `<br>` or space). |
| `panel` | Semantic Blockquote | Prefixed with type (e.g., `> **Info**: ...`). |
| `decisionList`| Decision List | Continuous blockquote. |
| `decisionItem`| Decision Item | Prefixed with state (e.g., `> **✓ Decision**: ...`). |
| `expand` | Expander | `<details>` (HTML) or Blockquote (Text). |
| `nestedExpand`| Nested Expander | Same as `expand`. |
| `emoji` | Emoji | Shortcode (e.g., `:smile:`) or fallback text. |
| `mention` | User Mention | `Name (accountId:...)`. |
| `status` | Status Badge | `[Status: TEXT]`. |
| `date` | Date | ISO 8601 (`YYYY-MM-DD`). |
| `inlineCard` | Smart Link | Link or JSONLD extraction. |
| `media` | Media Item | Image `![alt](url)` or Placeholder `[Type: id]`. |
| `mediaSingle` | Media Container | Wrapper for single media. |
| `mediaGroup` | Media Gallery | List of media items. |

### Complex Layouts

#### Tables
*   **Headers**: Automatically detected. If the first row contains only data cells, an empty header row is generated (`| | |...`).
*   **Multiline Content**: Paragraphs within cells are joined with `<br>` tags to preserve line breaks while staying within a single table cell.
*   **Complex Content**: Lists, code blocks, and panels inside cells are preserved (though rendering may vary by markdown viewer).

#### Panels
Panels are converted to blockquotes with a bold label indicating their type:
*   `info` → `> **Info**: ...`
*   `note` → `> **Note**: ...`
*   `success` → `> **Success**: ...`
*   `warning` → `> **Warning**: ...`
*   `error` → `> **Error**: ...`

#### Decision Lists
Decisions are rendered as a continuous blockquote with status icons:
*   `DECIDED` → `> **✓ Decision**: ...`
*   `UNDECIDED` → `> **? Decision**: ...`

### Rich Media & Interactive Elements

#### Expanders
Expandable sections (`expand`, `nestedExpand`) adapt to the configuration:
*   **Default**: Rendered as a blockquote with a bold title (`> **Title**`).
*   **AllowHTML**: Rendered as `<details><summary>Title</summary>...`.

#### Inline Metadata
*   **Emoji**: Converted to shortcodes (e.g., `:smile:`) if available, otherwise uses the fallback text.
*   **Mentions**: Displayed as `User Name (accountId:12345)` to preserve identity without accidentally triggering GitHub user notifications or relying on internal Jira IDs.
*   **Status**: Rendered as `[Status: IN PROGRESS]`. Color attributes are ignored to maintain clean markdown.
*   **Date**: Converted to standard `YYYY-MM-DD` format.

#### Media
*   **Images**: Rendered as standard markdown images `![Alt](url)`.
*   **Files/Placeholders**: If no URL is available (e.g., internal Jira attachments), rendered as a placeholder `[Image: id]` or `[File: id]`.

#### Smart Links
*   **Inline Cards**: Converted to standard links `[Title](url)`. If JSON-LD data is present, the title is extracted from it.

## Mark Support (Formatting)

Marks apply formatting to text nodes. Some marks behave differently based on the `AllowHTML` configuration.

| ADF Mark | Default (Pure Markdown) | With `AllowHTML: true` |
|----------|-------------------------|------------------------|
| `strong` | `**text**` | `**text**` |
| `em` | `*text*` (or `_text_` if mixed) | `*text*` |
| `strike` | `~~text~~` | `~~text~~` |
| `code` | `` `text` `` | `` `text` `` |
| `link` | `[text](url "title")` | `[text](url "title")` |
| `underline`| `text` (Formatting dropped) | `<u>text</u>` |
| `subsup` (sub) | `text` (Formatting dropped) | `<sub>text</sub>` |
| `subsup` (sup) | `^text` | `<sup>text</sup>` |

### Link Handling
*   Titles are supported: `[Text](url "Title")`.
*   Empty text links are handled: `[](url)`.
*   Missing URLs fallback to plain text.

### Runtime Link and Media Hooks

Both conversion directions support optional runtime hooks:

*   **ADF -> Markdown (`converter`)**
    *   `LinkHook(ctx, LinkRenderInput) -> LinkRenderOutput`
    *   `MediaHook(ctx, MediaRenderInput) -> MediaRenderOutput`
*   **Markdown -> ADF (`mdconverter`)**
    *   `LinkHook(ctx, LinkParseInput) -> LinkParseOutput`
    *   `MediaHook(ctx, MediaParseInput) -> MediaParseOutput`

Hook behavior:

*   Return `Handled: false` to preserve built-in behavior.
*   Return `converter.ErrUnresolved` / `mdconverter.ErrUnresolved` to trigger resolution-mode handling.
*   Hooks receive typed metadata (`PageID`, `SpaceKey`, `AttachmentID`, `Filename`, `Anchor`) and raw attrs payloads.

Reverse-path note:

*   Prefer `ConvertWithContext(..., ConvertOptions{SourcePath: ...})` so hooks can resolve relative markdown links (`../page.md`) and local media paths consistently.

## Configuration Options

### `AllowHTML`
*   **False (Default)**: Produces strict GFM. Unsupported formatting (underline, subscript) is dropped or approximated to ensure the output works everywhere.
*   **True**: Uses raw HTML tags (`<u>`, `<sub>`, `<sup>`, `<br>`, `<details>`) for features GFM doesn't support natively.

### `Strict`
*   **False (Default)**: Gracefully handles unknown nodes by outputting a placeholder `[Unknown node: type]` or ignoring unknown marks. Recommended for general use.
*   **True**: Returns an error immediately upon encountering an unknown node or mark. Useful for validation or ensuring 100% conversion fidelity.

### `ResolutionMode` (Hook unresolved behavior)

*   **`best_effort` (Default)**: unresolved hook lookups add warnings and fall back to existing conversion logic.
*   **`strict`**: unresolved hook lookups fail conversion.

### Concurrency Contract

*   Converters keep per-conversion state and can be called concurrently.
*   Hook closures are caller-owned code and must synchronize shared mutable state.
