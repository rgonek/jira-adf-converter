package mdconverter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newGoldenReverseConverter(t testing.TB, cfg ReverseConfig) *Converter {
	t.Helper()
	conv, err := New(cfg)
	require.NoError(t, err)
	return conv
}

func reverseGoldenConfigForPath(path string) ReverseConfig {
	cfg := ReverseConfig{
		MentionDetection:  MentionDetectLink,
		EmojiDetection:    EmojiDetectShortcode,
		StatusDetection:   StatusDetectBracket,
		DateDetection:     DateDetectISO,
		PanelDetection:    PanelDetectGitHub,
		ExpandDetection:   ExpandDetectHTML,
		DecisionDetection: DecisionDetectEmoji,
		DateFormat:        "2006-01-02",
		MentionRegistry: map[string]string{
			"User Name": "12345",
			"username":  "12345",
			"Jane":      "42",
		},
	}

	base := filepath.Base(path)
	if strings.Contains(path, string(filepath.Separator)+"panels"+string(filepath.Separator)) {
		cfg.PanelDetection = PanelDetectBold
	}
	if strings.Contains(base, "panel_bold") {
		cfg.PanelDetection = PanelDetectBold
	}
	if strings.Contains(base, "panel_title") {
		cfg.PanelDetection = PanelDetectTitle
	}
	if strings.Contains(base, "decision_text") {
		cfg.DecisionDetection = DecisionDetectText
	}
	if strings.Contains(base, "mention_html") {
		cfg.MentionDetection = MentionDetectHTML
	}
	if strings.Contains(base, "mention_text") || base == "mention.md" {
		cfg.MentionDetection = MentionDetectAt
	}
	if strings.Contains(base, "status_text") {
		cfg.StatusDetection = StatusDetectText
	}
	if strings.Contains(path, string(filepath.Separator)+"expanders"+string(filepath.Separator)) {
		cfg.ExpandDetection = ExpandDetectBlockquote
	}
	if strings.Contains(base, "expand_html") {
		cfg.ExpandDetection = ExpandDetectHTML
	}
	if strings.Contains(base, "language_map_cpp") {
		cfg.LanguageMap = map[string]string{
			"cpp": "c++",
		}
	}
	if strings.Contains(base, "heading_offset1") {
		cfg.HeadingOffset = -1
	}
	if strings.Contains(base, "expand_html_detection_none") {
		cfg.ExpandDetection = ExpandDetectNone
	}
	if strings.Contains(base, "expand_html_detection_blockquote") {
		cfg.ExpandDetection = ExpandDetectBlockquote
	}
	if strings.Contains(base, "media_baseurl_strip_absolute") {
		cfg.MediaBaseURL = "https://example.com/media/"
	}
	if strings.Contains(base, "mention_boundary_retry") {
		cfg.MentionDetection = MentionDetectAt
	}

	return cfg
}

