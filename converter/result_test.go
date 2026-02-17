package converter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertReturnsWarnings(t *testing.T) {
	conv, err := New(Config{UnknownNodes: UnknownPlaceholder})
	require.NoError(t, err)

	result, err := conv.Convert([]byte(`{"type":"doc","content":[{"type":"mysteryNode"}]}`))
	require.NoError(t, err)

	require.Len(t, result.Warnings, 1)
	assert.Equal(t, WarningUnknownNode, result.Warnings[0].Type)
	assert.Equal(t, "mysteryNode", result.Warnings[0].NodeType)
	assert.Contains(t, result.Markdown, "[Unknown node: mysteryNode]")
}

func TestResultJSONSerialization(t *testing.T) {
	in := Result{
		Markdown: "hello\n",
		Warnings: []Warning{
			{Type: WarningDroppedFeature, NodeType: "extension", Message: "dropped"},
		},
	}

	data, err := json.Marshal(in)
	require.NoError(t, err)

	var out Result
	require.NoError(t, json.Unmarshal(data, &out))
	assert.Equal(t, in, out)
}
