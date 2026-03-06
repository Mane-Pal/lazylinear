package gui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/mane-pal/lazylinear/pkg/gui/styles"
	"github.com/mane-pal/lazylinear/pkg/gui/types"
)

// Filter options for the sidebar (state filters moved to States section)
var filterOptions = []struct {
	key   string
	label string
}{
	{"all", "All"},
	{"my_issues", "My Issues"},
}

// View renders the UI
func (g *Gui) View() string {
	if g.width == 0 {
		return "Loading..."
	}

	if g.showHelp {
		return g.renderHelp()
	}

	// Add margin around the entire UI (top, right, bottom, left)
	marginTop := 2
	marginSide := 1
	availableWidth := g.width - (marginSide * 2)
	availableHeight := g.height - marginTop - 1 // Top margin + bottom margin

	// Calculate dimensions - give more space to sidebar and detail
	sidebarWidth := 18
	detailWidth := availableWidth * 35 / 100 // 35% for detail
	issuesWidth := availableWidth - sidebarWidth - detailWidth - 6 // Account for borders

	contentHeight := availableHeight - 2 // Leave room for status bar

	// Render panels
	sidebar := g.renderSidebar(sidebarWidth, contentHeight)
	var middlePane string
	if g.middlePaneView == ViewProjects {
		middlePane = g.renderProjectsList(issuesWidth, contentHeight)
	} else {
		middlePane = g.renderIssuesList(issuesWidth, contentHeight)
	}
	detail := g.renderDetail(detailWidth, contentHeight)

	// Combine horizontally
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebar,
		middlePane,
		detail,
	)

	// Add status bar (pass available width for correct sizing)
	statusBar := g.renderStatusBar(availableWidth)

	view := lipgloss.JoinVertical(lipgloss.Left, content, statusBar)

	// Apply margin (top, right, bottom, left)
	marginStyle := lipgloss.NewStyle().Padding(marginTop, marginSide, 1, marginSide)
	view = marginStyle.Render(view)

	// Overlay form if open
	if g.formMode != FormNone {
		form := g.renderForm()
		view = g.overlayMenu(view, form)
	}

	// Overlay menu if open
	if g.menuType != MenuNone {
		var menu string
		if g.menuType == MenuConfirm {
			menu = g.renderConfirm()
		} else {
			menu = g.renderMenu()
		}
		view = g.overlayMenu(view, menu)
	}

	// Overlay search input if active
	if g.searchMode {
		search := g.renderSearchBar()
		view = g.overlayBottom(view, search)
	}

	// Overlay command input if active
	if g.commandMode {
		cmd := g.renderCommandLine()
		view = g.overlayBottom(view, cmd)
	}

	return view
}

func (g *Gui) getTeamIssueCount(teamID string) int {
	count := 0
	for _, issue := range g.issues {
		if issue.Team != nil && issue.Team.ID == teamID {
			count++
		}
	}
	return count
}

