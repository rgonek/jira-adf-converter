package converter_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/rgonek/jira-adf-converter/mdconverter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPandocRoundTripFixtures(t *testing.T) {
	tests := []struct {
		name           string
		fixturePath    string
		tableMode      converter.TableMode
		expectWarnings bool
	}{
		{name: "underline", fixturePath: "marks/underline_pandoc.json"},
		{name: "subscript", fixturePath: "marks/subscript_pandoc.json"},
		{name: "superscript", fixturePath: "marks/superscript_pandoc.json"},
		{name: "text color", fixturePath: "marks/text_color_pandoc.json"},
		{name: "background color", fixturePath: "marks/background_color_pandoc.json"},
		{name: "mention", fixturePath: "inline/mention_with_account_id_pandoc.json"},
		{name: "inline card", fixturePath: "inline/inline_card_with_title_pandoc.json"},
		{name: "paragraph alignment", fixturePath: "blocks/paragraph_aligned_center_pandoc.json"},
		{name: "expand with title", fixturePath: "expanders/expand_with_title_pandoc.json"},
		{name: "expand without title", fixturePath: "expanders/expand_without_title_pandoc.json"},
		{name: "nested expand", fixturePath: "expanders/nested_expand_pandoc.json"},
		{name: "simple table grid", fixturePath: "tables/simple_table_pandoc.json", tableMode: converter.TablePandoc},
		{name: "complex table fallback", fixturePath: "tables/complex_table_autopandoc_fallback.json", tableMode: converter.TableAutoPandoc, expectWarnings: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputPath := filepath.Join("..", "testdata", filepath.FromSlash(tt.fixturePath))
			adfInput, err := os.ReadFile(inputPath)
			require.NoError(t, err)

			forwardResult, reverseResult := runPandocRoundTrip(t, adfInput, tt.tableMode)
			if tt.expectWarnings {
				require.NotEmpty(t, forwardResult.Warnings)
			}
			assert.Empty(t, reverseResult.Warnings)
		})
	}
}

func TestPandocRoundTripCombinedInlineFeatures(t *testing.T) {
	adfInput := []byte(`{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"combo","marks":[{"type":"underline"},{"type":"textColor","attrs":{"color":"#0000ff"}}]},{"type":"text","text":" and "},{"type":"mention","attrs":{"id":"abc123","text":"Alice"}}]}]}`)
	runPandocRoundTrip(t, adfInput, converter.TableAutoPandoc)
}

func runPandocRoundTrip(t *testing.T, adfInput []byte, tableMode converter.TableMode) (converter.Result, mdconverter.Result) {
	t.Helper()

	forwardCfg := converter.Config{
		UnderlineStyle:       converter.UnderlinePandoc,
		SubSupStyle:          converter.SubSupPandoc,
		TextColorStyle:       converter.ColorPandoc,
		BackgroundColorStyle: converter.ColorPandoc,
		MentionStyle:         converter.MentionPandoc,
		AlignmentStyle:       converter.AlignPandoc,
		ExpandStyle:          converter.ExpandPandoc,
		InlineCardStyle:      converter.InlineCardPandoc,
		TableMode:            tableMode,
	}
	if forwardCfg.TableMode == "" {
		forwardCfg.TableMode = converter.TableAutoPandoc
	}

	forward, err := converter.New(forwardCfg)
	require.NoError(t, err)

	forwardResult, err := forward.Convert(adfInput)
	require.NoError(t, err)

	reverse, err := mdconverter.New(mdconverter.ReverseConfig{
		UnderlineDetection:  mdconverter.UnderlineDetectPandoc,
		SubSupDetection:     mdconverter.SubSupDetectPandoc,
		ColorDetection:      mdconverter.ColorDetectPandoc,
		AlignmentDetection:  mdconverter.AlignDetectPandoc,
		MentionDetection:    mdconverter.MentionDetectPandoc,
		ExpandDetection:     mdconverter.ExpandDetectPandoc,
		InlineCardDetection: mdconverter.InlineCardDetectPandoc,
		TableGridDetection:  true,
	})
	require.NoError(t, err)

	reverseResult, err := reverse.Convert(forwardResult.Markdown)
	require.NoError(t, err)

	var originalDoc converter.Doc
	var roundTripDoc converter.Doc
	require.NoError(t, json.Unmarshal(adfInput, &originalDoc))
	require.NoError(t, json.Unmarshal(reverseResult.ADF, &roundTripDoc))

	normalizeRoundTripDoc(&originalDoc)
	normalizeRoundTripDoc(&roundTripDoc)
	assert.Equal(t, originalDoc, roundTripDoc)

	return forwardResult, reverseResult
}

func normalizeRoundTripDoc(doc *converter.Doc) {
	if doc.Version == 0 && doc.Type == "doc" {
		doc.Version = 1
	}
	doc.Content = normalizeRoundTripNodes(doc.Content)
}

func normalizeRoundTripNodes(nodes []converter.Node) []converter.Node {
	if len(nodes) == 0 {
		return nil
	}

	normalized := make([]converter.Node, 0, len(nodes))
	for _, node := range nodes {
		node.Content = normalizeRoundTripNodes(node.Content)
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
			if node.Type == "inlineCard" {
				dataMap, _ := node.Attrs["data"].(map[string]interface{})
				if dataMap == nil {
					dataMap = map[string]interface{}{}
				}
				if topURL, ok := node.Attrs["url"].(string); ok && strings.TrimSpace(topURL) != "" {
					if _, exists := dataMap["url"]; !exists {
						dataMap["url"] = topURL
					}
				}
				delete(node.Attrs, "url")
				delete(dataMap, "@type")
				if len(dataMap) > 0 {
					node.Attrs["data"] = dataMap
				}
			}
		}
		if len(node.Attrs) == 0 {
			node.Attrs = nil
		}
		normalized = append(normalized, node)
	}

	return normalized
}
