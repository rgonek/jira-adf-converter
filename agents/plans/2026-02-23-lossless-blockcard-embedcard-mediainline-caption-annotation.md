# Lossless Round-Trip: blockCard, embedCard, mediaInline, caption, annotation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add lossless (round-trip) Pandoc-syntax support for five ADF node/mark types that are currently forward-only: `blockCard`, `embedCard`, `mediaInline`, `caption`, and `annotation`.

**Architecture:** Each type gets a new `*Pandoc` style constant in `converter/config.go`, a matching `*Detection` type in `mdconverter/config.go`, forward rendering in the appropriate converter file, reverse parsing in the appropriate mdconverter file, golden fixtures, and a round-trip test entry. All new options are opt-in (backward-compatible defaults). The `pandoc` preset in `cmd/jac/main.go` is extended.

**Tech Stack:** Go, goldmark (Markdown parser), testify (assertions), existing `PandocSpanNode` / `PandocDivNode` infrastructure.

---

## Overview of Pandoc Syntax Chosen

| ADF type | Forward Markdown | Reverse detection key |
|---|---|---|
| `annotation` (mark) | `[text]{.annotation annotation-id="annot-1" annotation-type="inlineComment"}` | Pandoc span, class `annotation` |
| `mediaInline` | `[File: abc-123]{.media-inline media-id="abc-123" media-type="file"}` | Pandoc span, class `media-inline` |
| `blockCard` | `[Title]{.block-card url="https://..."}` (block paragraph) | Pandoc span inside paragraph, class `block-card` |
| `embedCard` | `[Title]{.embed-card url="https://..." layout="wide"}` (block paragraph) | Pandoc span inside paragraph, class `embed-card` |
| `caption` | `[Caption text]{.media-caption}` (following the image on same line or separate line) | Pandoc span, class `media-caption` |

> **Note on caption:** The caption is emitted on the same line immediately after the image link, wrapped in `[...]{.media-caption}`. During reverse parsing, any `media-caption` span that is the *last* inline element of a paragraph that also contains an image is converted into a `caption` node inside the surrounding `mediaSingle`.

---

## Task 1: `annotation` mark — config constants

**Files:**
- Modify: `converter/config.go`
- Modify: `mdconverter/config.go`

**Step 1: Add `AnnotationStyle` type and constants to `converter/config.go`**

After the `InlineCardStyle` block (around line 108), add:

```go
// AnnotationStyle controls how annotation marks are rendered.
type AnnotationStyle string

const (
	AnnotationIgnore  AnnotationStyle = "ignore"
	AnnotationPandoc  AnnotationStyle = "pandoc"
)
```

Add `AnnotationStyle AnnotationStyle` field to the `Config` struct.

In `applyDefaults()` add:
```go
if c.AnnotationStyle == "" {
    c.AnnotationStyle = AnnotationIgnore
}
```

In `Validate()` add:
```go
if c.AnnotationStyle != AnnotationIgnore && c.AnnotationStyle != AnnotationPandoc {
    return fmt.Errorf("invalid annotationStyle %q", c.AnnotationStyle)
}
```

**Step 2: Add `AnnotationDetection` type and constants to `mdconverter/config.go`**

```go
// AnnotationDetection controls how annotation marks are reconstructed.
type AnnotationDetection string

const (
	AnnotationDetectNone   AnnotationDetection = "none"
	AnnotationDetectPandoc AnnotationDetection = "pandoc"
)
```

Add `AnnotationDetection AnnotationDetection` field to `ReverseConfig`.

In `applyDefaults()` add:
```go
if c.AnnotationDetection == "" {
    c.AnnotationDetection = AnnotationDetectNone
}
```

In `Validate()` add:
```go
if c.AnnotationDetection != AnnotationDetectNone && c.AnnotationDetection != AnnotationDetectPandoc {
    return fmt.Errorf("invalid annotationDetection %q", c.AnnotationDetection)
}
```

Add helper to `needsPandocInlineExtension()`:
```go
c.AnnotationDetection == AnnotationDetectPandoc
```

**Step 3: Run tests to verify nothing is broken**

```bash
make test
```
Expected: all tests pass (no new tests yet, just config plumbing).

**Step 4: Commit**
```bash
git add converter/config.go mdconverter/config.go
git commit -m "feat: add AnnotationStyle/AnnotationDetection config for Pandoc round-trip"
```

---

## Task 2: `annotation` mark — forward rendering

**Files:**
- Modify: `converter/marks.go` (line ~290, the `"annotation"` case in `convertMarkFull`)

**Step 1: Write the golden fixture for the Pandoc annotation output**

File: `testdata/marks/annotation_pandoc.json` (already exists as `annotation.json` but with different output)

Create `testdata/marks/annotation_pandoc.json`:
```json
{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"Annotated text","marks":[{"type":"annotation","attrs":{"id":"annot-1","annotationType":"inlineComment"}}]},{"type":"text","text":" normal text"}]}]}
```

