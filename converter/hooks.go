package converter

import (
	"context"
	"errors"
)

// ErrUnresolved indicates that a link or media reference could not be resolved by a hook.
var ErrUnresolved = errors.New("unresolved link or media reference")

// ResolutionMode controls how unresolved hook results are handled.
type ResolutionMode string

const (
	// ResolutionBestEffort continues conversion and falls back to built-in behavior.
	ResolutionBestEffort ResolutionMode = "best_effort"
	// ResolutionStrict fails conversion when a hook returns ErrUnresolved.
	ResolutionStrict ResolutionMode = "strict"
)

// ConvertOptions carries optional per-conversion context.
type ConvertOptions struct {
	SourcePath string
}

// LinkMetadata exposes common typed metadata for link hooks.
type LinkMetadata struct {
	PageID       string
	SpaceKey     string
	AttachmentID string
	Filename     string
	Anchor       string
}

// MediaMetadata exposes common typed metadata for media hooks.
type MediaMetadata struct {
	PageID       string
	SpaceKey     string
	AttachmentID string
	Filename     string
	Anchor       string
}

// LinkRenderHook can rewrite link output during ADF -> Markdown conversion.
type LinkRenderHook func(ctx context.Context, in LinkRenderInput) (LinkRenderOutput, error)

// MediaRenderHook can override media output during ADF -> Markdown conversion.
type MediaRenderHook func(ctx context.Context, in MediaRenderInput) (MediaRenderOutput, error)

// LinkRenderInput describes a link surface being rendered.
type LinkRenderInput struct {
	Source     string
	SourcePath string
	Href       string
	Title      string
	Text       string
	Meta       LinkMetadata
	Attrs      map[string]any
}

// LinkRenderOutput contains hook-provided link rendering data.
type LinkRenderOutput struct {
	Href     string
	Title    string
	TextOnly bool
	Handled  bool
}

// MediaRenderInput describes a media node being rendered.
type MediaRenderInput struct {
	SourcePath string
	MediaType  string
	ID         string
	URL        string
	Alt        string
	Meta       MediaMetadata
	Attrs      map[string]any
}

// MediaRenderOutput contains hook-provided markdown for media rendering.
type MediaRenderOutput struct {
	Markdown string
	Handled  bool
}
