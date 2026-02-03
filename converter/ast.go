package converter

// Doc represents the root document node of a Jira ADF JSON.
type Doc struct {
	Version int    `json:"version"`
	Type    string `json:"type"`
	Content []Node `json:"content,omitempty"`
}

// Node represents any node in the ADF tree (e.g., paragraph, text, etc.).
type Node struct {
	Type    string                 `json:"type"`
	Text    string                 `json:"text,omitempty"`
	Content []Node                 `json:"content,omitempty"`
	Marks   []Mark                 `json:"marks,omitempty"`
	Attrs   map[string]interface{} `json:"attrs,omitempty"`
}

// Mark represents text formatting applied to a node (e.g., strong, em, etc.).
type Mark struct {
	Type  string                 `json:"type"`
	Attrs map[string]interface{} `json:"attrs,omitempty"`
}
