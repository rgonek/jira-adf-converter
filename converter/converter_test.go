package converter

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

func TestGoldenFiles(t *testing.T) {
	testDataDir := "../testdata"

	err := filepath.Walk(testDataDir, func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err)

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".json" {
			return nil
		}

		// Run test for this JSON file
		t.Run(path, func(t *testing.T) {
			input, err := os.ReadFile(path)
			require.NoError(t, err)

			// Determine expected output file path
			goldenPath := strings.TrimSuffix(path, ".json") + ".md"

			// Configure converter based on filename or defaults
			// For now, default config.
			// The plan mentions {feature}_html.json for AllowHTML: true tests.
			cfg := Config{
				AllowHTML: strings.Contains(filepath.Base(path), "_html"),
				Strict:    false, // Default to false unless we want to test strict specifically
			}

			// For unknown_node.json, we expect specific output in non-strict mode.
			// If we wanted to test strict mode failure, we might need a separate naming convention or test logic.
			// But for now, let's stick to the generated output verification.

			conv := New(cfg)
			output, err := conv.Convert(input)
			require.NoError(t, err)

			if *update {
				err := os.WriteFile(goldenPath, []byte(output), 0644)
				require.NoError(t, err)
				t.Logf("Updated golden file: %s", goldenPath)
			} else {
				// Read expected output
				// If .md file doesn't exist yet, fail or treat as empty
				expectedData, err := os.ReadFile(goldenPath)
				if os.IsNotExist(err) {
					// If strictly running without update, missing golden file should fail
					t.Fatalf("Golden file missing: %s. Run with -update to create it.", goldenPath)
				}
				require.NoError(t, err)

				assert.Equal(t, string(expectedData), output)
			}
		})

		return nil
	})
	require.NoError(t, err)
}

func TestStrictMode(t *testing.T) {
	// Test that strict mode returns an error for unknown node types
	input := []byte(`{"type":"doc","content":[{"type":"unknownNode","content":[{"type":"text","text":"test"}]}]}`)

	cfg := Config{
		AllowHTML: false,
		Strict:    true,
	}
	conv := New(cfg)

	output, err := conv.Convert(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown node type")
	assert.Empty(t, output)
}

func TestNonStrictMode(t *testing.T) {
	// Test that non-strict mode handles unknown nodes gracefully
	input := []byte(`{"type":"doc","content":[{"type":"unknownNode","content":[{"type":"text","text":"test"}]}]}`)

	cfg := Config{
		AllowHTML: false,
		Strict:    false,
	}
	conv := New(cfg)

	output, err := conv.Convert(input)
	require.NoError(t, err)
	assert.Contains(t, output, "[Unknown node: unknownNode]")
}

func TestStrictModeWithUnknownMark(t *testing.T) {
	// Test that strict mode returns error for truly unknown marks (not underline)
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"colored","marks":[{"type":"textColor"}]}]}]}`)

	cfg := Config{
		AllowHTML: false,
		Strict:    true,
	}
	conv := New(cfg)

	output, err := conv.Convert(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown mark type: textColor")
	assert.Empty(t, output)
}

func TestNonStrictModeWithUnknownMark(t *testing.T) {
	// Test that non-strict mode handles unknown marks by preserving text without formatting
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"underlined","marks":[{"type":"underline"}]}]}]}`)

	cfg := Config{
		AllowHTML: false,
		Strict:    false,
	}
	conv := New(cfg)

	output, err := conv.Convert(input)
	require.NoError(t, err)
	// Text is preserved, but formatting is lost (no placeholder notation)
	assert.Contains(t, output, "underlined")
	assert.NotContains(t, output, "[underline:")
}

