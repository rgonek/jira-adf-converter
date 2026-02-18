package mdconverter

import (
	"encoding/json"
	"strings"

	"github.com/rgonek/jira-adf-converter/converter"
)

func (s *state) parseExtensionFence(language, body string) (converter.Node, bool, error) {
	language = strings.TrimSpace(language)
	switch strings.ToLower(language) {
	case "adf:extension":
		var payload converter.Node
		if err := json.Unmarshal([]byte(body), &payload); err != nil {
			s.addWarning(
				converter.WarningExtensionFallback,
				"adf:extension",
				"invalid extension payload, preserving as code block",
			)
			return converter.Node{}, false, nil
		}
		if strings.TrimSpace(payload.Type) == "" {
			s.addWarning(
				converter.WarningExtensionFallback,
				"adf:extension",
				"extension payload missing type, preserving as code block",
			)
			return converter.Node{}, false, nil
		}
		return payload, true, nil

	case "adf:inlinecard":
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(body), &payload); err != nil {
			s.addWarning(
				converter.WarningExtensionFallback,
				"adf:inlineCard",
				"invalid inline card payload, preserving as code block",
			)
			return converter.Node{}, false, nil
		}
		return converter.Node{
			Type:  "inlineCard",
			Attrs: payload,
		}, true, nil
	}

	return converter.Node{}, false, nil
}
