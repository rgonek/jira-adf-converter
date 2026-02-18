package main

import (
	"testing"

	"github.com/rgonek/jira-adf-converter/converter"
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
}

func TestPresetConfigInvalid(t *testing.T) {
	_, err := presetConfig("unknown")
	require.Error(t, err)
	assert.Equal(t, `unknown preset "unknown" (allowed: balanced, strict, readable, lossy)`, err.Error())
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