func TestReverseGoldenFiles(t *testing.T) {
	sharedTestDataDir := filepath.Join("..", "testdata")
	reverseOnlyTestDataDir := filepath.Join("testdata")
	fixtures := []string{
		"simple/basic_text",
		"nodes/blockquote",
		"nodes/heading",
		"nodes/rule",
		"nodes/hard_break",
		"codeblocks/basic",
		"codeblocks/language_map_cpp",
		"lists/bullet",
		"lists/ordered",
		"lists/task",
		"lists/mixed",
		"lists/task_rich",
		"lists/task_nested_bug",
		"marks/bold",
		"marks/italic",
		"marks/strike",
		"marks/link_with_title",
		"marks/nested_marks",
		"marks/formatting_html",
		"marks/subsup_html",
		"marks/underline_html",
		"marks/color_html",
		"tables/table_with_headers",
		"inline/emoji",
		"inline/status",
		"inline/mention_link",
		"inline/mention_html",
		"inline/mention",
		"inline/mention_text",
		"inline/inlinecard_embed",
		"inline/inlinecard_embed_with_text",
		"blocks/panel_github",
		"blocks/panel_bold",
		"blocks/panel_title",
		"blocks/align_html",
		"blocks/heading_align_html",
		"blocks/heading_align_html_marks",
		"blocks/heading_offset1",
		"panels/panel_info",
		"panels/panel_warning",
		"panels/panel_multiline",
		"decisions/decision_decided",
		"decisions/decision_undecided",
		"decisions/decision_multiline",
		"decisions/decision_list_multiple",
		"decisions/decision_no_state",
		"expanders/expand",
		"expanders/expand_html",
		"expanders/expand_in_list",
		"expanders/expand_nested",
		"expanders/expand_no_title",
		"tables/table_html",
		"media/media_image_url",
		"inline/inline_card",
		"reverse/expanders/expand_html_in_list",
		"reverse/expanders/expand_html_nested_details",
		"reverse/expanders/expand_html_detection_none",
		"reverse/expanders/expand_html_detection_blockquote",
		"reverse/lists/task_loose_multiblock",
		"reverse/lists/task_inline_patterns_mention_text",
		"reverse/inline/span_nested_lifo_mention_html",
		"reverse/blocks/heading_offset1_align_html",
		"reverse/media/media_baseurl_strip_absolute",
		"reverse/inline/mention_boundary_retry",
		"reverse/smoke/empty",
		"extensions/ext_json",
		"extensions/inline_extension_with_text",
	}

	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture, func(t *testing.T) {
			testDataDir := sharedTestDataDir
			if strings.HasPrefix(fixture, "reverse/") {
				testDataDir = reverseOnlyTestDataDir
			}

			mdPath := filepath.Join(testDataDir, filepath.FromSlash(fixture+".md"))
			jsonPath := filepath.Join(testDataDir, filepath.FromSlash(fixture+".json"))

			markdown, err := os.ReadFile(mdPath)
			require.NoError(t, err)

			expectedJSON, err := os.ReadFile(jsonPath)
			require.NoError(t, err)

			cfg := reverseGoldenConfigForPath(mdPath)
			conv := newGoldenReverseConverter(t, cfg)
			result, err := conv.Convert(string(markdown))
			require.NoError(t, err)

			var actualDoc converter.Doc
			var expectedDoc converter.Doc
			require.NoError(t, json.Unmarshal(result.ADF, &actualDoc))
			require.NoError(t, json.Unmarshal(expectedJSON, &expectedDoc))

			normalizeDoc(&actualDoc)
			normalizeDoc(&expectedDoc)

			assert.Equal(t, expectedDoc, actualDoc)
		})
	}
}

func normalizeDoc(doc *converter.Doc) {
	if doc.Version == 0 && doc.Type == "doc" {
		doc.Version = 1
	}
	doc.Content = normalizeNodes(doc.Content)
}

func normalizeNodes(nodes []converter.Node) []converter.Node {
	if len(nodes) == 0 {
		return nil
	}

	normalized := make([]converter.Node, 0, len(nodes))
	for _, node := range nodes {
		node.Content = normalizeNodes(node.Content)
		node.Text = strings.TrimSuffix(node.Text, "\r")
		if node.Type == "heading" && node.Level > 0 {
			if node.Attrs == nil {
				node.Attrs = map[string]interface{}{}
			}
			if _, hasLevel := node.Attrs["level"]; !hasLevel {
				node.Attrs["level"] = float64(node.Level)
			}
			node.Level = 0
		}
		if len(node.Marks) == 0 {
			node.Marks = nil
		} else {
			for idx := range node.Marks {
				if len(node.Marks[idx].Attrs) == 0 {
					node.Marks[idx].Attrs = nil
				}
			}
		}
		if node.Attrs != nil {
			delete(node.Attrs, "localId")
			if node.Type == "media" {
				delete(node.Attrs, "collection")
			}
		}
		if len(node.Attrs) == 0 {
			node.Attrs = nil
		}
		normalized = append(normalized, node)
	}
	return normalized
}
