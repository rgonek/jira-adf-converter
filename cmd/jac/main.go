package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rgonek/jira-adf-converter/converter"
	"github.com/rgonek/jira-adf-converter/mdconverter"
)

const (
	presetBalanced = "balanced"
	presetStrict   = "strict"
	presetReadable = "readable"
	presetLossy    = "lossy"
	presetPandoc   = "pandoc"

	exitCodeOK        = 0
	exitCodeError     = 1
	exitCodeWarnError = 2
)

func presetConfig(preset string) (converter.Config, error) {
	switch strings.ToLower(strings.TrimSpace(preset)) {
	case "", presetBalanced:
		return converter.Config{}, nil
	case presetStrict:
		return converter.Config{
			UnknownNodes: converter.UnknownError,
			UnknownMarks: converter.UnknownError,
			MentionStyle: converter.MentionLink,
			Extensions: converter.ExtensionRules{
				Default: converter.ExtensionJSON,
			},
		}, nil
	case presetReadable:
		return converter.Config{
			MentionStyle:         converter.MentionText,
			TextColorStyle:       converter.ColorIgnore,
			BackgroundColorStyle: converter.ColorIgnore,
			AlignmentStyle:       converter.AlignIgnore,
			ExpandStyle:          converter.ExpandBlockquote,
			Extensions: converter.ExtensionRules{
				Default: converter.ExtensionText,
			},
		}, nil
	case presetLossy:
		return converter.Config{
			MentionStyle:         converter.MentionText,
			TextColorStyle:       converter.ColorIgnore,
			BackgroundColorStyle: converter.ColorIgnore,
			InlineCardStyle:      converter.InlineCardURL,
			Extensions: converter.ExtensionRules{
				Default: converter.ExtensionStrip,
			},
		}, nil
	case presetPandoc:
		return converter.Config{
			UnderlineStyle:       converter.UnderlinePandoc,
			SubSupStyle:          converter.SubSupPandoc,
			TextColorStyle:       converter.ColorPandoc,
			BackgroundColorStyle: converter.ColorPandoc,
			MentionStyle:         converter.MentionPandoc,
			AlignmentStyle:       converter.AlignPandoc,
			ExpandStyle:          converter.ExpandPandoc,
			InlineCardStyle:      converter.InlineCardPandoc,
			AnnotationStyle:      converter.AnnotationPandoc,
			MediaInlineStyle:     converter.MediaInlinePandoc,
			BlockCardStyle:       converter.BlockCardPandoc,
			EmbedCardStyle:       converter.EmbedCardPandoc,
			CaptionStyle:         converter.CaptionPandoc,

			LayoutSectionStyle: converter.LayoutSectionPandoc,
			TableMode:          converter.TableAutoPandoc,
		}, nil
	default:
		return converter.Config{}, fmt.Errorf("unknown preset %q (allowed: balanced, strict, readable, lossy, pandoc)", preset)
	}
}

func resolveConfig(preset string, allowHTML, strict bool) (converter.Config, error) {
	cfg, err := presetConfig(preset)
	if err != nil {
		return converter.Config{}, err
	}

	if allowHTML {
		cfg.UnderlineStyle = converter.UnderlineHTML
		cfg.SubSupStyle = converter.SubSupHTML
		cfg.HardBreakStyle = converter.HardBreakHTML
		cfg.ExpandStyle = converter.ExpandHTML
		cfg.LayoutSectionStyle = converter.LayoutSectionHTML

	}
	if strict {
		cfg.UnknownNodes = converter.UnknownError
		cfg.UnknownMarks = converter.UnknownError
	}

	return cfg, nil
}

