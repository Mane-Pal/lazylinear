package linear

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mane-pal/lazylinear/pkg/auth"
)

const apiURL = "https://api.linear.app/graphql"

// Client is the Linear API client
type Client struct {
	httpClient *http.Client
	creds      *auth.Credentials

	// Sub-clients for different resources
	Issues    *IssueService
	Teams     *TeamService
	Users     *UserService
	Comments  *CommentService
	Projects  *ProjectService
	Cycles    *CycleService
	Relations *RelationService
}

// NewClient creates a new Linear API client
func NewClient(creds *auth.Credentials) (*Client, error) {
	if creds == nil {
		return nil, fmt.Errorf("credentials required")
	}

	c := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		creds:      creds,
	}

	// Initialize sub-clients
	c.Issues = &IssueService{client: c}
	c.Teams = &TeamService{client: c}
	c.Users = &UserService{client: c}
	c.Comments = &CommentService{client: c}
	c.Projects = &ProjectService{client: c}
	c.Cycles = &CycleService{client: c}
	c.Relations = &RelationService{client: c}

	return c, nil
}

type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphQLError  `json:"errors,omitempty"`
}

type graphQLError struct {
	Message string `json:"message"`
}

func (c *Client) do(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	// Auto-refresh expired OAuth tokens
	if err := c.creds.Refresh(); err != nil {
		return fmt.Errorf("refresh credentials: %w", err)
	}

	reqBody := graphQLRequest{
		Query:     query,
		Variables: variables,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", c.creds.HeaderValue)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "lazylinear")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var gqlResp graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return fmt.Errorf("GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	if result != nil {
		if err := json.Unmarshal(gqlResp.Data, result); err != nil {
			return fmt.Errorf("unmarshal data: %w", err)
		}
	}

	return nil
}
