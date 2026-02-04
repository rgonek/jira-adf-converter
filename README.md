# Jira ADF to GFM Converter

A Go library to convert Jira [Atlassian Document Format (ADF)](https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/) to GitHub Flavored Markdown (GFM).

The primary goal is to generate markdown that is easily readable by AI agents and humans alike, preserving text content while gracefully handling unsupported formatting.

**Note**: This converter follows the [GitHub Flavored Markdown (GFM) specification](https://github.github.com/gfm/) as closely as possible. Where GFM lacks native support for certain ADF features (e.g., subscript/superscript, underline, panels), we provide fallback representations that prioritize readability and semantic preservation.

## Features

*   **Granular Support**: Incremental support for ADF nodes.
*   **Text Preservation**: All text content is preserved. Minor formatting (colors) may be lost for unsupported features.
*   **Configurable Output**:
    *   **Pure Markdown**: (Default) Uses text-based formatting.
    *   **HTML Support**: Optional flag to use raw HTML for features not supported by GFM.
*   **CLI Tool**: Includes a command-line utility for converting files.

## Installation

### Library

```bash
go get github.com/rgonek/jira-adf-converter
```

### CLI Tool (`jac`)

```bash
go install github.com/rgonek/jira-adf-converter/cmd/jac@latest
```

**Usage:**

```bash
jac input.json
jac --strict input.json
jac --allow-html input.json
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/rgonek/jira-adf-converter/converter"
)

func main() {
    jsonData := []byte(`{"version": 1, "type": "doc", "content": [...]}`)

    // Default conversion (Pure Markdown)
    cfg := converter.Config{
        AllowHTML: false,
        Strict:    false,
    }
    conv := converter.New(cfg)
    
    markdown, err := conv.Convert(jsonData)
    if err != nil {
        panic(err)
    }
    fmt.Println(markdown)
}
```

## Configuration

### Strict Mode

When `Strict: true`, the converter will return an error if it encounters:
- Unknown node types
- Unknown mark types

This is useful for ensuring all content is properly handled during conversion.

### Non-Strict Mode (Default)

When `Strict: false`, the converter will gracefully handle unknown elements:
- **Unknown nodes**: Rendered as `[Unknown node: type]` to indicate missing content
- **Unknown marks** (e.g., colors, custom marks): Silently ignored (text preserved, formatting lost)
- **Underline mark**: Supported - uses HTML `<u>` tag when `AllowHTML: true`, otherwise dropped

This mode prioritizes text content preservation. Minor semantic formatting like colors and other unsupported marks are dropped, but the actual text content is always preserved.

### Allow HTML

When `AllowHTML: true`, the converter will use raw HTML tags for features that don't have native markdown equivalents:
- **Underline**: Uses `<u>text</u>` HTML tag (renders correctly on GitHub and most markdown platforms)
- **Subscript/Superscript**: Uses `<sub>text</sub>` and `<sup>text</sup>` HTML tags

When `AllowHTML: false` (default), these features are handled gracefully:
- **Underline**: Formatting is dropped, text is preserved
- **Superscript**: Rendered with `^` prefix (e.g., `^text`)
- **Subscript**: Formatting is dropped, text is preserved (to avoid conflicts with GFM syntax)

## Current Implementation (Phase 2)

### Supported Nodes
- `doc` - Root document node
- `paragraph` - Text paragraphs
- `text` - Text content with formatting
- `heading` - Headings (H1-H6) with support for nested inline marks
- `blockquote` - Block quotes with proper nesting support (nested quotes use `>>` format per GFM spec)
- `rule` - Horizontal rules (---)
- `hardBreak` - Hard line breaks within paragraphs

### Supported Marks
- `strong` - Bold text (**text**)
- `em` - Italic text (*text*)
- `strike` - Strikethrough (~~text~~)
- `code` - Inline code (`text`)
- `underline` - Underlined text
  - With `AllowHTML: true`: `<u>text</u>`
  - With `AllowHTML: false`: Text preserved, formatting dropped
- `link` - Hyperlinks with optional titles
  - Format: `[text](url)` or `[text](url "title")`
  - Handles edge cases: empty text, missing href, quotes in title
- `subsup` - Subscript and superscript
  - With `AllowHTML: true`: `<sub>text</sub>` and `<sup>text</sup>`
  - With `AllowHTML: false`: 
    - Superscript: `^text` (carat prefix)
    - Subscript: plain text (no indicator to avoid GFM syntax conflicts)

### Mark Continuity

The converter implements intelligent mark continuity across adjacent text nodes within a paragraph. This means that marks are only closed and reopened when necessary, producing cleaner markdown output.

**Example:**
```json
{
  "type": "paragraph",
  "content": [
    {"type": "text", "text": "bold ", "marks": [{"type": "strong"}]},
    {"type": "text", "text": "bold+italic", "marks": [{"type": "strong"}, {"type": "em"}]},
    {"type": "text", "text": " end", "marks": [{"type": "strong"}]}
  ]
}
```

**Output:** `**bold _bold+italic_ end**`

The `strong` mark remains open across all three text nodes, while `em` is only applied to the middle node.

### Mark Delimiter Selection

When a paragraph contains text with both `strong` and `em` marks applied simultaneously, the converter automatically uses underscores for `em` to avoid delimiter conflicts:
- `strong` alone: `**text**`
- `em` alone: `*text*`
- `strong` + `em`: `**_text_**` (underscore used for em)

This paragraph-wide detection ensures consistent delimiter usage and proper markdown rendering.

## Limitations

### Features Not Yet Supported

The following ADF features are not yet supported and will be added in future phases:
- Lists and task lists (Phase 3)
- Tables (Phase 4)
- Code blocks (Phase 3)
- Panels (Phase 4)
- Expandable sections (Phase 5)
- Media/images (Phase 5)
- Mention nodes (Phase 5)
- Emoji nodes (Phase 5)
- Inline cards (Phase 5)

See [agents/plans/jira-to-gfm.md](agents/plans/jira-to-gfm.md) for the complete development roadmap.

## Development

See [agents/plans/jira-to-gfm.md](agents/plans/jira-to-gfm.md) for the development roadmap and [agents/plans/phase2-detailed.md](agents/plans/phase2-detailed.md) for Phase 2 implementation details.