func (g *Gui) renderSidebar(width, height int) string {
	style := styles.Panel
	if g.focusedPanel == types.TeamsContext {
		style = styles.PanelFocused
	}

	var b strings.Builder

	// Teams section header
	header := styles.PanelTitle.Render("─ Teams ─")
	b.WriteString(header)
	b.WriteString("\n\n")

	if len(g.teams) == 0 {
		b.WriteString(styles.Subtle.Render("  Loading..."))
	} else {
		for i, team := range g.teams {
			// Show issue count for selected team
			countStr := ""
			if i == g.selectedTeam {
				count := len(g.getDisplayIssues())
				countStr = fmt.Sprintf(" (%d)", count)
			}

			isSelected := i == g.selectedTeam && g.sidebarSection == 0
			if isSelected {
				line := fmt.Sprintf("> %s%s", team.Key, countStr)
				b.WriteString(styles.Selected.Render(line))
			} else if i == g.selectedTeam {
				// Selected team but filter section is focused
				b.WriteString(fmt.Sprintf("  %s%s", styles.Title.Render(team.Key), countStr))
			} else {
				b.WriteString(fmt.Sprintf("  %s", team.Key))
			}
			b.WriteString("\n")
		}
	}

	// Filters section
	b.WriteString("\n")
	filterHeader := styles.PanelTitle.Render("─ Filters ─")
	b.WriteString(filterHeader)
	b.WriteString("\n\n")

	for i, filter := range filterOptions {
		isActive := g.activeFilter == filter.key
		isSelected := g.sidebarSection == 1 && g.selectedFilter == i && g.focusedPanel == types.TeamsContext

		prefix := "  "
		if isActive {
			prefix = "● "
		}

		label := filter.label
		if isSelected {
			line := fmt.Sprintf("> %s", label)
			b.WriteString(styles.Selected.Render(line))
		} else if isActive {
			b.WriteString(styles.Title.Render(prefix + label))
		} else {
			b.WriteString(styles.Subtle.Render(prefix + label))
		}
		b.WriteString("\n")
	}

	// States section (multi-select)
	b.WriteString("\n")
	statesHeader := styles.PanelTitle.Render("─ States ─")
	b.WriteString(statesHeader)
	b.WriteString("\n\n")

	if len(g.teamStates) == 0 {
		b.WriteString(styles.Subtle.Render("  Loading..."))
		b.WriteString("\n")
	} else {
		for i, state := range g.teamStates {
			count := g.countIssuesByState(state.ID)
			isActive := g.activeStateFilters[state.ID]
			isSelected := g.sidebarSection == 2 && g.selectedStateIdx == i && g.focusedPanel == types.TeamsContext

			stateStyle := lipgloss.NewStyle().Foreground(styles.StateColor(state.Type))
			icon := styles.StateIcon(state.Type)
			label := fmt.Sprintf("%s %s (%d)", icon, state.Name, count)

			if isSelected {
				line := fmt.Sprintf("> %s", label)
				b.WriteString(styles.Selected.Render(line))
			} else if isActive {
				b.WriteString(stateStyle.Bold(true).Render("● " + label))
			} else {
				b.WriteString(stateStyle.Render("  " + label))
			}
			b.WriteString("\n")
		}
	}

	// Priority section
	b.WriteString("\n")
	prioHeader := styles.PanelTitle.Render("─ Priority ─")
	b.WriteString(prioHeader)
	b.WriteString("\n\n")

	priorityLabels := []struct {
		level int
		label string
		color lipgloss.Color
	}{
		{1, "Urgent", styles.PriorityUrgent},
		{2, "High", styles.PriorityHigh},
		{3, "Medium", styles.PriorityMedium},
		{4, "Low", styles.PriorityLow},
	}

	for i, prio := range priorityLabels {
		// Count issues with this priority
		count := g.countIssuesByPriority(prio.level)
		isActive := g.activePriorityFilter == prio.level
		isSelected := g.sidebarSection == 3 && g.selectedPriority == i && g.focusedPanel == types.TeamsContext

		prioStyle := lipgloss.NewStyle().Foreground(prio.color)
		label := fmt.Sprintf("%s (%d)", prio.label, count)

		if isSelected {
			line := fmt.Sprintf("> %s", label)
			b.WriteString(styles.Selected.Render(line))
		} else if isActive {
			b.WriteString(prioStyle.Bold(true).Render("● " + label))
		} else {
			b.WriteString(prioStyle.Render("  " + label))
		}
		b.WriteString("\n")
	}

	return style.Width(width).Height(height).Render(b.String())
}

func (g *Gui) countIssuesByPriority(priority int) int {
	count := 0
	for _, issue := range g.issues {
		if issue.Priority == priority {
			count++
		}
	}
	return count
}

func (g *Gui) countIssuesByState(stateID string) int {
	count := 0
	for _, issue := range g.issues {
		if issue.State != nil && issue.State.ID == stateID {
			count++
		}
	}
	return count
}

func (g *Gui) renderIssuesList(width, height int) string {
	style := styles.Panel
	if g.focusedPanel == types.IssuesContext {
		style = styles.PanelFocused
	}

	var b strings.Builder

	// Header with count and search indicator
	issues := g.getDisplayIssues()
	issueCount := len(issues)
	header := fmt.Sprintf("─ Issues (%d) ─", issueCount)
	if g.searchQuery != "" {
		header = fmt.Sprintf("─ Issues (%d) [/%s] ─", issueCount, g.searchQuery)
	}
	b.WriteString(styles.PanelTitle.Render(header))
	b.WriteString("\n\n")

	if g.loading {
		b.WriteString(styles.Subtle.Render("  Loading..."))
	} else if g.err != nil {
		b.WriteString(styles.ErrorStyle.Render("  Error: " + g.err.Error()))
	} else if len(issues) == 0 {
		if g.searchQuery != "" {
			b.WriteString(styles.Subtle.Render("  No matching issues"))
		} else {
			b.WriteString(styles.Subtle.Render("  No issues"))
		}
	} else {
		// Calculate visible range for scrolling
		visibleLines := height - 6
		if visibleLines < 1 {
			visibleLines = 1
		}
		start := 0
		if g.selectedIssue >= visibleLines {
			start = g.selectedIssue - visibleLines + 1
		}
		end := min(len(issues), start+visibleLines)

		// Calculate max width for title (leave room for icon + identifier + padding)
		identifierWidth := 12 // e.g., "DEVOPS-1234"
		titleWidth := width - identifierWidth - 8
		if titleWidth < 10 {
			titleWidth = 10
		}

		for i := start; i < end; i++ {
			issue := issues[i]

			// State icon with color
			stateIcon := "○"
			stateColor := styles.Secondary
			if issue.State != nil {
				stateIcon = styles.StateIcon(issue.State.Type)
				stateColor = styles.StateColor(issue.State.Type)
			}
			iconStyled := lipgloss.NewStyle().Foreground(stateColor).Render(stateIcon)

			// Format: icon identifier title
			title := truncate(issue.Title, titleWidth)

			// Highlight search matches in title
			if g.searchQuery != "" {
				title = g.highlightMatch(title, g.searchQuery)
			}

			if i == g.selectedIssue {
				// Selected line - full highlight
				line := fmt.Sprintf("> %s %s %s", iconStyled, issue.Identifier, title)
				b.WriteString(styles.Selected.Render(line))
			} else {
				line := fmt.Sprintf("  %s %s %s", iconStyled, styles.Subtle.Render(issue.Identifier), title)
				b.WriteString(line)
			}
			b.WriteString("\n")
		}

		// Scroll indicator
		if len(issues) > visibleLines {
			scrollInfo := fmt.Sprintf("\n  %d/%d", g.selectedIssue+1, len(issues))
			b.WriteString(styles.Subtle.Render(scrollInfo))
		}
	}

	return style.Width(width).Height(height).Render(b.String())
}

