package context

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mane-pal/lazylinear/pkg/gui/controllers/helpers"
	"github.com/mane-pal/lazylinear/pkg/gui/state"
	"github.com/mane-pal/lazylinear/pkg/gui/styles"
	"github.com/mane-pal/lazylinear/pkg/gui/types"
)

// IssuesContext is the issues/projects list panel.
type IssuesContext struct {
	BaseContext
	State        *state.AppState
	FilterHelper *helpers.FilterHelper

	// Closures
	LoadSelectedIssueDetails func() tea.Cmd
	OpenInBrowser            func(url string) tea.Cmd
	CopyToClipboard          func(text string) tea.Cmd
	OpenStateMenu            func() tea.Cmd
	OpenAssignMenu           func() tea.Cmd
	OpenPriorityMenu         func()
	OpenCreateFormCmd        func() tea.Cmd
	OpenEditForm             func(issueID string)
	OpenArchiveConfirm       func(issueID string)
	OpenLabelsMenu           func() tea.Cmd
	OpenCycleMenu            func() tea.Cmd
}

func NewIssuesContext(st *state.AppState) *IssuesContext {
	return &IssuesContext{
		BaseContext:   NewBaseContext(types.IssuesContext, types.MainContext),
		State:         st,
		FilterHelper:  helpers.NewFilterHelper(),
	}
}

func (c *IssuesContext) HandleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	if c.State.MiddlePaneView == types.ViewProjects {
		return c.handleProjectsKey(key)
	}

	switch key {
	case "j", "down":
		issues := c.State.GetDisplayIssues()
		if c.State.SelectedIssue < len(issues)-1 {
			c.State.SelectedIssue++
			if c.LoadSelectedIssueDetails != nil {
				return c.LoadSelectedIssueDetails()
			}
		}
	case "k", "up":
		if c.State.SelectedIssue > 0 {
			c.State.SelectedIssue--
			if c.LoadSelectedIssueDetails != nil {
				return c.LoadSelectedIssueDetails()
			}
		}
	case "g":
		c.State.SelectedIssue = 0
		if c.LoadSelectedIssueDetails != nil {
			return c.LoadSelectedIssueDetails()
		}
	case "G":
		issues := c.State.GetDisplayIssues()
		c.State.SelectedIssue = max(0, len(issues)-1)
		if c.LoadSelectedIssueDetails != nil {
			return c.LoadSelectedIssueDetails()
		}
	case "enter", "l":
		c.State.FocusedPanel = types.DetailContext
		if c.LoadSelectedIssueDetails != nil {
			return c.LoadSelectedIssueDetails()
		}
	case "h":
		c.State.FocusedPanel = types.TeamsContext
	case "o":
		if issue := c.State.GetSelectedIssue(); issue != nil && c.OpenInBrowser != nil {
			return c.OpenInBrowser(issue.URL)
		}
	case "y":
		if issue := c.State.GetSelectedIssue(); issue != nil && c.CopyToClipboard != nil {
			return c.CopyToClipboard(issue.URL)
		}
	case "Y":
		if issue := c.State.GetSelectedIssue(); issue != nil && c.CopyToClipboard != nil {
			return c.CopyToClipboard(issue.Identifier)
		}
	case "s":
		if c.OpenStateMenu != nil {
			return c.OpenStateMenu()
		}
	case "a":
		if c.OpenAssignMenu != nil {
			return c.OpenAssignMenu()
		}
	case "p":
		if c.OpenPriorityMenu != nil {
			c.OpenPriorityMenu()
		}
	case "n":
		if c.OpenCreateFormCmd != nil {
			return c.OpenCreateFormCmd()
		}
	case "e":
		if issue := c.State.GetSelectedIssue(); issue != nil && c.OpenEditForm != nil {
			c.OpenEditForm(issue.ID)
		}
	case "d":
		if issue := c.State.GetSelectedIssue(); issue != nil && c.OpenArchiveConfirm != nil {
			c.OpenArchiveConfirm(issue.ID)
		}
	case "L":
		if c.OpenLabelsMenu != nil {
			return c.OpenLabelsMenu()
		}
	case "c":
		if c.OpenCycleMenu != nil {
			return c.OpenCycleMenu()
		}
	}
	return nil
}

func (c *IssuesContext) handleProjectsKey(key string) tea.Cmd {
	switch key {
	case "j", "down":
		if c.State.SelectedProject < len(c.State.Projects)-1 {
			c.State.SelectedProject++
		}
	case "k", "up":
		if c.State.SelectedProject > 0 {
			c.State.SelectedProject--
		}
	case "g":
		c.State.SelectedProject = 0
	case "G":
		c.State.SelectedProject = max(0, len(c.State.Projects)-1)
	case "enter", "l":
		// No-op for projects currently
	case "h":
		c.State.FocusedPanel = types.TeamsContext
	case "o":
		if c.State.SelectedProject < len(c.State.Projects) && c.OpenInBrowser != nil {
			return c.OpenInBrowser(c.State.Projects[c.State.SelectedProject].URL)
		}
	}
	return nil
}

