package converter

import (
	"fmt"
	"strings"
	"time"
)

// UnderlineStyle controls how underline marks are rendered.
type UnderlineStyle string

const (
	UnderlineIgnore UnderlineStyle = "ignore"
	UnderlineBold   UnderlineStyle = "bold"
	UnderlineHTML   UnderlineStyle = "html"
	UnderlinePandoc UnderlineStyle = "pandoc"
)

// SubSupStyle controls how subscript/superscript marks are rendered.
type SubSupStyle string

const (
	SubSupIgnore SubSupStyle = "ignore"
	SubSupHTML   SubSupStyle = "html"
	SubSupLaTeX  SubSupStyle = "latex"
	SubSupPandoc SubSupStyle = "pandoc"
)

// ColorStyle controls how text/background colors are rendered.
type ColorStyle string

const (
	ColorIgnore ColorStyle = "ignore"
	ColorHTML   ColorStyle = "html"
	ColorPandoc ColorStyle = "pandoc"
)

// MentionStyle controls how user mentions are rendered.
type MentionStyle string

const (
	MentionText   MentionStyle = "text"
	MentionLink   MentionStyle = "link"
	MentionHTML   MentionStyle = "html"
	MentionPandoc MentionStyle = "pandoc"
)

// EmojiStyle controls how emoji nodes are rendered.
type EmojiStyle string

const (
	EmojiShortcode EmojiStyle = "shortcode"
	EmojiUnicode   EmojiStyle = "unicode"
)

// PanelStyle controls how Info/Note/Warning panels are rendered.
type PanelStyle string

const (
	PanelNone   PanelStyle = "none"
	PanelBold   PanelStyle = "bold"
	PanelGitHub PanelStyle = "github"
	PanelTitle  PanelStyle = "title"
)

// AlignmentStyle controls how block alignment is rendered.
type AlignmentStyle string

const (
	AlignIgnore AlignmentStyle = "ignore"
	AlignHTML   AlignmentStyle = "html"
	AlignPandoc AlignmentStyle = "pandoc"
)

// HardBreakStyle controls how hard line breaks are rendered.
type HardBreakStyle string

const (
	HardBreakBackslash HardBreakStyle = "backslash"
	HardBreakHTML      HardBreakStyle = "html"
)

// ExpandStyle controls how expand/collapse sections are rendered.
type ExpandStyle string

const (
	ExpandBlockquote ExpandStyle = "blockquote"
	ExpandHTML       ExpandStyle = "html"
	ExpandPandoc     ExpandStyle = "pandoc"
)

// StatusStyle controls how status badges are rendered.
type StatusStyle string

const (
	StatusBracket StatusStyle = "bracket"
	StatusText    StatusStyle = "text"
)

// InlineCardStyle controls how smart links / inline cards are rendered.
type InlineCardStyle string

const (
	InlineCardLink   InlineCardStyle = "link"
	InlineCardURL    InlineCardStyle = "url"
	InlineCardEmbed  InlineCardStyle = "embed"
	InlineCardPandoc InlineCardStyle = "pandoc"
)

// DecisionStyle controls the prefix for decision items.
type DecisionStyle string

const (
	DecisionEmoji DecisionStyle = "emoji"
	DecisionText  DecisionStyle = "text"
)

// OrderedListStyle controls ordered list numbering.
type OrderedListStyle string

const (
	OrderedIncremental OrderedListStyle = "incremental"
	OrderedLazy        OrderedListStyle = "lazy"
)

// LayoutSectionStyle controls how layout sections are rendered.
type LayoutSectionStyle string

const (
	LayoutSectionStandard LayoutSectionStyle = "standard"
	LayoutSectionHTML     LayoutSectionStyle = "html"
	LayoutSectionPandoc   LayoutSectionStyle = "pandoc"
)

// BodiedExtensionStyle controls how bodied extensions are rendered.
type BodiedExtensionStyle string

const (
	BodiedExtensionStandard BodiedExtensionStyle = "standard"
	BodiedExtensionHTML     BodiedExtensionStyle = "html"
	BodiedExtensionPandoc   BodiedExtensionStyle = "pandoc"
	BodiedExtensionJSON     BodiedExtensionStyle = "json"
)

// TableMode controls how tables are rendered.
type TableMode string

const (
	TableAuto       TableMode = "auto"
	TablePipe       TableMode = "pipe"
	TableHTML       TableMode = "html"
	TablePandoc     TableMode = "pandoc"
	TableAutoPandoc TableMode = "autopandoc"
)

