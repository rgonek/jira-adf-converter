package mdconverter

import (
	"fmt"
	"strings"
	"time"

	"github.com/rgonek/jira-adf-converter/converter"
)

// MentionDetection controls how mention nodes are reconstructed.
type MentionDetection string

const (
	MentionDetectNone   MentionDetection = "none"
	MentionDetectLink   MentionDetection = "link"
	MentionDetectAt     MentionDetection = "at"
	MentionDetectHTML   MentionDetection = "html"
	MentionDetectPandoc MentionDetection = "pandoc"
	MentionDetectAll    MentionDetection = "all"
)

// UnderlineDetection controls how underline marks are reconstructed.
type UnderlineDetection string

const (
	UnderlineDetectNone   UnderlineDetection = "none"
	UnderlineDetectHTML   UnderlineDetection = "html"
	UnderlineDetectPandoc UnderlineDetection = "pandoc"
	UnderlineDetectAll    UnderlineDetection = "all"
)

// SubSupDetection controls how subscript/superscript marks are reconstructed.
type SubSupDetection string

const (
	SubSupDetectNone   SubSupDetection = "none"
	SubSupDetectHTML   SubSupDetection = "html"
	SubSupDetectPandoc SubSupDetection = "pandoc"
	SubSupDetectAll    SubSupDetection = "all"
)

// ColorDetection controls how color marks are reconstructed.
type ColorDetection string

const (
	ColorDetectNone   ColorDetection = "none"
	ColorDetectHTML   ColorDetection = "html"
	ColorDetectPandoc ColorDetection = "pandoc"
	ColorDetectAll    ColorDetection = "all"
)

// AlignmentDetection controls how alignment is reconstructed.
type AlignmentDetection string

const (
	AlignDetectNone   AlignmentDetection = "none"
	AlignDetectHTML   AlignmentDetection = "html"
	AlignDetectPandoc AlignmentDetection = "pandoc"
	AlignDetectAll    AlignmentDetection = "all"
)

// EmojiDetection controls how emoji nodes are reconstructed.
type EmojiDetection string

const (
	EmojiDetectNone      EmojiDetection = "none"
	EmojiDetectShortcode EmojiDetection = "shortcode"
	EmojiDetectUnicode   EmojiDetection = "unicode"
	EmojiDetectAll       EmojiDetection = "all"
)

// StatusDetection controls how status nodes are reconstructed.
type StatusDetection string

const (
	StatusDetectNone    StatusDetection = "none"
	StatusDetectBracket StatusDetection = "bracket"
	StatusDetectText    StatusDetection = "text"
	StatusDetectAll     StatusDetection = "all"
)

// DateDetection controls how date nodes are reconstructed.
type DateDetection string

const (
	DateDetectNone DateDetection = "none"
	DateDetectISO  DateDetection = "iso"
	DateDetectAll  DateDetection = "all"
)

// PanelDetection controls how panel blocks are reconstructed.
type PanelDetection string

const (
	PanelDetectNone   PanelDetection = "none"
	PanelDetectBold   PanelDetection = "bold"
	PanelDetectGitHub PanelDetection = "github"
	PanelDetectTitle  PanelDetection = "title"
	PanelDetectAll    PanelDetection = "all"
)

// ExpandDetection controls how expand blocks are reconstructed.
type ExpandDetection string

const (
	ExpandDetectNone       ExpandDetection = "none"
	ExpandDetectBlockquote ExpandDetection = "blockquote"
	ExpandDetectHTML       ExpandDetection = "html"
	ExpandDetectPandoc     ExpandDetection = "pandoc"
	ExpandDetectAll        ExpandDetection = "all"
)

// InlineCardDetection controls how inline cards are reconstructed.
type InlineCardDetection string

const (
	InlineCardDetectNone   InlineCardDetection = "none"
	InlineCardDetectLink   InlineCardDetection = "link"
	InlineCardDetectPandoc InlineCardDetection = "pandoc"
	InlineCardDetectAll    InlineCardDetection = "all"
)

