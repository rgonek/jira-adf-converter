# Plan: Production Readiness Hardening

## Goal
Move `jira-adf-converter` from "near production-ready" to a repeatable, auditable production posture for public consumption and long-term maintenance.

This plan covers all previously proposed improvements:
1. Markdown escaping/literal safety
2. Reverse date-detection parity with configurable `DateFormat`
3. CLI warning visibility + warning-based failure mode
4. Reverse-converter concurrency hardening tests
5. Module/toolchain hygiene (`go mod tidy` discipline)
6. CI security/static-analysis hardening (`staticcheck`, `govulncheck`, matrix checks)
7. Release/governance essentials (license, security policy, contributing guidance, changelog/tags)

---

## Scope

### In Scope
- Library correctness improvements for ADF -> Markdown and Markdown -> ADF
- CLI ergonomics for warning handling
- Test coverage improvements (especially concurrency for `mdconverter`)
- CI pipeline hardening and dependency hygiene checks
- Repository governance/release documentation needed for production adoption

### Out of Scope
- Building a hosted service/API wrapper around this library
- New conversion features unrelated to production hardening
- Large refactors of converter architecture that are not needed for these goals

---

## Deliverables
1. Literal-safe Markdown emission strategy with fixtures and regression tests
2. `DateFormat`-aware reverse date detection/parsing strategy with tests
3. CLI warning surface (`stderr`) plus `--fail-on-warning`
4. Reverse concurrency tests proving safe repeated concurrent `Convert` usage
5. Hardened CI pipeline (fmt, staticcheck, vuln scan, matrix, tidy guard, race reliability)
6. Repository production docs (`LICENSE`, `SECURITY.md`, `CONTRIBUTING.md`, optional `CODEOWNERS`)
7. Release process docs + automation (SemVer, changelog, tags, release workflow)
8. Explicit bidirectional task-list regression fixtures for canonical checkbox syntax

---

## Step-by-Step Implementation Plan

### Task 1: Markdown Literal Safety and Escaping
**Goal**: Ensure plain text and text-adjacent contexts render as intended Markdown literals without accidental formatting/structure changes.

**Primary Files**:
- `converter/converter.go`
- `converter/marks.go`
- `converter/inline.go`
- `converter/tables.go`
- New helper(s): `converter/markdown_escape.go` (or equivalent)

**Implementation Details**:
1. Define escaping rules for raw text in inline flow (outside code spans/fences).
2. Apply escaping in contexts where plain text currently emits directly.
3. Ensure link text/title/url escaping is safe and deterministic.
4. Validate table-cell escaping remains correct and does not double-escape existing behavior.
5. Document escaping strategy in `docs/features.md`.

**Test Additions**:
- New fixtures under `testdata/marks/` and `testdata/edge_cases/` for literal characters:
  - `*`, `_`, `[`, `]`, `(`, `)`, `#`, `>`, backticks, backslashes, and mixed combinations.
- Roundtrip-oriented tests for literals that should survive ADF -> MD -> ADF.

**Acceptance Criteria**:
- Literal markdown control characters in text no longer mutate document semantics unexpectedly.
- No regressions in existing golden files unless intentionally updated and reviewed.
- Added docs explain escaping behavior and known constraints.

---

### Task 2: Reverse Date Detection Parity with `DateFormat`
**Goal**: Remove ISO-only detection limitation and support configurable reverse parsing aligned with `ReverseConfig.DateFormat`.

**Primary Files**:
- `mdconverter/patterns.go`
- `mdconverter/config.go`
- `mdconverter/golden_test.go`
- `docs/features.md`

**Implementation Details**:
1. Replace/augment hardcoded `dateISORe` matching with layout-aware detection.
2. Parse candidate text segments using `time.Parse` with `ReverseConfig.DateFormat`.
3. Keep ISO fallback behavior where desirable for backward compatibility (explicitly documented).
4. Guard against over-detection (false positives in normal prose).

**Test Additions**:
- New reverse fixtures for non-ISO dates (for example `02 Jan 2006`, `2006/01/02`, custom separators).
- Negative fixtures where similar-looking strings should remain plain text.

**Acceptance Criteria**:
- Reverse date parsing works for configured layout(s), not only ISO regex patterns.
- Existing ISO behavior still passes unless intentionally changed.
- False positives are controlled by tests.

---

### Task 3: CLI Warning Visibility and Fail Mode
**Goal**: Make warnings operationally visible and optionally enforce warning-free conversion in strict pipelines.

**Primary Files**:
- `cmd/jac/main.go`
- `cmd/jac/preset_test.go`
- New tests: `cmd/jac/main_test.go` (or equivalent)
- `README.md`

**Implementation Details**:
1. Emit conversion warnings to `stderr` in both forward and reverse modes.
2. Add `--fail-on-warning` flag:
   - If warnings exist, exit non-zero after writing output/warnings according to defined behavior.