func (g *Gui) renderProjectsList(width, height int) string {
	style := styles.Panel
	if g.focusedPanel == types.IssuesContext {
		style = styles.PanelFocused
	}

	var b strings.Builder

	// Header with count
	header := fmt.Sprintf("─ Projects (%d) ─", len(g.projects))
	b.WriteString(styles.PanelTitle.Render(header))
	b.WriteString("\n\n")

	if g.loading {
		b.WriteString(styles.Subtle.Render("  Loading..."))
	} else if g.err != nil {
		b.WriteString(styles.ErrorStyle.Render("  Error: " + g.err.Error()))
	} else if len(g.projects) == 0 {
		b.WriteString(styles.Subtle.Render("  No projects"))
	} else {
		// Calculate visible range for scrolling
		visibleLines := height - 6
		if visibleLines < 1 {
			visibleLines = 1
		}
		start := 0
		if g.selectedProject >= visibleLines {
			start = g.selectedProject - visibleLines + 1
		}
		end := min(len(g.projects), start+visibleLines)

		// Calculate max width for name
		nameWidth := width - 20
		if nameWidth < 10 {
			nameWidth = 10
		}

		for i := start; i < end; i++ {
			project := g.projects[i]

			// Project state icon
			stateIcon := "○"
			stateColor := styles.Secondary
			switch project.State {
			case "started":
				stateIcon = "●"
				stateColor = styles.Primary
			case "planned":
				stateIcon = "◐"
				stateColor = styles.Warning
			case "completed":
				stateIcon = "✓"
				stateColor = styles.Success
			case "canceled":
				stateIcon = "✗"
				stateColor = styles.Error
			case "paused":
				stateIcon = "◑"
				stateColor = styles.Secondary
			}
			iconStyled := lipgloss.NewStyle().Foreground(stateColor).Render(stateIcon)

			// Progress indicator
			progress := fmt.Sprintf("%3.0f%%", project.Progress*100)

			// Format: icon name progress (issue count)
			name := truncate(project.Name, nameWidth)
			issueInfo := fmt.Sprintf("(%d)", project.IssueCount)

			if i == g.selectedProject {
				line := fmt.Sprintf("> %s %s %s %s", iconStyled, name, progress, issueInfo)
				b.WriteString(styles.Selected.Render(line))
			} else {
				line := fmt.Sprintf("  %s %s %s %s", iconStyled, name, styles.Subtle.Render(progress), styles.Subtle.Render(issueInfo))
				b.WriteString(line)
			}
			b.WriteString("\n")
		}

		// Scroll indicator
		if len(g.projects) > visibleLines {
			scrollInfo := fmt.Sprintf("\n  %d/%d", g.selectedProject+1, len(g.projects))
			b.WriteString(styles.Subtle.Render(scrollInfo))
		}
	}

	return style.Width(width).Height(height).Render(b.String())
}

