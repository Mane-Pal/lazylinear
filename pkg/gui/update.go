package gui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mane-pal/lazylinear/pkg/gui/styles"
	"github.com/mane-pal/lazylinear/pkg/gui/types"
	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

// Update handles messages
func (g *Gui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Help overlay intercepts all keys
		if g.showHelp {
			g.showHelp = false
			return g, nil
		}
		return g.handleKey(msg)

	case tea.WindowSizeMsg:
		g.width = msg.Width
		g.height = msg.Height
		return g, nil

	case teamsLoadedMsg:
		g.teams = msg.teams
		// Pre-select team by name if provided via CLI
		if g.cliOpts.TeamName != "" {
			for i, team := range g.teams {
				if strings.EqualFold(team.Name, g.cliOpts.TeamName) || strings.EqualFold(team.Key, g.cliOpts.TeamName) {
					g.selectedTeam = i
					break
				}
			}
		}
		// Load states for the selected team
		if len(g.teams) > 0 {
			return g, g.loadTeamStates()
		}
		return g, nil

	case issuesLoadedMsg:
		g.issues = msg.issues
		g.loading = false
		// Reset selection if out of bounds
		if g.selectedIssue >= len(g.issues) {
			g.selectedIssue = max(0, len(g.issues)-1)
		}
		return g, nil

	case projectsLoadedMsg:
		g.projects = msg.projects
		g.loading = false
		// Reset selection if out of bounds
		if g.selectedProject >= len(g.projects) {
			g.selectedProject = max(0, len(g.projects)-1)
		}
		return g, nil

	case userLoadedMsg:
		g.currentUser = msg.user
		return g, nil

	case errMsg:
		g.err = msg.err
		g.loading = false
		return g, nil

	case statusMsg:
		g.statusMsg = string(msg)
		return g, clearStatusAfter(3 * time.Second)

	case clearStatusMsg:
		g.statusMsg = ""
		return g, nil

	case teamStatesLoadedMsg:
		g.teamStates = msg.states
		// Auto-select active states (backlog, unstarted, started) for initial load
		g.activeStateFilters = make(map[string]bool)
		for _, state := range g.teamStates {
			if state.Type == "backlog" || state.Type == "unstarted" || state.Type == "started" {
				g.activeStateFilters[state.ID] = true
			}
		}

		// If CLI requested create issue mode, open the form now
		if g.cliOpts.CreateIssue {
			g.cliOpts.CreateIssue = false // Only trigger once
			g.loading = false
			return g, g.openCreateFormCmd()
		}

		// Reload issues with the active state filter
		g.loading = true
		return g, g.loadIssues()

	case teamMembersLoadedMsg:
		g.teamMembers = msg.members
		return g, nil

	case issueUpdatedMsg:
		// Update the issue in our list
		for i, issue := range g.issues {
			if issue.ID == msg.issue.ID {
				g.issues[i] = msg.issue
				break
			}
		}
		return g, nil

	case detailedIssueLoadedMsg:
		g.detailedIssue = msg.issue
		g.selectedComment = -1
		g.detailScroll = 0
		return g, nil

	case commentCreatedMsg:
		// Reload detailed issue to refresh comments
		if g.detailedIssue != nil {
			g.closeForm()
			return g, g.loadDetailedIssue(g.detailedIssue.ID)
		}
		return g, nil

	case commentDeletedMsg:
		g.selectedComment = -1
		if msg.issueID != "" {
			return g, g.loadDetailedIssue(msg.issueID)
		}
		return g, nil

	case backgroundSyncMsg:
		// Background sync - reload issues silently
		g.lastSynced = time.Now()
		return g, tea.Batch(g.loadIssues(), backgroundSyncTick())

	case stateMenuMsg:
		// Convert to menu items
		items := make([]MenuItem, len(msg.states))
		for i, state := range msg.states {
			items[i] = MenuItem{
				ID:    state.ID,
				Label: state.Name,
				Color: styles.StateColor(state.Type),
			}
		}

		g.menuType = MenuState
		g.menuTitle = "Change State"
		g.menuItems = items
		g.menuSelected = 0

		// Pre-select current state
		if msg.issue.State != nil {
			for i, item := range items {
				if item.ID == msg.issue.State.ID {
					g.menuSelected = i
					break
				}
			}
		}
		return g, nil

	case reloadIssuesMsg:
		g.loading = true
		return g, tea.Batch(
			g.loadIssues(),
			func() tea.Msg { return statusMsg("Updated") },
		)

	case assignMenuMsg:
		// Convert to menu items, with "Unassigned" option first
		items := make([]MenuItem, len(msg.members)+1)
		items[0] = MenuItem{
			ID:    "",
			Label: "Unassigned",
			Color: styles.Secondary,
		}
		for i, member := range msg.members {
			items[i+1] = MenuItem{
				ID:    member.ID,
				Label: member.DisplayName,
				Color: styles.Primary,
			}
		}

		g.menuType = MenuAssign
		g.menuTitle = "Assign User"
		g.menuItems = items
		g.menuSelected = 0

		// Pre-select current assignee
		if msg.issue.Assignee != nil {
			for i, item := range items {
				if item.ID == msg.issue.Assignee.ID {
					g.menuSelected = i
					break
				}
			}
		}
		return g, nil

	case labelsMenuMsg:
		// Store labels for rendering
		g.teamLabels = msg.labels

		// Build selected labels map from current issue labels
		g.selectedLabels = make(map[string]bool)
		if msg.issue.Labels != nil {
			for _, label := range msg.issue.Labels.Nodes {
				g.selectedLabels[label.ID] = true
			}
		}

		// Convert to menu items with checkmarks for selected
		items := make([]MenuItem, len(msg.labels))
		for i, label := range msg.labels {
			items[i] = MenuItem{
				ID:       label.ID,
				Label:    label.Name,
				Color:    lipgloss.Color(label.Color),
				Selected: g.selectedLabels[label.ID],
			}
		}

		g.menuType = MenuLabels
		g.menuTitle = "Select Labels (Space to toggle)"
		g.menuItems = items
		g.menuSelected = 0
		return g, nil

	case cyclesMenuMsg:
		// Store cycles for reference
		g.teamCycles = msg.cycles

		// Convert to menu items with "No cycle" option first
		items := make([]MenuItem, len(msg.cycles)+1)
		items[0] = MenuItem{
			ID:    "",
			Label: "No Cycle",
			Color: styles.Secondary,
		}
		for i, cycle := range msg.cycles {
			items[i+1] = MenuItem{
				ID:    cycle.ID,
				Label: cycle.Name,
				Color: styles.Primary,
			}
		}

		g.menuType = MenuCycle
		g.menuTitle = "Assign to Cycle"
		g.menuItems = items
		g.menuSelected = 0

		// Pre-select current cycle
		if msg.issue.Cycle != nil {
			for i, item := range items {
				if item.ID == msg.issue.Cycle.ID {
					g.menuSelected = i
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
	if g.formMode != FormNone {
		return g.handleFormKey(msg)
	}

	// Search mode intercepts keys
	if g.searchMode {
		return g.handleSearchKey(msg)
	}

	// Command mode intercepts keys
	if g.commandMode {
		return g.handleCommandKey(msg)
	}

	// Menu intercepts keys when open
	if g.menuType != MenuNone {
		return g.handleMenuKey(key)
	}

	// Global keys
	switch key {
	case "q", "ctrl+c":
		return g, tea.Quit

	case "?":
		g.showHelp = !g.showHelp
		return g, nil

	case "/":
		g.openSearch()
		return g, nil

	case ":":
		g.openCommandMode()
		return g, nil

	case "tab":
		// Tab cycles within sidebar, or returns to sidebar from other panels
		if g.focusedPanel == types.TeamsContext {
			g.cycleSidebarSection(1)
		} else {
			g.focusedPanel = types.TeamsContext
		}
		return g, nil

	case "shift+tab":
		// Shift+Tab cycles backwards within sidebar, or returns to sidebar
		if g.focusedPanel == types.TeamsContext {
			g.cycleSidebarSection(-1)
		} else {
			g.focusedPanel = types.TeamsContext
		}
		return g, nil

	case "r":
		g.loading = true
		g.lastSynced = time.Now()
		if g.middlePaneView == ViewProjects {
			return g, g.loadProjects()
		}
		return g, g.loadIssues()

	case "P":
		// Toggle between Issues and Projects view
		if g.middlePaneView == ViewIssues {
			g.middlePaneView = ViewProjects
			g.loading = true
			return g, g.loadProjects()
		} else {
			g.middlePaneView = ViewIssues
			g.loading = true
			return g, g.loadIssues()
		}

	case "m":
		// Toggle My Issues filter and reload from API
		if g.activeFilter == "my_issues" {
			g.activeFilter = "all"
		} else {
			g.activeFilter = "my_issues"
		}
		g.selectedIssue = 0
		g.selectedFilter = g.getFilterIndex(g.activeFilter)
		g.loading = true
		return g, g.loadIssues()

	case "x":
		// Clear all filters and reload with active states only
		g.activeFilter = "all"
		g.activeStateFilters = make(map[string]bool)
		// Re-select active states (backlog, unstarted, started)
		for _, state := range g.teamStates {
			if state.Type == "backlog" || state.Type == "unstarted" || state.Type == "started" {
				g.activeStateFilters[state.ID] = true
			}
		}
		g.activePriorityFilter = -1
		g.searchQuery = ""
		g.filteredIssues = nil
		g.selectedIssue = 0
		g.selectedFilter = 0
		g.selectedStateIdx = 0
		g.selectedPriority = 0
		g.loading = true
		return g, g.loadIssues()

	case "esc":
		// Clear search filter
		if g.searchQuery != "" {
			g.searchQuery = ""
			g.filteredIssues = nil
			g.selectedIssue = 0
			return g, nil
		}
	}

	// Panel-specific keys
	switch g.focusedPanel {
	case types.TeamsContext:
		return g.handleTeamsKey(key)
	case types.IssuesContext:
		return g.handleIssuesKey(key)
	case types.DetailContext:
		return g.handleDetailKey(key)
	}

	return g, nil
}

func (g *Gui) handleMenuKey(key string) (tea.Model, tea.Cmd) {
	// Handle confirm dialog
	if g.menuType == MenuConfirm {
		switch key {
		case "y", "Y", "enter":
			cmd := g.confirmAction()
			g.closeMenu()
			return g, cmd
		case "n", "N", "esc", "q":
			g.closeMenu()
		}
		return g, nil
	}

	// Handle label multi-select menu
	if g.menuType == MenuLabels {
		switch key {
		case "j", "down":
			if g.menuSelected < len(g.menuItems)-1 {
				g.menuSelected++
			}
		case "k", "up":
			if g.menuSelected > 0 {
				g.menuSelected--
			}
		case " ": // Space toggles selection
			if g.menuSelected < len(g.menuItems) {
				item := &g.menuItems[g.menuSelected]
				item.Selected = !item.Selected
				g.selectedLabels[item.ID] = item.Selected
				if !item.Selected {
					delete(g.selectedLabels, item.ID)
				}
			}
		case "enter": // Confirm selection
			issue := g.getSelectedIssue()
			if issue != nil {
				labelIDs := make([]string, 0, len(g.selectedLabels))
				for id := range g.selectedLabels {
					labelIDs = append(labelIDs, id)
				}
				cmd := g.updateIssueLabels(issue.ID, labelIDs)
				g.closeMenu()
				return g, cmd
			}
			g.closeMenu()
		case "esc", "q":
			g.closeMenu()
		}
		return g, nil
	}

	// Handle regular menus
	switch key {
	case "j", "down":
		if g.menuSelected < len(g.menuItems)-1 {
			g.menuSelected++
		}
	case "k", "up":
		if g.menuSelected > 0 {
			g.menuSelected--
		}
	case "enter":
		return g.selectMenuItem()
	case "esc", "q":
		g.closeMenu()
	}
	return g, nil
}

func (g *Gui) closeMenu() {
	g.menuType = MenuNone
	g.menuItems = nil
	g.menuSelected = 0
	g.menuTitle = ""
}

func (g *Gui) selectMenuItem() (tea.Model, tea.Cmd) {
	if g.menuSelected >= len(g.menuItems) {
		g.closeMenu()
		return g, nil
	}

	item := g.menuItems[g.menuSelected]
	issue := g.getSelectedIssue()
	if issue == nil {
		g.closeMenu()
		return g, nil
	}

	var cmd tea.Cmd

	switch g.menuType {
	case MenuState:
		cmd = g.updateIssueState(issue.ID, item.ID)
	case MenuAssign:
		cmd = g.updateIssueAssignee(issue.ID, item.ID)
	case MenuPriority:
		cmd = g.updateIssuePriority(issue.ID, item.ID)
	case MenuCycle:
		var cycleID *string
		if item.ID != "" {
			cycleID = &item.ID
		}
		cmd = g.updateIssueCycle(issue.ID, cycleID)
	case MenuRelationType:
		// Store the relation type and open issue picker
		g.closeMenu()
		// TODO: Open issue picker for selecting the related issue
		return g, func() tea.Msg { return statusMsg("Relation type selected: " + item.ID) }
	}

	g.closeMenu()
	return g, cmd
}

func (g *Gui) handleTeamsKey(key string) (tea.Model, tea.Cmd) {
	// Sections: 0=teams, 1=filters, 2=states, 3=priority
	switch key {
	case "j", "down":
		switch g.sidebarSection {
		case 0: // Teams section
			if g.selectedTeam < len(g.teams)-1 {
				g.selectedTeam++
			} else {
				g.sidebarSection = 1
				g.selectedFilter = 0
			}
		case 1: // Filters section
			if g.selectedFilter < len(filterOptions)-1 {
				g.selectedFilter++
			} else {
				g.sidebarSection = 2
				g.selectedStateIdx = 0
			}
		case 2: // States section
			if g.selectedStateIdx < len(g.teamStates)-1 {
				g.selectedStateIdx++
			} else {
				g.sidebarSection = 3
				g.selectedPriority = 0
			}
		case 3: // Priority section
			if g.selectedPriority < 3 {
				g.selectedPriority++
			}
		}
	case "k", "up":
		switch g.sidebarSection {
		case 0: // Teams section
			if g.selectedTeam > 0 {
				g.selectedTeam--
			}
		case 1: // Filters section
			if g.selectedFilter > 0 {
				g.selectedFilter--
			} else {
				g.sidebarSection = 0
			}
		case 2: // States section
			if g.selectedStateIdx > 0 {
				g.selectedStateIdx--
			} else {
				g.sidebarSection = 1
				g.selectedFilter = len(filterOptions) - 1
			}
		case 3: // Priority section
			if g.selectedPriority > 0 {
				g.selectedPriority--
			} else {
				g.sidebarSection = 2
				if len(g.teamStates) > 0 {
					g.selectedStateIdx = len(g.teamStates) - 1
				}
			}
		}
	case " ": // Space toggles filters/states/priority and reloads from API
		switch g.sidebarSection {
		case 1: // Toggle filter
			g.activeFilter = filterOptions[g.selectedFilter].key
			g.selectedIssue = 0
			g.loading = true
			return g, g.loadIssues()
		case 2: // Toggle state filter (multi-select)
			if g.selectedStateIdx < len(g.teamStates) {
				stateID := g.teamStates[g.selectedStateIdx].ID
				if g.activeStateFilters[stateID] {
					delete(g.activeStateFilters, stateID)
				} else {
					g.activeStateFilters[stateID] = true
				}
				g.selectedIssue = 0
				g.loading = true
				return g, g.loadIssues()
			}
		case 3: // Toggle priority filter
			priority := g.selectedPriority + 1
			if g.activePriorityFilter == priority {
				g.activePriorityFilter = -1
			} else {
				g.activePriorityFilter = priority
			}
			g.selectedIssue = 0
			g.loading = true
			return g, g.loadIssues()
		}
	case "enter", "l":
		// Enter selects team and moves to Issues panel
		if g.sidebarSection == 0 {
			g.selectedIssue = 0
			g.activeStateFilters = make(map[string]bool) // Clear state filters
			g.loading = true
			g.focusedPanel = types.IssuesContext
			return g, tea.Batch(g.loadIssues(), g.loadTeamStates())
		}
		// For other sections, just move to Issues panel
		g.focusedPanel = types.IssuesContext
	case "m":
		// Quick toggle for My Issues and reload
		if g.activeFilter == "my_issues" {
			g.activeFilter = "all"
		} else {
			g.activeFilter = "my_issues"
		}
		g.selectedIssue = 0
		g.selectedFilter = g.getFilterIndex(g.activeFilter)
		g.loading = true
		return g, g.loadIssues()
	case "x":
		// Clear all filters and reload with active states
		g.activeFilter = "all"
		g.activeStateFilters = make(map[string]bool)
		for _, state := range g.teamStates {
			if state.Type == "backlog" || state.Type == "unstarted" || state.Type == "started" {
				g.activeStateFilters[state.ID] = true
			}
		}
		g.activePriorityFilter = -1
		g.selectedIssue = 0
		g.selectedFilter = 0
		g.selectedStateIdx = 0
		g.selectedPriority = 0
		g.loading = true
		return g, g.loadIssues()
	}
	return g, nil
}

func (g *Gui) getFilterIndex(filter string) int {
	for i, f := range filterOptions {
		if f.key == filter {
			return i
		}
	}
	return 0
}

func (g *Gui) handleIssuesKey(key string) (tea.Model, tea.Cmd) {
	// Handle projects view navigation
	if g.middlePaneView == ViewProjects {
		switch key {
		case "j", "down":
			if g.selectedProject < len(g.projects)-1 {
				g.selectedProject++
			}
		case "k", "up":
			if g.selectedProject > 0 {
				g.selectedProject--
			}
		case "g":
			g.selectedProject = 0
		case "G":
			g.selectedProject = max(0, len(g.projects)-1)
		case "enter", "l":
			// Could show project detail in future
			// For now, show project issues?
			return g, nil
		case "h":
			g.focusedPanel = types.TeamsContext
		case "o":
			// Open project in browser
			if g.selectedProject < len(g.projects) {
				return g, g.openInBrowser(g.projects[g.selectedProject].URL)
			}
		}
		return g, nil
	}

	// Handle issues view navigation
	switch key {
	case "j", "down":
		issues := g.getDisplayIssues()
		if g.selectedIssue < len(issues)-1 {
			g.selectedIssue++
			return g, g.loadSelectedIssueDetails()
		}
	case "k", "up":
		if g.selectedIssue > 0 {
			g.selectedIssue--
			return g, g.loadSelectedIssueDetails()
		}
	case "g":
		g.selectedIssue = 0
		return g, g.loadSelectedIssueDetails()
	case "G":
		issues := g.getDisplayIssues()
		g.selectedIssue = max(0, len(issues)-1)
		return g, g.loadSelectedIssueDetails()
	case "enter", "l":
		g.focusedPanel = types.DetailContext
		return g, g.loadSelectedIssueDetails()
	case "h":
		g.focusedPanel = types.TeamsContext

	// Quick actions
	case "o":
		// Open in browser
		if issue := g.getSelectedIssue(); issue != nil {
			return g, g.openInBrowser(issue.URL)
		}
	case "y":
		// Yank URL
		if issue := g.getSelectedIssue(); issue != nil {
			return g, g.copyToClipboard(issue.URL)
		}
	case "Y":
		// Yank identifier
		if issue := g.getSelectedIssue(); issue != nil {
			return g, g.copyToClipboard(issue.Identifier)
		}

	// Menus
	case "s":
		// Change state
		return g, g.openStateMenu()
	case "a":
		// Assign user
		return g, g.openAssignMenu()
	case "p":
		// Set priority
		g.openPriorityMenu()

	// Create/Edit/Archive
	case "n":
		// New issue
		return g, g.openCreateFormCmd()
	case "e":
		// Edit issue (also used for estimate in detail panel)
		if issue := g.getSelectedIssue(); issue != nil {
			g.openEditForm(issue)
		}
	case "d":
		// Archive issue
		if issue := g.getSelectedIssue(); issue != nil {
			g.openArchiveConfirm(issue)
		}

	// New features
	case "L":
		// Open labels menu (multi-select)
		return g, g.openLabelsMenu()
	case "c":
		// Assign to cycle
		return g, g.openCycleMenu()
	}
	return g, nil
}

func (g *Gui) handleDetailKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "h", "esc":
		g.focusedPanel = types.IssuesContext
		g.selectedComment = -1
	case "j", "down":
		// Navigate comments
		if g.detailedIssue != nil {
			comments := g.detailedIssue.GetComments()
			if len(comments) > 0 {
				if g.selectedComment < len(comments)-1 {
					g.selectedComment++
				}
			}
		}
	case "k", "up":
		// Navigate comments
		if g.selectedComment > -1 {
			g.selectedComment--
		}

	// Quick actions (same as issues panel)
	case "o":
		if issue := g.getSelectedIssue(); issue != nil {
			return g, g.openInBrowser(issue.URL)
		}
	case "y":
		if issue := g.getSelectedIssue(); issue != nil {
			return g, g.copyToClipboard(issue.URL)
		}
	case "Y":
		if issue := g.getSelectedIssue(); issue != nil {
			return g, g.copyToClipboard(issue.Identifier)
		}

	// Menus
	case "s":
		return g, g.openStateMenu()
	case "a":
		return g, g.openAssignMenu()
	case "p":
		g.openPriorityMenu()

	// Labels (multi-select)
	case "L":
		return g, g.openLabelsMenu()

	// Cycle
	case "c":
		// In detail, c is used for both comment and cycle
		// If no detailed issue, add comment, otherwise assign cycle
		if g.selectedComment == -1 && g.detailedIssue != nil {
			// Check if this could be for add comment (if not showing cycle info)
			// For now, let's use Ctrl+C for cycle and c for comment
			g.openCommentForm()
		}

	// Due date
	case "D":
		if issue := g.getSelectedIssue(); issue != nil {
			g.openDueDateInput(issue)
		}

	// Estimate
	case "E":
		if issue := g.getSelectedIssue(); issue != nil {
			g.openEstimateInput(issue)
		}

	// Parent issue
	case "P":
		if issue := g.getSelectedIssue(); issue != nil {
			return g, g.openParentPicker(issue)
		}

	// Relations
	case "R":
		g.openRelationTypeMenu()

	// Cycle (alternative key)
	case "C":
		return g, g.openCycleMenu()

	// Edit - context dependent
	case "e":
		if g.selectedComment >= 0 && g.detailedIssue != nil {
			// Edit selected comment
			comments := g.detailedIssue.GetComments()
			if g.selectedComment < len(comments) {
				comment := comments[g.selectedComment]
				// Only allow editing own comments
				if g.currentUser != nil && comment.User != nil && comment.User.ID == g.currentUser.ID {
					g.openEditCommentForm(comment)
				} else {
					return g, func() tea.Msg { return statusMsg("Can only edit your own comments") }
				}
			}
		} else if issue := g.getSelectedIssue(); issue != nil {
			// Edit issue
			g.openEditForm(issue)
		}

	// Delete - context dependent
	case "d":
		if g.selectedComment >= 0 && g.detailedIssue != nil {
			// Delete selected comment
			comments := g.detailedIssue.GetComments()
			if g.selectedComment < len(comments) {
				comment := comments[g.selectedComment]
				// Only allow deleting own comments
				if g.currentUser != nil && comment.User != nil && comment.User.ID == g.currentUser.ID {
					g.openDeleteCommentConfirm(comment)
				} else {
					return g, func() tea.Msg { return statusMsg("Can only delete your own comments") }
				}
			}
		} else if issue := g.getSelectedIssue(); issue != nil {
			// Archive issue
			g.openArchiveConfirm(issue)
		}

	// Sub-issue creation
	case "N":
		if issue := g.getSelectedIssue(); issue != nil {
			return g, g.openCreateSubIssueFormCmd(issue)
		}
	}
	return g, nil
}

func (g *Gui) cycleFocus(direction int) {
	panels := []types.ContextKey{
		types.TeamsContext,
		types.IssuesContext,
		types.DetailContext,
	}

	current := 0
	for i, p := range panels {
		if p == g.focusedPanel {
			current = i
			break
		}
	}

	next := (current + direction + len(panels)) % len(panels)
	g.focusedPanel = panels[next]
}

func (g *Gui) cycleSidebarSection(direction int) {
	// Sections: 0=teams, 1=filters, 2=states, 3=priority
	numSections := 4
	g.sidebarSection = (g.sidebarSection + direction + numSections) % numSections

	// Reset selection index for new section
	switch g.sidebarSection {
	case 0:
		// Keep selectedTeam as is
	case 1:
		g.selectedFilter = 0
	case 2:
		g.selectedStateIdx = 0
	case 3:
		g.selectedPriority = 0
	}
}

func (g *Gui) getDisplayIssues() []*models.Issue {
	// State/priority/assignee filtering is now done server-side via API
	// This function only handles search filtering (client-side for interactive search)
	if g.filteredIssues != nil {
		return g.filteredIssues
	}
	return g.issues
}

func (g *Gui) getSelectedIssue() *models.Issue {
	issues := g.getDisplayIssues()
	if g.selectedIssue < len(issues) {
		return issues[g.selectedIssue]
	}
	return nil
}
