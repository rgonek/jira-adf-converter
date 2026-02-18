package mdconverter

import (
	"fmt"
	"strings"
	"time"
)

// MentionDetection controls how mention nodes are reconstructed.
type MentionDetection string

const (
	MentionDetectNone MentionDetection = "none"
	MentionDetectLink MentionDetection = "link"
	MentionDetectAt   MentionDetection = "at"
	MentionDetectHTML MentionDetection = "html"
	MentionDetectAll  MentionDetection = "all"
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
	ExpandDetectAll        ExpandDetection = "all"
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
	MentionDetection  MentionDetection  `json:"mentionDetection,omitempty"`
	EmojiDetection    EmojiDetection    `json:"emojiDetection,omitempty"`
	StatusDetection   StatusDetection   `json:"statusDetection,omitempty"`
	DateDetection     DateDetection     `json:"dateDetection,omitempty"`
	PanelDetection    PanelDetection    `json:"panelDetection,omitempty"`
	ExpandDetection   ExpandDetection   `json:"expandDetection,omitempty"`
	DecisionDetection DecisionDetection `json:"decisionDetection,omitempty"`

	DateFormat      string            `json:"dateFormat,omitempty"`
	HeadingOffset   int               `json:"headingOffset,omitempty"`
	LanguageMap     map[string]string `json:"languageMap,omitempty"`
	MediaBaseURL    string            `json:"mediaBaseURL,omitempty"`
	MentionRegistry map[string]string `json:"mentionRegistry,omitempty"`
	EmojiRegistry   map[string]string `json:"emojiRegistry,omitempty"`
	ResolutionMode  ResolutionMode    `json:"resolutionMode,omitempty"`
	LinkHook        LinkParseHook     `json:"-"`
	MediaHook       MediaParseHook    `json:"-"`
}

func (c ReverseConfig) applyDefaults() ReverseConfig {
	if c.MentionDetection == "" {
		c.MentionDetection = MentionDetectLink
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
	return cloned
}

// Validate checks that config values are valid.
func (c ReverseConfig) Validate() error {
	if c.MentionDetection != MentionDetectNone &&
		c.MentionDetection != MentionDetectLink &&
		c.MentionDetection != MentionDetectAt &&
		c.MentionDetection != MentionDetectHTML &&
		c.MentionDetection != MentionDetectAll {
		return fmt.Errorf("invalid mentionDetection %q", c.MentionDetection)
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
		c.ExpandDetection != ExpandDetectAll {
		return fmt.Errorf("invalid expandDetection %q", c.ExpandDetection)
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