func (g *Gui) renderDetail(width, height int) string {
	style := styles.Panel
	if g.focusedPanel == types.DetailContext {
		style = styles.PanelFocused
	}

	var b strings.Builder

	// Header
	b.WriteString(styles.PanelTitle.Render("─ Detail ─"))
	b.WriteString("\n\n")

	issue := g.getSelectedIssue()
	if issue == nil {
		b.WriteString(styles.Subtle.Render("  Select an issue"))
		return style.Width(width).Height(height).Render(b.String())
	}

	// Use detailed issue if available (has comments)
	displayIssue := issue
	if g.detailedIssue != nil && g.detailedIssue.ID == issue.ID {
		displayIssue = g.detailedIssue
	}

	// Identifier (large, colored)
	b.WriteString(styles.Title.Render(displayIssue.Identifier))
	b.WriteString("\n\n")

	// Title (wrapped)
	titleWidth := width - 4
	if titleWidth < 20 {
		titleWidth = 20
	}
	wrappedTitle := wrapText(displayIssue.Title, titleWidth)
	b.WriteString(wrappedTitle)
	b.WriteString("\n\n")

	// Metadata section
	b.WriteString(styles.Subtle.Render("─────────────────────────"))
	b.WriteString("\n\n")

	if displayIssue.State != nil {
		stateStyle := lipgloss.NewStyle().Foreground(styles.StateColor(displayIssue.State.Type))
		icon := styles.StateIcon(displayIssue.State.Type)
		b.WriteString(fmt.Sprintf("State:    %s %s\n", stateStyle.Render(icon), stateStyle.Render(displayIssue.State.Name)))
	}

	if displayIssue.Priority > 0 {
		prioStyle := lipgloss.NewStyle().Foreground(styles.PriorityColor(displayIssue.Priority))
		b.WriteString(fmt.Sprintf("Priority: %s\n", prioStyle.Render(styles.PriorityLabel(displayIssue.Priority))))
	} else {
		b.WriteString(fmt.Sprintf("Priority: %s\n", styles.Subtle.Render("None")))
	}

	if displayIssue.Assignee != nil {
		b.WriteString(fmt.Sprintf("Assignee: @%s\n", displayIssue.Assignee.DisplayName))
	} else {
		b.WriteString(fmt.Sprintf("Assignee: %s\n", styles.Subtle.Render("Unassigned")))
	}

	if labels := displayIssue.GetLabels(); len(labels) > 0 {
		labelNames := make([]string, len(labels))
		for i, l := range labels {
			labelNames[i] = l.Name
		}
		b.WriteString(fmt.Sprintf("Labels:   %s\n", strings.Join(labelNames, ", ")))
	}

	// Description
	if displayIssue.Description != "" {
		b.WriteString("\n")
		b.WriteString(styles.Subtle.Render("─────────────────────────"))
		b.WriteString("\n\n")
		// Wrap description to fit panel
		desc := wrapText(displayIssue.Description, titleWidth)
		// Limit description length for now
		lines := strings.Split(desc, "\n")
		if len(lines) > 6 {
			desc = strings.Join(lines[:6], "\n") + "..."
		}
		b.WriteString(desc)
	}

	// Comments section
	comments := displayIssue.GetComments()
	if len(comments) > 0 || g.focusedPanel == types.DetailContext {
		b.WriteString("\n\n")
		b.WriteString(styles.Subtle.Render("─ Comments ─"))
		if g.focusedPanel == types.DetailContext {
			b.WriteString(styles.Subtle.Render(" (c: add, j/k: navigate)"))
		}
		b.WriteString("\n\n")

		if len(comments) == 0 {
			b.WriteString(styles.Subtle.Render("  No comments yet"))
		} else {
			for i, comment := range comments {
				isSelected := g.focusedPanel == types.DetailContext && i == g.selectedComment

				// Author and time
				author := "Unknown"
				if comment.User != nil {
					author = comment.User.DisplayName
				}
				timeAgo := formatTimeAgo(comment.CreatedAt)

				header := fmt.Sprintf("@%s · %s", author, timeAgo)
				if isSelected {
					b.WriteString(styles.Selected.Render("> " + header))
				} else {
					b.WriteString(styles.Subtle.Render("  " + header))
				}
				b.WriteString("\n")

				// Comment body (truncated)
				body := wrapText(comment.Body, titleWidth-2)
				lines := strings.Split(body, "\n")
				if len(lines) > 3 {
					body = strings.Join(lines[:3], "\n") + "..."
				}
				for _, line := range strings.Split(body, "\n") {
					b.WriteString("  " + line + "\n")
				}
				b.WriteString("\n")
			}
		}
	}

	return style.Width(width).Height(height).Render(b.String())
}

