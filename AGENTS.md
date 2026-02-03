# AI Agents Documentation

This repository contains a Go library for converting Jira ADF to Markdown.

## Plans & Roadmap

The current development plan is stored in:
*   [agents/plans/jira-to-gfm.md](agents/plans/jira-to-gfm.md) (general)
*   [agents/plans/phase1-detailed.md](agents/plans/phase1-detailed.md) (phase 1)

## Quick Commands

Use the Makefile for common tasks:

```bash
# Build the CLI binary
make build

# Run all tests
make test

# Update golden files after intentional changes
make test-update

# Run linter
make lint

# Format code
make fmt

# Run all checks (fmt, lint, test)
make check

# Clean build artifacts
make clean
```

## Context

*   **Goal**: Create a high-fidelity converter that preserves semantics for AI consumption.
*   **Architecture**:
    *   Input: JSON (ADF)
    *   Output: String (GFM)
    *   Testing: Golden file approach (`testdata/**/*.json` vs `testdata/**/*.md`).
*   **Test Organization**: Tests can be organized in subdirectories under `testdata/`. The harness recursively finds all `*.json` files and matches them with corresponding `.md` files.
