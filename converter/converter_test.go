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

func newTestConverter(t testing.TB, cfg Config) *Converter {
	t.Helper()

	conv, err := New(cfg)
	require.NoError(t, err)

	return conv
}

func goldenConfigForPath(path string) Config {
	cfg := Config{
		UnderlineStyle: UnderlineIgnore,
		SubSupStyle:    SubSupIgnore,
		MentionStyle:   MentionText,
		ExpandStyle:    ExpandBlockquote,
		UnknownNodes:   UnknownPlaceholder,
		UnknownMarks:   UnknownSkip,
	}

	base := filepath.Base(path)
	if strings.Contains(base, "_html") {
		// Legacy support for existing _html tests, might need refinement
		cfg.UnderlineStyle = UnderlineHTML
		cfg.SubSupStyle = SubSupHTML
		cfg.HardBreakStyle = HardBreakHTML
		cfg.ExpandStyle = ExpandHTML
	}

	// Marks
	if strings.Contains(base, "underline_bold") {
		cfg.UnderlineStyle = UnderlineBold
	}
	if strings.Contains(base, "underline_html") {
		cfg.UnderlineStyle = UnderlineHTML
	}
	if strings.Contains(base, "subsup_latex") {
		cfg.SubSupStyle = SubSupLaTeX
	}
	if strings.Contains(base, "subsup_html") {
		cfg.SubSupStyle = SubSupHTML
	}
	if strings.Contains(base, "color_html") {
		cfg.TextColorStyle = ColorHTML
	}
	if strings.Contains(base, "color_ignore") {
		cfg.TextColorStyle = ColorIgnore
	}
	if strings.Contains(base, "bgcolor_html") {
		cfg.BackgroundColorStyle = ColorHTML
	}

	// Blocks
	if strings.Contains(base, "panel_bold") {
		cfg.PanelStyle = PanelBold
	}
	if strings.Contains(base, "panel_github") {
		cfg.PanelStyle = PanelGitHub
	}
	if strings.Contains(base, "panel_title") {
		cfg.PanelStyle = PanelTitle
	}
	if strings.Contains(base, "align_html") {
		cfg.AlignmentStyle = AlignHTML
	}
	if strings.Contains(base, "expand_html") {
		cfg.ExpandStyle = ExpandHTML
	}
	if strings.Contains(base, "heading_offset1") {
		cfg.HeadingOffset = 1
	}

	// Lists
	if strings.Contains(base, "bullet_star") {
		cfg.BulletMarker = '*'
	}
	if strings.Contains(base, "ordered_lazy") {
		cfg.OrderedListStyle = OrderedLazy
	}

	// Tables
	if strings.Contains(base, "table_auto") {
		cfg.TableMode = TableAuto
	}
	if strings.Contains(base, "table_html") {
		cfg.TableMode = TableHTML
	}

	// Extensions
	if strings.Contains(base, "ext_json") {
		cfg.Extensions.Default = ExtensionJSON
	}
	if strings.Contains(base, "ext_strip") {
		cfg.Extensions.Default = ExtensionStrip
	}
	if strings.Contains(base, "ext_text") {
		cfg.Extensions.Default = ExtensionText
	}

	// Inline
	if strings.Contains(base, "mention_text") {
		cfg.MentionStyle = MentionText
	}
	if strings.Contains(base, "mention_link") {
		cfg.MentionStyle = MentionLink
	}
	if strings.Contains(base, "mention_html") {
		cfg.MentionStyle = MentionHTML
	}
	if strings.Contains(base, "emoji_unicode") {
		cfg.EmojiStyle = EmojiUnicode
	}
	if strings.Contains(base, "status_text") {
		cfg.StatusStyle = StatusText
	}
	if strings.Contains(base, "date_iso") {
		cfg.DateFormat = "2006-01-02"
	}
	if strings.Contains(base, "inlinecard_embed") {
		cfg.InlineCardStyle = InlineCardEmbed
	}

	// Media
	if strings.Contains(base, "media_baseurl") {
		cfg.MediaBaseURL = "https://example.com/media/"
	}

	return cfg
}

func normalizeNewlines(value string) string {
	return strings.ReplaceAll(value, "\r\n", "\n")
}

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

			cfg := goldenConfigForPath(path)
			conv := newTestConverter(t, cfg)
			result, err := conv.Convert(input)
			require.NoError(t, err)
			output := result.Markdown

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

				assert.Equal(t, normalizeNewlines(string(expectedData)), normalizeNewlines(output))
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
		UnknownNodes: UnknownError,
	}
	conv := newTestConverter(t, cfg)

	result, err := conv.Convert(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown node type")
	assert.Empty(t, result.Markdown)
}

