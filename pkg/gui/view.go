package gui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/mane-pal/lazylinear/pkg/gui/styles"
	"github.com/mane-pal/lazylinear/pkg/gui/types"
)

func (g *Gui) View() string {
	if g.width == 0 {
		return "Loading..."
	}

	if g.State.ShowHelp {
		return g.renderHelp()
	}

	marginTop := 2
	marginSide := 1
	availableWidth := g.width - (marginSide * 2)
	availableHeight := g.height - marginTop - 1

	sidebarWidth := 18
	detailWidth := availableWidth * 35 / 100
	issuesWidth := availableWidth - sidebarWidth - detailWidth - 6

	contentHeight := availableHeight - 2

	sidebar := g.SidebarCtx.View(sidebarWidth, contentHeight)
	middlePane := g.IssuesCtx.View(issuesWidth, contentHeight)
	detail := g.DetailCtx.View(detailWidth, contentHeight)

	content := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, middlePane, detail)
	statusBar := g.renderStatusBar(availableWidth)
	view := lipgloss.JoinVertical(lipgloss.Left, content, statusBar)

	marginStyle := lipgloss.NewStyle().Padding(marginTop, marginSide, 1, marginSide)
	view = marginStyle.Render(view)

	if g.formMode != types.FormNone {
		form := g.renderForm()
		view = g.overlayMenu(view, form)
	}

	if g.MenuPopup.IsOpen() {
		var menu string
		if g.MenuPopup.MenuType == types.MenuConfirm {
			menu = g.MenuPopup.RenderConfirm()
		} else {
			menu = g.MenuPopup.RenderMenu()
		}
		view = g.overlayMenu(view, menu)
	}

	if g.SearchPopup.SearchMode {
		search := g.SearchPopup.RenderSearchBar(g.width)
		view = g.overlayBottom(view, search)
	}

	if g.SearchPopup.CommandMode {
		cmd := g.SearchPopup.RenderCommandLine(g.width)
		view = g.overlayBottom(view, cmd)
	}

	return view
}

func (g *Gui) renderStatusBar(width int) string {
	left := styles.Logo.Render("▲ LazyLinear")
	if g.State.CurrentUser != nil {
		left += styles.Subtle.Render(" │ " + g.State.CurrentUser.DisplayName)
	}

	if g.State.StatusMsg != "" {
		left += styles.Subtle.Render(" │ ") + styles.WarningStyle.Render(g.State.StatusMsg)
	}

	middle := ""
	if !g.State.LastSynced.IsZero() {
		middle = styles.Subtle.Render("synced " + formatTimeAgo(g.State.LastSynced))
	}

	help := []string{
		styles.HelpKey.Render("/") + styles.HelpDesc.Render(" search"),
		styles.HelpKey.Render(":") + styles.HelpDesc.Render(" cmd"),
		styles.HelpKey.Render("?") + styles.HelpDesc.Render(" help"),
	}
	right := strings.Join(help, "  ")

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
	var b strings.Builder

	title := styles.Title.Render("LazyLinear Help")
	separator := styles.Subtle.Render("═══════════════════════════════════════════════════════")

	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(separator)
	b.WriteString("\n\n")

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
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

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
	return overlay
}

