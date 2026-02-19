package main

import (
	"testing"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/rgonek/jira-adf-converter/mdconverter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPresetConfig(t *testing.T) {
	t.Run("balanced", func(t *testing.T) {
		cfg, err := presetConfig(presetBalanced)
		require.NoError(t, err)
		assert.Equal(t, converter.Config{}, cfg)
	})

	t.Run("empty defaults to balanced", func(t *testing.T) {
		cfg, err := presetConfig("")
		require.NoError(t, err)
		assert.Equal(t, converter.Config{}, cfg)
	})

	t.Run("strict", func(t *testing.T) {
		cfg, err := presetConfig(presetStrict)
		require.NoError(t, err)
		assert.Equal(t, converter.UnknownError, cfg.UnknownNodes)
		assert.Equal(t, converter.UnknownError, cfg.UnknownMarks)
		assert.Equal(t, converter.MentionLink, cfg.MentionStyle)
		assert.Equal(t, converter.ExtensionJSON, cfg.Extensions.Default)
	})

	t.Run("readable", func(t *testing.T) {
		cfg, err := presetConfig(presetReadable)
		require.NoError(t, err)
		assert.Equal(t, converter.MentionText, cfg.MentionStyle)
		assert.Equal(t, converter.ColorIgnore, cfg.TextColorStyle)
		assert.Equal(t, converter.ColorIgnore, cfg.BackgroundColorStyle)
		assert.Equal(t, converter.AlignIgnore, cfg.AlignmentStyle)
		assert.Equal(t, converter.ExtensionText, cfg.Extensions.Default)
		assert.Equal(t, converter.ExpandBlockquote, cfg.ExpandStyle)
	})

	t.Run("lossy", func(t *testing.T) {
		cfg, err := presetConfig(presetLossy)
		require.NoError(t, err)
		assert.Equal(t, converter.MentionText, cfg.MentionStyle)
		assert.Equal(t, converter.ColorIgnore, cfg.TextColorStyle)
		assert.Equal(t, converter.ColorIgnore, cfg.BackgroundColorStyle)
		assert.Equal(t, converter.InlineCardURL, cfg.InlineCardStyle)
		assert.Equal(t, converter.ExtensionStrip, cfg.Extensions.Default)
	})

	t.Run("pandoc", func(t *testing.T) {
		cfg, err := presetConfig(presetPandoc)
		require.NoError(t, err)
		assert.Equal(t, converter.UnderlinePandoc, cfg.UnderlineStyle)
		assert.Equal(t, converter.SubSupPandoc, cfg.SubSupStyle)
		assert.Equal(t, converter.ColorPandoc, cfg.TextColorStyle)
		assert.Equal(t, converter.ColorPandoc, cfg.BackgroundColorStyle)
		assert.Equal(t, converter.MentionPandoc, cfg.MentionStyle)
		assert.Equal(t, converter.AlignPandoc, cfg.AlignmentStyle)
		assert.Equal(t, converter.ExpandPandoc, cfg.ExpandStyle)
		assert.Equal(t, converter.InlineCardPandoc, cfg.InlineCardStyle)
		assert.Equal(t, converter.TableAutoPandoc, cfg.TableMode)
	})
}

func TestPresetConfigInvalid(t *testing.T) {
	_, err := presetConfig("unknown")
	require.Error(t, err)
	assert.Equal(t, `unknown preset "unknown" (allowed: balanced, strict, readable, lossy, pandoc)`, err.Error())
}

func TestResolveConfigPresetPrecedence(t *testing.T) {
	cfg, err := resolveConfig(presetReadable, true, true)
	require.NoError(t, err)

	assert.Equal(t, converter.MentionText, cfg.MentionStyle)
	assert.Equal(t, converter.ExtensionText, cfg.Extensions.Default)
	assert.Equal(t, converter.ExpandHTML, cfg.ExpandStyle)
	assert.Equal(t, converter.UnderlineHTML, cfg.UnderlineStyle)
	assert.Equal(t, converter.SubSupHTML, cfg.SubSupStyle)
	assert.Equal(t, converter.HardBreakHTML, cfg.HardBreakStyle)
	assert.Equal(t, converter.UnknownError, cfg.UnknownNodes)
	assert.Equal(t, converter.UnknownError, cfg.UnknownMarks)
}

func TestReversePresetConfig(t *testing.T) {
	t.Run("balanced", func(t *testing.T) {
		cfg, err := reversePresetConfig(presetBalanced)
		require.NoError(t, err)
		assert.Equal(t, mdconverter.ReverseConfig{}, cfg)
	})

	t.Run("strict", func(t *testing.T) {
		cfg, err := reversePresetConfig(presetStrict)
		require.NoError(t, err)
		assert.Equal(t, mdconverter.MentionDetectLink, cfg.MentionDetection)
		assert.Equal(t, mdconverter.EmojiDetectShortcode, cfg.EmojiDetection)
		assert.Equal(t, mdconverter.StatusDetectBracket, cfg.StatusDetection)
		assert.Equal(t, mdconverter.DateDetectISO, cfg.DateDetection)
		assert.Equal(t, mdconverter.PanelDetectGitHub, cfg.PanelDetection)
		assert.Equal(t, mdconverter.ExpandDetectHTML, cfg.ExpandDetection)
		assert.Equal(t, mdconverter.DecisionDetectEmoji, cfg.DecisionDetection)
	})

	t.Run("readable", func(t *testing.T) {
		cfg, err := reversePresetConfig(presetReadable)
		require.NoError(t, err)
		assert.Equal(t, mdconverter.MentionDetectAt, cfg.MentionDetection)
		assert.Equal(t, mdconverter.StatusDetectText, cfg.StatusDetection)
		assert.Equal(t, mdconverter.PanelDetectBold, cfg.PanelDetection)
		assert.Equal(t, mdconverter.ExpandDetectBlockquote, cfg.ExpandDetection)
		assert.Equal(t, mdconverter.DecisionDetectText, cfg.DecisionDetection)
	})

	t.Run("lossy", func(t *testing.T) {
		cfg, err := reversePresetConfig(presetLossy)
		require.NoError(t, err)
		assert.Equal(t, mdconverter.MentionDetectNone, cfg.MentionDetection)
		assert.Equal(t, mdconverter.EmojiDetectNone, cfg.EmojiDetection)
		assert.Equal(t, mdconverter.StatusDetectNone, cfg.StatusDetection)
		assert.Equal(t, mdconverter.DateDetectNone, cfg.DateDetection)
		assert.Equal(t, mdconverter.PanelDetectNone, cfg.PanelDetection)
		assert.Equal(t, mdconverter.ExpandDetectNone, cfg.ExpandDetection)
		assert.Equal(t, mdconverter.DecisionDetectNone, cfg.DecisionDetection)
	})

	t.Run("pandoc", func(t *testing.T) {
		cfg, err := reversePresetConfig(presetPandoc)
		require.NoError(t, err)
		assert.Equal(t, mdconverter.UnderlineDetectPandoc, cfg.UnderlineDetection)
		assert.Equal(t, mdconverter.SubSupDetectPandoc, cfg.SubSupDetection)
		assert.Equal(t, mdconverter.ColorDetectPandoc, cfg.ColorDetection)
		assert.Equal(t, mdconverter.AlignDetectPandoc, cfg.AlignmentDetection)
		assert.Equal(t, mdconverter.MentionDetectPandoc, cfg.MentionDetection)
		assert.Equal(t, mdconverter.ExpandDetectPandoc, cfg.ExpandDetection)
		assert.Equal(t, mdconverter.InlineCardDetectPandoc, cfg.InlineCardDetection)
		assert.True(t, cfg.TableGridDetection)
	})
}

func TestReversePresetConfigInvalid(t *testing.T) {
	_, err := reversePresetConfig("unknown")
	require.Error(t, err)
	assert.Equal(t, `unknown preset "unknown" (allowed: balanced, strict, readable, lossy, pandoc)`, err.Error())
}

func TestResolveReverseConfigPresetPrecedence(t *testing.T) {
	cfg, err := resolveReverseConfig(presetReadable, true, true)
	require.NoError(t, err)

	assert.Equal(t, mdconverter.MentionDetectLink, cfg.MentionDetection)
	assert.Equal(t, mdconverter.EmojiDetectShortcode, cfg.EmojiDetection)
	assert.Equal(t, mdconverter.StatusDetectBracket, cfg.StatusDetection)
	assert.Equal(t, mdconverter.DateDetectISO, cfg.DateDetection)
	assert.Equal(t, mdconverter.PanelDetectGitHub, cfg.PanelDetection)
	assert.Equal(t, mdconverter.ExpandDetectHTML, cfg.ExpandDetection)
	assert.Equal(t, mdconverter.AlignDetectHTML, cfg.AlignmentDetection)
	assert.Equal(t, mdconverter.UnderlineDetectHTML, cfg.UnderlineDetection)
	assert.Equal(t, mdconverter.SubSupDetectHTML, cfg.SubSupDetection)
	assert.Equal(t, mdconverter.ColorDetectHTML, cfg.ColorDetection)
	assert.Equal(t, mdconverter.InlineCardDetectLink, cfg.InlineCardDetection)
	assert.Equal(t, mdconverter.DecisionDetectEmoji, cfg.DecisionDetection)
}
