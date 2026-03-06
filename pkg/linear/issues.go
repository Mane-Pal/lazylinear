package linear

import (
	"context"

	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

// IssueService handles issue operations
type IssueService struct {
	client *Client
}

const listIssuesQuery = `
query ListIssues($filter: IssueFilter, $first: Int, $after: String) {
    issues(filter: $filter, first: $first, after: $after, orderBy: updatedAt) {
        pageInfo {
            hasNextPage
            endCursor
        }
        nodes {
            id
            identifier
            title
            description
            priority
            estimate
            dueDate
            url
            createdAt
            updatedAt
            state {
                id
                name
                color
                type
            }
            assignee {
                id
                name
                displayName
                email
            }
            team {
                id
                key
                name
            }
            labels {
                nodes {
                    id
                    name
                    color
                }
            }
            cycle {
                id
                name
                number
            }
            parent {
                id
                identifier
                title
            }
        }
    }
}
`

// List returns issues matching the filter
func (s *IssueService) List(ctx context.Context, filter models.IssueFilter) ([]*models.Issue, error) {
	variables := map[string]interface{}{
		"first": 100, // Fetch up to 100 issues per request
	}

	// Build filter
	filterMap := make(map[string]interface{})

	if filter.TeamID != "" {
		filterMap["team"] = map[string]interface{}{
			"id": map[string]interface{}{"eq": filter.TeamID},
		}
	}

	if filter.AssigneeID != "" {
		filterMap["assignee"] = map[string]interface{}{
			"id": map[string]interface{}{"eq": filter.AssigneeID},
		}
	}

	// State filter (multiple states using "in" operator)
	if len(filter.StateIDs) > 0 {
		filterMap["state"] = map[string]interface{}{
			"id": map[string]interface{}{"in": filter.StateIDs},
		}
	}

	// Priority filter
	if filter.Priority != nil {
		filterMap["priority"] = map[string]interface{}{"eq": *filter.Priority}
	}

	if len(filterMap) > 0 {
		variables["filter"] = filterMap
	}

	var result struct {
		Issues struct {
			PageInfo struct {
				HasNextPage bool    `json:"hasNextPage"`
				EndCursor   *string `json:"endCursor"`
			} `json:"pageInfo"`
			Nodes []*models.Issue `json:"nodes"`
		} `json:"issues"`
	}

	if err := s.client.do(ctx, listIssuesQuery, variables, &result); err != nil {
		return nil, err
	}

	return result.Issues.Nodes, nil
}

const getIssueQuery = `
query GetIssue($id: String!) {
    issue(id: $id) {
        id
        identifier
        title
        description
        priority
        estimate
        dueDate
        url
        createdAt
        updatedAt
        state {
            id
            name
            color
            type
        }
        assignee {
            id
            name
            displayName
            email
        }
        team {
            id
            key
            name
        }
        labels {
            nodes {
                id
                name
                color
            }
        }
        cycle {
            id
            name
            number
            startsAt
            endsAt
        }
        parent {
            id
            identifier
            title
        }
        children {
            nodes {
                id
                identifier
                title
                state {
                    id
                    name
                    type
                }
            }
        }
        relations {
            nodes {
                id
                type
                relatedIssue {
                    id
                    identifier
                    title
                }
            }
        }
        comments {
            nodes {
                id
                body
                createdAt
                user {
                    id
                    name
                    displayName
                }
            }
        }
    }
}
`

// Get returns a single issue by ID
func (s *IssueService) Get(ctx context.Context, id string) (*models.Issue, error) {
	var result struct {
		Issue *models.Issue `json:"issue"`
	}

	err := s.client.do(ctx, getIssueQuery, map[string]interface{}{"id": id}, &result)
	return result.Issue, err
}

const updateIssueMutation = `
mutation UpdateIssue($id: String!, $input: IssueUpdateInput!) {
    issueUpdate(id: $id, input: $input) {
        success
        issue {
            id
            identifier
            state {
                id
                name
            }
        }
    }
}
`

// UpdateState changes an issue's workflow state
func (s *IssueService) UpdateState(ctx context.Context, issueID, stateID string) error {
	var result struct {
		IssueUpdate struct {
			Success bool `json:"success"`
		} `json:"issueUpdate"`
	}

	variables := map[string]interface{}{
		"id":    issueID,
		"input": map[string]interface{}{"stateId": stateID},
	}

	return s.client.do(ctx, updateIssueMutation, variables, &result)
}

// UpdateAssignee changes an issue's assignee
func (s *IssueService) UpdateAssignee(ctx context.Context, issueID string, assigneeID *string) error {
	var result struct {
		IssueUpdate struct {
			Success bool `json:"success"`
		} `json:"issueUpdate"`
	}

	input := map[string]interface{}{}
	if assigneeID != nil {
		input["assigneeId"] = *assigneeID
	} else {
		input["assigneeId"] = nil
	}

	variables := map[string]interface{}{
		"id":    issueID,
		"input": input,
	}

	return s.client.do(ctx, updateIssueMutation, variables, &result)
}

// UpdatePriority changes an issue's priority
func (s *IssueService) UpdatePriority(ctx context.Context, issueID string, priority int) error {
	var result struct {
		IssueUpdate struct {
			Success bool `json:"success"`
		} `json:"issueUpdate"`
	}

	variables := map[string]interface{}{
		"id":    issueID,
		"input": map[string]interface{}{"priority": priority},
	}

	return s.client.do(ctx, updateIssueMutation, variables, &result)
}

// Update updates an issue's title and/or description
func (s *IssueService) Update(ctx context.Context, issueID string, title, description *string) error {
	var result struct {
		IssueUpdate struct {
			Success bool `json:"success"`
		} `json:"issueUpdate"`
	}

	input := map[string]interface{}{}
	if title != nil {
		input["title"] = *title
	}
	if description != nil {
		input["description"] = *description
	}

	variables := map[string]interface{}{
		"id":    issueID,
		"input": input,
	}

	return s.client.do(ctx, updateIssueMutation, variables, &result)
}

const createIssueMutation = `
mutation CreateIssue($input: IssueCreateInput!) {
    issueCreate(input: $input) {
        success
        issue {
            id
            identifier
            title
            url
        }
    }
}
`

// IssueCreateInput holds the input for creating an issue
// IssueCreateInput holds the input for creating an issue
type IssueCreateInput struct {
	TeamID      string
	Title       string
	Description string
	Priority    int
	AssigneeID  *string // Optional
	StateID     *string // Optional
	ProjectID   *string // Optional
}

// Create creates a new issue
func (s *IssueService) Create(ctx context.Context, input IssueCreateInput) (*models.Issue, error) {
	var result struct {
		IssueCreate struct {
			Success bool          `json:"success"`
			Issue   *models.Issue `json:"issue"`
		} `json:"issueCreate"`
	}

	inputMap := map[string]interface{}{
		"teamId":      input.TeamID,
		"title":       input.Title,
		"description": input.Description,
		"priority":    input.Priority,
	}

	// Add optional fields
	if input.AssigneeID != nil {
		inputMap["assigneeId"] = *input.AssigneeID
	}
	if input.StateID != nil {
		inputMap["stateId"] = *input.StateID
	}
	if input.ProjectID != nil {
		inputMap["projectId"] = *input.ProjectID
	}

	variables := map[string]interface{}{
		"input": inputMap,
	}

	if err := s.client.do(ctx, createIssueMutation, variables, &result); err != nil {
		return nil, err
	}

	return result.IssueCreate.Issue, nil
}

const archiveIssueMutation = `
mutation ArchiveIssue($id: String!) {
    issueArchive(id: $id) {
        success
    }
}
`

// Archive archives an issue
func (s *IssueService) Archive(ctx context.Context, issueID string) error {
	var result struct {
		IssueArchive struct {
			Success bool `json:"success"`
		} `json:"issueArchive"`
	}

	return s.client.do(ctx, archiveIssueMutation, map[string]interface{}{"id": issueID}, &result)
}

// UpdateLabels updates an issue's labels
func (s *IssueService) UpdateLabels(ctx context.Context, issueID string, labelIDs []string) error {
	var result struct {
		IssueUpdate struct {
			Success bool `json:"success"`
		} `json:"issueUpdate"`
	}

	variables := map[string]interface{}{
		"id":    issueID,
		"input": map[string]interface{}{"labelIds": labelIDs},
	}

	return s.client.do(ctx, updateIssueMutation, variables, &result)
}

// UpdateDueDate updates an issue's due date (ISO date string or nil to clear)
func (s *IssueService) UpdateDueDate(ctx context.Context, issueID string, dueDate *string) error {
	var result struct {
		IssueUpdate struct {
			Success bool `json:"success"`
		} `json:"issueUpdate"`
	}

	input := map[string]interface{}{}
	if dueDate != nil {
		input["dueDate"] = *dueDate
	} else {
		input["dueDate"] = nil
	}

	variables := map[string]interface{}{
		"id":    issueID,
		"input": input,
	}

	return s.client.do(ctx, updateIssueMutation, variables, &result)
}

// UpdateEstimate updates an issue's estimate (story points)
func (s *IssueService) UpdateEstimate(ctx context.Context, issueID string, estimate *float64) error {
	var result struct {
		IssueUpdate struct {
			Success bool `json:"success"`
		} `json:"issueUpdate"`
	}

	input := map[string]interface{}{}
	if estimate != nil {
		input["estimate"] = *estimate
	} else {
		input["estimate"] = nil
	}

	variables := map[string]interface{}{
		"id":    issueID,
		"input": input,
	}

	return s.client.do(ctx, updateIssueMutation, variables, &result)
}

// UpdateCycle updates an issue's cycle
func (s *IssueService) UpdateCycle(ctx context.Context, issueID string, cycleID *string) error {
	var result struct {
		IssueUpdate struct {
			Success bool `json:"success"`
		} `json:"issueUpdate"`
	}

	input := map[string]interface{}{}
	if cycleID != nil {
		input["cycleId"] = *cycleID
	} else {
		input["cycleId"] = nil
	}

	variables := map[string]interface{}{
		"id":    issueID,
		"input": input,
	}

	return s.client.do(ctx, updateIssueMutation, variables, &result)
}

// UpdateParent updates an issue's parent
func (s *IssueService) UpdateParent(ctx context.Context, issueID string, parentID *string) error {
	var result struct {
		IssueUpdate struct {
			Success bool `json:"success"`
		} `json:"issueUpdate"`
	}

	input := map[string]interface{}{}
	if parentID != nil {
		input["parentId"] = *parentID
	} else {
		input["parentId"] = nil
	}

	variables := map[string]interface{}{
		"id":    issueID,
		"input": input,
	}

	return s.client.do(ctx, updateIssueMutation, variables, &result)
}