func reversePresetConfig(preset string) (mdconverter.ReverseConfig, error) {
	switch strings.ToLower(strings.TrimSpace(preset)) {
	case "", presetBalanced:
		return mdconverter.ReverseConfig{}, nil
	case presetStrict:
		return mdconverter.ReverseConfig{
			MentionDetection: mdconverter.MentionDetectLink,
			EmojiDetection:   mdconverter.EmojiDetectShortcode,
			StatusDetection:  mdconverter.StatusDetectBracket,
			DateDetection:    mdconverter.DateDetectISO,
			PanelDetection:   mdconverter.PanelDetectGitHub,

			LayoutSectionDetection: mdconverter.LayoutSectionDetectHTML,
			ExpandDetection:        mdconverter.ExpandDetectHTML,
			DecisionDetection:      mdconverter.DecisionDetectEmoji,
		}, nil
	case presetReadable:
		return mdconverter.ReverseConfig{
			MentionDetection: mdconverter.MentionDetectAt,
			EmojiDetection:   mdconverter.EmojiDetectShortcode,
			StatusDetection:  mdconverter.StatusDetectText,
			DateDetection:    mdconverter.DateDetectISO,
			PanelDetection:   mdconverter.PanelDetectBold,

			LayoutSectionDetection: mdconverter.LayoutSectionDetectNone,
			ExpandDetection:        mdconverter.ExpandDetectBlockquote,
			DecisionDetection:      mdconverter.DecisionDetectText,
		}, nil
	case presetLossy:
		return mdconverter.ReverseConfig{
			MentionDetection: mdconverter.MentionDetectNone,
			EmojiDetection:   mdconverter.EmojiDetectNone,
			StatusDetection:  mdconverter.StatusDetectNone,
			DateDetection:    mdconverter.DateDetectNone,
			PanelDetection:   mdconverter.PanelDetectNone,

			LayoutSectionDetection: mdconverter.LayoutSectionDetectNone,
			ExpandDetection:        mdconverter.ExpandDetectNone,
			DecisionDetection:      mdconverter.DecisionDetectNone,
		}, nil
	case presetPandoc:
		return mdconverter.ReverseConfig{
			UnderlineDetection:   mdconverter.UnderlineDetectPandoc,
			SubSupDetection:      mdconverter.SubSupDetectPandoc,
			ColorDetection:       mdconverter.ColorDetectPandoc,
			AlignmentDetection:   mdconverter.AlignDetectPandoc,
			MentionDetection:     mdconverter.MentionDetectPandoc,
			ExpandDetection:      mdconverter.ExpandDetectPandoc,
			InlineCardDetection:  mdconverter.InlineCardDetectPandoc,
			AnnotationDetection:  mdconverter.AnnotationDetectPandoc,
			MediaInlineDetection: mdconverter.MediaInlineDetectPandoc,
			BlockCardDetection:   mdconverter.BlockCardDetectPandoc,
			EmbedCardDetection:   mdconverter.EmbedCardDetectPandoc,
			CaptionDetection:     mdconverter.CaptionDetectPandoc,

			LayoutSectionDetection: mdconverter.LayoutSectionDetectPandoc,
			TableGridDetection:     true,
		}, nil
	default:
		return mdconverter.ReverseConfig{}, fmt.Errorf("unknown preset %q (allowed: balanced, strict, readable, lossy, pandoc)", preset)
	}
}

func resolveReverseConfig(preset string, allowHTML, strict bool) (mdconverter.ReverseConfig, error) {
	cfg, err := reversePresetConfig(preset)
	if err != nil {
		return mdconverter.ReverseConfig{}, err
	}

	if allowHTML {
		cfg.UnderlineDetection = mdconverter.UnderlineDetectAll
		cfg.SubSupDetection = mdconverter.SubSupDetectAll
		cfg.ColorDetection = mdconverter.ColorDetectAll
		cfg.AlignmentDetection = mdconverter.AlignDetectAll
		cfg.MentionDetection = mdconverter.MentionDetectAll
		cfg.ExpandDetection = mdconverter.ExpandDetectAll
		cfg.InlineCardDetection = mdconverter.InlineCardDetectAll
	}
	if strict {
		cfg.MentionDetection = mdconverter.MentionDetectLink
		cfg.EmojiDetection = mdconverter.EmojiDetectShortcode
		cfg.StatusDetection = mdconverter.StatusDetectBracket
		cfg.DateDetection = mdconverter.DateDetectISO
		cfg.PanelDetection = mdconverter.PanelDetectGitHub
		cfg.ExpandDetection = mdconverter.ExpandDetectHTML
		cfg.AlignmentDetection = mdconverter.AlignDetectHTML
		cfg.UnderlineDetection = mdconverter.UnderlineDetectHTML
		cfg.SubSupDetection = mdconverter.SubSupDetectHTML
		cfg.ColorDetection = mdconverter.ColorDetectHTML
		cfg.InlineCardDetection = mdconverter.InlineCardDetectLink
		cfg.DecisionDetection = mdconverter.DecisionDetectEmoji
	}

	return cfg, nil
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("jac", flag.ContinueOnError)
	flags.SetOutput(stderr)

	reverse := flags.Bool("reverse", false, "Convert Markdown to ADF JSON")
	allowHTML := flags.Bool("allow-html", false, "Enable HTML output")
	strict := flags.Bool("strict", false, "Return error on unknown nodes")
	failOnWarning := flags.Bool("fail-on-warning", false, "Exit non-zero when conversion emits warnings")
	preset := flags.String("preset", presetBalanced, "Preset: balanced|strict|readable|lossy|pandoc")

	flags.Usage = func() {
		fmt.Fprintf(stderr, "Usage: jac [options] <input-file>\n")
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return exitCodeOK
		}
		return exitCodeError
	}

	if flags.NArg() < 1 {
		flags.Usage()
		return exitCodeError
	}

	inputFile := flags.Arg(0)
	data, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(stderr, "Error reading file: %v\n", err)
		return exitCodeError
	}

	if *reverse {
		return runReverse(string(data), *preset, *allowHTML, *strict, *failOnWarning, stdout, stderr)
	}

	return runForward(data, *preset, *allowHTML, *strict, *failOnWarning, stdout, stderr)
}

