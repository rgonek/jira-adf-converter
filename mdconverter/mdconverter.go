package mdconverter

import (
	"encoding/json"
	"fmt"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

// Converter converts GFM markdown to Jira ADF JSON.
type Converter struct {
	config ReverseConfig
	parser goldmark.Markdown
}

type state struct {
	config           ReverseConfig
	source           []byte
	parser           goldmark.Markdown
	warnings         []converter.Warning
	htmlMentionStack []string
}

// New creates a new reverse Converter with the given config.
func New(config ReverseConfig) (*Converter, error) {
	cfg := config.applyDefaults().clone()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Converter{
		config: cfg,
		parser: goldmark.New(
			goldmark.WithExtensions(extension.GFM),
		),
	}, nil
}

// Convert takes a markdown document and returns ADF JSON.
func (c *Converter) Convert(markdown string) (Result, error) {
	s := &state{
		config: c.config,
		source: []byte(markdown),
		parser: c.parser,
	}

	root := c.parser.Parser().Parse(text.NewReader(s.source))
	doc, err := s.convertDocument(root)
	if err != nil {
		return Result{}, err
	}

	adf, err := json.Marshal(doc)
	if err != nil {
		return Result{}, fmt.Errorf("failed to marshal ADF JSON: %w", err)
	}

	return Result{
		ADF:      adf,
		Warnings: s.warnings,
	}, nil
}

func (s *state) addWarning(warnType converter.WarningType, nodeType, message string) {
	s.warnings = append(s.warnings, converter.Warning{
		Type:     warnType,
		NodeType: nodeType,
		Message:  message,
	})
}