func TestUnderlineWithAllowHTML(t *testing.T) {
	// Test that underline uses <u> tag when AllowHTML is enabled
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"underlined","marks":[{"type":"underline"}]}]}]}`)

	cfg := Config{
		AllowHTML: true,
		Strict:    false,
	}
	conv := New(cfg)

	output, err := conv.Convert(input)
	require.NoError(t, err)
	assert.Contains(t, output, "<u>underlined</u>")
}

func TestUnderlineWithoutAllowHTML(t *testing.T) {
	// Test that underline is dropped when AllowHTML is disabled
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"underlined","marks":[{"type":"underline"}]}]}]}`)

	cfg := Config{
		AllowHTML: false,
		Strict:    false,
	}
	conv := New(cfg)

	output, err := conv.Convert(input)
	require.NoError(t, err)
	// Should contain the text but no underline markup
	// Note: doc adds single \n at end, paragraph adds \n\n during processing
	assert.Equal(t, "underlined\n", output)
}

func TestUnderlineStrictMode(t *testing.T) {
	// Test that strict mode does NOT error for underline (it's a known mark now)
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"underlined","marks":[{"type":"underline"}]}]}]}`)

	cfg := Config{
		AllowHTML: false,
		Strict:    true,
	}
	conv := New(cfg)

	output, err := conv.Convert(input)
	require.NoError(t, err)
	assert.Equal(t, "underlined\n", output)
}

// Unit tests for helper methods

func TestGetMarksToClose(t *testing.T) {
	conv := New(Config{})

	tests := []struct {
		name         string
		activeMarks  []Mark
		currentMarks []Mark
		expected     []Mark
	}{
		{
			name:         "no active marks",
			activeMarks:  []Mark{},
			currentMarks: []Mark{{Type: "strong"}},
			expected:     nil,
		},
		{
			name:         "same marks",
			activeMarks:  []Mark{{Type: "strong"}, {Type: "em"}},
			currentMarks: []Mark{{Type: "strong"}, {Type: "em"}},
			expected:     nil,
		},
		{
			name:         "close all marks",
			activeMarks:  []Mark{{Type: "strong"}, {Type: "em"}},
			currentMarks: []Mark{},
			expected:     []Mark{{Type: "strong"}, {Type: "em"}},
		},
		{
			name:         "close one mark",
			activeMarks:  []Mark{{Type: "strong"}, {Type: "em"}},
			currentMarks: []Mark{{Type: "strong"}},
			expected:     []Mark{{Type: "em"}},
		},
		{
			name:         "different mark at same position",
			activeMarks:  []Mark{{Type: "strong"}},
			currentMarks: []Mark{{Type: "em"}},
			expected:     []Mark{{Type: "strong"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := conv.getMarksToCloseFull(tt.activeMarks, tt.currentMarks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMarksToOpen(t *testing.T) {
	conv := New(Config{})

	tests := []struct {
		name         string
		activeMarks  []Mark
		currentMarks []Mark
		expected     []Mark
	}{
		{
			name:         "no current marks",
			activeMarks:  []Mark{{Type: "strong"}},
			currentMarks: []Mark{},
			expected:     nil,
		},
		{
			name:         "same marks",
			activeMarks:  []Mark{{Type: "strong"}, {Type: "em"}},
			currentMarks: []Mark{{Type: "strong"}, {Type: "em"}},
			expected:     nil,
		},
		{
			name:         "open all marks",
			activeMarks:  []Mark{},
			currentMarks: []Mark{{Type: "strong"}, {Type: "em"}},
			expected:     []Mark{{Type: "strong"}, {Type: "em"}},
		},
		{
			name:         "open one mark",
			activeMarks:  []Mark{{Type: "strong"}},
			currentMarks: []Mark{{Type: "strong"}, {Type: "em"}},
			expected:     []Mark{{Type: "em"}},
		},
		{
			name:         "different mark at same position",
			activeMarks:  []Mark{{Type: "strong"}},
			currentMarks: []Mark{{Type: "em"}},
			expected:     []Mark{{Type: "em"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := conv.getMarksToOpenFull(tt.activeMarks, tt.currentMarks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertMark(t *testing.T) {
	conv := New(Config{})

	tests := []struct {
		name               string
		mark               Mark
		useUnderscoreForEm bool
		expectedOpen       string
		expectedClose      string
	}{
		{
			name:               "strong",
			mark:               Mark{Type: "strong"},
			useUnderscoreForEm: false,
			expectedOpen:       "**",
			expectedClose:      "**",
		},
		{
			name:               "em with asterisk",
			mark:               Mark{Type: "em"},
			useUnderscoreForEm: false,
			expectedOpen:       "*",
			expectedClose:      "*",
		},
		{
			name:               "em with underscore",
			mark:               Mark{Type: "em"},
			useUnderscoreForEm: true,
			expectedOpen:       "_",
			expectedClose:      "_",
		},
		{
			name:               "strike",
			mark:               Mark{Type: "strike"},
			useUnderscoreForEm: false,
			expectedOpen:       "~~",
			expectedClose:      "~~",
		},
		{
			name:               "code",
			mark:               Mark{Type: "code"},
			useUnderscoreForEm: false,
			expectedOpen:       "`",
			expectedClose:      "`",
		},
		{
			name:               "underline without HTML",
			mark:               Mark{Type: "underline"},
			useUnderscoreForEm: false,
			expectedOpen:       "",
			expectedClose:      "",
		},
		{
			name:               "unknown mark",
			mark:               Mark{Type: "unknown"},
			useUnderscoreForEm: false,
			expectedOpen:       "",
			expectedClose:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			open, close, err := conv.convertMarkFull(tt.mark, tt.useUnderscoreForEm)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOpen, open)
			assert.Equal(t, tt.expectedClose, close)
		})
	}
}

