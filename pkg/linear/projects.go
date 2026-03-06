package linear

import (
	"context"

	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

// ProjectService handles project operations
type ProjectService struct {
	client *Client
}

const listProjectsQuery = `
query ListProjects($filter: ProjectFilter, $first: Int) {
    projects(filter: $filter, first: $first, orderBy: updatedAt) {
        nodes {
            id
            name
            description
            state
            progress
            startDate
            targetDate
            url
            createdAt
            updatedAt
            lead {
                id
                name
                displayName
            }
            issues {
                nodes {
                    id
                }
            }
        }
    }
}
`

// List returns projects matching the filter
func (s *ProjectService) List(ctx context.Context, filter models.ProjectFilter) ([]*models.Project, error) {
	variables := map[string]interface{}{
		"first": 50,
	}

	// Build filter
	filterMap := make(map[string]interface{})

	if filter.TeamID != "" {
		filterMap["accessibleTeams"] = map[string]interface{}{
			"id": map[string]interface{}{"eq": filter.TeamID},
		}
	}

	if len(filterMap) > 0 {
		variables["filter"] = filterMap
	}

	var result struct {
		Projects struct {
			Nodes []struct {
				models.Project
				Issues struct {
					Nodes []struct {
						ID string `json:"id"`
					} `json:"nodes"`
				} `json:"issues"`
			} `json:"nodes"`
		} `json:"projects"`
	}

	if err := s.client.do(ctx, listProjectsQuery, variables, &result); err != nil {
		return nil, err
	}

	// Convert to slice of pointers and compute issue count
	projects := make([]*models.Project, len(result.Projects.Nodes))
	for i, node := range result.Projects.Nodes {
		p := node.Project
		p.IssueCount = len(node.Issues.Nodes)
		projects[i] = &p
	}

	return projects, nil
}

const getProjectQuery = `
query GetProject($id: String!) {
    project(id: $id) {
        id
        name
        description
        state
        progress
        startDate
        targetDate
        url
        createdAt
        updatedAt
        lead {
            id
            name
            displayName
        }
    }
}
`

// Get returns a single project by ID
func (s *ProjectService) Get(ctx context.Context, id string) (*models.Project, error) {
	var result struct {
		Project *models.Project `json:"project"`
	}

	err := s.client.do(ctx, getProjectQuery, map[string]interface{}{"id": id}, &result)
	return result.Project, err
}
