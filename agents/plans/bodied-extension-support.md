# Bodied Extension Multi-Strategy Support

## Overview

ADF `bodiedExtension` nodes contain both extension metadata (extensionKey, extensionType, parameters) AND child content (block nodes like paragraphs, task lists, etc.). Currently all extension types (`extension`, `inlineExtension`, `bodiedExtension`) share the same `ExtensionRules` config and `convertExtension()` code path, defaulting to JSON serialization. This loses readability for bodied extensions whose children are standard ADF blocks.

This plan adds three dedicated rendering strategies for `bodiedExtension` nodes, matching the pattern used by `layoutSection` and `expand`: **Standard** (lossy GFM), **HTML** (lossless), and **Pandoc** (lossless, default). Bidirectional roundtrip for HTML and Pandoc strategies. ExtensionHandlers still take priority when registered.

## Deliverables

1. New `BodiedExtensionStyle` forward config with 4 strategies (standard, html, pandoc, json)
2. Forward rendering for all strategies with children converted to readable markdown
3. New `BodiedExtensionDetection` reverse config with HTML and Pandoc detection
4. Reverse Pandoc: `.adf-bodied-extension` fenced div reconstruction
5. Reverse HTML: `<div class="adf-bodied-extension">` block consumption
6. Golden test coverage for all strategies (both directions)

---

## Step-by-Step Implementation Plan

### Task 1: Create Golden Test Data (TDD)

**Goal**: Create golden files for all bodied extension strategies before implementation.

**Directory**: `testdata/extensions/`

**Test Cases to Create**:

1. **`bodied_ext_pandoc.json/md`**
   - Input: `bodiedExtension` with `extensionKey="panel"`, parameters with title, content is a `taskList` with 2 items
   - Expected (Pandoc):
     ```markdown
     :::{ .adf-bodied-extension key="panel" extensionType="com.atlassian.confluence.macro.core" parameters="{\"macroParams\":{\"title\":{\"value\":\"Next you might want to:\"}}}" }

     - [ ] **Customise the overview page** - Click the pencil icon...
     - [ ] **Create additional pages** - Click the + in the left sidebar...

     :::
     ```

2. **`bodied_ext_html.json/md`**
   - Input: same as above
   - Expected (HTML):
     ```html
     <div class="adf-bodied-extension" data-extension-key="panel" data-extension-type="com.atlassian.confluence.macro.core" data-parameters="{&quot;macroParams&quot;:{&quot;title&quot;:{&quot;value&quot;:&quot;Next you might want to:&quot;}}}">

     - [ ] **Customise the overview page** - Click the pencil icon...
     - [ ] **Create additional pages** - Click the + in the left sidebar...

     </div>
     ```

3. **`bodied_ext_standard.json/md`**
   - Input: same as above
   - Expected (Standard/lossy): just the children rendered as markdown, no wrapper
     ```markdown
     - [ ] **Customise the overview page** - Click the pencil icon...
     - [ ] **Create additional pages** - Click the + in the left sidebar...
     ```

4. **`bodied_ext_json.json/md`**
   - Input: same as above (or use existing `ext_json` fixture)
   - Expected: JSON code fence (existing `ExtensionJSON` behavior, same as current `ext_json.md`)

5. **`bodied_ext_pandoc_no_params.json/md`**
   - Input: `bodiedExtension` with no `parameters` attr, just `extensionKey` and content
   - Expected: Pandoc div without `parameters` attribute

### Task 2: Forward Config — `BodiedExtensionStyle`

**Goal**: Add config type, field, defaults, and validation.

**File**: `converter/config.go`

**Implementation Details**:

- **New type and constants**:
  ```go
  type BodiedExtensionStyle string
  const (
      BodiedExtensionStandard BodiedExtensionStyle = "standard"
      BodiedExtensionHTML     BodiedExtensionStyle = "html"
      BodiedExtensionPandoc   BodiedExtensionStyle = "pandoc"
      BodiedExtensionJSON     BodiedExtensionStyle = "json"
  )
  ```

- **Config struct**: Add `BodiedExtensionStyle BodiedExtensionStyle` field (after `LayoutSectionStyle`)

- **`applyDefaults()`**: Default to `BodiedExtensionPandoc`

- **`Validate()`**: Accept all 4 values, reject others

**Acceptance Criteria**:
- `go build ./...` succeeds
- Validation accepts all 4 constants
- Invalid values are rejected

### Task 3: Forward Rendering — `convertBodiedExtension()`

**Goal**: Implement bodied extension rendering with all 3 new strategies.

**File**: `converter/extensions.go`

**Implementation Details**:

- **Modify `convertExtension()`**: After the ExtensionHandler check, add:
  ```go
  if node.Type == "bodiedExtension" && s.config.BodiedExtensionStyle != BodiedExtensionJSON {
      return s.convertBodiedExtension(node)
  }
  ```
  When `BodiedExtensionJSON`, fall through to existing `ExtensionRules` logic (backward compat).

