package helpers

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mane-pal/lazylinear/pkg/gui/types"
	"github.com/mane-pal/lazylinear/pkg/linear"
	"github.com/mane-pal/lazylinear/pkg/linear/models"
	"github.com/mane-pal/lazylinear/pkg/utils"
)

// IssuesHelper provides issue CRUD and mutation operations.
type IssuesHelper struct {
	Client *linear.Client
}

func NewIssuesHelper(client *linear.Client) *IssuesHelper {
	return &IssuesHelper{Client: client}
}

func (h *IssuesHelper) UpdateIssueState(issueID, stateID string) tea.Cmd {
	return func() tea.Msg {
		if err := h.Client.Issues.UpdateState(context.Background(), issueID, stateID); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update state: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

func (h *IssuesHelper) UpdateIssueAssignee(issueID, assigneeID string) tea.Cmd {
	return func() tea.Msg {
		var assignee *string
		if assigneeID != "" {
			assignee = &assigneeID
		}
		if err := h.Client.Issues.UpdateAssignee(context.Background(), issueID, assignee); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update assignee: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

func (h *IssuesHelper) UpdateIssuePriority(issueID, priorityStr string) tea.Cmd {
	return func() tea.Msg {
		priority := 0
		fmt.Sscanf(priorityStr, "%d", &priority)
		if err := h.Client.Issues.UpdatePriority(context.Background(), issueID, priority); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update priority: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

func (h *IssuesHelper) UpdateIssueLabels(issueID string, labelIDs []string) tea.Cmd {
	return func() tea.Msg {
		if err := h.Client.Issues.UpdateLabels(context.Background(), issueID, labelIDs); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update labels: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

func (h *IssuesHelper) UpdateIssueCycle(issueID string, cycleID *string) tea.Cmd {
	return func() tea.Msg {
		if err := h.Client.Issues.UpdateCycle(context.Background(), issueID, cycleID); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update cycle: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

func (h *IssuesHelper) UpdateIssueDueDate(issueID string, dueDate *string) tea.Cmd {
	return func() tea.Msg {
		if err := h.Client.Issues.UpdateDueDate(context.Background(), issueID, dueDate); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update due date: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

func (h *IssuesHelper) UpdateIssueEstimate(issueID string, estimate *float64) tea.Cmd {
	return func() tea.Msg {
		if err := h.Client.Issues.UpdateEstimate(context.Background(), issueID, estimate); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update estimate: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

func (h *IssuesHelper) UpdateIssueParent(issueID string, parentID *string) tea.Cmd {
	return func() tea.Msg {
		if err := h.Client.Issues.UpdateParent(context.Background(), issueID, parentID); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update parent: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

func (h *IssuesHelper) ArchiveIssue(issueID string) tea.Cmd {
	return func() tea.Msg {
		if err := h.Client.Issues.Archive(context.Background(), issueID); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to archive: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

func (h *IssuesHelper) CreateIssue(input linear.IssueCreateInput) tea.Cmd {
	return func() tea.Msg {
		issue, err := h.Client.Issues.Create(context.Background(), input)
		if err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to create issue: %v", err))
		}
		return types.StatusMsg(fmt.Sprintf("Created %s", issue.Identifier))
	}
}

func (h *IssuesHelper) EditIssue(issueID, title, description string) tea.Cmd {
	return func() tea.Msg {
		if err := h.Client.Issues.Update(context.Background(), issueID, &title, &description); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update issue: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

func (h *IssuesHelper) CreateComment(issueID, body string) tea.Cmd {
	return func() tea.Msg {
		comment, err := h.Client.Comments.Create(context.Background(), issueID, body)
		if err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to create comment: %v", err))
		}
		return types.CommentCreatedMsg{Comment: comment}
	}
}

func (h *IssuesHelper) UpdateComment(commentID, body string) tea.Cmd {
	return func() tea.Msg {
		if err := h.Client.Comments.Update(context.Background(), commentID, body); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update comment: %v", err))
		}
		return types.CommentCreatedMsg{}
	}
}

func (h *IssuesHelper) DeleteComment(commentID, issueID string) tea.Cmd {
	return func() tea.Msg {
		if err := h.Client.Comments.Delete(context.Background(), commentID); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to delete comment: %v", err))
		}
		return types.CommentDeletedMsg{IssueID: issueID}
	}
}

func (h *IssuesHelper) OpenInBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		if err := utils.OpenBrowser(url); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to open browser: %v", err))
		}
		return types.StatusMsg("Opened in browser")
	}
}

func (h *IssuesHelper) CopyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		if err := utils.CopyToClipboard(text); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to copy: %v", err))
		}
		return types.StatusMsg(fmt.Sprintf("Copied: %s", text))
	}
}

func (h *IssuesHelper) OpenParentPicker(issue *models.Issue) tea.Cmd {
	if issue.Team == nil {
		return func() tea.Msg {
			return types.StatusMsg("No team for issue")
		}
	}

	issueCopy := issue
	return func() tea.Msg {
		filter := models.IssueFilter{TeamID: issueCopy.Team.ID}
		issues, err := h.Client.Issues.List(context.Background(), filter)
		if err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to load issues: %v", err))
		}
		return types.ParentPickerMsg{Issues: issues, Issue: issueCopy}
	}
}
