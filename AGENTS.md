# AI Agents Documentation

This repository contains a Go library and CLI for bidirectional conversion between Jira ADF and GitHub Flavored Markdown.

## Plans & Roadmap

The current development plan is stored in:
*   [agents/plans/jira-to-gfm.md](agents/plans/jira-to-gfm.md) (forward direction overview)
*   [agents/plans/phase6-detailed.md](agents/plans/phase6-detailed.md) (granular config + presets)
*   [agents/plans/gfm-to-adf.md](agents/plans/gfm-to-adf.md) (reverse converter architecture)
*   [agents/plans/link-media-hooks.md](agents/plans/link-media-hooks.md) (runtime link/media hooks)

## Quick Commands

Use the Makefile for common tasks:

```bash
# Build the CLI binary
make build

# Run all tests
make test

# Run all checks (fmt, lint, test)
make check

# Update golden files after intentional changes
make test-update

# Run linter
make lint

# Format code
make fmt

# Clean build artifacts
make clean

# Forward conversion (ADF -> Markdown)
go run ./cmd/jac --preset=balanced testdata/simple/basic_text.json

# Reverse conversion (Markdown -> ADF JSON)
go run ./cmd/jac --reverse --preset=balanced testdata/simple/basic_text.md
```

## Context

*   **Goal**: Preserve ADF semantics with AI-friendly Markdown while supporting deterministic reverse parsing back to ADF.
*   **Core packages**:
    *   `converter`: ADF JSON input (`[]byte`) -> `converter.Result{Markdown, Warnings}`
    *   `mdconverter`: Markdown input (`string`) -> `mdconverter.Result{ADF, Warnings}`
*   **Runtime hooks**: Both directions support context-aware link/media hooks and `ResolutionMode` (`best_effort` or `strict`).
*   **Testing model**:
    *   Shared golden fixtures in `testdata/**` (`.json` <-> `.md` pairs)
    *   Reverse-only fixtures in `mdconverter/testdata/reverse/**`
    *   Unit tests for config, hooks, and edge cases
    *   Reverse fuzz/benchmark coverage in `mdconverter/`