Create `testdata/marks/annotation_pandoc.md`:
```
[Annotated text]{.annotation annotation-id="annot-1" annotation-type="inlineComment"} normal text
```

**Step 2: Run existing forward tests to confirm the fixture will FAIL with current code**

```bash
make test
```
The new fixture won't be picked up automatically until we add a test entry, but we'll add it in Task 5. Proceed.

**Step 3: Implement forward rendering in `converter/marks.go`**

Replace the `"annotation"` case (currently `return "", "", nil`) with:

```go
case "annotation":
    switch s.config.AnnotationStyle {
    case AnnotationPandoc:
        id := mark.GetStringAttr("id", "")
        annotationType := mark.GetStringAttr("annotationType", "")
        open := `[`
        close := `]{.annotation`
        if id != "" {
            close += fmt.Sprintf(` annotation-id=%q`, id)
        }
        if annotationType != "" {
            close += fmt.Sprintf(` annotation-type=%q`, annotationType)
        }
        close += `}`
        return open, close, nil
    default:
        // ignore — preserve text, drop mark
        return "", "", nil
    }
```

**Step 4: Run tests**
```bash
make test
```
Expected: all existing tests pass.

**Step 5: Commit**
```bash
git add converter/marks.go testdata/marks/annotation_pandoc.json testdata/marks/annotation_pandoc.md
git commit -m "feat: render annotation mark as Pandoc span when AnnotationStyle=pandoc"
```

---

## Task 3: `annotation` mark — reverse parsing

**Files:**
- Modify: `mdconverter/pandoc_inline_convert.go` — `convertPandocSpanNode`, `hasUnknownPandocSpanClass`, `hasUnknownPandocSpanAttr`
- Modify: `mdconverter/config.go` — add `shouldDetectAnnotationPandoc()` helper

**Step 1: Add detection helper to `mdconverter/config.go`**

```go
func (s *state) shouldDetectAnnotationPandoc() bool {
    return s.config.AnnotationDetection == AnnotationDetectPandoc
}
```

(Add to the `state` methods near the other `shouldDetect*` helpers.)

**Step 2: Whitelist `annotation` class and `annotation-id`/`annotation-type` attrs**

In `hasUnknownPandocSpanClass`:
```go
case "underline", "mention", "inline-card", "annotation":
    continue
```

In `hasUnknownPandocSpanAttr`:
```go
case "mention-id", "url", "color", "background-color", "style", "annotation-id", "annotation-type":
    continue
```

**Step 3: Add `annotation` branch in `convertPandocSpanNode`**

After the `inline-card` block (before the `underline` block), add:

```go
if hasPandocClass(node.Classes, "annotation") {
    if !s.shouldDetectAnnotationPandoc() {
        return []converter.Node{newTextNode(literal, stack.current())}, nil
    }
    annotationID := strings.TrimSpace(node.Attrs["annotation-id"])
    annotationType := strings.TrimSpace(node.Attrs["annotation-type"])

    inlineContent, err := s.convertInlineFragment(node.Content)
    if err != nil {
        return nil, err
    }

    annotationMark := converter.Mark{
        Type: "annotation",
        Attrs: map[string]interface{}{
            "id":             annotationID,
            "annotationType": annotationType,
        },
    }
    if annotationID == "" {
        delete(annotationMark.Attrs, "id")
    }
    if annotationType == "" {
        delete(annotationMark.Attrs, "annotationType")
    }

    inlineContent = applyMarkToInlineNodes(inlineContent, annotationMark)
    return applyOuterMarksToInlineNodes(inlineContent, stack.current()), nil
}
```

**Step 4: Run tests**
```bash
make test
```
Expected: all pass.

**Step 5: Commit**
```bash
git add mdconverter/pandoc_inline_convert.go mdconverter/config.go
git commit -m "feat: reverse-parse annotation Pandoc span to ADF annotation mark"
```

---

## Task 4: `annotation` mark — round-trip test

**Files:**
- Modify: `converter/pandoc_roundtrip_test.go`

**Step 1: Add annotation to `runPandocRoundTrip` config**

In `runPandocRoundTrip`, add to `forwardCfg`:
```go
AnnotationStyle: converter.AnnotationPandoc,
```

Add to `reverse` config:
```go
AnnotationDetection: mdconverter.AnnotationDetectPandoc,
```

**Step 2: Add test entry**

```go
{name: "annotation mark", fixturePath: "marks/annotation_pandoc.json"},
```

**Step 3: Run tests**
```bash
make test
```
Expected: all pass including new `annotation mark` round-trip test.

**Step 4: Update `pandoc` preset in `cmd/jac/main.go`**

In `presetConfig`, `pandoc` case add:
```go
AnnotationStyle: converter.AnnotationPandoc,
```

In `reversePresetConfig`, `pandoc` case add:
```go
AnnotationDetection: mdconverter.AnnotationDetectPandoc,
```

**Step 5: Run tests and build**
```bash
make check
```

