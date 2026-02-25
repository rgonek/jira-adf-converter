package mdconverter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindDatePatternForLayout(t *testing.T) {
	match, ok := findDatePatternForLayout("Ship date: 02 Jan 2026", "02 Jan 2006")
	assert.True(t, ok)
	assert.Equal(t, "02 Jan 2026", match.value)
	assert.Equal(t, "1767312000", match.extra)

	_, ok = findDatePatternForLayout("Ship date: 02 Janitor 2026", "02 Jan 2006")
	assert.False(t, ok)

	_, ok = findDatePatternForLayout("Ref: x2026/01/02y", "2006/01/02")
	assert.False(t, ok)
}

func TestDateDetectionLayoutsIncludeISOFallback(t *testing.T) {
	s := &state{config: ReverseConfig{DateFormat: "02 Jan 2006"}}
	layouts := s.dateDetectionLayouts()

	assert.Equal(t, []string{"02 Jan 2006", "2006-01-02"}, layouts)
}
