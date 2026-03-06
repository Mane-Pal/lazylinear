package helpers

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mane-pal/lazylinear/pkg/gui/styles"
	"github.com/mane-pal/lazylinear/pkg/gui/types"
	"github.com/mane-pal/lazylinear/pkg/linear"
	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

// MenuHelper provides menu-opening logic (loads data, builds items).
type MenuHelper struct {
	Client *linear.Client
}

func NewMenuHelper(client *linear.Client) *MenuHelper {
	return &MenuHelper{Client: client}
}

func (h *MenuHelper) OpenStateMenu(issue *models.Issue) tea.Cmd {
	if issue == nil || issue.Team == nil {
		return func() tea.Msg {
			return types.StatusMsg("No issue selected")
		}
	}

	issueCopy := issue
	return func() tea.Msg {
		team, err := h.Client.Teams.GetWithStates(context.Background(), issueCopy.Team.ID)
		if err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to load states: %v", err))
		}
		return types.StateMenuMsg{States: team.States, Issue: issueCopy}
	}
}

func (h *MenuHelper) OpenAssignMenu(issue *models.Issue) tea.Cmd {
	if issue == nil || issue.Team == nil {
		return func() tea.Msg {
			return types.StatusMsg("No issue selected")
		}
	}

	issueCopy := issue
	return func() tea.Msg {
		members, err := h.Client.Users.TeamMembers(context.Background(), issueCopy.Team.ID)
		if err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to load members: %v", err))
		}
		return types.AssignMenuMsg{Members: members, Issue: issueCopy}
	}
}

func (h *MenuHelper) BuildPriorityMenuItems() []types.MenuItem {
	return []types.MenuItem{
		{ID: "0", Label: "No priority", Color: styles.Secondary},
		{ID: "1", Label: "Urgent", Color: styles.PriorityUrgent},
		{ID: "2", Label: "High", Color: styles.PriorityHigh},
		{ID: "3", Label: "Medium", Color: styles.PriorityMedium},
		{ID: "4", Label: "Low", Color: styles.PriorityLow},
	}
}

func (h *MenuHelper) OpenLabelsMenu(issue *models.Issue) tea.Cmd {
	if issue == nil || issue.Team == nil {
		return func() tea.Msg {
			return types.StatusMsg("No issue selected")
		}
	}

	issueCopy := issue
	return func() tea.Msg {
		team, err := h.Client.Teams.GetWithStates(context.Background(), issueCopy.Team.ID)
		if err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to load labels: %v", err))
		}
		return types.LabelsMenuMsg{Labels: team.Labels, Issue: issueCopy}
	}
}

func (h *MenuHelper) OpenCycleMenu(issue *models.Issue) tea.Cmd {
	if issue == nil || issue.Team == nil {
		return func() tea.Msg {
			return types.StatusMsg("No issue selected")
		}
	}

	issueCopy := issue
	return func() tea.Msg {
		cycles, err := h.Client.Cycles.List(context.Background(), issueCopy.Team.ID)
		if err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to load cycles: %v", err))
		}
		return types.CyclesMenuMsg{Cycles: cycles, Issue: issueCopy}
	}
}

func (h *MenuHelper) BuildRelationTypeMenuItems() []types.MenuItem {
	return []types.MenuItem{
		{ID: "blocks", Label: "Blocks", Color: styles.PriorityUrgent},
		{ID: "blockedBy", Label: "Blocked by", Color: styles.PriorityHigh},
		{ID: "related", Label: "Related to", Color: styles.Primary},
		{ID: "duplicate", Label: "Duplicate of", Color: styles.Secondary},
	}
}
