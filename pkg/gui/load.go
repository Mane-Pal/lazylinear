package gui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mane-pal/lazylinear/pkg/gui/types"
	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

func (g *Gui) loadTeams() tea.Cmd {
	return func() tea.Msg {
		teams, err := g.client.Teams.List(context.Background())
		if err != nil {
			return types.ErrMsg{Err: err}
		}
		return types.TeamsLoadedMsg{Teams: teams}
	}
}

func (g *Gui) loadIssues() tea.Cmd {
	return func() tea.Msg {
		filter := models.IssueFilter{}

		if g.State.SelectedTeam < len(g.State.Teams) && g.State.Teams[g.State.SelectedTeam] != nil {
			filter.TeamID = g.State.Teams[g.State.SelectedTeam].ID
		}

		if g.State.ActiveFilter == "my_issues" && g.State.CurrentUser != nil {
			filter.AssigneeID = g.State.CurrentUser.ID
		}

		if len(g.State.ActiveStateFilters) > 0 {
			filter.StateIDs = make([]string, 0, len(g.State.ActiveStateFilters))
			for stateID := range g.State.ActiveStateFilters {
				filter.StateIDs = append(filter.StateIDs, stateID)
			}
		}

		if g.State.ActivePriorityFilter > 0 {
			filter.Priority = &g.State.ActivePriorityFilter
		}

		issues, err := g.client.Issues.List(context.Background(), filter)
		if err != nil {
			return types.ErrMsg{Err: err}
		}
		return types.IssuesLoadedMsg{Issues: issues}
	}
}

func (g *Gui) loadProjects() tea.Cmd {
	return func() tea.Msg {
		filter := models.ProjectFilter{}

		if g.State.SelectedTeam < len(g.State.Teams) && g.State.Teams[g.State.SelectedTeam] != nil {
			filter.TeamID = g.State.Teams[g.State.SelectedTeam].ID
		}

		projects, err := g.client.Projects.List(context.Background(), filter)
		if err != nil {
			return types.ErrMsg{Err: err}
		}
		return types.ProjectsLoadedMsg{Projects: projects}
	}
}

func (g *Gui) loadUser() tea.Cmd {
	return func() tea.Msg {
		user, err := g.client.Users.Viewer(context.Background())
		if err != nil {
			return types.ErrMsg{Err: err}
		}
		return types.UserLoadedMsg{User: user}
	}
}

func (g *Gui) loadTeamStates() tea.Cmd {
	return func() tea.Msg {
		if g.State.SelectedTeam >= len(g.State.Teams) {
			return nil
		}
		team, err := g.client.Teams.GetWithStates(context.Background(), g.State.Teams[g.State.SelectedTeam].ID)
		if err != nil {
			return types.ErrMsg{Err: err}
		}
		return types.TeamStatesLoadedMsg{States: team.States}
	}
}

func (g *Gui) loadTeamMembers() tea.Cmd {
	return func() tea.Msg {
		teamIdx := g.formTeam
		if teamIdx < 0 || teamIdx >= len(g.State.Teams) {
			teamIdx = g.State.SelectedTeam
		}
		if teamIdx >= len(g.State.Teams) {
			return nil
		}
		members, err := g.client.Users.TeamMembers(context.Background(), g.State.Teams[teamIdx].ID)
		if err != nil {
			return types.ErrMsg{Err: err}
		}
		return types.TeamMembersLoadedMsg{Members: members}
	}
}

func (g *Gui) loadFormTeamStates() tea.Cmd {
	return func() tea.Msg {
		teamIdx := g.formTeam
		if teamIdx < 0 || teamIdx >= len(g.State.Teams) {
			return nil
		}
		team, err := g.client.Teams.GetWithStates(context.Background(), g.State.Teams[teamIdx].ID)
		if err != nil {
			return types.ErrMsg{Err: err}
		}
		return types.TeamStatesLoadedMsg{States: team.States}
	}
}

func (g *Gui) loadDetailedIssue(issueID string) tea.Cmd {
	return func() tea.Msg {
		issue, err := g.client.Issues.Get(context.Background(), issueID)
		if err != nil {
			return types.ErrMsg{Err: err}
		}
		return types.DetailedIssueLoadedMsg{Issue: issue}
	}
}

func (g *Gui) loadSelectedIssueDetails() tea.Cmd {
	if issue := g.State.GetSelectedIssue(); issue != nil {
		g.State.DetailedIssue = nil
		return g.loadDetailedIssue(issue.ID)
	}
	return nil
}