**Step 6: Commit**
```bash
git add converter/pandoc_roundtrip_test.go cmd/jac/main.go
git commit -m "feat: wire annotation round-trip test and pandoc preset"
```

---

## Task 5: `mediaInline` — config constants

**Files:**
- Modify: `converter/config.go`
- Modify: `mdconverter/config.go`

**Step 1: Add `MediaInlineStyle` type and constants to `converter/config.go`**

```go
// MediaInlineStyle controls how mediaInline nodes are rendered.
type MediaInlineStyle string

const (
	MediaInlineDefault MediaInlineStyle = "default"
	MediaInlinePandoc  MediaInlineStyle = "pandoc"
)
```

Add `MediaInlineStyle MediaInlineStyle` to `Config` struct.

In `applyDefaults()`:
```go
if c.MediaInlineStyle == "" {
    c.MediaInlineStyle = MediaInlineDefault
}
```

In `Validate()`:
```go
if c.MediaInlineStyle != MediaInlineDefault && c.MediaInlineStyle != MediaInlinePandoc {
    return fmt.Errorf("invalid mediaInlineStyle %q", c.MediaInlineStyle)
}
```

**Step 2: Add `MediaInlineDetection` type and constants to `mdconverter/config.go`**

```go
// MediaInlineDetection controls how mediaInline nodes are reconstructed.
type MediaInlineDetection string

const (
	MediaInlineDetectNone   MediaInlineDetection = "none"
	MediaInlineDetectPandoc MediaInlineDetection = "pandoc"
)
```

Add `MediaInlineDetection MediaInlineDetection` to `ReverseConfig`.

In `applyDefaults()`:
```go
if c.MediaInlineDetection == "" {
    c.MediaInlineDetection = MediaInlineDetectNone
}
```

In `Validate()`:
```go
if c.MediaInlineDetection != MediaInlineDetectNone && c.MediaInlineDetection != MediaInlineDetectPandoc {
    return fmt.Errorf("invalid mediaInlineDetection %q", c.MediaInlineDetection)
}
```

Add to `needsPandocInlineExtension()`:
```go
c.MediaInlineDetection == MediaInlineDetectPandoc
```

**Step 3: Run tests**
```bash
make test
```

**Step 4: Commit**
```bash
git add converter/config.go mdconverter/config.go
git commit -m "feat: add MediaInlineStyle/MediaInlineDetection config for Pandoc round-trip"
```

---

## Task 6: `mediaInline` — forward rendering

**Files:**
- Modify: `converter/media.go` — `convertMediaInline` (line 124)

**Step 1: Create golden fixtures**

`testdata/media/media_inline_pandoc.json`:
```json
{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"Before "},{"type":"mediaInline","attrs":{"id":"abc-123","type":"file","collection":""}},{"type":"text","text":" after"}]}]}
```

`testdata/media/media_inline_pandoc.md`:
```
Before [File: abc-123]{.media-inline media-id="abc-123" media-type="file"} after
```

**Step 2: Implement forward rendering in `converter/media.go`**

Replace `convertMediaInline`:

```go
// convertMediaInline converts a mediaInline node
func (s *state) convertMediaInline(node Node) (string, error) {
    if s.config.MediaInlineStyle == MediaInlinePandoc {
        id := s.getMediaID(node)
        mediaType := node.GetStringAttr("type", "file")
        // Use the same display text as the default path
        defaultText, err := s.convertMedia(node)
        if err != nil {
            return "", err
        }
        span := fmt.Sprintf("[%s]{.media-inline", defaultText)
        if id != "" {
            span += fmt.Sprintf(` media-id=%q`, id)
        }
        span += fmt.Sprintf(` media-type=%q`, mediaType)
        span += "}"
        return span, nil
    }
    return s.convertMedia(node)
}
```

**Step 3: Run tests**
```bash
make test
```

**Step 4: Commit**
```bash
git add converter/media.go testdata/media/media_inline_pandoc.json testdata/media/media_inline_pandoc.md
git commit -m "feat: render mediaInline as Pandoc span when MediaInlineStyle=pandoc"
```

---

## Task 7: `mediaInline` — reverse parsing

**Files:**
- Modify: `mdconverter/pandoc_inline_convert.go`
- Modify: `mdconverter/inline.go` — `*ast.Image` handler (currently always makes `mediaSingle`)

**Step 1: Add detection helper**

In `mdconverter/config.go` (or inline in the state receiver file):

```go
func (s *state) shouldDetectMediaInlinePandoc() bool {
    return s.config.MediaInlineDetection == MediaInlineDetectPandoc
}
```

**Step 2: Whitelist in `hasUnknownPandocSpanClass` / `hasUnknownPandocSpanAttr`**

Class: add `"media-inline"` to the allowed list.

Attrs: add `"media-id"`, `"media-type"` to the allowed list.

**Step 3: Add `media-inline` branch in `convertPandocSpanNode`**

