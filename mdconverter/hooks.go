package mdconverter

import (
	"context"

	"github.com/rgonek/jira-adf-converter/converter"
)

// ErrUnresolved indicates that a link or media reference could not be resolved by a hook.
var ErrUnresolved = converter.ErrUnresolved

// ResolutionMode controls how unresolved hook results are handled.
type ResolutionMode = converter.ResolutionMode

const (
	ResolutionBestEffort ResolutionMode = converter.ResolutionBestEffort
	ResolutionStrict     ResolutionMode = converter.ResolutionStrict
)

// ConvertOptions carries optional per-conversion context.
type ConvertOptions struct {
	SourcePath string
}

// LinkMetadata exposes common typed metadata for link hooks.
type LinkMetadata = converter.LinkMetadata

// MediaMetadata exposes common typed metadata for media hooks.
type MediaMetadata = converter.MediaMetadata

// LinkParseHook can rewrite markdown links during Markdown -> ADF conversion.
type LinkParseHook func(ctx context.Context, in LinkParseInput) (LinkParseOutput, error)

// MediaParseHook can map markdown image/file destinations to media attributes.
type MediaParseHook func(ctx context.Context, in MediaParseInput) (MediaParseOutput, error)

// LinkParseInput describes a markdown link being parsed.
type LinkParseInput struct {
	SourcePath  string
	Destination string
	Title       string
	Text        string
	Meta        LinkMetadata
	Raw         map[string]any
}

// LinkParseOutput contains hook-provided link parsing overrides.
type LinkParseOutput struct {
	Destination string
	Title       string
	ForceLink   bool
	ForceCard   bool
	Handled     bool
}

// MediaParseInput describes a markdown image/file being parsed.
type MediaParseInput struct {
	SourcePath  string
	Destination string
	Alt         string
	Meta        MediaMetadata
	Raw         map[string]any
}

// MediaParseOutput contains hook-provided media parsing overrides.
type MediaParseOutput struct {
	MediaType string
	ID        string
	URL       string
	Alt       string
	Handled   bool
}
