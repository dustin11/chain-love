package vo

type PluginFileNode struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"` // "file" or "folder"
	Content  string            `json:"content,omitempty"`
	Children []*PluginFileNode `json:"children,omitempty"`
	IsOpen   bool              `json:"isOpen,omitempty"`
}