```go
if hasPandocClass(node.Classes, "media-inline") {
    if !s.shouldDetectMediaInlinePandoc() {
        return []converter.Node{newTextNode(literal, stack.current())}, nil
    }
    mediaID := strings.TrimSpace(node.Attrs["media-id"])
    mediaType := strings.TrimSpace(node.Attrs["media-type"])
    if mediaType == "" {
        mediaType = "file"
    }
    if mediaID == "" {
        s.addWarning(converter.WarningMissingAttribute, "pandocSpan", "pandoc media-inline span missing media-id")
        return []converter.Node{newTextNode(literal, stack.current())}, nil
    }
    attrs := map[string]interface{}{
        "id":   mediaID,
        "type": mediaType,
    }
    return []converter.Node{{Type: "mediaInline", Attrs: attrs}}, nil
}
```

**Step 4: Run tests**
```bash
make test
```

**Step 5: Commit**
```bash
git add mdconverter/pandoc_inline_convert.go
git commit -m "feat: reverse-parse media-inline Pandoc span to ADF mediaInline node"
```

---

## Task 8: `mediaInline` — round-trip test

**Files:**
- Modify: `converter/pandoc_roundtrip_test.go`

**Step 1: Add `MediaInlineStyle` to `runPandocRoundTrip` forward config**
```go
MediaInlineStyle: converter.MediaInlinePandoc,
```

Add `MediaInlineDetection` to reverse config:
```go
MediaInlineDetection: mdconverter.MediaInlineDetectPandoc,
```

**Step 2: Add test entry**
```go
{name: "media inline", fixturePath: "media/media_inline_pandoc.json"},
```

**Step 3: Update `normalizeRoundTripNodes` to ignore `collection` on `mediaInline`**

The `media` type already strips `collection`. Check if `mediaInline` also needs it:
```go
if node.Type == "media" || node.Type == "mediaInline" {
    delete(node.Attrs, "collection")
}
```

**Step 4: Update pandoc preset**

`cmd/jac/main.go` — `pandoc` case:
```go
MediaInlineStyle: converter.MediaInlinePandoc,
```
Reverse pandoc case:
```go
MediaInlineDetection: mdconverter.MediaInlineDetectPandoc,
```

**Step 5: Run tests and build**
```bash
make check
```

**Step 6: Commit**
```bash
git add converter/pandoc_roundtrip_test.go cmd/jac/main.go
git commit -m "feat: wire mediaInline round-trip test and pandoc preset"
```

---

## Task 9: `blockCard` / `embedCard` — config constants

**Files:**
- Modify: `converter/config.go`
- Modify: `mdconverter/config.go`

**Step 1: Add `BlockCardStyle` and `EmbedCardStyle` types to `converter/config.go`**

```go
// BlockCardStyle controls how blockCard nodes are rendered.
type BlockCardStyle string

const (
	BlockCardDefault BlockCardStyle = "default"
	BlockCardPandoc  BlockCardStyle = "pandoc"
)

// EmbedCardStyle controls how embedCard nodes are rendered.
type EmbedCardStyle string

const (
	EmbedCardDefault EmbedCardStyle = "default"
	EmbedCardPandoc  EmbedCardStyle = "pandoc"
)
```

Add both fields to `Config` struct.

In `applyDefaults()`:
```go
if c.BlockCardStyle == "" {
    c.BlockCardStyle = BlockCardDefault
}
if c.EmbedCardStyle == "" {
    c.EmbedCardStyle = EmbedCardDefault
}
```

In `Validate()`:
```go
if c.BlockCardStyle != BlockCardDefault && c.BlockCardStyle != BlockCardPandoc {
    return fmt.Errorf("invalid blockCardStyle %q", c.BlockCardStyle)
}
if c.EmbedCardStyle != EmbedCardDefault && c.EmbedCardStyle != EmbedCardPandoc {
    return fmt.Errorf("invalid embedCardStyle %q", c.EmbedCardStyle)
}
```

**Step 2: Add `BlockCardDetection` and `EmbedCardDetection` to `mdconverter/config.go`**

```go
// BlockCardDetection controls how blockCard nodes are reconstructed.
type BlockCardDetection string

const (
	BlockCardDetectNone   BlockCardDetection = "none"
	BlockCardDetectPandoc BlockCardDetection = "pandoc"
)

// EmbedCardDetection controls how embedCard nodes are reconstructed.
type EmbedCardDetection string

const (
	EmbedCardDetectNone   EmbedCardDetection = "none"
	EmbedCardDetectPandoc EmbedCardDetection = "pandoc"
)
```

Add both fields to `ReverseConfig`.

In `applyDefaults()`:
```go
if c.BlockCardDetection == "" {
    c.BlockCardDetection = BlockCardDetectNone
}
if c.EmbedCardDetection == "" {
    c.EmbedCardDetection = EmbedCardDetectNone
}
```

In `Validate()` — add the two validation checks.

Add to `needsPandocInlineExtension()`:
```go
c.BlockCardDetection == BlockCardDetectPandoc ||
c.EmbedCardDetection == EmbedCardDetectPandoc
```

