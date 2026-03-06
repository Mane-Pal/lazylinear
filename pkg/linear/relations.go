package linear

import (
	"context"
)

// RelationService handles issue relation operations
type RelationService struct {
	client *Client
}

// RelationType represents the type of relation between issues
type RelationType string

const (
	RelationBlocks    RelationType = "blocks"
	RelationBlockedBy RelationType = "blockedBy"
	RelationRelated   RelationType = "related"
	RelationDuplicate RelationType = "duplicate"
)

const createRelationMutation = `
mutation CreateIssueRelation($issueId: String!, $relatedIssueId: String!, $type: String!) {
    issueRelationCreate(input: {
        issueId: $issueId
        relatedIssueId: $relatedIssueId
        type: $type
    }) {
        success
        issueRelation {
            id
        }
    }
}
`

// Create creates a relation between two issues
func (s *RelationService) Create(ctx context.Context, issueID, relatedIssueID string, relType RelationType) error {
	var result struct {
		IssueRelationCreate struct {
			Success bool `json:"success"`
		} `json:"issueRelationCreate"`
	}

	variables := map[string]interface{}{
		"issueId":        issueID,
		"relatedIssueId": relatedIssueID,
		"type":           string(relType),
	}

	return s.client.do(ctx, createRelationMutation, variables, &result)
}

const deleteRelationMutation = `
mutation DeleteIssueRelation($id: String!) {
    issueRelationDelete(id: $id) {
        success
    }
}
`

// Delete deletes an issue relation
func (s *RelationService) Delete(ctx context.Context, relationID string) error {
	var result struct {
		IssueRelationDelete struct {
			Success bool `json:"success"`
		} `json:"issueRelationDelete"`
	}

	return s.client.do(ctx, deleteRelationMutation, map[string]interface{}{"id": relationID}, &result)
}