// ExtensionMode controls how extension nodes are handled.
type ExtensionMode string

const (
	ExtensionJSON  ExtensionMode = "json"
	ExtensionText  ExtensionMode = "text"
	ExtensionStrip ExtensionMode = "strip"
)

// ExtensionRules allows per-extension-type configuration.
type ExtensionRules struct {
	Default ExtensionMode            `json:"default"`
	ByType  map[string]ExtensionMode `json:"byType,omitempty"`
}

// ModeFor resolves extension mode for a specific extension type.
func (r ExtensionRules) ModeFor(extensionType string) ExtensionMode {
	if extensionType != "" && r.ByType != nil {
		if mode, ok := r.ByType[extensionType]; ok {
			return mode
		}
	}
	return r.Default
}

// UnknownPolicy controls behavior for unrecognized ADF elements.
type UnknownPolicy string

const (
	UnknownError       UnknownPolicy = "error"
	UnknownSkip        UnknownPolicy = "skip"
	UnknownPlaceholder UnknownPolicy = "placeholder"
)

// Config holds all converter configuration options.
type Config struct {
	UnderlineStyle       UnderlineStyle              `json:"underlineStyle,omitempty"`
	SubSupStyle          SubSupStyle                 `json:"subSupStyle,omitempty"`
	TextColorStyle       ColorStyle                  `json:"textColorStyle,omitempty"`
	BackgroundColorStyle ColorStyle                  `json:"backgroundColorStyle,omitempty"`
	MentionStyle         MentionStyle                `json:"mentionStyle,omitempty"`
	EmojiStyle           EmojiStyle                  `json:"emojiStyle,omitempty"`
	PanelStyle           PanelStyle                  `json:"panelStyle,omitempty"`
	HeadingOffset        int                         `json:"headingOffset,omitempty"`
	HardBreakStyle       HardBreakStyle              `json:"hardBreakStyle,omitempty"`
	AlignmentStyle       AlignmentStyle              `json:"alignmentStyle,omitempty"`
	ExpandStyle          ExpandStyle                 `json:"expandStyle,omitempty"`
	StatusStyle          StatusStyle                 `json:"statusStyle,omitempty"`
	InlineCardStyle      InlineCardStyle             `json:"inlineCardStyle,omitempty"`
	LayoutSectionStyle   LayoutSectionStyle          `json:"layoutSectionStyle,omitempty"`
	BodiedExtensionStyle BodiedExtensionStyle        `json:"bodiedExtensionStyle,omitempty"`
	DecisionStyle        DecisionStyle               `json:"decisionStyle,omitempty"`
	DateFormat           string                      `json:"dateFormat,omitempty"`
	TableMode            TableMode                   `json:"tableMode,omitempty"`
	BulletMarker         rune                        `json:"bulletMarker,omitempty"`
	OrderedListStyle     OrderedListStyle            `json:"orderedListStyle,omitempty"`
	Extensions           ExtensionRules              `json:"extensions,omitempty"`
	MediaBaseURL         string                      `json:"mediaBaseURL,omitempty"`
	ResolutionMode       ResolutionMode              `json:"resolutionMode,omitempty"`
	LanguageMap          map[string]string           `json:"languageMap,omitempty"`
	UnknownNodes         UnknownPolicy               `json:"unknownNodes,omitempty"`
	UnknownMarks         UnknownPolicy               `json:"unknownMarks,omitempty"`
	LinkHook             LinkRenderHook              `json:"-"`
	MediaHook            MediaRenderHook             `json:"-"`
	ExtensionHandlers    map[string]ExtensionHandler `json:"-"`
}

