# Phase 1: Infrastructure & Basic Text

## Overview
This phase establishes the foundational infrastructure for the Jira ADF to GFM converter. We will set up the Go module structure, define the AST types, implement the core converter logic with configuration, create the CLI tool, and build a robust testing harness. By the end of this phase, the converter will support basic text nodes with common formatting marks.

## Deliverables
1. Go module initialized with proper structure
2. `converter/` package with AST definitions and conversion logic
3. `cmd/jac/` CLI tool with `--allow-html` and `--strict` flags
4. Golden file test harness with normal and update modes (supports subdirectories)
5. `Makefile` with standard build, test, and lint targets
6. Support for: `doc`, `paragraph`, `text` nodes and `strong`, `em`, `strike`, `code` marks

---

## Step-by-Step Implementation Plan

### Task 1: Initialize Go Module
**Goal**: Set up the Go module structure and Makefile.

**Steps**:
1. Run `go mod init github.com/rgonek/jira-adf-converter`
2. Run `go get github.com/stretchr/testify/assert` for testing
3. Create directory structure:
   - `converter/` - Library package
   - `cmd/jac/` - CLI tool
   - `testdata/` - Test fixtures
4. Create `Makefile` with standard targets (see [Makefile](#makefile) section)

**Acceptance Criteria**:
- `go.mod` exists with correct module path
- Directory structure exists
- `Makefile` exists with all standard targets
- Can run `make test` without errors

---

### Task 2: Define AST Types
**Goal**: Create Go structs that represent the ADF JSON structure.

**File**: `converter/ast.go`

**ADF JSON Structure to Model**:
```json
{
  "version": 1,
  "type": "doc",
  "content": [
    {
      "type": "paragraph",
      "content": [
        {
          "type": "text",
          "text": "Hello",
          "marks": [
            {"type": "strong"}
          ]
        }
      ]
    }
  ]
}
```

**Required Types**:
- `Doc` - Root document node (fields: Version int, Type string, Content []Node)
- `Node` - Represents any ADF node (fields: Type string, Text string, Content []Node, Marks []Mark, Attrs map[string]interface{})
- `Mark` - Represents text formatting (fields: Type string, Attrs map[string]interface{})

**Acceptance Criteria**:
- Structs can be unmarshaled from JSON using `json.Unmarshal`
- All fields have proper JSON tags
- No business logic in this file - just type definitions

---

### Task 3: Implement Core Converter
**Goal**: Create the Converter struct with configuration and conversion logic.

**File**: `converter/converter.go`

**Implementation Details**:

```go
// Config holds converter configuration
type Config struct {
    AllowHTML bool // If true, use HTML for unsupported features
    Strict    bool // If true, return error on unknown nodes
}

// Converter converts ADF to GFM
type Converter struct {
    config Config
}

// New creates a new Converter with the given config
func New(config Config) *Converter

// Convert takes an ADF JSON document and returns GFM markdown
func (c *Converter) Convert(input []byte) (string, error)
```

**Internal Conversion Methods**:
- `convertNode(node Node) (string, error)` - Main dispatcher, handles all node types
- `convertMark(mark Mark) (string, string)` - Returns prefix and suffix for marks

**Node Type Handlers** (Phase 1):
- `doc` - Process content, concatenate results
- `paragraph` - Wrap content in newlines
- `text` - Return text content with marks applied

**Mark Handlers** (Phase 1):
- `strong` -> `**text**`
- `em` -> `*text*`
- `strike` -> `~~text~~`
- `code` -> `` `text` ``
- Mixed marks use different delimiters (e.g., `strong` + `em` -> `**_text_**`)
- Nested marks are fully supported

**Paragraph Trailing Newlines**:
- Standard paragraphs: Two trailing newlines (`\n\n`)
- Last paragraph in document: Can have fewer trailing newlines (one or none)

**Unknown Node Handling**:
- If `Strict: true` -> return error
- If `Strict: false` -> return `[Unknown node: {type}]` inline (may break markdown flow)

**AllowHTML Flag** (Phase 1):
- Flag is present and passed through Config
- No HTML output in Phase 1 (deferred to Phase 2)
- Phase 1 uses pure GFM markdown only

**Acceptance Criteria**:
- Can convert simple ADF JSON with text and marks
- Respects Strict config for unknown nodes
- All node types return string output
- Mixed marks use alternating delimiters (`**_text_**`)
- Nested marks work correctly
- Last paragraph has appropriate trailing newlines

---

### Task 4: Create CLI Tool
**Goal**: Build the command-line interface.

**File**: `cmd/jac/main.go`

**Command**: `jac` (Jira ADF Converter)

**Flags**:
- `--allow-html` - Enable HTML output (default: false)
- `--strict` - Return error on unknown nodes (default: false)

**Usage**:
```bash
jac input.json                    # Convert file to stdout
jac --allow-html input.json       # Allow HTML in output
jac --strict input.json           # Strict mode
```

**Implementation**:
1. Parse flags using `flag` package
2. Read input file (os.Args last argument)
3. Create Converter with config from flags
4. Convert and print to stdout
5. Handle errors appropriately (exit code 1 on error)

**Acceptance Criteria**:
- `go run cmd/jac/main.go test.json` works
- Flags are properly parsed and passed to converter
- Errors are printed to stderr with exit code 1

---

### Task 5: Create Golden File Test Harness
**Goal**: Implement data-driven tests with golden file support.

**File**: `converter/converter_test.go`

**Test Structure**:
```
testdata/
  basic_text.json
  basic_text.md
  bold_italic.json
  bold_italic.md
  unknown_node.json
  unknown_node.md
```

**Test Runner Logic**:
1. Recursively find all `*.json` files in `testdata/` (support subdirectories)
2. For each JSON file:
   - Parse JSON
   - Convert using default config (or specific config for HTML variants)
   - Read corresponding `.md` file from the same directory
   - Compare output to expected
3. Support `-update` flag to regenerate `.md` files

**Test Cases to Create**:

1. **basic_text** - Simple paragraph with plain text
   ```json
   {"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello World"}]}]}
   ```
   Expected: `Hello World\n\n`

2. **bold** - Strong/bold text
   ```json
   {"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"bold","marks":[{"type":"strong"}]}]}]}
   ```
   Expected: `**bold**\n\n`

3. **italic** - Emphasis/italic text
   ```json
   {"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"italic","marks":[{"type":"em"}]}]}]}
   ```
   Expected: `*italic*\n\n`

4. **strike** - Strikethrough text
   ```json
   {"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"strike","marks":[{"type":"strike"}]}]}]}
   ```
   Expected: `~~strike~~\n\n`

5. **inline_code** - Inline code
   ```json
   {"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"code","marks":[{"type":"code"}]}]}]}
   ```
   Expected: `` `code`\n\n ``

6. **mixed_marks** - Multiple marks on same text
   ```json
   {"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"bold italic","marks":[{"type":"strong"},{"type":"em"}]}]}]}
   ```
   Expected: `**_bold italic_**\n\n` (use mixed delimiters)

7. **unknown_node** - Unknown node type (non-strict mode)
   ```json
   {"type":"doc","content":[{"type":"unknownNode","content":[{"type":"text","text":"test"}]}]}
   ```
   Expected: `[Unknown node: unknownNode]` (inline, may break markdown flow)

8. **multiple_paragraphs** - Multiple paragraphs
   ```json
   {"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Para 1"}]},{"type":"paragraph","content":[{"type":"text","text":"Para 2"}]}]}
   ```
   Expected: `Para 1\n\nPara 2\n\n` (or `Para 1\n\nPara 2\n` for last paragraph)

9. **nested_marks** - Test nested marks support
   ```json
   {"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"bold ","marks":[{"type":"strong"}]},{"type":"text","text":"bold+italic","marks":[{"type":"strong"},{"type":"em"}]},{"type":"text","text":" end","marks":[{"type":"strong"}]}]}]}
   ```
   Expected: `**bold _bold+italic_ end**\n\n`

**Update Mode**:
- Check for `-update` flag in test
- If set, write actual output to `.md` files instead of comparing
- Print message: "Updated testdata/{name}.md"

**Acceptance Criteria**:
- `go test` runs all test cases
- `go test -update` regenerates all `.md` files
- Tests use testify/assert for clear failure messages
- Testdata directory is created with all test fixtures

---

### Task 6: Integration Testing
**Goal**: Ensure the CLI and library work together correctly.

**Test Approach**: Use `os/exec` in tests to run the CLI binary.

**File**: `cmd/jac/main_test.go` (optional, or integrate into converter_test.go)

**Test Cases**:
1. CLI successfully converts a file
2. CLI respects `--allow-html` flag
3. CLI respects `--strict` flag
4. CLI exits with error code on invalid input
5. CLI prints help with `-h`

**Acceptance Criteria**:
- Can build CLI: `make build` (or `go build -o bin/jac cmd/jac/main.go`)
- `./bin/jac testdata/basic_text.json` outputs correct markdown
- Exit codes are correct
- `make lint` passes without errors

---

## Makefile

Create a `Makefile` at project root with these targets:

```makefile
.PHONY: build test test-update lint fmt clean install

# Build the CLI binary
build:
	go build -o bin/jac cmd/jac/main.go

# Run all tests
test:
	go test ./...

# Update golden files
test-update:
	go test ./... -update

# Run linter (go vet)
lint:
	go vet ./...

# Format all Go code
fmt:
	go fmt ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
install:
	go mod download

# Run all checks (fmt, lint, test)
check: fmt lint test

# Default target
all: build
```

## Testing Strategy

### Running Tests

**Normal mode** (validates against expected output):
```bash
make test
# or
go test ./...
```

**Update mode** (regenerates golden files):
```bash
make test-update
# or
go test ./... -update
```

### Test File Naming Convention

- `{feature}.json` - Input ADF JSON
- `{feature}.md` - Expected GFM output
- `{feature}_html.json` - Input for HTML variant tests (uses AllowHTML: true)
- `{feature}_html.md` - Expected HTML output

### Test Data Organization

The test harness must support subdirectories within `testdata/`:

```
testdata/
  phase1/
    basic_text.json
    basic_text.md
    bold.json
    bold.md
  phase2/
    heading.json
    heading.md
```

The harness should recursively find all `*.json` files and match them with corresponding `.md` files in the same directory.

---

## Development Order

Recommended implementation order to minimize blockers:

1. **Task 1** - Initialize module (prerequisite for everything)
2. **Task 2** - Define AST types (needed for unmarshaling)
3. **Task 5** - Create test harness early (create testdata files first, implement harness)
4. **Task 3** - Implement core converter (start with basic_text, expand to marks)
5. **Task 4** - Create CLI tool (depends on converter working)
6. **Task 6** - Integration testing (depends on CLI)

---

## Success Criteria for Phase 1

The phase is complete when:

- [ ] `make test` passes all tests
- [ ] `make lint` passes without errors
- [ ] `make build` creates working binary
- [ ] All testdata files exist and are correct
- [ ] CLI tool can be built and converts JSON to markdown
- [ ] All Phase 1 node types are supported
- [ ] All Phase 1 marks are supported
- [ ] Mixed marks use alternating delimiters (`**_text_**`)
- [ ] Nested marks are fully supported
- [ ] Last paragraph has appropriate trailing newlines
- [ ] Unknown nodes output `[Unknown node: {type}]` inline in non-strict mode
- [ ] Strict mode and AllowHTML flags work (AllowHTML deferred to Phase 2)
- [ ] Test harness supports subdirectories in testdata/
- [ ] Documentation exists (README with basic usage)

---

## Next Phase Preview

Phase 2 will add:
- `heading` (H1-H6)
- `blockquote`
- `rule` (horizontal rule)
- `hardBreak`
- `link` mark
- `subsup` mark with config
- `underline` mark with config
