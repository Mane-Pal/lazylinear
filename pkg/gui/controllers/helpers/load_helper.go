package helpers

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mane-pal/lazylinear/pkg/gui/state"
	"github.com/mane-pal/lazylinear/pkg/gui/types"
	"github.com/mane-pal/lazylinear/pkg/linear"
	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

// LoadHelper provides async data loading commands.
type LoadHelper struct {
	Client *linear.Client
	State  *state.AppState
}

func NewLoadHelper(client *linear.Client, st *state.AppState) *LoadHelper {
	return &LoadHelper{Client: client, State: st}
}

func (h *LoadHelper) LoadTeams() tea.Cmd {
	return func() tea.Msg {
		teams, err := h.Client.Teams.List(context.Background())
		if err != nil {
			return types.ErrMsg{Err: err}
		}
		return types.TeamsLoadedMsg{Teams: teams}
	}
}

func (h *LoadHelper) LoadIssues() tea.Cmd {
	return func() tea.Msg {
		filter := models.IssueFilter{}

		if h.State.SelectedTeam < len(h.State.Teams) && h.State.Teams[h.State.SelectedTeam] != nil {
			filter.TeamID = h.State.Teams[h.State.SelectedTeam].ID
		}

		if h.State.ActiveFilter == "my_issues" && h.State.CurrentUser != nil {
			filter.AssigneeID = h.State.CurrentUser.ID
		}

		if len(h.State.ActiveStateFilters) > 0 {
			filter.StateIDs = make([]string, 0, len(h.State.ActiveStateFilters))
			for stateID := range h.State.ActiveStateFilters {
				filter.StateIDs = append(filter.StateIDs, stateID)
			}
		}

		if h.State.ActivePriorityFilter > 0 {
			filter.Priority = &h.State.ActivePriorityFilter
		}

		issues, err := h.Client.Issues.List(context.Background(), filter)
		if err != nil {
			return types.ErrMsg{Err: err}
		}
		return types.IssuesLoadedMsg{Issues: issues}
	}
}

func (h *LoadHelper) LoadProjects() tea.Cmd {
	return func() tea.Msg {
		filter := models.ProjectFilter{}

		if h.State.SelectedTeam < len(h.State.Teams) && h.State.Teams[h.State.SelectedTeam] != nil {
			filter.TeamID = h.State.Teams[h.State.SelectedTeam].ID
		}

		projects, err := h.Client.Projects.List(context.Background(), filter)
		if err != nil {
			return types.ErrMsg{Err: err}
		}
		return types.ProjectsLoadedMsg{Projects: projects}
	}
}

func (h *LoadHelper) LoadUser() tea.Cmd {
	return func() tea.Msg {
		user, err := h.Client.Users.Viewer(context.Background())
		if err != nil {
			return types.ErrMsg{Err: err}
		}
		return types.UserLoadedMsg{User: user}
	}
}

func (h *LoadHelper) LoadTeamStates() tea.Cmd {
	return func() tea.Msg {
		if h.State.SelectedTeam >= len(h.State.Teams) {
			return nil
		}
		team, err := h.Client.Teams.GetWithStates(context.Background(), h.State.Teams[h.State.SelectedTeam].ID)
		if err != nil {
			return types.ErrMsg{Err: err}
		}
		return types.TeamStatesLoadedMsg{States: team.States}
	}
}

func (h *LoadHelper) LoadTeamMembers(teamIdx int) tea.Cmd {
	return func() tea.Msg {
		if teamIdx < 0 || teamIdx >= len(h.State.Teams) {
			teamIdx = h.State.SelectedTeam
		}
		if teamIdx >= len(h.State.Teams) {
			return nil
		}
		members, err := h.Client.Users.TeamMembers(context.Background(), h.State.Teams[teamIdx].ID)
		if err != nil {
			return types.ErrMsg{Err: err}
		}
		return types.TeamMembersLoadedMsg{Members: members}
	}
}

func (h *LoadHelper) LoadFormTeamStates(teamIdx int) tea.Cmd {
	return func() tea.Msg {
		if teamIdx < 0 || teamIdx >= len(h.State.Teams) {
			return nil
		}
		team, err := h.Client.Teams.GetWithStates(context.Background(), h.State.Teams[teamIdx].ID)
		if err != nil {
			return types.ErrMsg{Err: err}
		}
		return types.TeamStatesLoadedMsg{States: team.States}
	}
}

func (h *LoadHelper) LoadDetailedIssue(issueID string) tea.Cmd {
	return func() tea.Msg {
		issue, err := h.Client.Issues.Get(context.Background(), issueID)
		if err != nil {
			return types.ErrMsg{Err: err}
		}
		return types.DetailedIssueLoadedMsg{Issue: issue}
	}
}

func (h *LoadHelper) LoadSelectedIssueDetails() tea.Cmd {
	if issue := h.State.GetSelectedIssue(); issue != nil {
		h.State.DetailedIssue = nil
		return h.LoadDetailedIssue(issue.ID)
	}
	return nil
}