// DecisionDetection controls how decision blocks are reconstructed.
type DecisionDetection string

const (
	DecisionDetectNone  DecisionDetection = "none"
	DecisionDetectEmoji DecisionDetection = "emoji"
	DecisionDetectText  DecisionDetection = "text"
	DecisionDetectAll   DecisionDetection = "all"
)

// ReverseConfig configures Markdown to ADF conversion behavior.
type ReverseConfig struct {
	MentionDetection    MentionDetection    `json:"mentionDetection,omitempty"`
	UnderlineDetection  UnderlineDetection  `json:"underlineDetection,omitempty"`
	SubSupDetection     SubSupDetection     `json:"subSupDetection,omitempty"`
	ColorDetection      ColorDetection      `json:"colorDetection,omitempty"`
	AlignmentDetection  AlignmentDetection  `json:"alignmentDetection,omitempty"`
	EmojiDetection      EmojiDetection      `json:"emojiDetection,omitempty"`
	StatusDetection     StatusDetection     `json:"statusDetection,omitempty"`
	DateDetection       DateDetection       `json:"dateDetection,omitempty"`
	PanelDetection      PanelDetection      `json:"panelDetection,omitempty"`
	ExpandDetection     ExpandDetection     `json:"expandDetection,omitempty"`
	InlineCardDetection InlineCardDetection `json:"inlineCardDetection,omitempty"`
	TableGridDetection  bool                `json:"tableGridDetection,omitempty"`
	DecisionDetection   DecisionDetection   `json:"decisionDetection,omitempty"`

	DateFormat        string                                `json:"dateFormat,omitempty"`
	HeadingOffset     int                                   `json:"headingOffset,omitempty"`
	LanguageMap       map[string]string                     `json:"languageMap,omitempty"`
	MediaBaseURL      string                                `json:"mediaBaseURL,omitempty"`
	MentionRegistry   map[string]string                     `json:"mentionRegistry,omitempty"`
	EmojiRegistry     map[string]string                     `json:"emojiRegistry,omitempty"`
	ResolutionMode    ResolutionMode                        `json:"resolutionMode,omitempty"`
	LinkHook          LinkParseHook                         `json:"-"`
	MediaHook         MediaParseHook                        `json:"-"`
	ExtensionHandlers map[string]converter.ExtensionHandler `json:"-"`
}

func (c ReverseConfig) applyDefaults() ReverseConfig {
	if c.MentionDetection == "" {
		c.MentionDetection = MentionDetectLink
	}
	if c.UnderlineDetection == "" {
		c.UnderlineDetection = UnderlineDetectHTML
	}
	if c.SubSupDetection == "" {
		c.SubSupDetection = SubSupDetectHTML
	}
	if c.ColorDetection == "" {
		c.ColorDetection = ColorDetectHTML
	}
	if c.AlignmentDetection == "" {
		c.AlignmentDetection = AlignDetectHTML
	}
	if c.EmojiDetection == "" {
		c.EmojiDetection = EmojiDetectShortcode
	}
	if c.StatusDetection == "" {
		c.StatusDetection = StatusDetectBracket
	}
	if c.DateDetection == "" {
		c.DateDetection = DateDetectISO
	}
	if c.PanelDetection == "" {
		c.PanelDetection = PanelDetectGitHub
	}
	if c.ExpandDetection == "" {
		c.ExpandDetection = ExpandDetectHTML
	}
	if c.InlineCardDetection == "" {
		c.InlineCardDetection = InlineCardDetectNone
	}
	if c.DecisionDetection == "" {
		c.DecisionDetection = DecisionDetectEmoji
	}
	if c.DateFormat == "" {
		c.DateFormat = "2006-01-02"
	}
	if c.ResolutionMode == "" {
		c.ResolutionMode = ResolutionBestEffort
	}

	return c
}

