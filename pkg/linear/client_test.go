package linear

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mane-pal/lazylinear/pkg/auth"
	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

// newTestClient creates a Client pointed at a test HTTP server.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	c := &Client{
		httpClient: server.Client(),
		creds:      &auth.Credentials{HeaderValue: "test-key"},
	}
	c.Issues = &IssueService{client: c}
	c.Teams = &TeamService{client: c}
	c.Users = &UserService{client: c}
	c.Comments = &CommentService{client: c}
	c.Projects = &ProjectService{client: c}
	c.Cycles = &CycleService{client: c}
	c.Relations = &RelationService{client: c}
	// Override the API URL by patching the do method isn't possible directly,
	// so we use a custom transport approach instead.
	return c, server
}

// doWithURL is a helper that lets tests call the client's do method against a custom URL.
func doWithURL(c *Client, url string, ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	// Temporarily override the constant by using the do method's internals
	// Since apiURL is a const, we test via the httptest server approach.
	// We'll use a custom round-tripper instead.
	return c.do(ctx, query, variables, result)
}

func TestNewClient(t *testing.T) {
	t.Run("requires credentials", func(t *testing.T) {
		_, err := NewClient(nil)
		if err == nil {
			t.Fatal("expected error for nil credentials")
		}
	})

	t.Run("creates client with valid credentials", func(t *testing.T) {
		creds := &auth.Credentials{HeaderValue: "lin_api_test123"}
		c, err := NewClient(creds)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Issues == nil || c.Teams == nil || c.Users == nil {
			t.Fatal("sub-clients not initialized")
		}
		if c.Comments == nil || c.Projects == nil || c.Cycles == nil || c.Relations == nil {
			t.Fatal("sub-clients not initialized")
		}
	})
}

// graphQLHandler creates an HTTP handler that validates the request and returns canned responses.
func graphQLHandler(t *testing.T, wantAuthHeader string, response interface{}) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		// Verify method
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Verify headers
		if r.Header.Get("Authorization") != wantAuthHeader {
			t.Errorf("expected Authorization %q, got %q", wantAuthHeader, r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("User-Agent") != "lazylinear" {
			t.Errorf("expected User-Agent lazylinear, got %q", r.Header.Get("User-Agent"))
		}

		// Decode request body
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if req.Query == "" {
			t.Error("empty query")
		}

		// Return canned response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// transportFunc allows using a function as an http.RoundTripper.
type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// newTestClientWithServer creates a client that routes requests to a test server.
func newTestClientWithServer(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	c := &Client{
		httpClient: &http.Client{
			Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				// Rewrite the URL to point at the test server
				r.URL.Scheme = "http"
				r.URL.Host = server.Listener.Addr().String()
				return http.DefaultTransport.RoundTrip(r)
			}),
		},
		creds: &auth.Credentials{HeaderValue: "test-key"},
	}
	c.Issues = &IssueService{client: c}
	c.Teams = &TeamService{client: c}
	c.Users = &UserService{client: c}
	c.Comments = &CommentService{client: c}
	c.Projects = &ProjectService{client: c}
	c.Cycles = &CycleService{client: c}
	c.Relations = &RelationService{client: c}
	return c
}

func TestTeamsList(t *testing.T) {
	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"teams": map[string]interface{}{
				"nodes": []map[string]interface{}{
					{"id": "team-1", "key": "ENG", "name": "Engineering"},
					{"id": "team-2", "key": "DES", "name": "Design"},
				},
			},
		},
	}

	server := httptest.NewServer(graphQLHandler(t, "test-key", resp))
	defer server.Close()

	client := newTestClientWithServer(t, server)
	teams, err := client.Teams.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(teams) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(teams))
	}
	if teams[0].Key != "ENG" {
		t.Errorf("expected key ENG, got %s", teams[0].Key)
	}
	if teams[1].Name != "Design" {
		t.Errorf("expected name Design, got %s", teams[1].Name)
	}
}