func (g *Gui) renderStatusBar(width int) string {
	// Left side: logo and context
	left := styles.Logo.Render("▲ LazyLinear")
	if g.currentUser != nil {
		left += styles.Subtle.Render(" │ " + g.currentUser.DisplayName)
	}

	// Show status message if any
	if g.statusMsg != "" {
		left += styles.Subtle.Render(" │ ") + styles.WarningStyle.Render(g.statusMsg)
	}

	// Middle: last synced
	middle := ""
	if !g.lastSynced.IsZero() {
		middle = styles.Subtle.Render("synced " + formatTimeAgo(g.lastSynced))
	}

	// Right side: help hints
	help := []string{
		styles.HelpKey.Render("/") + styles.HelpDesc.Render(" search"),
		styles.HelpKey.Render(":") + styles.HelpDesc.Render(" cmd"),
		styles.HelpKey.Render("?") + styles.HelpDesc.Render(" help"),
	}
	right := strings.Join(help, "  ")

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	middleWidth := lipgloss.Width(middle)
	rightWidth := lipgloss.Width(right)
	totalContent := leftWidth + middleWidth + rightWidth

	spacing1 := (width - totalContent) / 2
	spacing2 := width - totalContent - spacing1
	if spacing1 < 1 {
		spacing1 = 1
	}
	if spacing2 < 1 {
		spacing2 = 1
	}

	bar := left + strings.Repeat(" ", spacing1) + middle + strings.Repeat(" ", spacing2) + right

	return styles.StatusBar.Width(width).Render(bar)
}

func (g *Gui) renderHelp() string {
	// Build help with proper column alignment
	var b strings.Builder

	title := styles.Title.Render("LazyLinear Help")
	separator := styles.Subtle.Render("═══════════════════════════════════════════════════════")

	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(separator)
	b.WriteString("\n\n")

	// Helper to format a row with two columns
	// Left: key (10 chars) + desc (16 chars) = 26 chars
	// Right: key (8 chars) + desc (18 chars) = 26 chars
	row := func(lKey, lDesc, rKey, rDesc string) string {
		left := fmt.Sprintf("%-10s %-16s", lKey, lDesc)
		if rKey != "" {
			right := fmt.Sprintf("%-8s %s", rKey, rDesc)
			return left + "   " + right + "\n"
		}
		return left + "\n"
	}

	header := func(left, right string) string {
		l := styles.Title.Render(left)
		lPad := 27 - lipgloss.Width(l)
		if lPad < 0 {
			lPad = 0
		}
		if right != "" {
			r := styles.Title.Render(right)
			return l + strings.Repeat(" ", lPad) + "   " + r + "\n"
		}
		return l + "\n"
	}

	underline := func() string {
		return styles.Subtle.Render("─────────────────────────") + "   " + styles.Subtle.Render("─────────────────────────") + "\n"
	}

	// NAVIGATION | ACTIONS
	b.WriteString(header("NAVIGATION", "ACTIONS"))
	b.WriteString(underline())
	b.WriteString(row("j/k ↑/↓", "Move up/down", "n", "New issue"))
	b.WriteString(row("g/G", "Top/Bottom", "e", "Edit issue/comment"))
	b.WriteString(row("Tab", "Next section", "d", "Archive/delete"))
	b.WriteString(row("Shift+Tab", "Prev section", "s", "Change state"))
	b.WriteString(row("Space", "Toggle filter", "a", "Assign user"))
	b.WriteString(row("Enter/l", "Select/Forward", "p", "Set priority"))
	b.WriteString(row("Esc/h", "Back", "o", "Open in browser"))
	b.WriteString(row("", "", "y", "Yank URL"))
	b.WriteString(row("", "", "Y", "Yank issue ID"))
	b.WriteString(row("", "", "r", "Refresh"))
	b.WriteString("\n")

	// GLOBAL | COMMENTS
	b.WriteString(header("GLOBAL", "COMMENTS (Detail)"))
	b.WriteString(underline())
	b.WriteString(row("/", "Search", "c", "Add comment"))
	b.WriteString(row(":", "Command mode", "j/k", "Navigate comments"))
	b.WriteString(row("?", "Toggle help", "e", "Edit (own)"))
	b.WriteString(row("m", "My issues", "d", "Delete (own)"))
	b.WriteString(row("x", "Clear filters", "", ""))
	b.WriteString(row("P", "Issues/Projects", "", ""))
	b.WriteString(row("q Ctrl+c", "Quit", "", ""))
	b.WriteString("\n")

	// VIM COMMANDS
	b.WriteString(header("VIM COMMANDS (:)", ""))
	b.WriteString(styles.Subtle.Render("─────────────────────────") + "\n")
	b.WriteString(row(":q", "Quit", "", ""))
	b.WriteString(row(":w", "Refresh", "", ""))
	b.WriteString(row(":help", "Show help", "", ""))
	b.WriteString(row(":clear", "Clear search", "", ""))
	b.WriteString(row(":ENG-123", "Jump to issue", "", ""))
	b.WriteString("\n")

	b.WriteString(styles.Subtle.Render("           Press any key to close"))

	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2)

	return lipgloss.Place(
		g.width, g.height,
		lipgloss.Center, lipgloss.Center,
		helpStyle.Render(b.String()),
	)
}

