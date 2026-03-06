package models

import "time"

// Issue represents a Linear issue
type Issue struct {
	ID          string    `json:"id"`
	Identifier  string    `json:"identifier"` // e.g., "ENG-123"
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"` // 0=none, 1=urgent, 2=high, 3=medium, 4=low
	Estimate    *float64  `json:"estimate"` // Story points
	DueDate     *string   `json:"dueDate"`  // ISO date string
	URL         string    `json:"url"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Relations (populated from nested JSON)
	State     *WorkflowState  `json:"state"`
	Assignee  *User           `json:"assignee"`
	Team      *Team           `json:"team"`
	Labels    *LabelsNodes    `json:"labels"`
	Project   *Project        `json:"project"`
	Cycle     *Cycle          `json:"cycle"`
	Parent    *Issue          `json:"parent"`
	Children  *ChildrenNodes  `json:"children"`
	Relations *RelationsNodes `json:"relations"`
	Comments  *CommentsNodes  `json:"comments"`
}

// LabelsNodes wraps the labels array for GraphQL response
type LabelsNodes struct {
	Nodes []*Label `json:"nodes"`
}

// CommentsNodes wraps the comments array for GraphQL response
type CommentsNodes struct {
	Nodes []*Comment `json:"nodes"`
}

// ChildrenNodes wraps the children issues array for GraphQL response
type ChildrenNodes struct {
	Nodes []*Issue `json:"nodes"`
}

// RelationsNodes wraps the issue relations array for GraphQL response
type RelationsNodes struct {
	Nodes []*IssueRelation `json:"nodes"`
}

// IssueRelation represents a relation between two issues
type IssueRelation struct {
	ID           string `json:"id"`
	Type         string `json:"type"` // "blocks", "blocked_by", "related", "duplicate"
	RelatedIssue *Issue `json:"relatedIssue"`
}

// GetLabels returns the labels slice
func (i *Issue) GetLabels() []*Label {
	if i.Labels == nil {
		return nil
	}
	return i.Labels.Nodes
}

// GetComments returns the comments slice
func (i *Issue) GetComments() []*Comment {
	if i.Comments == nil {
		return nil
	}
	return i.Comments.Nodes
}

// GetChildren returns the children issues slice
func (i *Issue) GetChildren() []*Issue {
	if i.Children == nil {
		return nil
	}
	return i.Children.Nodes
}

// GetRelations returns the issue relations slice
func (i *Issue) GetRelations() []*IssueRelation {
	if i.Relations == nil {
		return nil
	}
	return i.Relations.Nodes
}

// IssueFilter defines filters for listing issues
type IssueFilter struct {
	TeamID     string
	AssigneeID string
	StateIDs   []string
	LabelIDs   []string
	Priority   *int
}