**Step 3: Run tests**
```bash
make test
```

**Step 4: Commit**
```bash
git add converter/config.go mdconverter/config.go
git commit -m "feat: add BlockCardStyle/EmbedCardStyle/detection config for Pandoc round-trip"
```

---

## Task 10: `blockCard` / `embedCard` — forward rendering

**Files:**
- Modify: `converter/blocks.go` — `convertBlockCard` (line 252), `convertEmbedCard` (line 268)
- Modify: `converter/inline.go` — `convertInlineCard` (to pass node type down for Pandoc path)

**Step 1: Create golden fixtures**

`testdata/inline/block_card_pandoc.json`:
```json
{"type":"doc","version":1,"content":[{"type":"blockCard","attrs":{"url":"https://example.com"}}]}
```

`testdata/inline/block_card_pandoc.md`:
```
[https://example.com]{.block-card url="https://example.com"}
```

`testdata/inline/embed_card_pandoc.json`:
```json
{"type":"doc","version":1,"content":[{"type":"embedCard","attrs":{"url":"https://embedded.example.com","layout":"wide"}}]}
```

`testdata/inline/embed_card_pandoc.md`:
```
[https://embedded.example.com]{.embed-card url="https://embedded.example.com" layout="wide"}
```

**Step 2: Implement Pandoc path in `convertBlockCard`**

```go
func (s *state) convertBlockCard(node Node) (string, error) {
    if s.config.BlockCardStyle == BlockCardPandoc {
        _, url := s.getInlineCardLinkData(node)
        if url == "" {
            return "", nil
        }
        return fmt.Sprintf("[%s]{.block-card url=%q}\n\n", url, url), nil
    }
    // existing default path
    content, err := s.convertInlineCard(node)
    ...
}
```

**Step 3: Implement Pandoc path in `convertEmbedCard`**

```go
func (s *state) convertEmbedCard(node Node) (string, error) {
    if s.config.EmbedCardStyle == EmbedCardPandoc {
        _, url := s.getInlineCardLinkData(node)
        if url == "" {
            return "", nil
        }
        layout := node.GetStringAttr("layout", "")
        span := fmt.Sprintf("[%s]{.embed-card url=%q", url, url)
        if layout != "" {
            span += fmt.Sprintf(` layout=%q`, layout)
        }
        span += "}\n\n"
        return span, nil
    }
    // existing default path
    content, err := s.convertInlineCard(node)
    ...
}
```

**Step 4: Run tests**
```bash
make test
```

**Step 5: Commit**
```bash
git add converter/blocks.go testdata/inline/block_card_pandoc.json testdata/inline/block_card_pandoc.md testdata/inline/embed_card_pandoc.json testdata/inline/embed_card_pandoc.md
git commit -m "feat: render blockCard/embedCard as Pandoc spans when style=pandoc"
```

---

## Task 11: `blockCard` / `embedCard` — reverse parsing

**Files:**
- Modify: `mdconverter/pandoc_inline_convert.go`
- Modify: `mdconverter/blocks.go` — paragraph block handler (to promote a standalone Pandoc span paragraph to a block card)

> **Note:** `blockCard` and `embedCard` are block-level ADF nodes, but their Pandoc representation is a Pandoc span inside a paragraph (since Pandoc spans are inline). The reverse parser must detect when a paragraph contains *only* a single `block-card` or `embed-card` span and emit the block node directly instead of wrapping in a paragraph.

**Step 1: Whitelist classes and attrs**

Classes: add `"block-card"`, `"embed-card"`.

Attrs: add `"layout"`.

**Step 2: Add branches in `convertPandocSpanNode`**

```go
if hasPandocClass(node.Classes, "block-card") {
    if !s.shouldDetectBlockCardPandoc() {
        return []converter.Node{newTextNode(literal, stack.current())}, nil
    }
    url := strings.TrimSpace(node.Attrs["url"])
    if url == "" {
        s.addWarning(converter.WarningMissingAttribute, "pandocSpan", "block-card span missing url")
        return []converter.Node{newTextNode(literal, stack.current())}, nil
    }
    return []converter.Node{{
        Type:  "blockCard",
        Attrs: map[string]interface{}{"url": url},
    }}, nil
}

if hasPandocClass(node.Classes, "embed-card") {
    if !s.shouldDetectEmbedCardPandoc() {
        return []converter.Node{newTextNode(literal, stack.current())}, nil
    }
    url := strings.TrimSpace(node.Attrs["url"])
    if url == "" {
        s.addWarning(converter.WarningMissingAttribute, "pandocSpan", "embed-card span missing url")
        return []converter.Node{newTextNode(literal, stack.current())}, nil
    }
    attrs := map[string]interface{}{"url": url}
    if layout := strings.TrimSpace(node.Attrs["layout"]); layout != "" {
        attrs["layout"] = layout
    }
    return []converter.Node{{
        Type:  "embedCard",
        Attrs: attrs,
    }}, nil
}
```

