package converter

import (
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strings"
)

func (s *state) convertExtension(node Node) (string, error) {
	extensionKey := node.GetStringAttr("extensionKey", "")
	if extensionKey != "" && s.config.ExtensionHandlers != nil {
		if handler, ok := s.config.ExtensionHandlers[extensionKey]; ok {
			input := ExtensionRenderInput{
				SourcePath: s.options.SourcePath,
				Node:       node,
			}
			output, err := handler.ToMarkdown(s.ctx, input)
			if err != nil {
				return "", err
			}
			if output.Handled {
				var sb strings.Builder
				sb.WriteString("::: { .adf-extension ")
				sb.WriteString(fmt.Sprintf("key=%q", extensionKey))

				keys := make([]string, 0, len(output.Metadata))
				for k := range output.Metadata {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					sb.WriteString(fmt.Sprintf(" %s=%q", k, output.Metadata[k]))
				}
				sb.WriteString(" }\n")
				sb.WriteString(output.Markdown)
				if !strings.HasSuffix(output.Markdown, "\n") {
					sb.WriteString("\n")
				}
				sb.WriteString(":::\n\n")
				return sb.String(), nil
			}
		}
	}

	if node.Type == "bodiedExtension" && s.config.BodiedExtensionStyle != BodiedExtensionJSON {
		return s.convertBodiedExtension(node)
	}

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

func (s *state) convertBodiedExtension(node Node) (string, error) {
	children, err := s.convertChildren(node.Content)
	if err != nil {
		return "", err
	}

	switch s.config.BodiedExtensionStyle {
	case BodiedExtensionStandard:
		return children, nil

	case BodiedExtensionHTML:
		key := node.GetStringAttr("extensionKey", "")
		extType := node.GetStringAttr("extensionType", "")
		params := s.serializeBodiedExtensionParams(node.Attrs)

		var sb strings.Builder
		sb.WriteString("<div class=\"adf-bodied-extension\" ")
		sb.WriteString(fmt.Sprintf("data-extension-key=%q ", html.EscapeString(key)))
		sb.WriteString(fmt.Sprintf("data-extension-type=%q", html.EscapeString(extType)))
		if params != "" {
			sb.WriteString(fmt.Sprintf(" data-parameters=%q", html.EscapeString(params)))
		}
		sb.WriteString(">\n\n")
		sb.WriteString(children)
		if !strings.HasSuffix(children, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("</div>\n\n")
		return sb.String(), nil

	case BodiedExtensionPandoc:
		key := node.GetStringAttr("extensionKey", "")
		extType := node.GetStringAttr("extensionType", "")
		params := s.serializeBodiedExtensionParams(node.Attrs)

		var sb strings.Builder
		sb.WriteString("::: { .adf-bodied-extension ")
		sb.WriteString(fmt.Sprintf("key=%q ", key))
		sb.WriteString(fmt.Sprintf("extensionType=%q", extType))
		if params != "" {
			sb.WriteString(fmt.Sprintf(" parameters=%q", params))
		}
		sb.WriteString(" }\n\n")
		sb.WriteString(children)
		if !strings.HasSuffix(children, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString(":::\n\n")
		return sb.String(), nil

	default:
		return s.renderExtensionJSON(node)
	}
}

func (s *state) serializeBodiedExtensionParams(attrs map[string]interface{}) string {
	params, ok := attrs["parameters"]
	if !ok || params == nil {
		return ""
	}

	data, err := json.Marshal(params)
	if err != nil {
		return ""
	}

	return string(data)
}