func TestIssuesList(t *testing.T) {
	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"issues": map[string]interface{}{
				"pageInfo": map[string]interface{}{
					"hasNextPage": false,
					"endCursor":   nil,
				},
				"nodes": []map[string]interface{}{
					{
						"id":         "issue-1",
						"identifier": "ENG-1",
						"title":      "Fix login bug",
						"priority":   2,
						"state":      map[string]interface{}{"id": "state-1", "name": "In Progress", "color": "#f00", "type": "started"},
					},
				},
			},
		},
	}

	server := httptest.NewServer(graphQLHandler(t, "test-key", resp))
	defer server.Close()

	client := newTestClientWithServer(t, server)
	issues, err := client.Issues.List(context.Background(), models.IssueFilter{TeamID: "team-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Identifier != "ENG-1" {
		t.Errorf("expected identifier ENG-1, got %s", issues[0].Identifier)
	}
	if issues[0].Title != "Fix login bug" {
		t.Errorf("expected title 'Fix login bug', got %s", issues[0].Title)
	}
}

func TestIssuesListWithFilters(t *testing.T) {
	var capturedReq graphQLRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedReq)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"issues": map[string]interface{}{
					"pageInfo": map[string]interface{}{"hasNextPage": false},
					"nodes":    []interface{}{},
				},
			},
		})
	}))
	defer server.Close()

	client := newTestClientWithServer(t, server)

	priority := 1
	_, err := client.Issues.List(context.Background(), models.IssueFilter{
		TeamID:     "team-1",
		AssigneeID: "user-1",
		StateIDs:   []string{"state-1", "state-2"},
		Priority:   &priority,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify filter was included in variables
	filter, ok := capturedReq.Variables["filter"].(map[string]interface{})
	if !ok {
		t.Fatal("expected filter in variables")
	}
	if filter["team"] == nil {
		t.Error("expected team filter")
	}
	if filter["assignee"] == nil {
		t.Error("expected assignee filter")
	}
	if filter["state"] == nil {
		t.Error("expected state filter")
	}
	if filter["priority"] == nil {
		t.Error("expected priority filter")
	}
}

func TestIssueGet(t *testing.T) {
	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"issue": map[string]interface{}{
				"id":         "issue-1",
				"identifier": "ENG-42",
				"title":      "Implement dark mode",
				"priority":   3,
				"state": map[string]interface{}{
					"id": "state-1", "name": "Todo", "color": "#ccc", "type": "unstarted",
				},
				"comments": map[string]interface{}{
					"nodes": []map[string]interface{}{
						{"id": "comment-1", "body": "Looks good"},
					},
				},
			},
		},
	}

	server := httptest.NewServer(graphQLHandler(t, "test-key", resp))
	defer server.Close()

	client := newTestClientWithServer(t, server)
	issue, err := client.Issues.Get(context.Background(), "issue-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue.Identifier != "ENG-42" {
		t.Errorf("expected ENG-42, got %s", issue.Identifier)
	}
	comments := issue.GetComments()
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}
	if comments[0].Body != "Looks good" {
		t.Errorf("expected 'Looks good', got %s", comments[0].Body)
	}
}

func TestIssueCreate(t *testing.T) {
	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"issueCreate": map[string]interface{}{
				"success": true,
				"issue": map[string]interface{}{
					"id":         "issue-new",
					"identifier": "ENG-99",
					"title":      "New issue",
					"url":        "https://linear.app/eng/issue/ENG-99",
				},
			},
		},
	}

	server := httptest.NewServer(graphQLHandler(t, "test-key", resp))
	defer server.Close()

	client := newTestClientWithServer(t, server)
	assignee := "user-1"
	issue, err := client.Issues.Create(context.Background(), IssueCreateInput{
		TeamID:      "team-1",
		Title:       "New issue",
		Description: "Description",
		Priority:    2,
		AssigneeID:  &assignee,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue.Identifier != "ENG-99" {
		t.Errorf("expected ENG-99, got %s", issue.Identifier)
	}
}

func TestViewerQuery(t *testing.T) {
	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"viewer": map[string]interface{}{
				"id":          "user-1",
				"name":        "John Doe",
				"displayName": "John",
				"email":       "john@example.com",
			},
		},
	}

	server := httptest.NewServer(graphQLHandler(t, "test-key", resp))
	defer server.Close()

	client := newTestClientWithServer(t, server)
	user, err := client.Users.Viewer(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Name != "John Doe" {
		t.Errorf("expected John Doe, got %s", user.Name)
	}
	if user.Email != "john@example.com" {
		t.Errorf("expected john@example.com, got %s", user.Email)
	}
}

