package models

// Team represents a Linear team
type Team struct {
	ID     string           `json:"id"`
	Key    string           `json:"key"` // e.g., "ENG"
	Name   string           `json:"name"`
	States []*WorkflowState `json:"-"` // Loaded separately
	Labels []*Label         `json:"-"` // Loaded separately
}

// StatesResponse is used for JSON unmarshaling
type StatesResponse struct {
	Nodes []*WorkflowState `json:"nodes"`
}

// LabelsResponse is used for JSON unmarshaling
type LabelsResponse struct {
	Nodes []*Label `json:"nodes"`
}