func (c Config) applyDefaults() Config {
	if c.UnderlineStyle == "" {
		c.UnderlineStyle = UnderlineBold
	}
	if c.SubSupStyle == "" {
		c.SubSupStyle = SubSupHTML
	}
	if c.TextColorStyle == "" {
		c.TextColorStyle = ColorIgnore
	}
	if c.BackgroundColorStyle == "" {
		c.BackgroundColorStyle = ColorIgnore
	}
	if c.MentionStyle == "" {
		c.MentionStyle = MentionLink
	}
	if c.EmojiStyle == "" {
		c.EmojiStyle = EmojiShortcode
	}
	if c.PanelStyle == "" {
		c.PanelStyle = PanelGitHub
	}
	if c.HardBreakStyle == "" {
		c.HardBreakStyle = HardBreakBackslash
	}
	if c.AlignmentStyle == "" {
		c.AlignmentStyle = AlignIgnore
	}
	if c.ExpandStyle == "" {
		c.ExpandStyle = ExpandHTML
	}
	if c.StatusStyle == "" {
		c.StatusStyle = StatusBracket
	}
	if c.InlineCardStyle == "" {
		c.InlineCardStyle = InlineCardLink
	}
	if c.LayoutSectionStyle == "" {
		c.LayoutSectionStyle = LayoutSectionStandard
	}
	if c.BodiedExtensionStyle == "" {
		c.BodiedExtensionStyle = BodiedExtensionPandoc
	}
	if c.DecisionStyle == "" {
		c.DecisionStyle = DecisionEmoji
	}
	if c.DateFormat == "" {
		c.DateFormat = "2006-01-02"
	}
	if c.TableMode == "" {
		c.TableMode = TableAuto
	}
	if c.BulletMarker == 0 {
		c.BulletMarker = '-'
	}
	if c.OrderedListStyle == "" {
		c.OrderedListStyle = OrderedIncremental
	}
	if c.Extensions.Default == "" {
		c.Extensions.Default = ExtensionJSON
	}
	if c.UnknownNodes == "" {
		c.UnknownNodes = UnknownPlaceholder
	}
	if c.UnknownMarks == "" {
		c.UnknownMarks = UnknownSkip
	}
	if c.ResolutionMode == "" {
		c.ResolutionMode = ResolutionBestEffort
	}

	return c
}

// clone returns a deep copy of Config for map-backed fields.
func (c Config) clone() Config {
	cloned := c
	cloned.Extensions.ByType = cloneExtensionModeMap(c.Extensions.ByType)
	cloned.LanguageMap = cloneStringMap(c.LanguageMap)
	cloned.LinkHook = c.LinkHook
	cloned.MediaHook = c.MediaHook
	cloned.ExtensionHandlers = cloneExtensionHandlerMap(c.ExtensionHandlers)
	return cloned
}

