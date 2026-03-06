package popup

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mane-pal/lazylinear/pkg/gui/state"
	"github.com/mane-pal/lazylinear/pkg/gui/styles"
	"github.com/mane-pal/lazylinear/pkg/gui/types"
	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

// SearchPopup handles search and vim command mode overlays.
type SearchPopup struct {
	State *state.AppState

	// Search state
	SearchMode  bool
	SearchInput textinput.Model

	// Vim command mode
	CommandMode  bool
	CommandInput textinput.Model

	// Closures for actions
	LoadIssues               func() tea.Cmd
	LoadSelectedIssueDetails func() tea.Cmd
}

func NewSearchPopup(st *state.AppState) *SearchPopup {
	return &SearchPopup{State: st}
}

func (p *SearchPopup) IsOpen() bool {
	return p.SearchMode || p.CommandMode
}

func (p *SearchPopup) OpenSearch() {
	ti := textinput.New()
	ti.Placeholder = "Search issues..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40

	p.SearchMode = true
	p.SearchInput = ti
}

func (p *SearchPopup) CloseSearch() {
	p.SearchMode = false
}

func (p *SearchPopup) HandleSearchKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch key {
	case "esc":
		p.CloseSearch()
		return nil
	case "enter":
		p.applySearch()
		return nil
	}

	var cmd tea.Cmd
	p.SearchInput, cmd = p.SearchInput.Update(msg)
	p.applySearch()

	return cmd
}

func (p *SearchPopup) applySearch() {
	query := strings.ToLower(strings.TrimSpace(p.SearchInput.Value()))
	p.State.SearchQuery = query

	if query == "" {
		p.State.FilteredIssues = nil
		return
	}

	var filtered []*models.Issue
	for _, issue := range p.State.Issues {
		if matchesSearch(issue, query) {
			filtered = append(filtered, issue)
		}
	}
	p.State.FilteredIssues = filtered
	p.State.SelectedIssue = 0
}

func matchesSearch(issue *models.Issue, query string) bool {
	if strings.Contains(strings.ToLower(issue.Identifier), query) {
		return true
	}
	if strings.Contains(strings.ToLower(issue.Title), query) {
		return true
	}
	if strings.Contains(strings.ToLower(issue.Description), query) {
		return true
	}
	if issue.State != nil && strings.Contains(strings.ToLower(issue.State.Name), query) {
		return true
	}
	if issue.Assignee != nil && strings.Contains(strings.ToLower(issue.Assignee.DisplayName), query) {
		return true
	}
	return false
}

func (p *SearchPopup) RenderSearchBar(width int) string {
	searchStyle := lipgloss.NewStyle().
		Background(styles.BgDark).
		Foreground(styles.FgLight).
		Padding(0, 1).
		Width(width)

	prompt := lipgloss.NewStyle().Foreground(styles.Primary).Render("/")
	input := p.SearchInput.View()

	return searchStyle.Render(prompt + input)
}

// Command mode

func (p *SearchPopup) OpenCommandMode() {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 40

	p.CommandMode = true
	p.CommandInput = ti
}

func (p *SearchPopup) CloseCommandMode() {
	p.CommandMode = false
}

func (p *SearchPopup) HandleCommandKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch key {
	case "esc":
		p.CloseCommandMode()
		return nil
	case "enter":
		return p.executeCommand()
	}

	var cmd tea.Cmd
	p.CommandInput, cmd = p.CommandInput.Update(msg)
	return cmd
}

func (p *SearchPopup) executeCommand() tea.Cmd {
	cmd := strings.TrimSpace(p.CommandInput.Value())
	p.CloseCommandMode()

	switch cmd {
	case "q", "quit":
		return tea.Quit
	case "w", "write", "refresh":
		p.State.Loading = true
		p.State.LastSynced = time.Now()
		if p.LoadIssues != nil {
			return p.LoadIssues()
		}
		return nil
	case "help", "h":
		p.State.ShowHelp = true
		return nil
	case "clear", "c":
		p.State.SearchQuery = ""
		p.State.FilteredIssues = nil
		p.State.SelectedIssue = 0
		return func() tea.Msg { return types.StatusMsg("Search cleared") }
	default:
		if len(cmd) > 0 {
			for i, issue := range p.State.GetDisplayIssues() {
				if strings.EqualFold(issue.Identifier, cmd) {
					p.State.SelectedIssue = i
					p.State.FocusedPanel = types.IssuesContext
					if p.LoadSelectedIssueDetails != nil {
						return p.LoadSelectedIssueDetails()
					}
					return nil
				}
			}
			return func() tea.Msg { return types.StatusMsg(fmt.Sprintf("Unknown command: %s", cmd)) }
		}
	}

	return nil
}

func (p *SearchPopup) RenderCommandLine(width int) string {
	cmdStyle := lipgloss.NewStyle().
		Background(styles.BgDark).
		Foreground(styles.FgLight).
		Padding(0, 1).
		Width(width)

	prompt := lipgloss.NewStyle().Foreground(styles.Primary).Render(":")
	input := p.CommandInput.View()

	return cmdStyle.Render(prompt + input)
}
