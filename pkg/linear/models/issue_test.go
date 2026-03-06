package models

import (
	"encoding/json"
	"testing"
)

func TestIssueGetLabels(t *testing.T) {
	t.Run("nil labels", func(t *testing.T) {
		issue := &Issue{}
		if labels := issue.GetLabels(); labels != nil {
			t.Errorf("expected nil, got %v", labels)
		}
	})

	t.Run("empty labels", func(t *testing.T) {
		issue := &Issue{Labels: &LabelsNodes{Nodes: []*Label{}}}
		labels := issue.GetLabels()
		if labels == nil {
			t.Fatal("expected non-nil slice")
		}
		if len(labels) != 0 {
			t.Errorf("expected 0 labels, got %d", len(labels))
		}
	})

	t.Run("with labels", func(t *testing.T) {
		issue := &Issue{
			Labels: &LabelsNodes{
				Nodes: []*Label{
					{ID: "l1", Name: "Bug", Color: "#f00"},
					{ID: "l2", Name: "Feature", Color: "#0f0"},
				},
			},
		}
		labels := issue.GetLabels()
		if len(labels) != 2 {
			t.Fatalf("expected 2 labels, got %d", len(labels))
		}
		if labels[0].Name != "Bug" {
			t.Errorf("expected Bug, got %s", labels[0].Name)
		}
	})
}

func TestIssueGetComments(t *testing.T) {
	t.Run("nil comments", func(t *testing.T) {
		issue := &Issue{}
		if comments := issue.GetComments(); comments != nil {
			t.Errorf("expected nil, got %v", comments)
		}
	})

	t.Run("with comments", func(t *testing.T) {
		issue := &Issue{
			Comments: &CommentsNodes{
				Nodes: []*Comment{
					{ID: "c1", Body: "Hello"},
				},
			},
		}
		comments := issue.GetComments()
		if len(comments) != 1 {
			t.Fatalf("expected 1 comment, got %d", len(comments))
		}
		if comments[0].Body != "Hello" {
			t.Errorf("expected Hello, got %s", comments[0].Body)
		}
	})
}

func TestIssueGetChildren(t *testing.T) {
	t.Run("nil children", func(t *testing.T) {
		issue := &Issue{}
		if children := issue.GetChildren(); children != nil {
			t.Errorf("expected nil, got %v", children)
		}
	})

	t.Run("with children", func(t *testing.T) {
		issue := &Issue{
			Children: &ChildrenNodes{
				Nodes: []*Issue{
					{ID: "i1", Title: "Sub-task 1"},
					{ID: "i2", Title: "Sub-task 2"},
				},
			},
		}
		children := issue.GetChildren()
		if len(children) != 2 {
			t.Fatalf("expected 2 children, got %d", len(children))
		}
	})
}

func TestIssueGetRelations(t *testing.T) {
	t.Run("nil relations", func(t *testing.T) {
		issue := &Issue{}
		if relations := issue.GetRelations(); relations != nil {
			t.Errorf("expected nil, got %v", relations)
		}
	})

	t.Run("with relations", func(t *testing.T) {
		issue := &Issue{
			Relations: &RelationsNodes{
				Nodes: []*IssueRelation{
					{ID: "r1", Type: "blocks", RelatedIssue: &Issue{ID: "i2"}},
				},
			},
		}
		relations := issue.GetRelations()
		if len(relations) != 1 {
			t.Fatalf("expected 1 relation, got %d", len(relations))
		}
		if relations[0].Type != "blocks" {
			t.Errorf("expected blocks, got %s", relations[0].Type)
		}
	})
}

func TestIssueJSONUnmarshal(t *testing.T) {
	data := `{
		"id": "issue-1",
		"identifier": "ENG-123",
		"title": "Test issue",
		"priority": 2,
		"state": {"id": "s1", "name": "In Progress", "color": "#ff0", "type": "started"},
		"assignee": {"id": "u1", "name": "Alice", "displayName": "alice", "email": "alice@co.com"},
		"labels": {"nodes": [{"id": "l1", "name": "Bug", "color": "#f00"}]},
		"comments": {"nodes": [{"id": "c1", "body": "LGTM"}]}
	}`

	var issue Issue
	if err := json.Unmarshal([]byte(data), &issue); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if issue.Identifier != "ENG-123" {
		t.Errorf("expected ENG-123, got %s", issue.Identifier)
	}
	if issue.Priority != 2 {
		t.Errorf("expected priority 2, got %d", issue.Priority)
	}
	if issue.State == nil || issue.State.Name != "In Progress" {
		t.Error("state not unmarshaled correctly")
	}
	if issue.Assignee == nil || issue.Assignee.Name != "Alice" {
		t.Error("assignee not unmarshaled correctly")
	}
	if len(issue.GetLabels()) != 1 {
		t.Error("labels not unmarshaled correctly")
	}
	if len(issue.GetComments()) != 1 {
		t.Error("comments not unmarshaled correctly")
	}
}

func TestIssueFilterDefaults(t *testing.T) {
	filter := IssueFilter{}
	if filter.TeamID != "" {
		t.Error("expected empty TeamID")
	}
	if filter.Priority != nil {
		t.Error("expected nil Priority")
	}
	if filter.StateIDs != nil {
		t.Error("expected nil StateIDs")
	}
}
