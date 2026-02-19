# Plan: Registry-based Extension Hook System

This plan outlines the implementation of a Registry-based Extension Hook system to support custom transformations of ADF extensions (macros) in both forward (ADF -> Markdown) and reverse (Markdown -> ADF) directions.

## Objectives
- Allow users to register custom handlers for specific ADF extensions (e.g., `plantumlcloud`).
- Support rendering extensions as clean, human-readable Markdown (e.g., code blocks with specific languages).
- Support reconstructing original ADF extension nodes from Markdown patterns.
- Ensure round-trip safety and support for metadata.

## Proposed Changes

### 1. Define `ExtensionHandler` Interface
Location: `converter/extensions.go` (or a new file `converter/extension_handler.go`)

```go
package converter

// ExtensionHandler defines the interface for custom ADF extension transformations.
type ExtensionHandler interface {
	// ToMarkdown converts an ADF extension node's attributes and parameters to Markdown.
	// attrs contains the node's standard attributes (extensionKey, extensionType, etc.).
	// parameters contains the extension-specific parameters.
	ToMarkdown(attrs map[string]any, parameters map[string]any) (string, error)

	// FromMarkdown reconstructs the ADF extension's attributes/parameters from Markdown content and metadata.
	// markdownContent is the content of the detected pattern (e.g., code block body).
	// metadata can contain additional information extracted from the Markdown.
	// It returns a map of attributes to be merged into the ADF extension node.
	// If the map contains a "type" key, it will be used as the ADF node type (extension, inlineExtension).
	// Otherwise, it defaults to "extension".
	// NOTE: Initial implementation focuses on extension and inlineExtension. 
	// bodiedExtension support is deferred.
	FromMarkdown(markdownContent string, metadata map[string]any) (map[string]any, error)
}
```

### 2. Update Configuration Structs

#### `converter.Config` (Forward Direction)
Location: `converter/config.go`

- Add `ExtensionHandlers map[string]ExtensionHandler` to `Config`.
- Key: `extensionKey` (primary) or `extensionType` (fallback).
- Mark as `json:"-"` to exclude from serialization.

#### `mdconverter.ReverseConfig` (Reverse Direction)
Location: `mdconverter/config.go`

- Add `ExtensionHandlers map[string]ExtensionHandler` to `ReverseConfig`.
- Key: Identifier for the Markdown pattern (e.g., code block language like `puml`).
- Mark as `json:"-"`.

### 3. Modify ADF Renderer (Forward)
Location: `converter/extensions.go`

Modify `convertExtension(node Node)`:
1. Identify `extensionKey` (defaulting to `extensionType`).
2. Check if a handler exists in `s.config.ExtensionHandlers` for this key.
3. If found:
   - Extract `parameters` from `node.Attrs["parameters"]` (if they exist).
   - Call `handler.ToMarkdown(node.Attrs, parameters)`.
   - Return the resulting string.
4. If not found, proceed with existing `ExtensionMode` logic (`json`, `text`, `strip`).

### 4. Modify Markdown Parser (Reverse)
Location: `mdconverter/extensions.go`

Modify `parseExtensionFence(language, body string)`:
1. Normalize `language`.
2. Check if a handler exists in `s.config.ExtensionHandlers` for this language.
3. If found:
   - Call `handler.FromMarkdown(body, nil)` (metadata support can be added later if needed).
   - Construct a `converter.Node` using the returned attributes.
   - The `type` of the node should probably be part of what `FromMarkdown` returns or inferred.
     *Recommendation*: `FromMarkdown` should return a map that includes the `type` (extension, inlineExtension, bodiedExtension) and all necessary `attrs`.
4. If not found, proceed with existing `adf:extension` logic.

### 5. Metadata Support
To support metadata (like original filenames or layout settings) that doesn't fit in the code block body:
- For Forward: Handlers can include metadata in the Markdown (e.g., as HTML comments or specifically formatted text).
- For Reverse: We may need to update the `walker` or provide a way for handlers to peek at surrounding context. *Initial implementation will focus on code block body.*

## Verification Plan

### Automated Tests
1. **Unit Test for `plantumlcloud` handler**:
   - Create a mock `ExtensionHandler` for `plantumlcloud`.
   - Verify ADF -> Markdown conversion produces ` ```puml ` code block.
   - Verify Markdown -> ADF conversion reconstructs the original extension node.
2. **Round-trip tests**:
   - Add a new golden file pair in `testdata/extensions/custom_handler.json` and `.md`.
3. **Fallback tests**:
   - Ensure unknown extensions still use the `json` strategy (producing `adf:extension` blocks).

### Manual Verification
- Run `go run ./cmd/jac` with a custom registration (via a temporary test main) to verify CLI behavior.