// Validate checks that config values are valid.
func (c Config) Validate() error {
	if c.UnderlineStyle != UnderlineIgnore && c.UnderlineStyle != UnderlineBold && c.UnderlineStyle != UnderlineHTML && c.UnderlineStyle != UnderlinePandoc {
		return fmt.Errorf("invalid underlineStyle %q", c.UnderlineStyle)
	}
	if c.SubSupStyle != SubSupIgnore && c.SubSupStyle != SubSupHTML && c.SubSupStyle != SubSupLaTeX && c.SubSupStyle != SubSupPandoc {
		return fmt.Errorf("invalid subSupStyle %q", c.SubSupStyle)
	}
	if c.TextColorStyle != ColorIgnore && c.TextColorStyle != ColorHTML && c.TextColorStyle != ColorPandoc {
		return fmt.Errorf("invalid textColorStyle %q", c.TextColorStyle)
	}
	if c.BackgroundColorStyle != ColorIgnore && c.BackgroundColorStyle != ColorHTML && c.BackgroundColorStyle != ColorPandoc {
		return fmt.Errorf("invalid backgroundColorStyle %q", c.BackgroundColorStyle)
	}
	if c.MentionStyle != MentionText && c.MentionStyle != MentionLink && c.MentionStyle != MentionHTML && c.MentionStyle != MentionPandoc {
		return fmt.Errorf("invalid mentionStyle %q", c.MentionStyle)
	}
	if c.EmojiStyle != EmojiShortcode && c.EmojiStyle != EmojiUnicode {
		return fmt.Errorf("invalid emojiStyle %q", c.EmojiStyle)
	}
	if c.PanelStyle != PanelNone && c.PanelStyle != PanelBold && c.PanelStyle != PanelGitHub && c.PanelStyle != PanelTitle {
		return fmt.Errorf("invalid panelStyle %q", c.PanelStyle)
	}
	if c.HeadingOffset < 0 || c.HeadingOffset > 5 {
		return fmt.Errorf("headingOffset must be between 0 and 5, got %d", c.HeadingOffset)
	}
	if c.HardBreakStyle != HardBreakBackslash && c.HardBreakStyle != HardBreakHTML {
		return fmt.Errorf("invalid hardBreakStyle %q", c.HardBreakStyle)
	}
	if c.AlignmentStyle != AlignIgnore && c.AlignmentStyle != AlignHTML && c.AlignmentStyle != AlignPandoc {
		return fmt.Errorf("invalid alignmentStyle %q", c.AlignmentStyle)
	}
	if c.ExpandStyle != ExpandBlockquote && c.ExpandStyle != ExpandHTML && c.ExpandStyle != ExpandPandoc {
		return fmt.Errorf("invalid expandStyle %q", c.ExpandStyle)
	}
	if c.StatusStyle != StatusBracket && c.StatusStyle != StatusText {
		return fmt.Errorf("invalid statusStyle %q", c.StatusStyle)
	}
	if c.InlineCardStyle != InlineCardLink && c.InlineCardStyle != InlineCardURL && c.InlineCardStyle != InlineCardEmbed && c.InlineCardStyle != InlineCardPandoc {
		return fmt.Errorf("invalid inlineCardStyle %q", c.InlineCardStyle)
	}
	if c.LayoutSectionStyle != LayoutSectionStandard && c.LayoutSectionStyle != LayoutSectionHTML && c.LayoutSectionStyle != LayoutSectionPandoc {
		return fmt.Errorf("invalid layoutSectionStyle %q", c.LayoutSectionStyle)
	}
	if c.BodiedExtensionStyle != BodiedExtensionStandard && c.BodiedExtensionStyle != BodiedExtensionHTML && c.BodiedExtensionStyle != BodiedExtensionPandoc && c.BodiedExtensionStyle != BodiedExtensionJSON {
		return fmt.Errorf("invalid bodiedExtensionStyle %q", c.BodiedExtensionStyle)
	}
	if c.DecisionStyle != DecisionEmoji && c.DecisionStyle != DecisionText {
		return fmt.Errorf("invalid decisionStyle %q", c.DecisionStyle)
	}
	if c.DateFormat == "" || !hasDateReferenceTokens(c.DateFormat) {
		return fmt.Errorf("invalid dateFormat %q: must contain Go reference date components", c.DateFormat)
	}
	if c.TableMode != TableAuto && c.TableMode != TablePipe && c.TableMode != TableHTML && c.TableMode != TablePandoc && c.TableMode != TableAutoPandoc {
		return fmt.Errorf("invalid tableMode %q", c.TableMode)
	}
	if c.BulletMarker != '-' && c.BulletMarker != '*' && c.BulletMarker != '+' {
		return fmt.Errorf("invalid bulletMarker %q: must be one of -, *, +", c.BulletMarker)
	}
	if c.OrderedListStyle != OrderedIncremental && c.OrderedListStyle != OrderedLazy {
		return fmt.Errorf("invalid orderedListStyle %q", c.OrderedListStyle)
	}
	if c.Extensions.Default != ExtensionJSON && c.Extensions.Default != ExtensionText && c.Extensions.Default != ExtensionStrip {
		return fmt.Errorf("invalid extensions.default %q", c.Extensions.Default)
	}
	for extensionType, mode := range c.Extensions.ByType {
		if strings.TrimSpace(extensionType) == "" {
			return fmt.Errorf("extensions.byType contains empty key")
		}
		if mode != ExtensionJSON && mode != ExtensionText && mode != ExtensionStrip {
			return fmt.Errorf("invalid extensions.byType mode %q for type %q", mode, extensionType)
		}
	}
	for from, to := range c.LanguageMap {
		if strings.TrimSpace(from) == "" || strings.TrimSpace(to) == "" {
			return fmt.Errorf("languageMap keys and values must be non-empty")
		}
	}
	if c.UnknownNodes != UnknownError && c.UnknownNodes != UnknownSkip && c.UnknownNodes != UnknownPlaceholder {
		return fmt.Errorf("invalid unknownNodes policy %q", c.UnknownNodes)
	}
	if c.UnknownMarks != UnknownError && c.UnknownMarks != UnknownSkip && c.UnknownMarks != UnknownPlaceholder {
		return fmt.Errorf("invalid unknownMarks policy %q", c.UnknownMarks)
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

func cloneExtensionModeMap(src map[string]ExtensionMode) map[string]ExtensionMode {
	if src == nil {
		return nil
	}

	dst := make(map[string]ExtensionMode, len(src))
	for key, value := range src {
		dst[key] = value
	}

	return dst
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

func cloneExtensionHandlerMap(src map[string]ExtensionHandler) map[string]ExtensionHandler {
	if src == nil {
		return nil
	}

	dst := make(map[string]ExtensionHandler, len(src))
	for key, value := range src {
		dst[key] = value
	}

	return dst
}
