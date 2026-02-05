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
*   `--allow-html`: Enable HTML tags for underline, subscript, superscript, and break tags in tables.
*   `--strict`: Return an error if unknown nodes are encountered (default is to render a placeholder).

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
        AllowHTML: false, // Use pure Markdown
        Strict:    false, // Graceful error handling
    }
    conv := converter.New(cfg)
    
    markdown, err := conv.Convert(jsonData)
    if err != nil {
        panic(err)
    }
    fmt.Println(markdown)
    // Output: Hello, **World!**
}
```

## ⚙️ Configuration

| Option | Description |
|--------|-------------|
| **`AllowHTML`** | **`false`**: (Default) Drops non-GFM formatting (underline, sub/sup) or uses text symbols. <br> **`true`**: Renders `<u>`, `<sub>`, `<sup>`, and `<br>` for better fidelity in HTML-enabled viewers. |
| **`Strict`** | **`false`**: (Default) Renders `[Unknown node: type]` for unsupported nodes. <br> **`true`**: Returns an error immediately on unsupported nodes. |