func TestTeamMembers(t *testing.T) {
	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"team": map[string]interface{}{
				"members": map[string]interface{}{
					"nodes": []map[string]interface{}{
						{"id": "user-1", "name": "Alice", "displayName": "alice", "email": "alice@co.com"},
						{"id": "user-2", "name": "Bob", "displayName": "bob", "email": "bob@co.com"},
					},
				},
			},
		},
	}

	server := httptest.NewServer(graphQLHandler(t, "test-key", resp))
	defer server.Close()

	client := newTestClientWithServer(t, server)
	members, err := client.Users.TeamMembers(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
}

func TestCommentCreate(t *testing.T) {
	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"commentCreate": map[string]interface{}{
				"success": true,
				"comment": map[string]interface{}{
					"id":   "comment-new",
					"body": "Nice work!",
					"user": map[string]interface{}{"id": "user-1", "name": "Alice"},
				},
			},
		},
	}

	server := httptest.NewServer(graphQLHandler(t, "test-key", resp))
	defer server.Close()

	client := newTestClientWithServer(t, server)
	comment, err := client.Comments.Create(context.Background(), "issue-1", "Nice work!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comment.Body != "Nice work!" {
		t.Errorf("expected 'Nice work!', got %s", comment.Body)
	}
}

func TestCommentUpdateAndDelete(t *testing.T) {
	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"commentUpdate": map[string]interface{}{"success": true},
			"commentDelete": map[string]interface{}{"success": true},
		},
	}

	server := httptest.NewServer(graphQLHandler(t, "test-key", resp))
	defer server.Close()

	client := newTestClientWithServer(t, server)

	if err := client.Comments.Update(context.Background(), "comment-1", "Updated body"); err != nil {
		t.Fatalf("update error: %v", err)
	}
	if err := client.Comments.Delete(context.Background(), "comment-1"); err != nil {
		t.Fatalf("delete error: %v", err)
	}
}

func TestProjectsList(t *testing.T) {
	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"projects": map[string]interface{}{
				"nodes": []map[string]interface{}{
					{
						"id":       "proj-1",
						"name":     "Launch v2",
						"state":    "started",
						"progress": 0.45,
						"issues": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{"id": "i1"},
								{"id": "i2"},
								{"id": "i3"},
							},
						},
					},
				},
			},
		},
	}

	server := httptest.NewServer(graphQLHandler(t, "test-key", resp))
	defer server.Close()

	client := newTestClientWithServer(t, server)
	projects, err := client.Projects.List(context.Background(), models.ProjectFilter{TeamID: "team-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Name != "Launch v2" {
		t.Errorf("expected 'Launch v2', got %s", projects[0].Name)
	}
	if projects[0].IssueCount != 3 {
		t.Errorf("expected issue count 3, got %d", projects[0].IssueCount)
	}
}

func TestCyclesList(t *testing.T) {
	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"team": map[string]interface{}{
				"cycles": map[string]interface{}{
					"nodes": []map[string]interface{}{
						{"id": "cycle-1", "name": "Sprint 1", "number": 1},
						{"id": "cycle-2", "name": "Sprint 2", "number": 2},
					},
				},
			},
		},
	}

	server := httptest.NewServer(graphQLHandler(t, "test-key", resp))
	defer server.Close()

	client := newTestClientWithServer(t, server)
	cycles, err := client.Cycles.List(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 2 {
		t.Fatalf("expected 2 cycles, got %d", len(cycles))
	}
}

func TestGetActiveCycle(t *testing.T) {
	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"team": map[string]interface{}{
				"activeCycle": map[string]interface{}{
					"id": "cycle-1", "name": "Sprint 1", "number": 1,
				},
			},
		},
	}

	server := httptest.NewServer(graphQLHandler(t, "test-key", resp))
	defer server.Close()

	client := newTestClientWithServer(t, server)
	cycle, err := client.Cycles.GetActive(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cycle.Name != "Sprint 1" {
		t.Errorf("expected 'Sprint 1', got %s", cycle.Name)
	}
}

func TestRelationCreateAndDelete(t *testing.T) {
	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"issueRelationCreate": map[string]interface{}{
				"success":       true,
				"issueRelation": map[string]interface{}{"id": "rel-1"},
			},
			"issueRelationDelete": map[string]interface{}{"success": true},
		},
	}

	server := httptest.NewServer(graphQLHandler(t, "test-key", resp))
	defer server.Close()

	client := newTestClientWithServer(t, server)

	if err := client.Relations.Create(context.Background(), "issue-1", "issue-2", RelationBlocks); err != nil {
		t.Fatalf("create relation error: %v", err)
	}
	if err := client.Relations.Delete(context.Background(), "rel-1"); err != nil {
		t.Fatalf("delete relation error: %v", err)
	}
}

