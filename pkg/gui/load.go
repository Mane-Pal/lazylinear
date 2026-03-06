package gui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

func (g *Gui) loadTeams() tea.Cmd {
	return func() tea.Msg {
		teams, err := g.client.Teams.List(context.Background())
		if err != nil {
			return errMsg{err}
		}
		return teamsLoadedMsg{teams}
	}
}

func (g *Gui) loadIssues() tea.Cmd {
	return func() tea.Msg {
		filter := models.IssueFilter{}

		// Team filter
		if g.selectedTeam < len(g.teams) && g.teams[g.selectedTeam] != nil {
			filter.TeamID = g.teams[g.selectedTeam].ID
		}

		// Assignee filter ("My Issues")
		if g.activeFilter == "my_issues" && g.currentUser != nil {
			filter.AssigneeID = g.currentUser.ID
		}

		// State filter (multi-select)
		if len(g.activeStateFilters) > 0 {
			filter.StateIDs = make([]string, 0, len(g.activeStateFilters))
			for stateID := range g.activeStateFilters {
				filter.StateIDs = append(filter.StateIDs, stateID)
			}
		}

		// Priority filter
		if g.activePriorityFilter > 0 {
			filter.Priority = &g.activePriorityFilter
		}

		issues, err := g.client.Issues.List(context.Background(), filter)
		if err != nil {
			return errMsg{err}
		}
		return issuesLoadedMsg{issues}
	}
}

func (g *Gui) loadProjects() tea.Cmd {
	return func() tea.Msg {
		filter := models.ProjectFilter{}

		// Team filter
		if g.selectedTeam < len(g.teams) && g.teams[g.selectedTeam] != nil {
			filter.TeamID = g.teams[g.selectedTeam].ID
		}

		projects, err := g.client.Projects.List(context.Background(), filter)
		if err != nil {
			return errMsg{err}
		}
		return projectsLoadedMsg{projects}
	}
}

func (g *Gui) loadUser() tea.Cmd {
	return func() tea.Msg {
		user, err := g.client.Users.Viewer(context.Background())
		if err != nil {
			return errMsg{err}
		}
		return userLoadedMsg{user}
	}
}

func (g *Gui) loadTeamStates() tea.Cmd {
	return func() tea.Msg {
		if g.selectedTeam >= len(g.teams) {
			return nil
		}
		team, err := g.client.Teams.GetWithStates(context.Background(), g.teams[g.selectedTeam].ID)
		if err != nil {
			return errMsg{err}
		}
		return teamStatesLoadedMsg{states: team.States}
	}
}

func (g *Gui) loadTeamMembers() tea.Cmd {
	return func() tea.Msg {
		teamIdx := g.formTeam
		if teamIdx < 0 || teamIdx >= len(g.teams) {
			teamIdx = g.selectedTeam
		}
		if teamIdx >= len(g.teams) {
			return nil
		}
		members, err := g.client.Users.TeamMembers(context.Background(), g.teams[teamIdx].ID)
		if err != nil {
			return errMsg{err}
		}
		return teamMembersLoadedMsg{members: members}
	}
}

func (g *Gui) loadFormTeamStates() tea.Cmd {
	return func() tea.Msg {
		teamIdx := g.formTeam
		if teamIdx < 0 || teamIdx >= len(g.teams) {
			return nil
		}
		team, err := g.client.Teams.GetWithStates(context.Background(), g.teams[teamIdx].ID)
		if err != nil {
			return errMsg{err}
		}
		return teamStatesLoadedMsg{states: team.States}
	}
}

func (g *Gui) loadDetailedIssue(issueID string) tea.Cmd {
	return func() tea.Msg {
		issue, err := g.client.Issues.Get(context.Background(), issueID)
		if err != nil {
			return errMsg{err}
		}
		return detailedIssueLoadedMsg{issue}
	}
}

func (g *Gui) loadSelectedIssueDetails() tea.Cmd {
	if issue := g.getSelectedIssue(); issue != nil {
		// Clear previous detailed issue when selection changes
		g.detailedIssue = nil
		return g.loadDetailedIssue(issue.ID)
	}
	return nil
}