3. Keep conversion output on `stdout` for pipe compatibility.
4. Add concise warning formatting (type/node/context/message).
5. Update CLI docs/examples.

**Test Additions**:
- Tests verifying:
  - warnings are reported,
  - `--fail-on-warning` changes exit behavior,
  - normal warning-free runs remain unchanged.

**Acceptance Criteria**:
- Users can see warnings without inspecting library structs.
- CI/publish pipelines can fail on warnings deterministically.
- Output stream behavior (`stdout` for converted content) remains script-friendly.

---

### Task 4: Reverse Converter Concurrency Hardening
**Goal**: Add explicit evidence that `mdconverter.Converter` is safe for concurrent use under documented conditions.

**Primary Files**:
- `mdconverter/*_test.go` (new concurrency-focused tests)
- `docs/features.md` (concurrency contract wording)

**Implementation Details**:
1. Add stress tests for concurrent `Convert` calls using shared converter instances.
2. Add config-isolation tests (map-backed fields) mirroring forward-side coverage quality.
3. Add hook-oriented concurrency test patterns to reinforce caller-owned synchronization expectations.

**Acceptance Criteria**:
- New reverse concurrency tests pass reliably with repeated runs.
- Concurrency contract is explicitly documented and consistent across both directions.

---

### Task 4A: Task List Bidirectional Coverage Guard
**Goal**: Ensure canonical task-list markdown (`- [ ] item`) is explicitly covered in both conversion directions and protected against regression.

**Primary Files**:
- Forward fixtures: `testdata/lists/task*.json`, `testdata/lists/task*.md`
- Reverse fixtures: `mdconverter/testdata/reverse/lists/task*.md`, `mdconverter/testdata/reverse/lists/task*.json`
- Test runners: `converter/converter_test.go`, `mdconverter/golden_test.go`

**Implementation Details**:
1. Add/confirm a minimal forward fixture pair with exactly:
   - `- [ ] first`
   - `- [ ] second`
2. Add/confirm a minimal reverse fixture pair with the same markdown input and expected `taskList/taskItem` ADF output (`state: TODO`).
3. Add a mixed-state fixture (`TODO` + `DONE`) in both directions to ensure checkbox state mapping remains stable.
4. Keep one nested task-list fixture to protect indentation/hierarchy behavior.
5. Include this fixture set in release smoke tests.

**Acceptance Criteria**:
- Both directions have explicit golden fixtures for the two-line unchecked task-list form.
- Mixed-state and nested task-list behavior is covered and passing.
- Any future regression in task checkbox syntax fails CI.

---

### Task 5: Module and Toolchain Hygiene
**Goal**: Enforce deterministic module hygiene and reduce drift in dependency metadata.

**Primary Files**:
- `go.mod`
- `go.sum`
- `Makefile`
- CI workflow(s)

**Implementation Details**:
1. Apply `go mod tidy` normalization (direct vs indirect correctness).
2. Add a `tidy-check` target (or equivalent CI step) to fail if `go.mod`/`go.sum` drift.
3. Ensure local/CI commands are documented in `README.md` and `AGENTS.md` references as needed.

**Acceptance Criteria**:
- `go mod tidy` produces no diffs in clean state.
- CI fails when module files are out of sync.

---

### Task 6: CI Security and Static Analysis Hardening
**Goal**: Expand CI from core tests to production-grade quality/security gates.

**Primary Files**:
- `.github/workflows/ci.yml`
- Optional new workflows: `.github/workflows/security.yml`, `.github/workflows/codeql.yml`
- Optional dependency automation: `.github/dependabot.yml`
- `Makefile`
- `README.md`

**Implementation Details**:
1. Add a primary matrix check job:
   - OS: at least `ubuntu-latest` and `windows-latest`.
   - Go versions: minimum supported + current target.
2. Add deterministic hygiene gates:
   - `go mod tidy` + fail-on-diff for `go.mod`/`go.sum`.
   - `gofmt` check (fail if any file would be reformatted).
3. Strengthen test execution:
   - `go test ./... -shuffle=on -count=1` in standard test jobs.
4. Add static-analysis and security checks:
   - `go vet ./...`
   - `staticcheck ./...`
   - `govulncheck ./...` (PR and/or scheduled job)
5. Keep race detector job explicit and reliable:
   - Linux runner,
   - `CGO_ENABLED=1`,
   - required C toolchain setup before `go test -race ./...`.
6. Optional but recommended:
   - Separate `security.yml` for scheduled CodeQL + govulncheck scans.
   - Add Dependabot for Go modules and GitHub Actions updates.

**Acceptance Criteria**:
- CI gates include gofmt, tidy-check, vet, test, race, staticcheck, and govulncheck.
- Matrix coverage confirms expected multi-platform and multi-Go-version behavior.
- Security workflow and dependency automation are configured (if enabled).

