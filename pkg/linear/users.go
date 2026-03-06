package linear

import (
	"context"

	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

// UserService handles user operations
type UserService struct {
	client *Client
}

const viewerQuery = `
query Viewer {
    viewer {
        id
        name
        displayName
        email
    }
}
`

// Viewer returns the authenticated user
func (s *UserService) Viewer(ctx context.Context) (*models.User, error) {
	var result struct {
		Viewer *models.User `json:"viewer"`
	}

	if err := s.client.do(ctx, viewerQuery, nil, &result); err != nil {
		return nil, err
	}

	return result.Viewer, nil
}

const teamMembersQuery = `
query TeamMembers($teamId: String!) {
    team(id: $teamId) {
        members {
            nodes {
                id
                name
                displayName
                email
            }
        }
    }
}
`

// TeamMembers returns members of a team
func (s *UserService) TeamMembers(ctx context.Context, teamID string) ([]*models.User, error) {
	var result struct {
		Team struct {
			Members struct {
				Nodes []*models.User `json:"nodes"`
			} `json:"members"`
		} `json:"team"`
	}

	if err := s.client.do(ctx, teamMembersQuery, map[string]interface{}{"teamId": teamID}, &result); err != nil {
		return nil, err
	}

	return result.Team.Members.Nodes, nil
}
