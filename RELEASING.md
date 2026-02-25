# Releasing

This project uses Semantic Versioning (`MAJOR.MINOR.PATCH`).

## Versioning Policy

- `MAJOR`: breaking API or behavior changes.
- `MINOR`: backward-compatible feature additions.
- `PATCH`: backward-compatible fixes/docs/chore runtime fixes.
- Optional pre-release tags: `-rc.N`, `-beta.N`.

## Pre-Release Checklist

1. Confirm `main` is green in CI.
2. Update `CHANGELOG.md` (move relevant notes from `[Unreleased]`).
3. Run local checks:

```bash
make check
```

4. Run smoke conversions (forward + reverse + canonical task-lists):

```bash
go run ./cmd/jac --preset=balanced testdata/simple/basic_text.json > /tmp/basic_text.md
go run ./cmd/jac --reverse --preset=balanced testdata/simple/basic_text.md > /tmp/basic_text.adf.json

go run ./cmd/jac --preset=balanced testdata/lists/task_canonical_unchecked.json > /tmp/task_canonical.md
go run ./cmd/jac --reverse --preset=balanced mdconverter/testdata/reverse/lists/task_canonical_unchecked.md > /tmp/task_canonical.adf.json
go run ./cmd/jac --reverse --preset=balanced mdconverter/testdata/reverse/lists/task_mixed_states.md > /tmp/task_mixed.adf.json
go run ./cmd/jac --reverse --preset=balanced mdconverter/testdata/reverse/lists/task_nested.md > /tmp/task_nested.adf.json
```

5. Optionally validate release candidate automation:

```bash
git tag -a vX.Y.Z-rc.1 -m "vX.Y.Z-rc.1"
git push origin vX.Y.Z-rc.1
```

6. Verify generated artifacts/checksums from the release workflow.

## Final Release Steps

1. Create annotated tag on the release commit:

```bash
git tag -a vX.Y.Z -m "vX.Y.Z"
git push origin vX.Y.Z
```

2. The `release.yml` workflow builds binaries, generates checksums, and publishes a GitHub Release.
3. Perform a quick post-release smoke conversion with the published binary.
