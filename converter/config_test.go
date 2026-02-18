package converter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyDefaults(t *testing.T) {
	cfg := (Config{}).applyDefaults()

	assert.Equal(t, UnderlineBold, cfg.UnderlineStyle)
	assert.Equal(t, SubSupHTML, cfg.SubSupStyle)
	assert.Equal(t, ColorIgnore, cfg.TextColorStyle)
	assert.Equal(t, ColorIgnore, cfg.BackgroundColorStyle)
	assert.Equal(t, MentionLink, cfg.MentionStyle)
	assert.Equal(t, EmojiShortcode, cfg.EmojiStyle)
	assert.Equal(t, PanelGitHub, cfg.PanelStyle)
	assert.Equal(t, HardBreakBackslash, cfg.HardBreakStyle)
	assert.Equal(t, AlignIgnore, cfg.AlignmentStyle)
	assert.Equal(t, ExpandHTML, cfg.ExpandStyle)
	assert.Equal(t, StatusBracket, cfg.StatusStyle)
	assert.Equal(t, InlineCardLink, cfg.InlineCardStyle)
	assert.Equal(t, DecisionEmoji, cfg.DecisionStyle)
	assert.Equal(t, "2006-01-02", cfg.DateFormat)
	assert.Equal(t, TableAuto, cfg.TableMode)
	assert.Equal(t, rune('-'), cfg.BulletMarker)
	assert.Equal(t, OrderedIncremental, cfg.OrderedListStyle)
	assert.Equal(t, ExtensionJSON, cfg.Extensions.Default)
	assert.Equal(t, UnknownPlaceholder, cfg.UnknownNodes)
	assert.Equal(t, UnknownSkip, cfg.UnknownMarks)
}

func TestValidateValid(t *testing.T) {
	cfg := Config{
		UnderlineStyle:       UnderlineHTML,
		SubSupStyle:          SubSupLaTeX,
		TextColorStyle:       ColorHTML,
		BackgroundColorStyle: ColorIgnore,
		MentionStyle:         MentionHTML,
		EmojiStyle:           EmojiUnicode,
		PanelStyle:           PanelTitle,
		HeadingOffset:        2,
		HardBreakStyle:       HardBreakHTML,
		AlignmentStyle:       AlignHTML,
		ExpandStyle:          ExpandBlockquote,
		StatusStyle:          StatusText,
		InlineCardStyle:      InlineCardEmbed,
		DecisionStyle:        DecisionText,
		DateFormat:           "2006-01-02",
		TableMode:            TablePipe,
		BulletMarker:         '*',
		OrderedListStyle:     OrderedLazy,
		Extensions: ExtensionRules{
			Default: ExtensionJSON,
			ByType: map[string]ExtensionMode{
				"macro": ExtensionText,
			},
		},
		LanguageMap: map[string]string{
			"c++": "cpp",
		},
		UnknownNodes: UnknownSkip,
		UnknownMarks: UnknownError,
	}

	require.NoError(t, cfg.Validate())
}

func TestValidateInvalidEnum(t *testing.T) {
	cfg := (Config{}).applyDefaults()
	cfg.MentionStyle = MentionStyle("invalid")
	require.Error(t, cfg.Validate())
}

func TestValidateInvalidRange(t *testing.T) {
	cfg := (Config{}).applyDefaults()
	cfg.HeadingOffset = 9
	require.Error(t, cfg.Validate())

	cfg = (Config{}).applyDefaults()
	cfg.BulletMarker = 'x'
	require.Error(t, cfg.Validate())
}

func TestExtensionRulesLookup(t *testing.T) {
	rules := ExtensionRules{
		Default: ExtensionJSON,
		ByType: map[string]ExtensionMode{
			"jira-macro": ExtensionText,
		},
	}

	assert.Equal(t, ExtensionText, rules.ModeFor("jira-macro"))
	assert.Equal(t, ExtensionJSON, rules.ModeFor("unknown"))
}

func TestConfigSerialization(t *testing.T) {
	cfg := (Config{
		UnderlineStyle: UnderlineHTML,
		HeadingOffset:  1,
		BulletMarker:   '+',
		Extensions: ExtensionRules{
			Default: ExtensionJSON,
			ByType: map[string]ExtensionMode{
				"custom": ExtensionStrip,
			},
		},
		LanguageMap: map[string]string{
			"js": "javascript",
		},
	}).applyDefaults()

	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	var decoded Config
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, cfg.UnderlineStyle, decoded.UnderlineStyle)
	assert.Equal(t, cfg.HeadingOffset, decoded.HeadingOffset)
	assert.Equal(t, cfg.BulletMarker, decoded.BulletMarker)
	assert.Equal(t, cfg.Extensions.Default, decoded.Extensions.Default)
	assert.Equal(t, cfg.Extensions.ByType["custom"], decoded.Extensions.ByType["custom"])
	assert.Equal(t, cfg.LanguageMap["js"], decoded.LanguageMap["js"])
}

func TestZeroConfigUsable(t *testing.T) {
	conv, err := New(Config{})
	require.NoError(t, err)
	require.NoError(t, conv.config.Validate())
}

func TestValidateLanguageMapEmptyValue(t *testing.T) {
	cfg := (Config{}).applyDefaults()

	cfg.LanguageMap = map[string]string{
		"js": "",
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "languageMap")
}

func TestValidateExtensionsByTypeEmptyKey(t *testing.T) {
	cfg := (Config{}).applyDefaults()

	cfg.Extensions.ByType = map[string]ExtensionMode{
		"": ExtensionText,
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty key")
}
