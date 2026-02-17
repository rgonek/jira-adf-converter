package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rgonek/jira-adf-converter/converter"
)

const (
	presetBalanced = "balanced"
	presetStrict   = "strict"
	presetReadable = "readable"
	presetLossy    = "lossy"
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
	default:
		return converter.Config{}, fmt.Errorf("unknown preset %q (allowed: balanced, strict, readable, lossy)", preset)
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

func main() {
	allowHTML := flag.Bool("allow-html", false, "Enable HTML output")
	strict := flag.Bool("strict", false, "Return error on unknown nodes")
	preset := flag.String("preset", presetBalanced, "Preset: balanced|strict|readable|lossy")
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
