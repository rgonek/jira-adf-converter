package converter_test

import (
	"encoding/json"
	"testing"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/rgonek/jira-adf-converter/mdconverter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLiteralCharactersRoundTrip(t *testing.T) {
	adfInput := []byte(`{"version":1,"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Literal * _ [ ] ( ) ` + "`" + ` and \\"}]},{"type":"paragraph","content":[{"type":"text","text":"# literal heading"}]},{"type":"paragraph","content":[{"type":"text","text":"> literal quote"}]}]}`)

	forward, err := converter.New(converter.Config{})
	require.NoError(t, err)
	forwardResult, err := forward.Convert(adfInput)
	require.NoError(t, err)

	reverse, err := mdconverter.New(mdconverter.ReverseConfig{})
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
}
