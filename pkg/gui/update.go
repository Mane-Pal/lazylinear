package gui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	guicontext "github.com/mane-pal/lazylinear/pkg/gui/context"
	"github.com/mane-pal/lazylinear/pkg/gui/styles"
	"github.com/mane-pal/lazylinear/pkg/gui/types"
)

// Update handles messages
func (g *Gui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Help overlay intercepts all keys
		if g.State.ShowHelp {
			g.State.ShowHelp = false
			return g, nil
		}
		return g.handleKey(msg)

	case tea.WindowSizeMsg:
		g.width = msg.Width
		g.height = msg.Height
		return g, nil

	case types.TeamsLoadedMsg:
		g.State.Teams = msg.Teams
		// Pre-select team by name if provided via CLI
		if g.cliOpts.TeamName != "" {
			for i, team := range g.State.Teams {
				if strings.EqualFold(team.Name, g.cliOpts.TeamName) || strings.EqualFold(team.Key, g.cliOpts.TeamName) {
					g.State.SelectedTeam = i
					break
				}
			}
		}
		// Load states for the selected team
		if len(g.State.Teams) > 0 {
			return g, g.loadTeamStates()
		}
		return g, nil

	case types.IssuesLoadedMsg:
		g.State.Issues = msg.Issues
		g.State.Loading = false
		// Reset selection if out of bounds
		if g.State.SelectedIssue >= len(g.State.Issues) {
			g.State.SelectedIssue = max(0, len(g.State.Issues)-1)
		}
		return g, nil

	case types.ProjectsLoadedMsg:
		g.State.Projects = msg.Projects
		g.State.Loading = false
		// Reset selection if out of bounds
		if g.State.SelectedProject >= len(g.State.Projects) {
			g.State.SelectedProject = max(0, len(g.State.Projects)-1)
		}
		return g, nil

	case types.UserLoadedMsg:
		g.State.CurrentUser = msg.User
		return g, nil

	case types.ErrMsg:
		g.State.Err = msg.Err
		g.State.Loading = false
		return g, nil

	case types.StatusMsg:
		g.State.StatusMsg = string(msg)
		return g, clearStatusAfter(3 * time.Second)

	case types.ClearStatusMsg:
		g.State.StatusMsg = ""
		return g, nil

	case types.TeamStatesLoadedMsg:
		g.State.TeamStates = msg.States
		// Auto-select active states (backlog, unstarted, started) for initial load
		g.State.ActiveStateFilters = make(map[string]bool)
		for _, state := range g.State.TeamStates {
			if state.Type == "backlog" || state.Type == "unstarted" || state.Type == "started" {
				g.State.ActiveStateFilters[state.ID] = true
			}
		}

		// If CLI requested create issue mode, open the form now
		if g.cliOpts.CreateIssue {
			g.cliOpts.CreateIssue = false // Only trigger once
			g.State.Loading = false
			return g, g.openCreateFormCmd()
		}

		// Reload issues with the active state filter
		g.State.Loading = true
		return g, g.loadIssues()

	case types.TeamMembersLoadedMsg:
		g.State.TeamMembers = msg.Members
		return g, nil

	case types.IssueUpdatedMsg:
		// Update the issue in our list
		for i, issue := range g.State.Issues {
			if issue.ID == msg.Issue.ID {
				g.State.Issues[i] = msg.Issue
				break
			}
		}
		return g, nil

	case types.DetailedIssueLoadedMsg:
		g.State.DetailedIssue = msg.Issue
		g.State.SelectedComment = -1
		g.State.DetailScroll = 0
		return g, nil

	case types.CommentCreatedMsg:
		// Reload detailed issue to refresh comments
		if g.State.DetailedIssue != nil {
			g.closeForm()
			return g, g.loadDetailedIssue(g.State.DetailedIssue.ID)
		}
		return g, nil

	case types.CommentDeletedMsg:
		g.State.SelectedComment = -1
		if msg.IssueID != "" {
			return g, g.loadDetailedIssue(msg.IssueID)
		}
		return g, nil

	case types.BackgroundSyncMsg:
		// Background sync - reload issues silently
		g.State.LastSynced = time.Now()
		return g, tea.Batch(g.loadIssues(), backgroundSyncTick())

	case types.StateMenuMsg:
		// Convert to menu items
		items := make([]types.MenuItem, len(msg.States))
		for i, state := range msg.States {
			items[i] = types.MenuItem{
				ID:    state.ID,
				Label: state.Name,
				Color: styles.StateColor(state.Type),
			}
		}

		g.MenuPopup.MenuType = types.MenuState
		g.MenuPopup.MenuTitle = "Change State"
		g.MenuPopup.MenuItems = items
		g.MenuPopup.MenuSelected = 0

		// Pre-select current state
		if msg.Issue.State != nil {
			for i, item := range items {
				if item.ID == msg.Issue.State.ID {
					g.MenuPopup.MenuSelected = i
					break
				}
			}
		}
		return g, nil

	case types.ReloadIssuesMsg:
		g.State.Loading = true
		return g, tea.Batch(
			g.loadIssues(),
			func() tea.Msg { return types.StatusMsg("Updated") },
		)

	case types.AssignMenuMsg:
		// Convert to menu items, with "Unassigned" option first
		items := make([]types.MenuItem, len(msg.Members)+1)
		items[0] = types.MenuItem{
			ID:    "",
			Label: "Unassigned",
			Color: styles.Secondary,
		}
		for i, member := range msg.Members {
			items[i+1] = types.MenuItem{
				ID:    member.ID,
				Label: member.DisplayName,
				Color: styles.Primary,
			}
		}

		g.MenuPopup.MenuType = types.MenuAssign
		g.MenuPopup.MenuTitle = "Assign User"
		g.MenuPopup.MenuItems = items
		g.MenuPopup.MenuSelected = 0

		// Pre-select current assignee
		if msg.Issue.Assignee != nil {
			for i, item := range items {
				if item.ID == msg.Issue.Assignee.ID {
					g.MenuPopup.MenuSelected = i
					break
				}
			}
		}
		return g, nil

	case types.LabelsMenuMsg:
		// Store labels for rendering
		g.State.TeamLabels = msg.Labels

		// Build selected labels map from current issue labels
		g.State.SelectedLabels = make(map[string]bool)
		if msg.Issue.Labels != nil {
			for _, label := range msg.Issue.Labels.Nodes {
				g.State.SelectedLabels[label.ID] = true
			}
		}

		// Convert to menu items with checkmarks for selected
		items := make([]types.MenuItem, len(msg.Labels))
		for i, label := range msg.Labels {
			items[i] = types.MenuItem{
				ID:       label.ID,
				Label:    label.Name,
				Color:    lipgloss.Color(label.Color),
				Selected: g.State.SelectedLabels[label.ID],
			}
		}

		g.MenuPopup.MenuType = types.MenuLabels
		g.MenuPopup.MenuTitle = "Select Labels (Space to toggle)"
		g.MenuPopup.MenuItems = items
		g.MenuPopup.MenuSelected = 0
		return g, nil

	case types.CyclesMenuMsg:
		// Store cycles for reference
		g.State.TeamCycles = msg.Cycles

		// Convert to menu items with "No cycle" option first
		items := make([]types.MenuItem, len(msg.Cycles)+1)
		items[0] = types.MenuItem{
			ID:    "",
			Label: "No Cycle",
			Color: styles.Secondary,
		}
		for i, cycle := range msg.Cycles {
			items[i+1] = types.MenuItem{
				ID:    cycle.ID,
				Label: cycle.Name,
				Color: styles.Primary,
			}
		}

		g.MenuPopup.MenuType = types.MenuCycle
		g.MenuPopup.MenuTitle = "Assign to Cycle"
		g.MenuPopup.MenuItems = items
		g.MenuPopup.MenuSelected = 0

		// Pre-select current cycle
		if msg.Issue.Cycle != nil {
			for i, item := range items {
				if item.ID == msg.Issue.Cycle.ID {
					g.MenuPopup.MenuSelected = i
					break
				}
			}
		}
		return g, nil
	}

	return g, nil
}

