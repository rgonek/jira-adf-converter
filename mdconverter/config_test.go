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
