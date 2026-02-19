package converter

import (
	"context"
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
	assert.Equal(t, ResolutionBestEffort, cfg.ResolutionMode)
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
		UnknownNodes:   UnknownSkip,
		UnknownMarks:   UnknownError,
		ResolutionMode: ResolutionStrict,
	}

	require.NoError(t, cfg.Validate())
}

func TestValidateInvalidEnum(t *testing.T) {
	cfg := (Config{}).applyDefaults()
	cfg.MentionStyle = MentionStyle("invalid")
	require.Error(t, cfg.Validate())
}

func TestValidateAcceptsPandocStyles(t *testing.T) {
	cfg := (Config{}).applyDefaults()
	cfg.UnderlineStyle = UnderlinePandoc
	cfg.SubSupStyle = SubSupPandoc
	cfg.TextColorStyle = ColorPandoc
	cfg.BackgroundColorStyle = ColorPandoc
	cfg.MentionStyle = MentionPandoc
	cfg.AlignmentStyle = AlignPandoc
	cfg.ExpandStyle = ExpandPandoc
	cfg.InlineCardStyle = InlineCardPandoc
	cfg.TableMode = TablePandoc
	require.NoError(t, cfg.Validate())

	cfg.TableMode = TableAutoPandoc
	require.NoError(t, cfg.Validate())
}

func TestValidateRejectsInvalidUnderlineStyle(t *testing.T) {
	cfg := (Config{}).applyDefaults()
	cfg.UnderlineStyle = UnderlineStyle("invalid")
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
		ResolutionMode: ResolutionStrict,
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
	assert.Equal(t, cfg.ResolutionMode, decoded.ResolutionMode)
}

func TestConfigSerializationExcludesHooks(t *testing.T) {
	cfg := (Config{
		LinkHook: func(_ context.Context, _ LinkRenderInput) (LinkRenderOutput, error) {
			return LinkRenderOutput{}, nil
		},
		MediaHook: func(_ context.Context, _ MediaRenderInput) (MediaRenderOutput, error) {
			return MediaRenderOutput{}, nil
		},
	}).applyDefaults()

	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "LinkHook")
	assert.NotContains(t, string(data), "MediaHook")
	assert.NotContains(t, string(data), "linkHook")
	assert.NotContains(t, string(data), "mediaHook")
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

func TestValidateDateFormatRejectsLiteralWithoutLayoutTokens(t *testing.T) {
	cfg := (Config{}).applyDefaults()
	cfg.DateFormat = "build-release"

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid dateFormat")
}

func TestValidateDateFormatAcceptsSingleDigitLayoutTokens(t *testing.T) {
	cfg := (Config{}).applyDefaults()
	cfg.DateFormat = "1/2/06 3:4:5 PM"

	require.NoError(t, cfg.Validate())
}