func runReverse(markdown, preset string, allowHTML, strict, failOnWarning bool, stdout, stderr io.Writer) int {
	cfg, err := resolveReverseConfig(preset, allowHTML, strict)
	if err != nil {
		fmt.Fprintf(stderr, "Invalid preset: %v\n", err)
		return exitCodeError
	}

	conv, err := mdconverter.New(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "Invalid config: %v\n", err)
		return exitCodeError
	}

	result, err := conv.Convert(markdown)
	if err != nil {
		fmt.Fprintf(stderr, "Error converting file: %v\n", err)
		return exitCodeError
	}

	var parsed any
	if err := json.Unmarshal(result.ADF, &parsed); err != nil {
		fmt.Fprintf(stderr, "Error parsing converted ADF JSON: %v\n", err)
		return exitCodeError
	}

	pretty, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "Error formatting ADF JSON: %v\n", err)
		return exitCodeError
	}

	fmt.Fprintln(stdout, string(pretty))
	printWarnings(stderr, result.Warnings)
	if failOnWarning && len(result.Warnings) > 0 {
		return exitCodeWarnError
	}

	return exitCodeOK
}

func runForward(data []byte, preset string, allowHTML, strict, failOnWarning bool, stdout, stderr io.Writer) int {
	cfg, err := resolveConfig(preset, allowHTML, strict)
	if err != nil {
		fmt.Fprintf(stderr, "Invalid preset: %v\n", err)
		return exitCodeError
	}

	conv, err := converter.New(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "Invalid config: %v\n", err)
		return exitCodeError
	}

	result, err := conv.Convert(data)
	if err != nil {
		fmt.Fprintf(stderr, "Error converting file: %v\n", err)
		return exitCodeError
	}

	fmt.Fprint(stdout, result.Markdown)
	printWarnings(stderr, result.Warnings)
	if failOnWarning && len(result.Warnings) > 0 {
		return exitCodeWarnError
	}

	return exitCodeOK
}

func printWarnings(stderr io.Writer, warnings []converter.Warning) {
	for _, warning := range warnings {
		fmt.Fprintf(stderr, "warning: %s\n", formatWarning(warning))
	}
}

func formatWarning(warning converter.Warning) string {
	parts := []string{fmt.Sprintf("type=%s", warning.Type)}
	if warning.NodeType != "" {
		parts = append(parts, fmt.Sprintf("node=%s", warning.NodeType))
	}
	if warning.ParentType != "" {
		parts = append(parts, fmt.Sprintf("parent=%s", warning.ParentType))
	}
	if warning.Context != "" {
		parts = append(parts, fmt.Sprintf("context=%q", warning.Context))
	}
	parts = append(parts, fmt.Sprintf("message=%q", warning.Message))
	return strings.Join(parts, " ")
}