func (c ReverseConfig) clone() ReverseConfig {
	cloned := c
	cloned.LanguageMap = cloneStringMap(c.LanguageMap)
	cloned.MentionRegistry = cloneStringMap(c.MentionRegistry)
	cloned.EmojiRegistry = cloneStringMap(c.EmojiRegistry)
	cloned.LinkHook = c.LinkHook
	cloned.MediaHook = c.MediaHook
	cloned.ExtensionHandlers = cloneExtensionHandlerMap(c.ExtensionHandlers)
	return cloned
}

// Validate checks that config values are valid.
func (c ReverseConfig) Validate() error {
	if c.MentionDetection != MentionDetectNone &&
		c.MentionDetection != MentionDetectLink &&
		c.MentionDetection != MentionDetectAt &&
		c.MentionDetection != MentionDetectHTML &&
		c.MentionDetection != MentionDetectPandoc &&
		c.MentionDetection != MentionDetectAll {
		return fmt.Errorf("invalid mentionDetection %q", c.MentionDetection)
	}

	if c.UnderlineDetection != UnderlineDetectNone &&
		c.UnderlineDetection != UnderlineDetectHTML &&
		c.UnderlineDetection != UnderlineDetectPandoc &&
		c.UnderlineDetection != UnderlineDetectAll {
		return fmt.Errorf("invalid underlineDetection %q", c.UnderlineDetection)
	}

	if c.SubSupDetection != SubSupDetectNone &&
		c.SubSupDetection != SubSupDetectHTML &&
		c.SubSupDetection != SubSupDetectPandoc &&
		c.SubSupDetection != SubSupDetectAll {
		return fmt.Errorf("invalid subSupDetection %q", c.SubSupDetection)
	}

	if c.ColorDetection != ColorDetectNone &&
		c.ColorDetection != ColorDetectHTML &&
		c.ColorDetection != ColorDetectPandoc &&
		c.ColorDetection != ColorDetectAll {
		return fmt.Errorf("invalid colorDetection %q", c.ColorDetection)
	}

	if c.AlignmentDetection != AlignDetectNone &&
		c.AlignmentDetection != AlignDetectHTML &&
		c.AlignmentDetection != AlignDetectPandoc &&
		c.AlignmentDetection != AlignDetectAll {
		return fmt.Errorf("invalid alignmentDetection %q", c.AlignmentDetection)
	}

	if c.EmojiDetection != EmojiDetectNone &&
		c.EmojiDetection != EmojiDetectShortcode &&
		c.EmojiDetection != EmojiDetectUnicode &&
		c.EmojiDetection != EmojiDetectAll {
		return fmt.Errorf("invalid emojiDetection %q", c.EmojiDetection)
	}

	if c.StatusDetection != StatusDetectNone &&
		c.StatusDetection != StatusDetectBracket &&
		c.StatusDetection != StatusDetectText &&
		c.StatusDetection != StatusDetectAll {
		return fmt.Errorf("invalid statusDetection %q", c.StatusDetection)
	}

	if c.DateDetection != DateDetectNone &&
		c.DateDetection != DateDetectISO &&
		c.DateDetection != DateDetectAll {
		return fmt.Errorf("invalid dateDetection %q", c.DateDetection)
	}

	if c.PanelDetection != PanelDetectNone &&
		c.PanelDetection != PanelDetectBold &&
		c.PanelDetection != PanelDetectGitHub &&
		c.PanelDetection != PanelDetectTitle &&
		c.PanelDetection != PanelDetectAll {
		return fmt.Errorf("invalid panelDetection %q", c.PanelDetection)
	}

	if c.ExpandDetection != ExpandDetectNone &&
		c.ExpandDetection != ExpandDetectBlockquote &&
		c.ExpandDetection != ExpandDetectHTML &&
		c.ExpandDetection != ExpandDetectPandoc &&
		c.ExpandDetection != ExpandDetectAll {
		return fmt.Errorf("invalid expandDetection %q", c.ExpandDetection)
	}

	if c.InlineCardDetection != InlineCardDetectNone &&
		c.InlineCardDetection != InlineCardDetectLink &&
		c.InlineCardDetection != InlineCardDetectPandoc &&
		c.InlineCardDetection != InlineCardDetectAll {
		return fmt.Errorf("invalid inlineCardDetection %q", c.InlineCardDetection)
	}

	if c.DecisionDetection != DecisionDetectNone &&
		c.DecisionDetection != DecisionDetectEmoji &&
		c.DecisionDetection != DecisionDetectText &&
		c.DecisionDetection != DecisionDetectAll {
		return fmt.Errorf("invalid decisionDetection %q", c.DecisionDetection)
	}

	if c.HeadingOffset < -5 || c.HeadingOffset > 5 {
		return fmt.Errorf("headingOffset must be between -5 and 5, got %d", c.HeadingOffset)
	}

	if c.DateFormat == "" || !hasDateReferenceTokens(c.DateFormat) {
		return fmt.Errorf("invalid dateFormat %q: must contain Go reference date components", c.DateFormat)
	}

	for from, to := range c.LanguageMap {
		if strings.TrimSpace(from) == "" || strings.TrimSpace(to) == "" {
			return fmt.Errorf("languageMap keys and values must be non-empty")
		}
	}

	for name, id := range c.MentionRegistry {
		if strings.TrimSpace(name) == "" || strings.TrimSpace(id) == "" {
			return fmt.Errorf("mentionRegistry keys and values must be non-empty")
		}
	}

	for shortcode, id := range c.EmojiRegistry {
		if strings.TrimSpace(shortcode) == "" || strings.TrimSpace(id) == "" {
			return fmt.Errorf("emojiRegistry keys and values must be non-empty")
		}
	}

	if c.ResolutionMode != ResolutionBestEffort && c.ResolutionMode != ResolutionStrict {
		return fmt.Errorf("invalid resolutionMode %q", c.ResolutionMode)
	}

	return nil
}

