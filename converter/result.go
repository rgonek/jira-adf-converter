package converter

// Result holds the output of a conversion.
type Result struct {
	Markdown string    `json:"markdown"`
	Warnings []Warning `json:"warnings,omitempty"`
}

// WarningType categorizes conversion warnings.
type WarningType string

const (
	WarningUnknownNode         WarningType = "unknown_node"
	WarningUnknownMark         WarningType = "unknown_mark"
	WarningDroppedFeature      WarningType = "dropped_feature"
	WarningExtensionFallback   WarningType = "extension_fallback"
	WarningMissingAttribute    WarningType = "missing_attribute"
	WarningUnresolvedReference WarningType = "unresolved_reference"
)

// Warning represents a non-fatal issue encountered during conversion.
type Warning struct {
	Type     WarningType `json:"type"`
	NodeType string      `json:"nodeType,omitempty"`
	Message  string      `json:"message"`
}
