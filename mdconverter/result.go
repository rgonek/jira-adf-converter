package mdconverter

import "github.com/rgonek/jira-adf-converter/converter"

// Result holds the output of a reverse conversion.
type Result struct {
	ADF      []byte              `json:"adf"`
	Warnings []converter.Warning `json:"warnings,omitempty"`
}