func (g *Gui) renderMenu() string {
	var b strings.Builder

	// Title
	b.WriteString(styles.Title.Render(g.menuTitle))
	b.WriteString("\n\n")

	// Items
	for i, item := range g.menuItems {
		itemStyle := lipgloss.NewStyle().Foreground(item.Color)

		// For multi-select menus (labels), show checkbox
		prefix := "  "
		if g.menuType == MenuLabels {
			if item.Selected {
				prefix = "[✓] "
			} else {
				prefix = "[ ] "
			}
		}

		if i == g.menuSelected {
			line := fmt.Sprintf("> %s%s", prefix[2:], item.Label) // Remove first 2 chars as we add >
			b.WriteString(styles.Selected.Render(line))
		} else {
			line := fmt.Sprintf("%s%s", prefix, itemStyle.Render(item.Label))
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	// Hint - different for multi-select
	b.WriteString("\n")
	if g.menuType == MenuLabels {
		b.WriteString(styles.Subtle.Render("j/k: select  space: toggle  enter: save  esc: cancel"))
	} else {
		b.WriteString(styles.Subtle.Render("j/k: select  enter: confirm  esc: cancel"))
	}

	menuStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Width(40)

	return menuStyle.Render(b.String())
}

func (g *Gui) overlayMenu(base, menu string) string {
	return lipgloss.Place(
		g.width, g.height,
		lipgloss.Center, lipgloss.Center,
		menu,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#000000")),
	)
}

func (g *Gui) overlayBottom(base, overlay string) string {
	// Split base into lines
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	// Replace bottom lines with overlay
	startLine := len(baseLines) - len(overlayLines) - 1
	if startLine < 0 {
		startLine = 0
	}

	for i, line := range overlayLines {
		if startLine+i < len(baseLines) {
			baseLines[startLine+i] = line
		}
	}

	return strings.Join(baseLines, "\n")
}

func (g *Gui) overlayCenter(base, overlay string) string {
	// Simple overlay - just return the overlay for now
	// In a real implementation, you'd center the overlay on the base
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	baseHeight := len(baseLines)
	overlayHeight := len(overlayLines)
	baseWidth := lipgloss.Width(base)
	overlayWidth := lipgloss.Width(overlay)

	// Calculate position to center
	startY := (baseHeight - overlayHeight) / 2
	startX := (baseWidth - overlayWidth) / 2

	if startY < 0 {
		startY = 0
	}
	if startX < 0 {
		startX = 0
	}

	// Build result
	var result strings.Builder
	for i, line := range baseLines {
		if i >= startY && i < startY+overlayHeight {
			overlayLineIdx := i - startY
			if overlayLineIdx < len(overlayLines) {
				// Overlay this line
				overlayLine := overlayLines[overlayLineIdx]
				lineWidth := lipgloss.Width(line)
				overlayLineWidth := lipgloss.Width(overlayLine)

				if startX > 0 && startX < lineWidth {
					// Insert overlay into line
					result.WriteString(line[:min(startX, len(line))])
				}
				result.WriteString(overlayLine)
				endX := startX + overlayLineWidth
				if endX < lineWidth && endX < len(line) {
					// Add rest of base line (but this gets tricky with ANSI codes)
				}
			} else {
				result.WriteString(line)
			}
		} else {
			result.WriteString(line)
		}
		if i < len(baseLines)-1 {
			result.WriteString("\n")
		}
	}

	// Simpler approach: just return the overlay on top
	return overlay
}

func (g *Gui) renderSearchBar() string {
	searchStyle := lipgloss.NewStyle().
		Background(styles.BgDark).
		Foreground(styles.FgLight).
		Padding(0, 1).
		Width(g.width)

	prompt := lipgloss.NewStyle().Foreground(styles.Primary).Render("/")
	input := g.searchInput.View()

	return searchStyle.Render(prompt + input)
}

func (g *Gui) renderCommandLine() string {
	cmdStyle := lipgloss.NewStyle().
		Background(styles.BgDark).
		Foreground(styles.FgLight).
		Padding(0, 1).
		Width(g.width)

	prompt := lipgloss.NewStyle().Foreground(styles.Primary).Render(":")
	input := g.commandInput.View()

	return cmdStyle.Render(prompt + input)
}

func (g *Gui) renderForm() string {
	var b strings.Builder

	// Handle comment forms
	if g.formMode == FormComment || g.formMode == FormEditComment {
		if g.formMode == FormComment {
			b.WriteString(styles.Title.Render("Add Comment"))
		} else {
			b.WriteString(styles.Title.Render("Edit Comment"))
		}
		b.WriteString("\n\n")
		b.WriteString(g.commentBody.View())
		b.WriteString("\n\n")
		b.WriteString(styles.Subtle.Render("Ctrl+S: save  Esc: cancel"))

		formStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.Primary).
			Padding(1, 2).
			Width(60)

		return formStyle.Render(b.String())
	}

	// Dynamic form sizing (70% of width, max 100)
	formWidth := (g.width * 70) / 100
	if formWidth > 100 {
		formWidth = 100
	}
	if formWidth < 70 {
		formWidth = 70
	}
	colWidth := (formWidth - 10) / 2

	// Issue form title
	if g.formMode == FormCreate {
		b.WriteString(styles.Title.Render("Create Issue"))
	} else {
		b.WriteString(styles.Title.Render("Edit Issue"))
	}
	b.WriteString("\n\n")

	// Helper for field labels
	fieldLabel := func(label string, field FormField) string {
		if g.formField == field {
			return lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render("> " + label)
		}
		return "  " + label
	}

	// Helper for selector field value display
	selectorValue := func(value string, focused bool) string {
		if focused {
			return styles.Selected.Render(" " + value + " ⏎")
		}
		return styles.Subtle.Render(value)
	}

	// Title field
	b.WriteString(fieldLabel("Title", FieldTitle) + "\n")
	b.WriteString("  " + g.formTitle.View())
	b.WriteString("\n\n")

	// Description field
	b.WriteString(fieldLabel("Description", FieldDescription) + "\n")
	b.WriteString("  " + g.formDescription.View())
	b.WriteString("\n\n")

	// Two-column layout
	col1 := strings.Builder{}
	col2 := strings.Builder{}

	// Team (col1) - press Enter to open popup
	col1.WriteString(fieldLabel("Team", FieldTeam) + "\n")
	teamValue := "(none)"
	if g.formTeam >= 0 && g.formTeam < len(g.teams) {
		teamValue = g.teams[g.formTeam].Name
	}
	col1.WriteString("  " + selectorValue(teamValue, g.formField == FieldTeam) + "\n\n")

	// Assignee (col2) - press Enter to open popup
	col2.WriteString(fieldLabel("Assignee", FieldAssignee) + "\n")
	assigneeValue := "Unassigned"
	if g.formAssignee >= 0 && g.formAssignee < len(g.teamMembers) {
		assigneeValue = g.teamMembers[g.formAssignee].DisplayName
	}
	col2.WriteString("  " + selectorValue(assigneeValue, g.formField == FieldAssignee) + "\n\n")

	// State (col1) - inline selector with j/k
	col1.WriteString(fieldLabel("State", FieldState) + "\n")
	if g.formField == FieldState {
		states := []string{"Default"}
		for _, s := range g.teamStates {
			states = append(states, s.Name)
		}
		for i, s := range states {
			actualIdx := i - 1
			if actualIdx == g.formState {
				if actualIdx >= 0 && actualIdx < len(g.teamStates) {
					stateStyle := lipgloss.NewStyle().Background(styles.StateColor(g.teamStates[actualIdx].Type)).Foreground(lipgloss.Color("#000"))
					col1.WriteString("  " + stateStyle.Render(" "+s+" ") + "\n")
				} else {
					col1.WriteString("  " + styles.Selected.Render(" "+s+" ") + "\n")
				}
			} else {
				col1.WriteString("  " + styles.Subtle.Render("  "+s) + "\n")
			}
		}
	} else {
		stateValue := "Default"
		if g.formState >= 0 && g.formState < len(g.teamStates) {
			stateValue = g.teamStates[g.formState].Name
		}
		col1.WriteString("  " + styles.Subtle.Render(stateValue) + "\n")
	}
	col1.WriteString("\n")

	// Priority (col2) - inline selector with j/k
	col2.WriteString(fieldLabel("Priority", FieldPriority) + "\n")
	priorities := []string{"None", "Urgent", "High", "Medium", "Low"}
	if g.formField == FieldPriority {
		for i, p := range priorities {
			if i == g.formPriority {
				prioStyle := lipgloss.NewStyle().Foreground(styles.PriorityColor(i)).Bold(true)
				col2.WriteString("  " + prioStyle.Render("● "+p) + "\n")
			} else {
				col2.WriteString("  " + styles.Subtle.Render("○ "+p) + "\n")
			}
		}
	} else {
		prioStyle := lipgloss.NewStyle().Foreground(styles.PriorityColor(g.formPriority))
		col2.WriteString("  " + prioStyle.Render(priorities[g.formPriority]) + "\n")
	}
	col2.WriteString("\n")

	// Project (col1) - press Enter to open popup
	col1.WriteString(fieldLabel("Project", FieldProject) + "\n")
	projectValue := "None"
	if g.formProject >= 0 && g.formProject < len(g.projects) {
		projectValue = g.projects[g.formProject].Name
	}
	col1.WriteString("  " + selectorValue(projectValue, g.formField == FieldProject) + "\n")

	// Combine columns
	leftCol := lipgloss.NewStyle().Width(colWidth).Render(col1.String())
	rightCol := lipgloss.NewStyle().Width(colWidth).Render(col2.String())
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol))
	b.WriteString("\n")

	// Hints
	hints := "Tab: next  "
	if g.formField == FieldTeam || g.formField == FieldAssignee || g.formField == FieldProject {
		hints += "Enter: select  "
	} else if g.formField == FieldState || g.formField == FieldPriority {
		hints += "j/k: change  "
	}
	hints += "Ctrl+S: save  Esc: cancel"
	b.WriteString(styles.Subtle.Render(hints))

	// Calculate dynamic height (60% of window, min 30)
	formHeight := (g.height * 60) / 100
	if formHeight < 30 {
		formHeight = 30
	}

	formStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(2, 3).
		Width(formWidth).
		Height(formHeight)

	formContent := formStyle.Render(b.String())

	// If list popup is open, overlay it
	if g.formListOpen {
		popup := g.renderFormListPopup()
		return g.overlayCenter(formContent, popup)
	}

	return formContent
}

