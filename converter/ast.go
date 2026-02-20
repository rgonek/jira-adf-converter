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
	Level   int                    `json:"level,omitempty"`
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

// GetStringAttr retrieves a string attribute or returns the default value.
func (n Node) GetStringAttr(key, defaultValue string) string {
	if n.Attrs == nil {
		return defaultValue
	}
	if v, ok := n.Attrs[key].(string); ok {
		return v
	}
	return defaultValue
}

// GetIntAttr retrieves an integer attribute (from float64) or returns the default value.
func (n Node) GetIntAttr(key string, defaultValue int) int {
	if n.Attrs == nil {
		return defaultValue
	}
	if v, ok := n.Attrs[key].(float64); ok {
		return int(v)
	}
	return defaultValue
}

// GetFloat64Attr retrieves a float64 attribute or returns the default value.
func (n Node) GetFloat64Attr(key string, defaultValue float64) float64 {
	if n.Attrs == nil {
		return defaultValue
	}
	if v, ok := n.Attrs[key].(float64); ok {
		return v
	}
	return defaultValue
}

// GetStringAttr retrieves a string attribute from a Mark or returns the default value.
func (m Mark) GetStringAttr(key, defaultValue string) string {
	if m.Attrs == nil {
		return defaultValue
	}
	if v, ok := m.Attrs[key].(string); ok {
		return v
	}
	return defaultValue
}
