package mdconverter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Converter converts GFM markdown to Jira ADF JSON.
type Converter struct {
	config ReverseConfig
	parser goldmark.Markdown
}

type state struct {
	config            ReverseConfig
	ctx               context.Context
	options           ConvertOptions
	source            []byte
	parser            goldmark.Markdown
	warnings          []converter.Warning
	htmlMentionStack  []string
	htmlSpanStack     []htmlSpanContext
	pandocExpandDepth int
	htmlExpandDepth   int
}

// New creates a new reverse Converter with the given config.
func New(config ReverseConfig) (*Converter, error) {
	cfg := config.applyDefaults().clone()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	options := []goldmark.Option{
		goldmark.WithExtensions(extension.GFM),
	}
	if cfg.needsPandocInlineExtension() {
		options = append(options, goldmark.WithParserOptions(
			parser.WithInlineParsers(
				util.Prioritized(NewSubscriptParser(), 79),
				util.Prioritized(NewSuperscriptParser(), 79),
				util.Prioritized(NewPandocSpanParser(), 79),
			),
		))
	}
	if cfg.needsPandocBlockExtension() {
		options = append(options, goldmark.WithParserOptions(
			parser.WithBlockParsers(
				util.Prioritized(NewPandocDivParser(), 500),
			),
		))
	}
	if cfg.TableGridDetection {
		options = append(options, goldmark.WithParserOptions(
			parser.WithBlockParsers(
				util.Prioritized(NewPandocGridTableParser(), 501),
			),
		))
	}

	return &Converter{
		config: cfg,
		parser: goldmark.New(options...),
	}, nil
}

// Convert takes a markdown document and returns ADF JSON.
func (c *Converter) Convert(markdown string) (Result, error) {
	return c.ConvertWithContext(context.Background(), markdown, ConvertOptions{})
}

// ConvertWithContext takes a markdown document and returns ADF JSON.
func (c *Converter) ConvertWithContext(ctx context.Context, markdown string, opts ConvertOptions) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	s := &state{
		config:  c.config,
		ctx:     ctx,
		options: opts,
		source:  []byte(markdown),
		parser:  c.parser,
	}

	if err := s.checkContext(); err != nil {
		return Result{}, err
	}

	root := c.parser.Parser().Parse(text.NewReader(s.source))
	if err := s.checkContext(); err != nil {
		return Result{}, err
	}
	doc, err := s.convertDocument(root)
	if err != nil {
		return Result{}, err
	}
	if err := s.checkContext(); err != nil {
		return Result{}, err
	}

	adf, err := json.Marshal(doc)
	if err != nil {
		return Result{}, fmt.Errorf("failed to marshal ADF JSON: %w", err)
	}
	if err := s.checkContext(); err != nil {
		return Result{}, err
	}

	return Result{
		ADF:      adf,
		Warnings: s.warnings,
	}, nil
}

func (s *state) checkContext() error {
	if s.ctx == nil {
		return nil
	}

	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
		return nil
	}
}

func (s *state) addWarning(warnType converter.WarningType, nodeType, message string) {
	s.warnings = append(s.warnings, converter.Warning{
		Type:     warnType,
		NodeType: nodeType,
		Message:  message,
	})
}

func (s *state) shouldDetectBodiedExtensionHTML() bool {
	return s.config.BodiedExtensionDetection == BodiedExtensionDetectHTML ||
		s.config.BodiedExtensionDetection == BodiedExtensionDetectAll
}

func (s *state) shouldDetectBodiedExtensionPandoc() bool {
	return s.config.BodiedExtensionDetection == BodiedExtensionDetectPandoc ||
		s.config.BodiedExtensionDetection == BodiedExtensionDetectAll
}