**Step 3: Promote single-card paragraphs in block conversion**

In `mdconverter/blocks.go` (or wherever paragraphs are assembled into block nodes), after building a paragraph's inline content, check if the content is a single `blockCard` or `embedCard` node and if so emit it as a top-level block node rather than inside a `paragraph`.

Look for the `*ast.Paragraph` case in `convertBlockNode` (or `convertBlockChildren`). After resolving inline content:

```go
// Promote blockCard / embedCard nodes that ended up inside a paragraph
if len(inlineContent) == 1 {
    if inlineContent[0].Type == "blockCard" || inlineContent[0].Type == "embedCard" {
        return inlineContent[0], nil
    }
}
```

**Step 4: Run tests**
```bash
make test
```

**Step 5: Commit**
```bash
git add mdconverter/pandoc_inline_convert.go mdconverter/blocks.go
git commit -m "feat: reverse-parse block-card/embed-card Pandoc spans to ADF block nodes"
```

---

## Task 12: `blockCard` / `embedCard` — round-trip tests

**Files:**
- Modify: `converter/pandoc_roundtrip_test.go`

**Step 1: Add config to `runPandocRoundTrip`**

Forward:
```go
BlockCardStyle: converter.BlockCardPandoc,
EmbedCardStyle: converter.EmbedCardPandoc,
```

Reverse:
```go
BlockCardDetection: mdconverter.BlockCardDetectPandoc,
EmbedCardDetection: mdconverter.EmbedCardDetectPandoc,
```

**Step 2: Add test entries**

```go
{name: "block card", fixturePath: "inline/block_card_pandoc.json"},
{name: "embed card", fixturePath: "inline/embed_card_pandoc.json"},
```

**Step 3: Update `normalizeRoundTripNodes` for `blockCard` / `embedCard`**

The existing normalization already handles `inlineCard` url/data normalization. `blockCard` only has `url` so no special normalization needed.

**Step 4: Update pandoc preset**

`cmd/jac/main.go`:
```go
BlockCardStyle: converter.BlockCardPandoc,
EmbedCardStyle: converter.EmbedCardPandoc,
```
Reverse:
```go
BlockCardDetection: mdconverter.BlockCardDetectPandoc,
EmbedCardDetection: mdconverter.EmbedCardDetectPandoc,
```

**Step 5: Run tests and build**
```bash
make check
```

**Step 6: Commit**
```bash
git add converter/pandoc_roundtrip_test.go cmd/jac/main.go
git commit -m "feat: wire blockCard/embedCard round-trip tests and pandoc preset"
```

---

## Task 13: `caption` — config constants

**Files:**
- Modify: `converter/config.go`
- Modify: `mdconverter/config.go`

**Step 1: Add `CaptionStyle` to `converter/config.go`**

```go
// CaptionStyle controls how caption nodes are rendered.
type CaptionStyle string

const (
	CaptionDefault CaptionStyle = "default"
	CaptionPandoc  CaptionStyle = "pandoc"
)
```

Add `CaptionStyle CaptionStyle` to `Config` struct.

In `applyDefaults()`:
```go
if c.CaptionStyle == "" {
    c.CaptionStyle = CaptionDefault
}
```

In `Validate()`:
```go
if c.CaptionStyle != CaptionDefault && c.CaptionStyle != CaptionPandoc {
    return fmt.Errorf("invalid captionStyle %q", c.CaptionStyle)
}
```

**Step 2: Add `CaptionDetection` to `mdconverter/config.go`**

```go
// CaptionDetection controls how caption nodes are reconstructed.
type CaptionDetection string

const (
	CaptionDetectNone   CaptionDetection = "none"
	CaptionDetectPandoc CaptionDetection = "pandoc"
)
```

Add `CaptionDetection CaptionDetection` to `ReverseConfig`.

In `applyDefaults()`:
```go
if c.CaptionDetection == "" {
    c.CaptionDetection = CaptionDetectNone
}
```

In `Validate()`:
```go
if c.CaptionDetection != CaptionDetectNone && c.CaptionDetection != CaptionDetectPandoc {
    return fmt.Errorf("invalid captionDetection %q", c.CaptionDetection)
}
```

Add to `needsPandocInlineExtension()`:
```go
c.CaptionDetection == CaptionDetectPandoc
```

**Step 3: Run tests**
```bash
make test
```

**Step 4: Commit**
```bash
git add converter/config.go mdconverter/config.go
git commit -m "feat: add CaptionStyle/CaptionDetection config for Pandoc round-trip"
```

---

## Task 14: `caption` — forward rendering

**Files:**
- Modify: `converter/media.go` — `convertCaption` (line 130)
- Also update `convertMediaSingle` (the parent that calls `convertCaption`) if needed

**Step 1: Create golden fixture**

`testdata/media/media_caption_pandoc.json`:
```json
{"type":"doc","version":1,"content":[{"type":"mediaSingle","content":[{"type":"media","attrs":{"id":"img-id","type":"image","collection":""}},{"type":"caption","content":[{"type":"text","text":"A photo caption"}]}]}]}
```

