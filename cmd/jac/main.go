package main

import (
	"encoding/json"
	"flag"
	"fmt"
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
			TableMode:            converter.TableAutoPandoc,
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
			MentionDetection:  mdconverter.MentionDetectLink,
			EmojiDetection:    mdconverter.EmojiDetectShortcode,
			StatusDetection:   mdconverter.StatusDetectBracket,
			DateDetection:     mdconverter.DateDetectISO,
			PanelDetection:    mdconverter.PanelDetectGitHub,
			ExpandDetection:   mdconverter.ExpandDetectHTML,
			DecisionDetection: mdconverter.DecisionDetectEmoji,
		}, nil
	case presetReadable:
		return mdconverter.ReverseConfig{
			MentionDetection:  mdconverter.MentionDetectAt,
			EmojiDetection:    mdconverter.EmojiDetectShortcode,
			StatusDetection:   mdconverter.StatusDetectText,
			DateDetection:     mdconverter.DateDetectISO,
			PanelDetection:    mdconverter.PanelDetectBold,
			ExpandDetection:   mdconverter.ExpandDetectBlockquote,
			DecisionDetection: mdconverter.DecisionDetectText,
		}, nil
	case presetLossy:
		return mdconverter.ReverseConfig{
			MentionDetection:  mdconverter.MentionDetectNone,
			EmojiDetection:    mdconverter.EmojiDetectNone,
			StatusDetection:   mdconverter.StatusDetectNone,
			DateDetection:     mdconverter.DateDetectNone,
			PanelDetection:    mdconverter.PanelDetectNone,
			ExpandDetection:   mdconverter.ExpandDetectNone,
			DecisionDetection: mdconverter.DecisionDetectNone,
		}, nil
	case presetPandoc:
		return mdconverter.ReverseConfig{
			UnderlineDetection:  mdconverter.UnderlineDetectPandoc,
			SubSupDetection:     mdconverter.SubSupDetectPandoc,
			ColorDetection:      mdconverter.ColorDetectPandoc,
			AlignmentDetection:  mdconverter.AlignDetectPandoc,
			MentionDetection:    mdconverter.MentionDetectPandoc,
			ExpandDetection:     mdconverter.ExpandDetectPandoc,
			InlineCardDetection: mdconverter.InlineCardDetectPandoc,
			TableGridDetection:  true,
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
	reverse := flag.Bool("reverse", false, "Convert Markdown to ADF JSON")
	allowHTML := flag.Bool("allow-html", false, "Enable HTML output")
	strict := flag.Bool("strict", false, "Return error on unknown nodes")
	preset := flag.String("preset", presetBalanced, "Preset: balanced|strict|readable|lossy|pandoc")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: jac [options] <input-file>\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}
	inputFile := args[0]

	data, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	if *reverse {
		cfg, err := resolveReverseConfig(*preset, *allowHTML, *strict)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid preset: %v\n", err)
			os.Exit(1)
		}

		conv, err := mdconverter.New(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid config: %v\n", err)
			os.Exit(1)
		}

		result, err := conv.Convert(string(data))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error converting file: %v\n", err)
			os.Exit(1)
		}

		var parsed any
		if err := json.Unmarshal(result.ADF, &parsed); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing converted ADF JSON: %v\n", err)
			os.Exit(1)
		}
		pretty, err := json.MarshalIndent(parsed, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting ADF JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(pretty))
		return
	}

	cfg, err := resolveConfig(*preset, *allowHTML, *strict)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid preset: %v\n", err)
		os.Exit(1)
	}

	conv, err := converter.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid config: %v\n", err)
		os.Exit(1)
	}

	result, err := conv.Convert(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting file: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(result.Markdown)
}
