package mdconverter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rgonek/jira-adf-converter/converter"
)

func (s *state) parseExtensionFence(language, body string) (converter.Node, bool, error) {
	language = strings.TrimSpace(language)
	switch strings.ToLower(language) {
	case "adf:extension":
		var payload converter.Node
		if err := json.Unmarshal([]byte(body), &payload); err != nil {
			return converter.Node{}, false, fmt.Errorf("failed to parse adf:extension payload: %w", err)
		}
		if strings.TrimSpace(payload.Type) == "" {
			return converter.Node{}, false, fmt.Errorf("adf:extension payload missing type")
		}
		return payload, true, nil

	case "adf:inlinecard":
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(body), &payload); err != nil {
			return converter.Node{}, false, fmt.Errorf("failed to parse adf:inlineCard payload: %w", err)
		}
		return converter.Node{
			Type:  "inlineCard",
			Attrs: payload,
		}, true, nil
	}

	return converter.Node{}, false, nil
}
