package gui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mane-pal/lazylinear/pkg/gui/styles"
	"github.com/mane-pal/lazylinear/pkg/gui/types"
	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

// Search functions

func (g *Gui) openSearch() {
	ti := textinput.New()
	ti.Placeholder = "Search issues..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40

	g.searchMode = true
	g.searchInput = ti
}

func (g *Gui) closeSearch() {
	g.searchMode = false
}

func (g *Gui) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		g.closeSearch()
		return g, nil
	case "enter":
		g.applySearch()
		return g, nil
	}

	// Update search input
	var cmd tea.Cmd
	g.searchInput, cmd = g.searchInput.Update(msg)

	// Live search as you type
	g.applySearch()

	return g, cmd
}

func (g *Gui) applySearch() {
	query := strings.ToLower(strings.TrimSpace(g.searchInput.Value()))
	g.searchQuery = query

	if query == "" {
		g.filteredIssues = nil
		return
	}

	// Filter issues by title, identifier, or description
	var filtered []*models.Issue
	for _, issue := range g.issues {
		if g.matchesSearch(issue, query) {
			filtered = append(filtered, issue)
		}
	}
	g.filteredIssues = filtered
	g.selectedIssue = 0
}

func (g *Gui) matchesSearch(issue *models.Issue, query string) bool {
	// Match against identifier (e.g., "ENG-123")
	if strings.Contains(strings.ToLower(issue.Identifier), query) {
		return true
	}
	// Match against title
	if strings.Contains(strings.ToLower(issue.Title), query) {
		return true
	}
	// Match against description
	if strings.Contains(strings.ToLower(issue.Description), query) {
		return true
	}
	// Match against state name
	if issue.State != nil && strings.Contains(strings.ToLower(issue.State.Name), query) {
		return true
	}
	// Match against assignee name
	if issue.Assignee != nil && strings.Contains(strings.ToLower(issue.Assignee.DisplayName), query) {
		return true
	}
	return false
}

func (g *Gui) highlightMatch(text, query string) string {
	if query == "" {
		return text
	}
	lower := strings.ToLower(text)
	idx := strings.Index(lower, strings.ToLower(query))
	if idx == -1 {
		return text
	}
	// Highlight the matching portion
	before := text[:idx]
	match := text[idx : idx+len(query)]
	after := text[idx+len(query):]
	highlighted := lipgloss.NewStyle().Foreground(styles.Warning).Bold(true).Render(match)
	return before + highlighted + after
}

// Vim command mode functions

func (g *Gui) openCommandMode() {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 40

	g.commandMode = true
	g.commandInput = ti
}

func (g *Gui) closeCommandMode() {
	g.commandMode = false
}

func (g *Gui) handleCommandKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		g.closeCommandMode()
		return g, nil
	case "enter":
		return g.executeCommand()
	}

	// Update command input
	var cmd tea.Cmd
	g.commandInput, cmd = g.commandInput.Update(msg)
	return g, cmd
}

func (g *Gui) executeCommand() (tea.Model, tea.Cmd) {
	cmd := strings.TrimSpace(g.commandInput.Value())
	g.closeCommandMode()

	switch cmd {
	case "q", "quit":
		return g, tea.Quit
	case "w", "write", "refresh":
		g.loading = true
		g.lastSynced = time.Now()
		return g, g.loadIssues()
	case "help", "h":
		g.showHelp = true
		return g, nil
	case "clear", "c":
		g.searchQuery = ""
		g.filteredIssues = nil
		g.selectedIssue = 0
		return g, func() tea.Msg { return statusMsg("Search cleared") }
	default:
		// Check for number (go to issue by number)
		if len(cmd) > 0 {
			// Try to find issue by identifier
			for i, issue := range g.getDisplayIssues() {
				if strings.EqualFold(issue.Identifier, cmd) {
					g.selectedIssue = i
					g.focusedPanel = types.IssuesContext
					return g, g.loadSelectedIssueDetails()
				}
			}
			return g, func() tea.Msg { return statusMsg(fmt.Sprintf("Unknown command: %s", cmd)) }
		}
	}

	return g, nil
}
