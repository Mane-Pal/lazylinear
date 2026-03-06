package linear

import (
	"context"

	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

// TeamService handles team operations
type TeamService struct {
	client *Client
}

const listTeamsQuery = `
query Teams {
    teams {
        nodes {
            id
            key
            name
        }
    }
}
`

// List returns all teams
func (s *TeamService) List(ctx context.Context) ([]*models.Team, error) {
	var result struct {
		Teams struct {
			Nodes []*models.Team `json:"nodes"`
		} `json:"teams"`
	}

	if err := s.client.do(ctx, listTeamsQuery, nil, &result); err != nil {
		return nil, err
	}

	return result.Teams.Nodes, nil
}

const getTeamWithStatesQuery = `
query Team($id: String!) {
    team(id: $id) {
        id
        key
        name
        states {
            nodes {
                id
                name
                color
                type
                position
            }
        }
        labels {
            nodes {
                id
                name
                color
            }
        }
    }
}
`

// teamWithStatesResponse matches the GraphQL response structure
type teamWithStatesResponse struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Name   string `json:"name"`
	States struct {
		Nodes []*models.WorkflowState `json:"nodes"`
	} `json:"states"`
	Labels struct {
		Nodes []*models.Label `json:"nodes"`
	} `json:"labels"`
}

// GetWithStates returns a team with its workflow states and labels
func (s *TeamService) GetWithStates(ctx context.Context, id string) (*models.Team, error) {
	var result struct {
		Team teamWithStatesResponse `json:"team"`
	}

	if err := s.client.do(ctx, getTeamWithStatesQuery, map[string]interface{}{"id": id}, &result); err != nil {
		return nil, err
	}

	team := &models.Team{
		ID:     result.Team.ID,
		Key:    result.Team.Key,
		Name:   result.Team.Name,
		States: result.Team.States.Nodes,
		Labels: result.Team.Labels.Nodes,
	}

	return team, nil
}
