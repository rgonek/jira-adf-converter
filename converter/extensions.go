package converter

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (s *state) convertExtension(node Node) (string, error) {
	extensionType := node.GetStringAttr("extensionType", "")
	if extensionType == "" {
		extensionType = node.GetStringAttr("extensionKey", "")
	}
	if extensionType == "" {
		extensionType = node.Type
	}

	strategy := s.config.Extensions.ModeFor(extensionType)

	switch strategy {
	case ExtensionStrip:
		s.addWarning(WarningDroppedFeature, node.Type, fmt.Sprintf("extension %q stripped", extensionType))
		return "", nil
	case ExtensionText:
		text, err := s.getExtensionFallbackText(node)
		if err != nil {
			return "", err
		}
		if text == "" {
			s.addWarning(WarningExtensionFallback, node.Type, fmt.Sprintf("extension %q has no fallback text", extensionType))
		}
		return text, nil
	case ExtensionJSON:
		return s.renderExtensionJSON(node)
	default:
		return "", fmt.Errorf("unknown extension strategy: %s", strategy)
	}
}

type extensionJSONNode struct {
	Type    string                 `json:"type"`
	Attrs   map[string]interface{} `json:"attrs,omitempty"`
	Content []Node                 `json:"content,omitempty"`
}

func (s *state) renderExtensionJSON(node Node) (string, error) {
	payload := extensionJSONNode{
		Type:    node.Type,
		Attrs:   node.Attrs,
		Content: node.Content,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal extension node: %w", err)
	}

	return fmt.Sprintf("```adf:extension\n%s\n```\n\n", string(data)), nil
}

func (s *state) getExtensionFallbackText(node Node) (string, error) {
	if len(node.Content) > 0 {
		text, err := s.convertChildren(node.Content)
		if err != nil {
			return "", err
		}
		text = strings.TrimSpace(text)
		if text != "" {
			return text, nil
		}
	}

	return node.GetStringAttr("text", ""), nil
}
