package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rgonek/jira-adf-converter/converter"
)

func main() {
	allowHTML := flag.Bool("allow-html", false, "Enable HTML output")
	strict := flag.Bool("strict", false, "Return error on unknown nodes")
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

	cfg := converter.Config{}
	if *allowHTML {
		cfg.UnderlineStyle = converter.UnderlineHTML
		cfg.SubSupStyle = converter.SubSupHTML
		cfg.HardBreakStyle = converter.HardBreakHTML
		cfg.ExpandStyle = converter.ExpandHTML
	}
	if *strict {
		cfg.UnknownNodes = converter.UnknownError
		cfg.UnknownMarks = converter.UnknownError
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
