# Jira ADF <-> GFM Converter

A Go library and CLI for bidirectional conversion between Jira [Atlassian Document Format (ADF)](https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/) and GitHub Flavored Markdown (GFM).

The project focuses on semantic, AI-friendly output while preserving round-trip metadata where possible.

## Features

- Bidirectional conversion APIs:
  - `converter` package: ADF JSON -> Markdown
  - `mdconverter` package: Markdown -> ADF JSON
- Granular, JSON-serializable configuration for formatting, detection, unknown handling, and extensions.
- Structured conversion results with warnings (`Result{Markdown|ADF, Warnings}`).
- Runtime link/media hooks in both directions with context, source-path support, and strict/best-effort unresolved behavior.
- Registry-based Extension Hook system to serialize specific ADF extensions as custom Markdown.
- Pandoc-flavored Markdown support for maximum fidelity (bracketed spans, fenced divs, grid tables).
- CLI presets (`balanced`, `strict`, `readable`, `lossy`, `pandoc`) with directional mapping.

See `docs/features.md` for detailed node/mark coverage and parsing rules.

## Installation

### Library

```bash
go get github.com/rgonek/jira-adf-converter
```

### CLI (`jac`)

```bash
go install github.com/rgonek/jira-adf-converter/cmd/jac@latest
```

## CLI Usage

Forward conversion (ADF JSON -> Markdown):

```bash
jac input.adf.json
```

Reverse conversion (Markdown -> ADF JSON):

```bash
jac --reverse input.md
```

Reverse mode prints pretty-formatted ADF JSON.

Common options:

- `--preset=balanced|strict|readable|lossy|pandoc`
- `--allow-html` (compatibility override; in forward mode it forces HTML-oriented rendering for underline/subsup/hard breaks/expand, and in reverse mode it enables broad HTML mention/expand detection)
- `--strict` (compatibility override; in forward mode it enforces unknown-node/mark errors, and in reverse mode it applies strict detection defaults)

Example:

```bash
jac --preset=readable input.adf.json > output.md
jac --reverse --preset=strict input.md > output.adf.json
```

Preset precedence in CLI is deterministic: preset first, then compatibility overrides (`--allow-html`, `--strict`).

## Library Usage

### ADF -> Markdown (`converter`)

```go
package main

import (
    "fmt"
    "os"

    "github.com/rgonek/jira-adf-converter/converter"
)

func main() {
    input, err := os.ReadFile("input.adf.json")
    if err != nil {
        panic(err)
    }

    conv, err := converter.New(converter.Config{
        MentionStyle: converter.MentionLink,
        PanelStyle:   converter.PanelGitHub,
    })
    if err != nil {
        panic(err)
    }

    result, err := conv.Convert(input)
    if err != nil {
        panic(err)
    }

    fmt.Print(result.Markdown)
    for _, w := range result.Warnings {
        fmt.Printf("warning: %s (%s): %s\n", w.Type, w.NodeType, w.Message)
    }
}
```

### Markdown -> ADF (`mdconverter`)

```go
package main

import (
    "fmt"

    "github.com/rgonek/jira-adf-converter/mdconverter"
)

func main() {
    conv, err := mdconverter.New(mdconverter.ReverseConfig{
        MentionDetection: mdconverter.MentionDetectLink,
        PanelDetection:   mdconverter.PanelDetectGitHub,
    })
    if err != nil {
        panic(err)
    }

    result, err := conv.Convert("[Page](https://example.com)")
    if err != nil {
        panic(err)
    }

    fmt.Println(string(result.ADF))
}
```

## Context-Aware Conversion and Hooks

Use `ConvertWithContext` when you need cancellation/timeouts, deterministic relative-path resolution, or custom mapping for links, media, and extensions.

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

adfJSON := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Page","marks":[{"type":"link","attrs":{"href":"https://confluence.example/wiki/pages/123"}}]}]}]}`)

cfg := converter.Config{
    ResolutionMode: converter.ResolutionBestEffort,
    LinkHook: func(ctx context.Context, in converter.LinkRenderInput) (converter.LinkRenderOutput, error) {
        if strings.HasPrefix(in.Href, "https://confluence.example/wiki/pages/") {
            return converter.LinkRenderOutput{Href: "../pages/123.md", Handled: true}, nil
        }
        return converter.LinkRenderOutput{Handled: false}, nil
    },
    ExtensionHandlers: map[string]converter.ExtensionHandler{
        "plantumlcloud": &MyPlantUMLHandler{},
    },
}

conv, _ := converter.New(cfg)
result, err := conv.ConvertWithContext(ctx, adfJSON, converter.ConvertOptions{
    SourcePath: "docs/spec.adf.json",
})
_ = result
_ = err
```

Reverse hooks use the same model (`mdconverter.LinkHook` / `mdconverter.MediaHook` / `mdconverter.ExtensionHandler`) and receive `ConvertOptions{SourcePath: ...}` for consistent relative reference mapping.

### Resolution Modes

- `best_effort` (default): if a hook returns `ErrUnresolved`, conversion continues with fallback behavior and adds a warning.
- `strict`: `ErrUnresolved` fails conversion.

### Hook Validation Rules

- Forward link hook (`LinkRenderOutput`): handled output needs `Href` unless `TextOnly=true`.
- Forward media hook (`MediaRenderOutput`): handled output needs non-empty `Markdown`.
- Reverse link hook (`LinkParseOutput`): handled output needs non-empty `Destination`; `ForceLink` and `ForceCard` cannot both be true.
- Reverse media hook (`MediaParseOutput`): handled output requires supported `MediaType` (`image` or `file`) and exactly one of `ID` or `URL`.

## Configuration Highlights

### Forward (`converter.Config`) defaults

| Field | Default |
|---|---|
| `UnderlineStyle` | `bold` |
| `SubSupStyle` | `html` |
| `MentionStyle` | `link` |
| `PanelStyle` | `github` |
| `ExpandStyle` | `html` |
| `InlineCardStyle` | `link` |
| `TableMode` | `auto` |
| `Extensions.Default` | `json` |
| `UnknownNodes` | `placeholder` |
| `UnknownMarks` | `skip` |
| `ResolutionMode` | `best_effort` |

### Reverse (`mdconverter.ReverseConfig`) defaults

| Field | Default |
|---|---|
| `MentionDetection` | `link` |
| `EmojiDetection` | `shortcode` |
| `StatusDetection` | `bracket` |
| `DateDetection` | `iso` |
| `PanelDetection` | `github` |
| `ExpandDetection` | `html` |
| `DecisionDetection` | `emoji` |
| `ResolutionMode` | `best_effort` |

## CLI Presets

`jac` supports `balanced`, `strict`, `readable`, `lossy`, and `pandoc` presets in both directions.

- `balanced`: library defaults (recommended for most workflows).
- `strict`: stronger fidelity/parsing constraints.
- `readable`: favors cleaner human-facing markdown and lenient reverse patterns.
- `lossy`: favors compact output over metadata preservation.
- `pandoc`: produces Pandoc-flavored Markdown with bracketed spans and fenced divs for maximum metadata preservation.

## Concurrency

- Converter instances are safe for concurrent `Convert`/`ConvertWithContext` calls.
- Hook closures are caller-owned and must protect shared mutable state.

## Documentation

- Detailed feature matrix and syntax mapping: `docs/features.md`
- Development roadmap plans: `agents/plans/`


