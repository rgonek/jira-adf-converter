# Jira ADF to GFM Converter

A robust Go library and CLI tool for converting Jira [Atlassian Document Format (ADF)](https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/) to GitHub Flavored Markdown (GFM).

The primary goal is to generate markdown that is **semantic**, **clean**, and **AI-friendly**, preserving structural hierarchy and text content while gracefully handling unsupported formatting.

## 🚀 Features

*   **30+ ADF Nodes Supported**: Complete coverage from basic text to complex tables, panels, decision lists, and rich media.
*   **Comprehensive Node Support**: Handles paragraphs, headings, blockquotes, code blocks, rules, and hard breaks.
*   **Complex Layouts**: Full support for **Tables** (including nested content), **Lists** (bullet, ordered, task), **Panels**, and **Decision Lists**.
*   **Rich Formatting**: Supports bold, italic, strikethrough, inline code, and links.
*   **Rich Media & Interactive Elements**: Expandable sections, emojis, user mentions, status badges, dates, and media embeds (images, files).
*   **Smart Links**: Atlassian inline cards with URL/JSON-LD extraction.
*   **Runtime Link/Media Hooks**: Rewrite links/media in both directions (ADF -> Markdown and Markdown -> ADF) with context-aware callbacks.
*   **Configurable Output**:
    *   **Pure Markdown**: (Default) Strictly adheres to GFM.
    *   **HTML Fallback**: Optional flag to use HTML tags for features GFM lacks (underline, subscript/superscript).
*   **Intelligent Text Processing**:
    *   Handles nested marks naturally.
    *   Preserves nested content in table cells and lists.
    *   Maps semantic elements (like "Info" panels) to readable markdown equivalents.

👉 **[View Detailed Feature Support & Documentation](docs/features.md)**

## 📦 Installation

### Library
```bash
go get github.com/rgonek/jira-adf-converter
```

### CLI Tool (`jac`)
```bash
go install github.com/rgonek/jira-adf-converter/cmd/jac@latest
```

## 🛠️ Usage

### Command Line Interface

Convert a file and print to stdout:
```bash
jac input.json
```

Save to a file:
```bash
jac input.json > output.md
```

**Options:**
*   `--allow-html`: Enable HTML-oriented rendering for underline/subsup/line breaks/expand blocks.
*   `--strict`: Return an error for unknown nodes and unknown marks.

```bash
jac --allow-html --strict input.json
```

### Go Library

```go
package main

import (
    "fmt"
    "github.com/rgonek/jira-adf-converter/converter"
)

func main() {
    jsonData := []byte(`{
      "version": 1,
      "type": "doc",
      "content": [
        {
          "type": "paragraph",
          "content": [
            {"type": "text", "text": "Hello, "},
            {"type": "text", "text": "World!", "marks": [{"type": "strong"}]}
          ]
        }
      ]
    }`)

    // Configure the converter
    cfg := converter.Config{
        UnderlineStyle: converter.UnderlineIgnore,
        UnknownNodes:   converter.UnknownPlaceholder,
        UnknownMarks:   converter.UnknownSkip,
    }
    conv, err := converter.New(cfg)
    if err != nil {
        panic(err)
    }
    
    result, err := conv.Convert(jsonData)
    if err != nil {
        panic(err)
    }
    fmt.Println(result.Markdown)
    // Output: Hello, **World!**
}
```

### Context-Aware Conversion and Runtime Hooks

Use `ConvertWithContext` when you need source-path-aware rewriting, cancellation/timeouts, or custom link/media mapping.

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

cfg := converter.Config{
    ResolutionMode: converter.ResolutionBestEffort,
    LinkHook: func(ctx context.Context, in converter.LinkRenderInput) (converter.LinkRenderOutput, error) {
        if strings.HasPrefix(in.Href, "https://confluence.example/wiki/pages/") {
            return converter.LinkRenderOutput{
                Href:    "../pages/123.md",
                Handled: true,
            }, nil
        }
        return converter.LinkRenderOutput{Handled: false}, nil
    },
    MediaHook: func(ctx context.Context, in converter.MediaRenderInput) (converter.MediaRenderOutput, error) {
        if in.MediaType == "image" && in.ID != "" {
            return converter.MediaRenderOutput{
                Markdown: "![Image](./assets/" + in.ID + ".png)",
                Handled:  true,
            }, nil
        }
        return converter.MediaRenderOutput{Handled: false}, nil
    },
}

conv, _ := converter.New(cfg)
result, err := conv.ConvertWithContext(ctx, jsonData, converter.ConvertOptions{SourcePath: "docs/spec.adf.json"})
_ = result
_ = err
```

Reverse conversion supports the same model:

```go
reverseCfg := mdconverter.ReverseConfig{
    ResolutionMode: mdconverter.ResolutionStrict,
    LinkHook: func(ctx context.Context, in mdconverter.LinkParseInput) (mdconverter.LinkParseOutput, error) {
        if strings.HasPrefix(in.Destination, "../") {
            return mdconverter.LinkParseOutput{
                Destination: "https://confluence.example/wiki/pages/123",
                ForceLink:   true,
                Handled:     true,
            }, nil
        }
        return mdconverter.LinkParseOutput{Handled: false}, nil
    },
}

reverseConv, _ := mdconverter.New(reverseCfg)
reverseResult, err := reverseConv.ConvertWithContext(ctx, "[Page](../page.md)", mdconverter.ConvertOptions{SourcePath: "docs/page.md"})
_ = reverseResult
_ = err
```

Resolution modes:

*   `best_effort` (default): unresolved hook references (`ErrUnresolved`) produce warnings and fall back to built-in behavior.
*   `strict`: unresolved references fail conversion immediately.

Concurrency note:

*   Converter instances are safe for concurrent `Convert`/`ConvertWithContext` calls.
*   Hook closures are caller-provided and must be thread-safe when shared across goroutines.

## ⚙️ Configuration

| Option | Description |
|--------|-------------|
| **`UnderlineStyle`** | Controls underline rendering (`ignore`, `bold`, `html`). |
| **`UnknownNodes`** | Controls unknown node handling (`placeholder`, `skip`, `error`). |
| **`UnknownMarks`** | Controls unknown mark handling (`skip`, `placeholder`, `error`). |
| **`ResolutionMode`** | Hook unresolved handling (`best_effort`, `strict`). |