func TestConvertMarkWithHTML(t *testing.T) {
	conv := New(Config{AllowHTML: true})

	tests := []struct {
		name               string
		mark               Mark
		useUnderscoreForEm bool
		expectedOpen       string
		expectedClose      string
	}{
		{
			name:               "underline with HTML",
			mark:               Mark{Type: "underline"},
			useUnderscoreForEm: false,
			expectedOpen:       "<u>",
			expectedClose:      "</u>",
		},
		{
			name:               "strong still uses markdown",
			mark:               Mark{Type: "strong"},
			useUnderscoreForEm: false,
			expectedOpen:       "**",
			expectedClose:      "**",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			open, close, err := conv.convertMarkFull(tt.mark, tt.useUnderscoreForEm)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOpen, open)
			assert.Equal(t, tt.expectedClose, close)
		})
	}
}

// Benchmark tests

func BenchmarkConvertSimpleText(b *testing.B) {
	conv := New(Config{})
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello World"}]}]}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := conv.Convert(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConvertWithMarks(b *testing.B) {
	conv := New(Config{})
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"bold italic","marks":[{"type":"strong"},{"type":"em"}]}]}]}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := conv.Convert(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConvertNestedMarks(b *testing.B) {
	conv := New(Config{})
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"bold ","marks":[{"type":"strong"}]},{"type":"text","text":"bold+italic","marks":[{"type":"strong"},{"type":"em"}]},{"type":"text","text":" end","marks":[{"type":"strong"}]}]}]}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := conv.Convert(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConvertMultipleParagraphs(b *testing.B) {
	conv := New(Config{})
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Para 1"}]},{"type":"paragraph","content":[{"type":"text","text":"Para 2"}]},{"type":"paragraph","content":[{"type":"text","text":"Para 3"}]}]}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := conv.Convert(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConvertLargeDocument(b *testing.B) {
	conv := New(Config{})
	// Create a document with 100 paragraphs
	var sb strings.Builder
	sb.WriteString(`{"type":"doc","content":[`)
	for i := 0; i < 100; i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(`{"type":"paragraph","content":[{"type":"text","text":"Paragraph `)
		sb.WriteString(string(rune('0' + (i % 10))))
		sb.WriteString(`"}]}`)
	}
	sb.WriteString(`]}`)
	input := []byte(sb.String())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := conv.Convert(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}