func (g *Gui) renderFormListPopup() string {
	var b strings.Builder

	// Title based on field
	var title string
	var items []string
	var currentSelection int

	switch g.formListField {
	case FieldTeam:
		title = "Select Team"
		for _, t := range g.teams {
			items = append(items, t.Name)
		}
		currentSelection = g.formTeam
	case FieldAssignee:
		title = "Select Assignee"
		items = append(items, "Unassigned")
		for _, m := range g.teamMembers {
			items = append(items, m.DisplayName)
		}
		currentSelection = g.formAssignee + 1
	case FieldProject:
		title = "Select Project"
		items = append(items, "None")
		for _, p := range g.projects {
			items = append(items, p.Name)
		}
		currentSelection = g.formProject + 1
	}

	b.WriteString(styles.Title.Render(title))
	b.WriteString("\n\n")

	// Search input
	b.WriteString(g.formListSearch.View())
	b.WriteString("\n\n")

	// Filter items
	query := strings.ToLower(g.formListSearch.Value())
	var filtered []struct {
		idx  int
		name string
	}
	for i, item := range items {
		if query == "" || strings.Contains(strings.ToLower(item), query) {
			filtered = append(filtered, struct {
				idx  int
				name string
			}{i, item})
		}
	}

	// Show items with scrolling
	maxShow := 10
	if len(filtered) < maxShow {
		maxShow = len(filtered)
	}

	startIdx := 0
	if g.formListSelected >= maxShow {
		startIdx = g.formListSelected - maxShow + 1
	}
	if startIdx > len(filtered)-maxShow {
		startIdx = len(filtered) - maxShow
	}
	if startIdx < 0 {
		startIdx = 0
	}

	endIdx := startIdx + maxShow
	if endIdx > len(filtered) {
		endIdx = len(filtered)
	}

	// Show scroll indicator if needed
	if startIdx > 0 {
		b.WriteString(styles.Subtle.Render("  ↑ more above"))
		b.WriteString("\n")
	}

	for i := startIdx; i < endIdx; i++ {
		item := filtered[i]
		prefix := "  "
		if item.idx == currentSelection {
			prefix = "✓ "
		}
		if i == g.formListSelected {
			b.WriteString(styles.Selected.Render(" " + prefix + item.name + " "))
		} else {
			b.WriteString("  " + prefix + styles.Subtle.Render(item.name))
		}
		b.WriteString("\n")
	}

	if endIdx < len(filtered) {
		b.WriteString(styles.Subtle.Render("  ↓ more below"))
		b.WriteString("\n")
	}

	if len(filtered) == 0 {
		b.WriteString(styles.Subtle.Render("  No matches"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.Subtle.Render("Type to filter • j/k: navigate • Enter: select • Esc: cancel"))

	popupWidth := 50
	popupStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Width(popupWidth)

	return popupStyle.Render(b.String())
}

func (g *Gui) renderConfirm() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render(g.confirmTitle))
	b.WriteString("\n\n")
	b.WriteString(g.confirmMessage)
	b.WriteString("\n\n")
	b.WriteString(styles.Subtle.Render("y: confirm  n/esc: cancel"))

	confirmStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Warning).
		Padding(1, 2).
		Width(50)

	return confirmStyle.Render(b.String())
}