- **New `convertBodiedExtension(node Node) (string, error)`**:

  - **Standard**: Call `s.convertChildren(node.Content)`, return directly. Metadata silently dropped.

  - **HTML**: Emit `<div class="adf-bodied-extension"` with `data-extension-key`, `data-extension-type`, `data-parameters` (JSON + HTML-escaped). Call `s.convertChildren(node.Content)` for body. Close with `</div>`.

  - **Pandoc**: Emit `:::{ .adf-bodied-extension key="..." extensionType="..." parameters="..." }`. Call `s.convertChildren(node.Content)` for body. Close with `:::`.

- **Helper `serializeBodiedExtensionParams(attrs map[string]interface{}) string`**: Extract `parameters` from node attrs, JSON-marshal it, return as string. If `parameters` is nil/empty, return "".

**Acceptance Criteria**:
- Standard: children rendered, no metadata in output
- HTML: well-formed div with all attrs, children rendered inside
- Pandoc: well-formed fenced div with all attrs, children rendered inside
- JSON: existing behavior preserved (delegates to `ExtensionRules`)
- ExtensionHandlers still take priority over all built-in strategies

### Task 4: Reverse Config — `BodiedExtensionDetection`

**Goal**: Add reverse config type, field, defaults, validation, and detection helpers.

**File**: `mdconverter/config.go`

**Implementation Details**:

- **New type and constants**:
  ```go
  type BodiedExtensionDetection string
  const (
      BodiedExtensionDetectNone   BodiedExtensionDetection = "none"
      BodiedExtensionDetectHTML   BodiedExtensionDetection = "html"
      BodiedExtensionDetectPandoc BodiedExtensionDetection = "pandoc"
      BodiedExtensionDetectAll    BodiedExtensionDetection = "all"
  )
  ```

- **`ReverseConfig` struct**: Add `BodiedExtensionDetection BodiedExtensionDetection` field

- **`applyDefaults()`**: Default to `BodiedExtensionDetectPandoc`

- **`Validate()`**: Accept all 4 values

- **Detection helpers** (on `*state`):
  ```go
  func (s *state) shouldDetectBodiedExtensionHTML() bool {
      return s.config.BodiedExtensionDetection == BodiedExtensionDetectHTML ||
          s.config.BodiedExtensionDetection == BodiedExtensionDetectAll
  }
  func (s *state) shouldDetectBodiedExtensionPandoc() bool {
      return s.config.BodiedExtensionDetection == BodiedExtensionDetectPandoc ||
          s.config.BodiedExtensionDetection == BodiedExtensionDetectAll
  }
  ```

- **Update `needsPandocBlockExtension()`**: Add bodied extension pandoc detection check

- **Update `hasUnknownPandocDivClass()`**: Add `"adf-bodied-extension"` to known classes

**Acceptance Criteria**:
- Validation accepts all 4 values
- Detection helpers return correct booleans
- `needsPandocBlockExtension()` returns true when pandoc detection is enabled
- All existing tests pass unchanged

### Task 5: Reverse Pandoc — `.adf-bodied-extension` Div

**Goal**: Reconstruct `bodiedExtension` from Pandoc fenced div.

**File**: `mdconverter/pandoc_div_convert.go`

**Implementation Details**:

Add handler in `convertPandocDivNode()` **before** the `.adf-extension` handler:

```go
if hasPandocClass(node.Classes, "adf-bodied-extension") {
    if !s.shouldDetectBodiedExtensionPandoc() {
        return literalFallback, true, nil
    }

    extensionKey := node.Attrs["key"]
    extensionType := node.Attrs["extensionType"]
    paramsJSON := node.Attrs["parameters"]

    content, err := s.convertBlockFragment(node.Body())
    if err != nil {
        return converter.Node{}, false, err
    }

    attrs := map[string]interface{}{
        "extensionKey":  extensionKey,
        "extensionType": extensionType,
    }
    if paramsJSON != "" {
        var params interface{}
        if err := json.Unmarshal([]byte(paramsJSON), &params); err == nil {
            attrs["parameters"] = params
        }
    }

    return converter.Node{
        Type:    "bodiedExtension",
        Attrs:   attrs,
        Content: content,
    }, true, nil
}
```

**Acceptance Criteria**:
- Pandoc div → `bodiedExtension` node with correct attrs and content
- Parameters JSON round-trips correctly
- Missing parameters attr produces node without parameters
- When detection disabled, falls back to literal paragraph

### Task 6: Reverse HTML — `<div class="adf-bodied-extension">` Block

**Goal**: Reconstruct `bodiedExtension` from HTML div wrapper.

**Files**: `mdconverter/html_blocks.go`, `mdconverter/walker.go`

**Implementation Details**:

**html_blocks.go**:
- Add `bodiedExtensionOpenPattern` regex:
  ```go
  bodiedExtensionOpenPattern = regexp.MustCompile(
      `(?is)^<div\s+class="adf-bodied-extension"\s+` +
      `data-extension-key="([^"]*)"\s+` +
      `data-extension-type="([^"]*)"\s*` +
      `(?:data-parameters="([^"]*)")?\s*>\s*$`,
  )
  ```
- Add `parseBodiedExtensionOpenTag()` returning `(extensionKey, extensionType, parametersJSON string, ok bool)`
- Reuse existing `divClosePattern` for `</div>` detection

**walker.go**:
- In `convertBlockSlice()`, add bodied extension HTML detection (after layout section detection):
  ```go
  if s.shouldDetectBodiedExtensionHTML() {
      if opening, ok := children[index].(*ast.HTMLBlock); ok {
          if key, extType, params, ok := parseBodiedExtensionOpenTagFromHTMLBlock(opening, s.source); ok {
              node, consumed, consumedOK, err := s.consumeBodiedExtensionBlock(children, index, parent, key, extType, params)
              // ...
          }
      }
  }
  ```
- Add `consumeBodiedExtensionBlock()` following the depth-tracked `consumeLayoutSectionBlock()` pattern

**Acceptance Criteria**:
- HTML div → `bodiedExtension` node with correct attrs and content
- Data attributes are properly HTML-unescaped (especially `data-parameters`)
- Closing `</div>` at correct depth terminates the block
- When detection disabled, HTML block passes through as text

### Task 7: Config Test Coverage

**Goal**: Unit tests for new config types.

**Files**: `converter/config_test.go`, `mdconverter/config_test.go`

**Test Cases**:
- Valid `BodiedExtensionStyle` values accepted by `Config.Validate()`
- Invalid `BodiedExtensionStyle` value rejected
- Valid `BodiedExtensionDetection` values accepted by `ReverseConfig.Validate()`
- Invalid `BodiedExtensionDetection` value rejected

### Task 8: Final Verification

**Goal**: Ensure all tests pass and no regressions exist.

**Steps**:
1. Run `go test ./...` — all tests pass
2. Run `go vet ./...` — zero issues
3. Verify round-trip: pandoc `.json` → `.md` → `.json` produces same ADF

**Acceptance Criteria**:
- All existing golden tests pass unchanged (backward compatibility)
- All new golden tests pass
- `go build ./... && go test ./... && go vet ./...` all succeed

---

## Key Files

| File | Action |
|---|---|
| `converter/config.go` | Add `BodiedExtensionStyle` type, config field, defaults, validation |
| `converter/extensions.go` | Add bodied extension branching + 3 strategies + helper |
| `mdconverter/config.go` | Add `BodiedExtensionDetection` type, config field, defaults, validation, helpers |
| `mdconverter/pandoc_div_convert.go` | Handle `.adf-bodied-extension` pandoc div + update `hasUnknownPandocDivClass` |
| `mdconverter/html_blocks.go` | Add HTML open tag pattern + parser |
| `mdconverter/walker.go` | Add HTML block consumption for bodied extensions |
| `converter/config_test.go` | Validation tests for `BodiedExtensionStyle` |
| `mdconverter/config_test.go` | Validation tests for `BodiedExtensionDetection` |
| `testdata/extensions/` | 5 new golden test pairs |

## Existing Code to Reuse

- `s.convertChildren(node.Content)` — renders ADF children to markdown (`converter/converter.go`)
- `s.convertBlockFragment(body)` — parses markdown fragment to ADF nodes (`mdconverter/html_blocks.go`)
- `consumeLayoutSectionBlock()` — depth-tracked HTML block consumption pattern (`mdconverter/walker.go`)
- `hasPandocClass()` / `hasUnknownPandocDivClass()` — pandoc div class detection (`mdconverter/pandoc_div_convert.go`)
- `Node.GetStringAttr()` — safe attribute extraction (`converter/ast.go`)
- `divClosePattern` — existing `</div>` regex (`mdconverter/html_blocks.go`)
- Golden test infrastructure — `TestGoldenFiles` / `TestReverseGoldenFiles`

---

## Success Criteria

The implementation is complete when:
- [ ] `BodiedExtensionStyle` config with 4 strategies exists and validates
- [ ] Forward converter renders bodied extensions correctly for all strategies
- [ ] ExtensionHandlers still take priority over built-in strategies
- [ ] `BodiedExtensionDetection` config with 4 detection modes exists and validates
- [ ] Pandoc reverse: `.adf-bodied-extension` div → `bodiedExtension` node
- [ ] HTML reverse: `<div class="adf-bodied-extension">` → `bodiedExtension` node
- [ ] JSON reverse: existing `parseExtensionFence()` continues to work
- [ ] All golden tests pass for all strategies
- [ ] Round-trip (pandoc/HTML): `.json` → `.md` → `.json` preserves ADF structure
- [ ] All existing tests pass unchanged (no regressions)
- [ ] `go build ./... && go test ./... && go vet ./...` all succeed