---

### Task 7: Release and Governance Baseline
**Goal**: Add standard open-source production artifacts and release workflow clarity.

**Primary Files**:
- `LICENSE`
- `SECURITY.md`
- `CONTRIBUTING.md`
- Optional `CODEOWNERS`
- `CHANGELOG.md`
- `RELEASING.md`
- Optional release automation: `.github/workflows/release.yml`
- `README.md`

**Implementation Details**:
1. Add license file (project-selected SPDX-compatible license).
2. Add security reporting policy (`SECURITY.md`).
3. Add contribution workflow (`CONTRIBUTING.md`) including testing expectations.
4. Define release versioning policy (SemVer):
   - `MAJOR`: breaking API/behavior changes.
   - `MINOR`: backward-compatible feature additions.
   - `PATCH`: backward-compatible fixes/docs/chore-level runtime fixes.
   - Optional pre-releases (`-rc.N`, `-beta.N`) for candidate testing.
5. Define changelog process:
   - Keep `CHANGELOG.md` using Keep-a-Changelog style sections.
   - Require upgrade notes for breaking changes.
6. Add explicit release runbook in `RELEASING.md`:
   - Ensure all CI checks pass on `main`.
   - Update `CHANGELOG.md` for the target version.
   - Create annotated tag `vX.Y.Z` on release commit.
   - Push tag; release workflow publishes artifacts and notes.
   - Run post-release smoke checks (CLI forward/reverse conversion).
7. Optional release automation workflow (`release.yml`) on tag push:
   - Re-run critical checks.
   - Build cross-platform binaries (Linux/Windows/macOS; amd64/arm64 where applicable).
   - Generate checksums file.
   - Create GitHub Release and upload artifacts.

**Acceptance Criteria**:
- Repository includes standard governance files expected by production adopters.
- Release steps (SemVer selection, changelog, tagging, publishing) are documented and repeatable.
- Tag-triggered release automation exists (if enabled) and produces downloadable artifacts + checksums.

---

### Task 8: Final Verification and Production Gate
**Goal**: Validate all improvements as a coherent production readiness milestone.

**Verification Commands (target state)**:
- `go test ./...`
- `go test -race ./...` (in CI with toolchain support)
- `gofmt` check (no diffs)
- `go vet ./...`
- `staticcheck ./...`
- `govulncheck ./...`
- `go mod tidy` (no diff)

**Additional Validation**:
1. CLI smoke tests for both directions with and without warnings.
2. Confirm no unintended golden-file churn.
3. Review benchmark baselines to ensure no major performance regressions.
4. Tag a dry-run pre-release (`vX.Y.Z-rc.1`) and verify release workflow artifact integrity.
5. Run dedicated task-list smoke conversion for both directions.

**Acceptance Criteria**:
- All gates pass in CI.
- New docs and governance files are present and linked from `README.md`.
- Release checklist is complete and usable.

---

## Recommended Execution Order
1. Task 1 (escaping) and Task 2 (date parity) first for correctness.
2. Task 3 and Task 4 next for operability + runtime confidence.
3. Task 5 and Task 6 for automated quality/security enforcement.
4. Task 7 and Task 8 to finalize production posture and release readiness.

---

## Risks and Mitigations

1. **Escaping changes may alter existing fixtures unexpectedly**
   - Mitigation: introduce targeted fixtures first; review diffs carefully; avoid broad blind updates.

2. **Date detection may over-match prose**
   - Mitigation: add negative tests and boundary-aware candidate scanning.

3. **CLI warning output could break existing scripts**
   - Mitigation: keep converted payload on `stdout`; warnings only on `stderr`; document behavior clearly.

4. **New CI gates may initially fail due environment/tooling assumptions**
   - Mitigation: stage gates incrementally, then enforce as required checks once stable.

5. **Governance/release docs require policy choices**
   - Mitigation: define explicit owner decisions early (license type, versioning scheme, support window).

---

## Success Criteria
The hardening effort is complete when:
- [x] Literal markdown text safety is implemented and regression-tested.
- [x] Reverse date detection honors configurable `DateFormat` with strong tests.
- [x] CLI prints warnings to `stderr` and supports `--fail-on-warning`.
- [x] Reverse converter has explicit concurrent conversion test coverage.
- [x] Canonical task-list syntax (`- [ ]` lines) is covered by forward and reverse golden tests.
- [x] `go.mod`/`go.sum` hygiene is enforced and clean.
- [ ] CI includes gofmt, staticcheck, govulncheck, race, matrix coverage, and tidy checks.
- [ ] Governance files (`LICENSE`, `SECURITY.md`, `CONTRIBUTING.md`) exist and are linked.
- [ ] Release/tag/changelog workflow is documented and repeatable.
- [ ] Release automation (or documented manual release path) is validated with a dry-run pre-release.