func (c ReverseConfig) needsPandocInlineExtension() bool {
	return c.UnderlineDetection == UnderlineDetectPandoc || c.UnderlineDetection == UnderlineDetectAll ||
		c.SubSupDetection == SubSupDetectPandoc || c.SubSupDetection == SubSupDetectAll ||
		c.ColorDetection == ColorDetectPandoc || c.ColorDetection == ColorDetectAll ||
		c.MentionDetection == MentionDetectPandoc || c.MentionDetection == MentionDetectAll ||
		c.InlineCardDetection == InlineCardDetectPandoc || c.InlineCardDetection == InlineCardDetectAll
}

func (c ReverseConfig) needsPandocBlockExtension() bool {
	return c.ExpandDetection == ExpandDetectPandoc || c.ExpandDetection == ExpandDetectAll ||
		c.AlignmentDetection == AlignDetectPandoc || c.AlignmentDetection == AlignDetectAll ||
		len(c.ExtensionHandlers) > 0
}

func hasDateReferenceTokens(format string) bool {
	format = strings.TrimSpace(format)
	if format == "" {
		return false
	}

	_ = time.Now().Format(format)

	referenceTokens := []string{
		"2006", "06", "Jan", "January", "1", "01",
		"2", "02", "_2", "Mon", "Monday", "15", "3", "03", "4", "04",
		"5", "05", "PM", "pm", "MST", "-0700", "-07:00", "Z0700", "Z07:00", "Z07",
	}
	for _, token := range referenceTokens {
		if strings.Contains(format, token) {
			return true
		}
	}
	return false
}

func cloneStringMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}

	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}

	return dst
}

func cloneExtensionHandlerMap(src map[string]converter.ExtensionHandler) map[string]converter.ExtensionHandler {
	if src == nil {
		return nil
	}

	dst := make(map[string]converter.ExtensionHandler, len(src))
	for key, value := range src {
		dst[key] = value
	}

	return dst
}
