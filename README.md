# Jira ADF to GFM Converter

A Go library to convert Jira [Atlassian Document Format (ADF)](https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/) to GitHub Flavored Markdown (GFM).

The primary goal is to generate markdown that is easily readable by AI agents and humans alike, ensuring no data loss during conversion.

## Features (Planned)

*   **Granular Support**: Incremental support for ADF nodes.
*   **Data Preservation**: Fallbacks for complex nodes (e.g., Panels -> Blockquotes) to ensure no semantic information is lost.
*   **Configurable Output**:
    *   **Pure Markdown**: (Default) Uses text-based formatting.
    *   **HTML Support**: Optional flag to use raw HTML for features not supported by GFM (e.g., underlining, colors).
*   **Task Lists**: Support for Jira tasks to GFM task lists.

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
jac input.json > output.md
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/rgonek/jira-adf-converter/adf"
)

func main() {
    jsonData := []byte(`{"version": 1, "type": "doc", "content": [...]}`)

    // Default conversion (Pure Markdown)
    markdown, err := adf.Convert(jsonData)
    if err != nil {
        panic(err)
    }
    fmt.Println(markdown)

    // With HTML enabled
    converter := adf.NewConverter(adf.WithHTML())
    markdownHtml, _ := converter.Convert(jsonData)
    fmt.Println(markdownHtml)
}
```

## Development

See [agents/plans/jira-to-gfm.md](agents/plans/jira-to-gfm.md) for the development roadmap.