`testdata/media/media_caption_pandoc.md`:
```
![Image: img-id][Image: img-id][A photo caption]{.media-caption}
```

> **Design note:** The image is rendered as `![alt](url)` or `[Image: id]` depending on config. For the pandoc case we render the image normally then append `[caption text]{.media-caption}` immediately after (no space), on the same block line. The reverse parser looks for a `media-caption` span trailing an image in the same inline context.

The exact output will depend on what `convertMedia` produces. Read `convertMedia` first to confirm the exact image output before finalizing the golden file. The format `[Image: img-id][A photo caption]{.media-caption}` (no separator) is what the current default path produces for `media_caption.md` — we just wrap the caption part in a Pandoc span.

So `testdata/media/media_caption_pandoc.md` should be:
```
[Image: img-id][A photo caption]{.media-caption}
```

**Step 2: Implement Pandoc path in `convertCaption`**

```go
func (s *state) convertCaption(node Node) (string, error) {
    content, err := s.convertChildrenWithParent(node.Content, node.Type)
    if err != nil {
        return "", err
    }
    if content == "" {
        return "", nil
    }
    if s.config.CaptionStyle == CaptionPandoc {
        return fmt.Sprintf("[%s]{.media-caption}", content), nil
    }
    return content, nil
}
```

**Step 3: Run tests**
```bash
make test
```
The existing `media_caption.md` golden file should still pass (default style).

**Step 4: Commit**
```bash
git add converter/media.go testdata/media/media_caption_pandoc.json testdata/media/media_caption_pandoc.md
git commit -m "feat: render caption as Pandoc span when CaptionStyle=pandoc"
```

---

## Task 15: `caption` — reverse parsing

**Files:**
- Modify: `mdconverter/pandoc_inline_convert.go` — whitelist + add branch
- Modify: `mdconverter/blocks.go` or `mdconverter/media.go` — post-process `mediaSingle` paragraph to extract trailing `caption` span

> **Design:** The caption span `[text]{.media-caption}` appears inline after the image placeholder text in the same paragraph. After assembling the inline content of a paragraph, if the paragraph is ultimately a `mediaSingle` parent, we need to extract any `media-caption` node and attach it as a `caption` child of the `mediaSingle`.
>
> More concretely: when we detect a `mediaSingle` (a paragraph containing only an image or a `[Image: id]` placeholder), we scan the remaining inline nodes for a `media-caption` node and, if found, append it as a `caption` node inside `mediaSingle`.
>
> The simplest place to hook this is in `convertPandocSpanNode` — return a synthetic `caption` node type. Then in the block assembler for `mediaSingle`, pull out any `caption` nodes from inline content.

**Step 1: Whitelist**

Class: add `"media-caption"`.

**Step 2: Add `media-caption` branch**

```go
if hasPandocClass(node.Classes, "media-caption") {
    if !s.shouldDetectCaptionPandoc() {
        return []converter.Node{newTextNode(literal, stack.current())}, nil
    }
    captionContent, err := s.convertInlineFragment(node.Content)
    if err != nil {
        return nil, err
    }
    return []converter.Node{{
        Type:    "caption",
        Content: captionContent,
    }}, nil
}
```

**Step 3: Extract `caption` node from `mediaSingle` inline content**

In `mdconverter/blocks.go` (or wherever `mediaSingle` is assembled), after building the `mediaSingle` content array, look for any `caption` node in the inline content and attach it:

```go
// After building mediaSingle.Content = [media node]
// check if inlineContent contains a caption node
for _, n := range inlineContent {
    if n.Type == "caption" {
        mediaSingle.Content = append(mediaSingle.Content, n)
    }
}
```

This requires understanding how the existing `mediaSingle` reconstruction works. Look at `mdconverter/blocks.go` and `mdconverter/media.go` for `*ast.Image` handling. The `mediaSingle` is likely built in `mdconverter/blocks.go` when a paragraph contains only a single image. The inline parse returns `mediaInline` (or `mediaSingle` content). The caption node will appear as an additional inline node after the image.

**Step 4: Run tests**
```bash
make test
```

**Step 5: Commit**
```bash
git add mdconverter/pandoc_inline_convert.go mdconverter/blocks.go
git commit -m "feat: reverse-parse media-caption Pandoc span to ADF caption node"
```

---

## Task 16: `caption` — round-trip test

**Files:**
- Modify: `converter/pandoc_roundtrip_test.go`

**Step 1: Add `CaptionStyle` to `runPandocRoundTrip` forward config**
```go
CaptionStyle: converter.CaptionPandoc,
```

Reverse:
```go
CaptionDetection: mdconverter.CaptionDetectPandoc,
```

**Step 2: Add test entry**
```go
{name: "media caption", fixturePath: "media/media_caption_pandoc.json"},
```

**Step 3: Update `normalizeRoundTripNodes`**

