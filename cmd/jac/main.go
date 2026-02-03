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

	cfg := converter.Config{
		AllowHTML: *allowHTML,
		Strict:    *strict,
	}
	conv := converter.New(cfg)

	output, err := conv.Convert(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting file: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(output)
}
