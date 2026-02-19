package mdconverter

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReverseConfigDefaultsIncludeResolutionMode(t *testing.T) {
	cfg := (ReverseConfig{}).applyDefaults()
	assert.Equal(t, ResolutionBestEffort, cfg.ResolutionMode)
	assert.Equal(t, UnderlineDetectHTML, cfg.UnderlineDetection)
	assert.Equal(t, SubSupDetectHTML, cfg.SubSupDetection)
	assert.Equal(t, ColorDetectHTML, cfg.ColorDetection)
	assert.Equal(t, AlignDetectHTML, cfg.AlignmentDetection)
	assert.Equal(t, InlineCardDetectNone, cfg.InlineCardDetection)
}

func TestReverseConfigValidateRejectsInvalidResolutionMode(t *testing.T) {
	cfg := (ReverseConfig{}).applyDefaults()
	cfg.ResolutionMode = ResolutionMode("invalid")

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolutionMode")
}

func TestReverseConfigSerializationExcludesHooks(t *testing.T) {
	cfg := (ReverseConfig{
		ResolutionMode: ResolutionStrict,
		LinkHook: func(_ context.Context, _ LinkParseInput) (LinkParseOutput, error) {
			return LinkParseOutput{}, nil
		},
		MediaHook: func(_ context.Context, _ MediaParseInput) (MediaParseOutput, error) {
			return MediaParseOutput{}, nil
		},
	}).applyDefaults()

	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(data), "resolutionMode")
	assert.NotContains(t, string(data), "linkHook")
	assert.NotContains(t, string(data), "mediaHook")
	assert.NotContains(t, string(data), "LinkHook")
	assert.NotContains(t, string(data), "MediaHook")
}

func TestReverseConfigValidateAcceptsPandocDetections(t *testing.T) {
	cfg := (ReverseConfig{}).applyDefaults()
	cfg.MentionDetection = MentionDetectPandoc
	cfg.UnderlineDetection = UnderlineDetectPandoc
	cfg.SubSupDetection = SubSupDetectPandoc
	cfg.ColorDetection = ColorDetectPandoc
	cfg.AlignmentDetection = AlignDetectPandoc
	cfg.ExpandDetection = ExpandDetectPandoc
	cfg.InlineCardDetection = InlineCardDetectPandoc
	require.NoError(t, cfg.Validate())
}

func TestReverseConfigValidateRejectsInvalidDetectionValues(t *testing.T) {
	tests := []struct {
		name string
		mut  func(cfg *ReverseConfig)
	}{
		{
			name: "underline",
			mut: func(cfg *ReverseConfig) {
				cfg.UnderlineDetection = UnderlineDetection("invalid")
			},
		},
		{
			name: "subsup",
			mut: func(cfg *ReverseConfig) {
				cfg.SubSupDetection = SubSupDetection("invalid")
			},
		},
		{
			name: "color",
			mut: func(cfg *ReverseConfig) {
				cfg.ColorDetection = ColorDetection("invalid")
			},
		},
		{
			name: "alignment",
			mut: func(cfg *ReverseConfig) {
				cfg.AlignmentDetection = AlignmentDetection("invalid")
			},
		},
		{
			name: "mention",
			mut: func(cfg *ReverseConfig) {
				cfg.MentionDetection = MentionDetection("invalid")
			},
		},
		{
			name: "expand",
			mut: func(cfg *ReverseConfig) {
				cfg.ExpandDetection = ExpandDetection("invalid")
			},
		},
		{
			name: "inlineCard",
			mut: func(cfg *ReverseConfig) {
				cfg.InlineCardDetection = InlineCardDetection("invalid")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := (ReverseConfig{}).applyDefaults()
			tt.mut(&cfg)
			require.Error(t, cfg.Validate())
		})
	}
}
