package converter

import "context"

type ExtensionRenderInput struct {
	SourcePath string
	Node       Node
}

type ExtensionRenderOutput struct {
	Markdown string
	Metadata map[string]string // handler serializes values; framework stores as div attrs
	Handled  bool
}

type ExtensionParseInput struct {
	SourcePath   string
	ExtensionKey string
	Body         string            // raw markdown content inside the .adf-extension div
	Metadata     map[string]string // div attrs (minus key and .adf-extension class)
}

type ExtensionParseOutput struct {
	Node    Node
	Handled bool
}

type ExtensionHandler interface {
	ToMarkdown(ctx context.Context, in ExtensionRenderInput) (ExtensionRenderOutput, error)
	FromMarkdown(ctx context.Context, in ExtensionParseInput) (ExtensionParseOutput, error)
}
