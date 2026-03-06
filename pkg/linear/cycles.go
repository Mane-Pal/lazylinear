package linear

import (
	"context"

	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

// CycleService handles cycle operations
type CycleService struct {
	client *Client
}

const listCyclesQuery = `
query Cycles($teamId: String!) {
    team(id: $teamId) {
        cycles(orderBy: startsAt) {
            nodes {
                id
                name
                number
                startsAt
                endsAt
            }
        }
    }
}
`

// List returns all cycles for a team
func (s *CycleService) List(ctx context.Context, teamID string) ([]*models.Cycle, error) {
	var result struct {
		Team struct {
			Cycles struct {
				Nodes []*models.Cycle `json:"nodes"`
			} `json:"cycles"`
		} `json:"team"`
	}

	variables := map[string]interface{}{
		"teamId": teamID,
	}

	if err := s.client.do(ctx, listCyclesQuery, variables, &result); err != nil {
		return nil, err
	}

	return result.Team.Cycles.Nodes, nil
}

const getActiveCycleQuery = `
query ActiveCycle($teamId: String!) {
    team(id: $teamId) {
        activeCycle {
            id
            name
            number
            startsAt
            endsAt
        }
    }
}
`

// GetActive returns the active cycle for a team
func (s *CycleService) GetActive(ctx context.Context, teamID string) (*models.Cycle, error) {
	var result struct {
		Team struct {
			ActiveCycle *models.Cycle `json:"activeCycle"`
		} `json:"team"`
	}

	variables := map[string]interface{}{
		"teamId": teamID,
	}

	if err := s.client.do(ctx, getActiveCycleQuery, variables, &result); err != nil {
		return nil, err
	}

	return result.Team.ActiveCycle, nil
}
