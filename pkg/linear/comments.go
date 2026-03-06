package linear

import (
	"context"

	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

// CommentService handles comment operations
type CommentService struct {
	client *Client
}

const createCommentMutation = `
mutation CreateComment($input: CommentCreateInput!) {
    commentCreate(input: $input) {
        success
        comment {
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
`

// Create adds a comment to an issue
func (s *CommentService) Create(ctx context.Context, issueID, body string) (*models.Comment, error) {
	var result struct {
		CommentCreate struct {
			Success bool            `json:"success"`
			Comment *models.Comment `json:"comment"`
		} `json:"commentCreate"`
	}

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"issueId": issueID,
			"body":    body,
		},
	}

	if err := s.client.do(ctx, createCommentMutation, variables, &result); err != nil {
		return nil, err
	}

	return result.CommentCreate.Comment, nil
}

const updateCommentMutation = `
mutation UpdateComment($id: String!, $input: CommentUpdateInput!) {
    commentUpdate(id: $id, input: $input) {
        success
    }
}
`

// Update modifies a comment's body
func (s *CommentService) Update(ctx context.Context, commentID, body string) error {
	var result struct {
		CommentUpdate struct {
			Success bool `json:"success"`
		} `json:"commentUpdate"`
	}

	variables := map[string]interface{}{
		"id":    commentID,
		"input": map[string]interface{}{"body": body},
	}

	return s.client.do(ctx, updateCommentMutation, variables, &result)
}

const deleteCommentMutation = `
mutation DeleteComment($id: String!) {
    commentDelete(id: $id) {
        success
    }
}
`

// Delete removes a comment
func (s *CommentService) Delete(ctx context.Context, commentID string) error {
	var result struct {
		CommentDelete struct {
			Success bool `json:"success"`
		} `json:"commentDelete"`
	}

	return s.client.do(ctx, deleteCommentMutation, map[string]interface{}{"id": commentID}, &result)
}
