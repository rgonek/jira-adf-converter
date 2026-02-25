# Contributing

Thanks for your interest in improving `jira-adf-converter`.

## Development Setup

1. Fork and clone the repository.
2. Install Go (current target: `1.25.x`).
3. Build and run tests:

```bash
make build
make test
```

## Before Opening a Pull Request

Run the local quality gates:

```bash
make fmt-check
make lint
make test
make tidy-check
```

Optional stronger checks (if installed):

```bash
make staticcheck
make vuln-check
```

## Golden Fixtures

- Forward fixtures live under `testdata/**`.
- Reverse-only fixtures live under `mdconverter/testdata/reverse/**`.

If behavior changes intentionally, update goldens with:

```bash
make test-update
```

Review fixture diffs carefully before committing.

## Pull Request Guidelines

- Keep PRs focused and explain the "why".
- Add or update tests for behavioral changes.
- Update docs when flags/config/behavior change.
- Ensure CI is green before requesting review.