func (g *Gui) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Form mode intercepts all keys
	if g.formMode != types.FormNone {
		return g.handleFormKey(msg)
	}

	// Search mode intercepts keys
	if g.SearchPopup.SearchMode {
		cmd := g.SearchPopup.HandleSearchKey(msg)
		return g, cmd
	}

	// Command mode intercepts keys
	if g.SearchPopup.CommandMode {
		cmd := g.SearchPopup.HandleCommandKey(msg)
		return g, cmd
	}

	// Menu intercepts keys when open
	if g.MenuPopup.IsOpen() {
		cmd, _ := g.MenuPopup.HandleKey(key)
		return g, cmd
	}

	// Global keys
	switch key {
	case "q", "ctrl+c":
		return g, tea.Quit

	case "?":
		g.State.ShowHelp = !g.State.ShowHelp
		return g, nil

	case "/":
		g.SearchPopup.OpenSearch()
		return g, nil

	case ":":
		g.SearchPopup.OpenCommandMode()
		return g, nil

	case "tab":
		if g.State.FocusedPanel == types.TeamsContext {
			g.cycleSidebarSection(1)
		} else {
			g.State.FocusedPanel = types.TeamsContext
		}
		return g, nil

	case "shift+tab":
		if g.State.FocusedPanel == types.TeamsContext {
			g.cycleSidebarSection(-1)
		} else {
			g.State.FocusedPanel = types.TeamsContext
		}
		return g, nil

	case "r":
		g.State.Loading = true
		g.State.LastSynced = time.Now()
		if g.State.MiddlePaneView == types.ViewProjects {
			return g, g.loadProjects()
		}
		return g, g.loadIssues()

	case "P":
		if g.State.MiddlePaneView == types.ViewIssues {
			g.State.MiddlePaneView = types.ViewProjects
			g.State.Loading = true
			return g, g.loadProjects()
		} else {
			g.State.MiddlePaneView = types.ViewIssues
			g.State.Loading = true
			return g, g.loadIssues()
		}

	case "m":
		if g.State.ActiveFilter == "my_issues" {
			g.State.ActiveFilter = "all"
		} else {
			g.State.ActiveFilter = "my_issues"
		}
		g.State.SelectedIssue = 0
		g.State.SelectedFilter = guicontext.GetFilterIndex(g.State.ActiveFilter)
		g.State.Loading = true
		return g, g.loadIssues()

	case "x":
		g.State.ActiveFilter = "all"
		g.State.ActiveStateFilters = make(map[string]bool)
		for _, st := range g.State.TeamStates {
			if st.Type == "backlog" || st.Type == "unstarted" || st.Type == "started" {
				g.State.ActiveStateFilters[st.ID] = true
			}
		}
		g.State.ActivePriorityFilter = -1
		g.State.SearchQuery = ""
		g.State.FilteredIssues = nil
		g.State.SelectedIssue = 0
		g.State.SelectedFilter = 0
		g.State.SelectedStateIdx = 0
		g.State.SelectedPriority = 0
		g.State.Loading = true
		return g, g.loadIssues()

	case "esc":
		if g.State.SearchQuery != "" {
			g.State.SearchQuery = ""
			g.State.FilteredIssues = nil
			g.State.SelectedIssue = 0
			return g, nil
		}
	}

	// Panel-specific keys - delegated to contexts
	switch g.State.FocusedPanel {
	case types.TeamsContext:
		return g, g.SidebarCtx.HandleKey(msg)
	case types.IssuesContext:
		return g, g.IssuesCtx.HandleKey(msg)
	case types.DetailContext:
		return g, g.DetailCtx.HandleKey(msg)
	}

	return g, nil
}

func (g *Gui) cycleSidebarSection(direction int) {
	guicontext.CycleSidebarSection(g.State, direction)
}
