package gui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mane-pal/lazylinear/pkg/gui/styles"
	"github.com/mane-pal/lazylinear/pkg/gui/types"
	"github.com/mane-pal/lazylinear/pkg/linear/models"
	"github.com/mane-pal/lazylinear/pkg/utils"
)

// Browser and clipboard actions

func (g *Gui) openInBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		if err := utils.OpenBrowser(url); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to open browser: %v", err))
		}
		return types.StatusMsg("Opened in browser")
	}
}

func (g *Gui) copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		if err := utils.CopyToClipboard(text); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to copy: %v", err))
		}
		return types.StatusMsg(fmt.Sprintf("Copied: %s", text))
	}
}

// Menu openers

func (g *Gui) openStateMenu() tea.Cmd {
	issue := g.State.GetSelectedIssue()
	if issue == nil || issue.Team == nil {
		return func() tea.Msg {
			return types.StatusMsg("No issue selected")
		}
	}

	issueCopy := issue
	return func() tea.Msg {
		team, err := g.client.Teams.GetWithStates(context.Background(), issueCopy.Team.ID)
		if err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to load states: %v", err))
		}
		return types.StateMenuMsg{States: team.States, Issue: issueCopy}
	}
}

func (g *Gui) openAssignMenu() tea.Cmd {
	issue := g.State.GetSelectedIssue()
	if issue == nil || issue.Team == nil {
		return func() tea.Msg {
			return types.StatusMsg("No issue selected")
		}
	}

	issueCopy := issue
	return func() tea.Msg {
		members, err := g.client.Users.TeamMembers(context.Background(), issueCopy.Team.ID)
		if err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to load members: %v", err))
		}
		return types.AssignMenuMsg{Members: members, Issue: issueCopy}
	}
}

func (g *Gui) openPriorityMenu() {
	priorities := []types.MenuItem{
		{ID: "0", Label: "No priority", Color: styles.Secondary},
		{ID: "1", Label: "Urgent", Color: styles.PriorityUrgent},
		{ID: "2", Label: "High", Color: styles.PriorityHigh},
		{ID: "3", Label: "Medium", Color: styles.PriorityMedium},
		{ID: "4", Label: "Low", Color: styles.PriorityLow},
	}

	g.MenuPopup.MenuType = types.MenuPriority
	g.MenuPopup.MenuTitle = "Set Priority"
	g.MenuPopup.MenuItems = priorities
	g.MenuPopup.MenuSelected = 0

	if issue := g.State.GetSelectedIssue(); issue != nil {
		g.MenuPopup.MenuSelected = issue.Priority
	}
}

// Issue update commands

func (g *Gui) updateIssueState(issueID, stateID string) tea.Cmd {
	return func() tea.Msg {
		if err := g.client.Issues.UpdateState(context.Background(), issueID, stateID); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update state: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

func (g *Gui) updateIssueAssignee(issueID, assigneeID string) tea.Cmd {
	return func() tea.Msg {
		var assignee *string
		if assigneeID != "" {
			assignee = &assigneeID
		}
		if err := g.client.Issues.UpdateAssignee(context.Background(), issueID, assignee); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update assignee: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

func (g *Gui) updateIssuePriority(issueID, priorityStr string) tea.Cmd {
	return func() tea.Msg {
		priority := 0
		fmt.Sscanf(priorityStr, "%d", &priority)
		if err := g.client.Issues.UpdatePriority(context.Background(), issueID, priority); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update priority: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

// Archive confirm dialog

func (g *Gui) openArchiveConfirm(issue *models.Issue) {
	g.MenuPopup.MenuType = types.MenuConfirm
	g.MenuPopup.ConfirmTitle = "Archive Issue?"
	g.MenuPopup.ConfirmMessage = fmt.Sprintf("Are you sure you want to archive\n%s: %s?", issue.Identifier, truncate(issue.Title, 30))
	issueID := issue.ID
	g.MenuPopup.ConfirmAction = func() tea.Cmd {
		return g.archiveIssue(issueID)
	}
}

func (g *Gui) archiveIssue(issueID string) tea.Cmd {
	return func() tea.Msg {
		if err := g.client.Issues.Archive(context.Background(), issueID); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to archive: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

// Label menu

func (g *Gui) openLabelsMenu() tea.Cmd {
	issue := g.State.GetSelectedIssue()
	if issue == nil || issue.Team == nil {
		return func() tea.Msg {
			return types.StatusMsg("No issue selected")
		}
	}

	issueCopy := issue
	return func() tea.Msg {
		team, err := g.client.Teams.GetWithStates(context.Background(), issueCopy.Team.ID)
		if err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to load labels: %v", err))
		}
		return types.LabelsMenuMsg{Labels: team.Labels, Issue: issueCopy}
	}
}

func (g *Gui) updateIssueLabels(issueID string, labelIDs []string) tea.Cmd {
	return func() tea.Msg {
		if err := g.client.Issues.UpdateLabels(context.Background(), issueID, labelIDs); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update labels: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

// Cycle menu

func (g *Gui) openCycleMenu() tea.Cmd {
	issue := g.State.GetSelectedIssue()
	if issue == nil || issue.Team == nil {
		return func() tea.Msg {
			return types.StatusMsg("No issue selected")
		}
	}

	issueCopy := issue
	return func() tea.Msg {
		cycles, err := g.client.Cycles.List(context.Background(), issueCopy.Team.ID)
		if err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to load cycles: %v", err))
		}
		return types.CyclesMenuMsg{Cycles: cycles, Issue: issueCopy}
	}
}

func (g *Gui) updateIssueCycle(issueID string, cycleID *string) tea.Cmd {
	return func() tea.Msg {
		if err := g.client.Issues.UpdateCycle(context.Background(), issueID, cycleID); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update cycle: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

// Due date and estimate

func (g *Gui) updateIssueDueDate(issueID string, dueDate *string) tea.Cmd {
	return func() tea.Msg {
		if err := g.client.Issues.UpdateDueDate(context.Background(), issueID, dueDate); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update due date: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

func (g *Gui) updateIssueEstimate(issueID string, estimate *float64) tea.Cmd {
	return func() tea.Msg {
		if err := g.client.Issues.UpdateEstimate(context.Background(), issueID, estimate); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update estimate: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

// Parent issue

func (g *Gui) updateIssueParent(issueID string, parentID *string) tea.Cmd {
	return func() tea.Msg {
		if err := g.client.Issues.UpdateParent(context.Background(), issueID, parentID); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update parent: %v", err))
		}
		return types.ReloadIssuesMsg{}
	}
}

// Relations

func (g *Gui) openRelationTypeMenu() {
	g.MenuPopup.MenuType = types.MenuRelationType
	g.MenuPopup.MenuTitle = "Add Relation"
	g.MenuPopup.MenuItems = []types.MenuItem{
		{ID: "blocks", Label: "Blocks", Color: styles.PriorityUrgent},
		{ID: "blockedBy", Label: "Blocked by", Color: styles.PriorityHigh},
		{ID: "related", Label: "Related to", Color: styles.Primary},
		{ID: "duplicate", Label: "Duplicate of", Color: styles.Secondary},
	}
	g.MenuPopup.MenuSelected = 0
}