func TestNonStrictMode(t *testing.T) {
	// Test that non-strict mode handles unknown nodes gracefully
	input := []byte(`{"type":"doc","content":[{"type":"unknownNode","content":[{"type":"text","text":"test"}]}]}`)

	cfg := Config{
		UnknownNodes: UnknownPlaceholder,
	}
	conv := newTestConverter(t, cfg)

	result, err := conv.Convert(input)
	require.NoError(t, err)
	assert.Contains(t, result.Markdown, "[Unknown node: unknownNode]")
	assert.NotEmpty(t, result.Warnings)
}

func TestStrictModeWithUnknownMark(t *testing.T) {
	// Test that strict mode returns error for truly unknown marks (not underline)
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"colored","marks":[{"type":"textColor"}]}]}]}`)

	cfg := Config{
		UnknownMarks: UnknownError,
	}
	conv := newTestConverter(t, cfg)

	result, err := conv.Convert(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown mark type: textColor")
	assert.Empty(t, result.Markdown)
}

func TestNonStrictModeWithUnknownMark(t *testing.T) {
	// Test that non-strict mode handles unknown marks by preserving text without formatting
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"colored","marks":[{"type":"textColor"}]}]}]}`)

	cfg := Config{
		UnknownMarks: UnknownSkip,
	}
	conv := newTestConverter(t, cfg)

	result, err := conv.Convert(input)
	require.NoError(t, err)
	assert.Contains(t, result.Markdown, "colored")
	assert.NotEmpty(t, result.Warnings)
}

func TestUnderlineWithAllowHTML(t *testing.T) {
	// Test that underline uses <u> tag in HTML mode.
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"underlined","marks":[{"type":"underline"}]}]}]}`)

	cfg := Config{
		UnderlineStyle: UnderlineHTML,
	}
	conv := newTestConverter(t, cfg)

	result, err := conv.Convert(input)
	require.NoError(t, err)
	assert.Contains(t, result.Markdown, "<u>underlined</u>")
}

func TestUnderlineWithoutHTML(t *testing.T) {
	// Test that underline is dropped in ignore mode.
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"underlined","marks":[{"type":"underline"}]}]}]}`)

	cfg := Config{
		UnderlineStyle: UnderlineIgnore,
	}
	conv := newTestConverter(t, cfg)

	result, err := conv.Convert(input)
	require.NoError(t, err)
	assert.Equal(t, "underlined\n", result.Markdown)
}

func TestUnderlineStrictMode(t *testing.T) {
	// Test that strict mode does NOT error for underline (it's a known mark now)
	input := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"underlined","marks":[{"type":"underline"}]}]}]}`)

	cfg := Config{
		UnderlineStyle: UnderlineIgnore,
		UnknownMarks:   UnknownError,
	}
	conv := newTestConverter(t, cfg)

	result, err := conv.Convert(input)
	require.NoError(t, err)
	assert.Equal(t, "underlined\n", result.Markdown)
}

// Unit tests for helper methods

func TestGetMarksToClose(t *testing.T) {
	conv := newTestConverter(t, Config{})
	s := &state{config: conv.config}

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
			result := s.getMarksToCloseFull(tt.activeMarks, tt.currentMarks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMarksToOpen(t *testing.T) {
	conv := newTestConverter(t, Config{})
	s := &state{config: conv.config}

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
			result := s.getMarksToOpenFull(tt.activeMarks, tt.currentMarks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertMark(t *testing.T) {
	conv := newTestConverter(t, Config{UnderlineStyle: UnderlineIgnore})
	s := &state{config: conv.config}

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
			open, close, err := s.convertMarkFull(tt.mark, tt.useUnderscoreForEm)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOpen, open)
			assert.Equal(t, tt.expectedClose, close)
		})
	}
}

func TestConvertMarkWithHTML(t *testing.T) {
	conv := newTestConverter(t, Config{UnderlineStyle: UnderlineHTML})
	s := &state{config: conv.config}

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
			open, close, err := s.convertMarkFull(tt.mark, tt.useUnderscoreForEm)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOpen, open)
			assert.Equal(t, tt.expectedClose, close)
		})
	}
}

// Benchmark tests

func BenchmarkConvertSimpleText(b *testing.B) {
	conv, err := New(Config{})
	if err != nil {
		b.Fatal(err)
	}
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
	conv, err := New(Config{})
	if err != nil {
		b.Fatal(err)
	}
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
	conv, err := New(Config{})
	if err != nil {
		b.Fatal(err)
	}
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
	conv, err := New(Config{})
	if err != nil {
		b.Fatal(err)
	}
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
	conv, err := New(Config{})
	if err != nil {
		b.Fatal(err)
	}
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