func (c *IssuesContext) View(width, height int) string {
	if c.State.MiddlePaneView == types.ViewProjects {
		return c.renderProjectsList(width, height)
	}
	return c.renderIssuesList(width, height)
}

func (c *IssuesContext) renderIssuesList(width, height int) string {
	style := styles.Panel
	if c.State.FocusedPanel == types.IssuesContext {
		style = styles.PanelFocused
	}

	var b strings.Builder

	issues := c.State.GetDisplayIssues()
	issueCount := len(issues)
	header := fmt.Sprintf("─ Issues (%d) ─", issueCount)
	if c.State.SearchQuery != "" {
		header = fmt.Sprintf("─ Issues (%d) [/%s] ─", issueCount, c.State.SearchQuery)
	}
	b.WriteString(styles.PanelTitle.Render(header))
	b.WriteString("\n\n")

	if c.State.Loading {
		b.WriteString(styles.Subtle.Render("  Loading..."))
	} else if c.State.Err != nil {
		b.WriteString(styles.ErrorStyle.Render("  Error: " + c.State.Err.Error()))
	} else if len(issues) == 0 {
		if c.State.SearchQuery != "" {
			b.WriteString(styles.Subtle.Render("  No matching issues"))
		} else {
			b.WriteString(styles.Subtle.Render("  No issues"))
		}
	} else {
		visibleLines := height - 6
		if visibleLines < 1 {
			visibleLines = 1
		}
		start := 0
		if c.State.SelectedIssue >= visibleLines {
			start = c.State.SelectedIssue - visibleLines + 1
		}
		end := min(len(issues), start+visibleLines)

		identifierWidth := 12
		titleWidth := width - identifierWidth - 8
		if titleWidth < 10 {
			titleWidth = 10
		}

		for i := start; i < end; i++ {
			issue := issues[i]

			stateIcon := "○"
			stateColor := styles.Secondary
			if issue.State != nil {
				stateIcon = styles.StateIcon(issue.State.Type)
				stateColor = styles.StateColor(issue.State.Type)
			}
			iconStyled := lipgloss.NewStyle().Foreground(stateColor).Render(stateIcon)

			title := truncate(issue.Title, titleWidth)

			if c.State.SearchQuery != "" {
				title = c.FilterHelper.HighlightMatch(title, c.State.SearchQuery)
			}

			if i == c.State.SelectedIssue {
				line := fmt.Sprintf("> %s %s %s", iconStyled, issue.Identifier, title)
				b.WriteString(styles.Selected.Render(line))
			} else {
				line := fmt.Sprintf("  %s %s %s", iconStyled, styles.Subtle.Render(issue.Identifier), title)
				b.WriteString(line)
			}
			b.WriteString("\n")
		}

		if len(issues) > visibleLines {
			scrollInfo := fmt.Sprintf("\n  %d/%d", c.State.SelectedIssue+1, len(issues))
			b.WriteString(styles.Subtle.Render(scrollInfo))
		}
	}

	return style.Width(width).Height(height).Render(b.String())
}

func (c *IssuesContext) renderProjectsList(width, height int) string {
	style := styles.Panel
	if c.State.FocusedPanel == types.IssuesContext {
		style = styles.PanelFocused
	}

	var b strings.Builder

	header := fmt.Sprintf("─ Projects (%d) ─", len(c.State.Projects))
	b.WriteString(styles.PanelTitle.Render(header))
	b.WriteString("\n\n")

	if c.State.Loading {
		b.WriteString(styles.Subtle.Render("  Loading..."))
	} else if c.State.Err != nil {
		b.WriteString(styles.ErrorStyle.Render("  Error: " + c.State.Err.Error()))
	} else if len(c.State.Projects) == 0 {
		b.WriteString(styles.Subtle.Render("  No projects"))
	} else {
		visibleLines := height - 6
		if visibleLines < 1 {
			visibleLines = 1
		}
		start := 0
		if c.State.SelectedProject >= visibleLines {
			start = c.State.SelectedProject - visibleLines + 1
		}
		end := min(len(c.State.Projects), start+visibleLines)

		nameWidth := width - 20
		if nameWidth < 10 {
			nameWidth = 10
		}

		for i := start; i < end; i++ {
			project := c.State.Projects[i]

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

			progress := fmt.Sprintf("%3.0f%%", project.Progress*100)
			name := truncate(project.Name, nameWidth)
			issueInfo := fmt.Sprintf("(%d)", project.IssueCount)

			if i == c.State.SelectedProject {
				line := fmt.Sprintf("> %s %s %s %s", iconStyled, name, progress, issueInfo)
				b.WriteString(styles.Selected.Render(line))
			} else {
				line := fmt.Sprintf("  %s %s %s %s", iconStyled, name, styles.Subtle.Render(progress), styles.Subtle.Render(issueInfo))
				b.WriteString(line)
			}
			b.WriteString("\n")
		}

		if len(c.State.Projects) > visibleLines {
			scrollInfo := fmt.Sprintf("\n  %d/%d", c.State.SelectedProject+1, len(c.State.Projects))
			b.WriteString(styles.Subtle.Render(scrollInfo))
		}
	}

	return style.Width(width).Height(height).Render(b.String())
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