func TestGraphQLErrorHandling(t *testing.T) {
	resp := map[string]interface{}{
		"errors": []map[string]interface{}{
			{"message": "Entity not found"},
		},
	}

	server := httptest.NewServer(graphQLHandler(t, "test-key", resp))
	defer server.Close()

	client := newTestClientWithServer(t, server)
	_, err := client.Teams.List(context.Background())
	if err == nil {
		t.Fatal("expected error for GraphQL error response")
	}
	if err.Error() != "GraphQL error: Entity not found" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestHTTPErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := newTestClientWithServer(t, server)
	_, err := client.Teams.List(context.Background())
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}

func TestTeamGetWithStates(t *testing.T) {
	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"team": map[string]interface{}{
				"id":   "team-1",
				"key":  "ENG",
				"name": "Engineering",
				"states": map[string]interface{}{
					"nodes": []map[string]interface{}{
						{"id": "s1", "name": "Backlog", "color": "#ccc", "type": "backlog", "position": 0},
						{"id": "s2", "name": "In Progress", "color": "#0f0", "type": "started", "position": 1},
						{"id": "s3", "name": "Done", "color": "#00f", "type": "completed", "position": 2},
					},
				},
				"labels": map[string]interface{}{
					"nodes": []map[string]interface{}{
						{"id": "l1", "name": "Bug", "color": "#f00"},
						{"id": "l2", "name": "Feature", "color": "#0f0"},
					},
				},
			},
		},
	}

	server := httptest.NewServer(graphQLHandler(t, "test-key", resp))
	defer server.Close()

	client := newTestClientWithServer(t, server)
	team, err := client.Teams.GetWithStates(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if team.Name != "Engineering" {
		t.Errorf("expected Engineering, got %s", team.Name)
	}
	if len(team.States) != 3 {
		t.Fatalf("expected 3 states, got %d", len(team.States))
	}
	if len(team.Labels) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(team.Labels))
	}
}

func TestIssueUpdateOperations(t *testing.T) {
	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"issueUpdate": map[string]interface{}{
				"success": true,
				"issue":   map[string]interface{}{"id": "issue-1", "identifier": "ENG-1"},
			},
		},
	}

	server := httptest.NewServer(graphQLHandler(t, "test-key", resp))
	defer server.Close()

	client := newTestClientWithServer(t, server)
	ctx := context.Background()

	t.Run("update state", func(t *testing.T) {
		if err := client.Issues.UpdateState(ctx, "issue-1", "state-2"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("update assignee", func(t *testing.T) {
		assignee := "user-1"
		if err := client.Issues.UpdateAssignee(ctx, "issue-1", &assignee); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("unassign", func(t *testing.T) {
		if err := client.Issues.UpdateAssignee(ctx, "issue-1", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("update priority", func(t *testing.T) {
		if err := client.Issues.UpdatePriority(ctx, "issue-1", 1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("update title and description", func(t *testing.T) {
		title := "New title"
		desc := "New desc"
		if err := client.Issues.Update(ctx, "issue-1", &title, &desc); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("update labels", func(t *testing.T) {
		if err := client.Issues.UpdateLabels(ctx, "issue-1", []string{"l1", "l2"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("update due date", func(t *testing.T) {
		date := "2025-12-31"
		if err := client.Issues.UpdateDueDate(ctx, "issue-1", &date); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("clear due date", func(t *testing.T) {
		if err := client.Issues.UpdateDueDate(ctx, "issue-1", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("update estimate", func(t *testing.T) {
		est := 5.0
		if err := client.Issues.UpdateEstimate(ctx, "issue-1", &est); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("update cycle", func(t *testing.T) {
		cycleID := "cycle-1"
		if err := client.Issues.UpdateCycle(ctx, "issue-1", &cycleID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("update parent", func(t *testing.T) {
		parentID := "issue-parent"
		if err := client.Issues.UpdateParent(ctx, "issue-1", &parentID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("archive", func(t *testing.T) {
		archiveResp := map[string]interface{}{
			"data": map[string]interface{}{
				"issueArchive": map[string]interface{}{"success": true},
			},
		}
		archiveServer := httptest.NewServer(graphQLHandler(t, "test-key", archiveResp))
		defer archiveServer.Close()
		archiveClient := newTestClientWithServer(t, archiveServer)

		if err := archiveClient.Issues.Archive(ctx, "issue-1"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