The `caption` node in round-trip may need `collection` stripped from the `media` child and `caption` content normalized. Check after running tests.

**Step 4: Update pandoc preset**

`cmd/jac/main.go`:
```go
CaptionStyle: converter.CaptionPandoc,
```
Reverse:
```go
CaptionDetection: mdconverter.CaptionDetectPandoc,
```

**Step 5: Run tests and build**
```bash
make check
```

**Step 6: Commit**
```bash
git add converter/pandoc_roundtrip_test.go cmd/jac/main.go
git commit -m "feat: wire caption round-trip test and pandoc preset"
```

---

## Task 17: Final integration check

**Step 1: Run the full test suite**
```bash
make check
```
Expected: all tests pass, zero lint errors.

**Step 2: Smoke-test the CLI with each new feature**

```bash
# annotation
go run ./cmd/jac --preset=pandoc testdata/marks/annotation_pandoc.json

# mediaInline
go run ./cmd/jac --preset=pandoc testdata/media/media_inline_pandoc.json

# blockCard
go run ./cmd/jac --preset=pandoc testdata/inline/block_card_pandoc.json

# embedCard
go run ./cmd/jac --preset=pandoc testdata/inline/embed_card_pandoc.json

# caption
go run ./cmd/jac --preset=pandoc testdata/media/media_caption_pandoc.json
```

Verify output matches expected `.md` golden files.

**Step 3: Smoke-test reverse direction**

```bash
go run ./cmd/jac --reverse --preset=pandoc testdata/marks/annotation_pandoc.md
go run ./cmd/jac --reverse --preset=pandoc testdata/media/media_inline_pandoc.md
go run ./cmd/jac --reverse --preset=pandoc testdata/inline/block_card_pandoc.md
go run ./cmd/jac --reverse --preset=pandoc testdata/inline/embed_card_pandoc.md
go run ./cmd/jac --reverse --preset=pandoc testdata/media/media_caption_pandoc.md
```

Verify output matches original `.json` golden files (modulo `collection`/`localId` normalization).

**Step 4: Final commit if any cleanup needed**
```bash
make check
git add -A
git commit -m "chore: final integration cleanup for lossless round-trip features"
```

---

## Key File Reference

| File | Purpose |
|---|---|
| `converter/config.go` | Add new `*Style` types, `Config` fields, `applyDefaults`, `Validate` |
| `mdconverter/config.go` | Add new `*Detection` types, `ReverseConfig` fields, `applyDefaults`, `Validate`, `needsPandocInlineExtension` |
| `converter/marks.go:290` | `annotation` case in `convertMarkFull` |
| `converter/media.go:124` | `convertMediaInline` |
| `converter/media.go:130` | `convertCaption` |
| `converter/blocks.go:252` | `convertBlockCard` |
| `converter/blocks.go:268` | `convertEmbedCard` |
| `mdconverter/pandoc_inline_convert.go` | `convertPandocSpanNode`, `hasUnknownPandocSpanClass`, `hasUnknownPandocSpanAttr` |
| `mdconverter/blocks.go` | `mediaSingle` paragraph promotion + caption extraction |
| `converter/pandoc_roundtrip_test.go` | `runPandocRoundTrip` config + test table |
| `cmd/jac/main.go` | `pandoc` preset (forward + reverse) |
| `testdata/marks/annotation_pandoc.{json,md}` | New fixture |
| `testdata/media/media_inline_pandoc.{json,md}` | New fixture |
| `testdata/inline/block_card_pandoc.{json,md}` | New fixture |
| `testdata/inline/embed_card_pandoc.{json,md}` | New fixture |
| `testdata/media/media_caption_pandoc.{json,md}` | New fixture |

---

## Caution Points

1. **`blockCard`/`embedCard` node promotion:** These are block-level nodes, not inline. Their Pandoc representation is a span inside a paragraph. The reverse parser must "unwrap" the paragraph when it contains only a single `blockCard` or `embedCard` node. Find the exact paragraph-to-block promotion code in `mdconverter/blocks.go` before implementing.

2. **`caption` attachment:** The caption span appears inline after the image in the same paragraph. The existing `mediaSingle` detection in `mdconverter` wraps lone images in `mediaSingle > media`. After adding caption support, the code must also pick up the trailing `caption` node.

3. **Quoting in Pandoc span attrs:** The existing code uses `fmt.Sprintf(... url=%q ...)` which produces Go-style double-quoted strings. This is consistent with how `mention-id`, `url` etc. are produced in existing Pandoc spans. Ensure your new attrs follow the same convention.

4. **`mediaInline` collection attr:** The round-trip normalization already strips `collection` from `media` nodes. `mediaInline` nodes also typically have a `collection` attr that needs stripping. Update `normalizeRoundTripNodes` accordingly.

5. **Default styles are backward-compatible:** All new `*Style` defaults are `"ignore"` / `"default"`, meaning existing tests and presets are unaffected until explicitly opted in.