func (g *Gui) renderForm() string {
	var b strings.Builder

	if g.formMode == types.FormComment || g.formMode == types.FormEditComment {
		if g.formMode == types.FormComment {
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

	formWidth := (g.width * 70) / 100
	if formWidth > 100 {
		formWidth = 100
	}
	if formWidth < 70 {
		formWidth = 70
	}
	colWidth := (formWidth - 10) / 2

	if g.formMode == types.FormCreate {
		b.WriteString(styles.Title.Render("Create Issue"))
	} else {
		b.WriteString(styles.Title.Render("Edit Issue"))
	}
	b.WriteString("\n\n")

	fieldLabel := func(label string, field types.FormField) string {
		if g.formField == field {
			return lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render("> " + label)
		}
		return "  " + label
	}

	selectorValue := func(value string, focused bool) string {
		if focused {
			return styles.Selected.Render(" " + value + " ⏎")
		}
		return styles.Subtle.Render(value)
	}

	b.WriteString(fieldLabel("Title", types.FieldTitle) + "\n")
	b.WriteString("  " + g.formTitle.View())
	b.WriteString("\n\n")

	b.WriteString(fieldLabel("Description", types.FieldDescription) + "\n")
	b.WriteString("  " + g.formDescription.View())
	b.WriteString("\n\n")

	col1 := strings.Builder{}
	col2 := strings.Builder{}

	col1.WriteString(fieldLabel("Team", types.FieldTeam) + "\n")
	teamValue := "(none)"
	if g.formTeam >= 0 && g.formTeam < len(g.State.Teams) {
		teamValue = g.State.Teams[g.formTeam].Name
	}
	col1.WriteString("  " + selectorValue(teamValue, g.formField == types.FieldTeam) + "\n\n")

	col2.WriteString(fieldLabel("Assignee", types.FieldAssignee) + "\n")
	assigneeValue := "Unassigned"
	if g.formAssignee >= 0 && g.formAssignee < len(g.State.TeamMembers) {
		assigneeValue = g.State.TeamMembers[g.formAssignee].DisplayName
	}
	col2.WriteString("  " + selectorValue(assigneeValue, g.formField == types.FieldAssignee) + "\n\n")

	col1.WriteString(fieldLabel("State", types.FieldState) + "\n")
	if g.formField == types.FieldState {
		states := []string{"Default"}
		for _, s := range g.State.TeamStates {
			states = append(states, s.Name)
		}
		for i, s := range states {
			actualIdx := i - 1
			if actualIdx == g.formState {
				if actualIdx >= 0 && actualIdx < len(g.State.TeamStates) {
					stateStyle := lipgloss.NewStyle().Background(styles.StateColor(g.State.TeamStates[actualIdx].Type)).Foreground(lipgloss.Color("#000"))
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
		if g.formState >= 0 && g.formState < len(g.State.TeamStates) {
			stateValue = g.State.TeamStates[g.formState].Name
		}
		col1.WriteString("  " + styles.Subtle.Render(stateValue) + "\n")
	}
	col1.WriteString("\n")

	col2.WriteString(fieldLabel("Priority", types.FieldPriority) + "\n")
	priorities := []string{"None", "Urgent", "High", "Medium", "Low"}
	if g.formField == types.FieldPriority {
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

	col1.WriteString(fieldLabel("Project", types.FieldProject) + "\n")
	projectValue := "None"
	if g.formProject >= 0 && g.formProject < len(g.State.Projects) {
		projectValue = g.State.Projects[g.formProject].Name
	}
	col1.WriteString("  " + selectorValue(projectValue, g.formField == types.FieldProject) + "\n")

	leftCol := lipgloss.NewStyle().Width(colWidth).Render(col1.String())
	rightCol := lipgloss.NewStyle().Width(colWidth).Render(col2.String())
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol))
	b.WriteString("\n")

	hints := "Tab: next  "
	if g.formField == types.FieldTeam || g.formField == types.FieldAssignee || g.formField == types.FieldProject {
		hints += "Enter: select  "
	} else if g.formField == types.FieldState || g.formField == types.FieldPriority {
		hints += "j/k: change  "
	}
	hints += "Ctrl+S: save  Esc: cancel"
	b.WriteString(styles.Subtle.Render(hints))

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

	if g.formListOpen {
		popup := g.renderFormListPopup()
		return g.overlayCenter(formContent, popup)
	}

	return formContent
}

func (g *Gui) renderFormListPopup() string {
	var b strings.Builder

	var title string
	var items []string
	var currentSelection int

	switch g.formListField {
	case types.FieldTeam:
		title = "Select Team"
		for _, t := range g.State.Teams {
			items = append(items, t.Name)
		}
		currentSelection = g.formTeam
	case types.FieldAssignee:
		title = "Select Assignee"
		items = append(items, "Unassigned")
		for _, m := range g.State.TeamMembers {
			items = append(items, m.DisplayName)
		}
		currentSelection = g.formAssignee + 1
	case types.FieldProject:
		title = "Select Project"
		items = append(items, "None")
		for _, p := range g.State.Projects {
			items = append(items, p.Name)
		}
		currentSelection = g.formProject + 1
	}

	b.WriteString(styles.Title.Render(title))
	b.WriteString("\n\n")

	b.WriteString(g.formListSearch.View())
	b.WriteString("\n\n")

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

