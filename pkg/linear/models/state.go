package models

// WorkflowState represents an issue state in Linear
type WorkflowState struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Color    string  `json:"color"`
	Type     string  `json:"type"` // backlog, unstarted, started, completed, canceled
	Position float64 `json:"position"`
}
