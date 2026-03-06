package popup

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mane-pal/lazylinear/pkg/gui/state"
	"github.com/mane-pal/lazylinear/pkg/gui/styles"
	"github.com/mane-pal/lazylinear/pkg/gui/types"
)

// MenuPopup handles menu, confirm, and label overlay dialogs.
type MenuPopup struct {
	State *state.AppState

	MenuType     types.MenuType
	MenuItems    []types.MenuItem
	MenuSelected int
	MenuTitle    string

	// Confirm dialog
	ConfirmTitle   string
	ConfirmMessage string
	ConfirmAction  func() tea.Cmd

	// Closures for issue update actions
	UpdateIssueState    func(issueID, stateID string) tea.Cmd
	UpdateIssueAssignee func(issueID, assigneeID string) tea.Cmd
	UpdateIssuePriority func(issueID, priorityStr string) tea.Cmd
	UpdateIssueCycle    func(issueID string, cycleID *string) tea.Cmd
	UpdateIssueLabels   func(issueID string, labelIDs []string) tea.Cmd
}

func NewMenuPopup(st *state.AppState) *MenuPopup {
	return &MenuPopup{State: st}
}

func (p *MenuPopup) IsOpen() bool {
	return p.MenuType != types.MenuNone
}

func (p *MenuPopup) Close() {
	p.MenuType = types.MenuNone
	p.MenuItems = nil
	p.MenuSelected = 0
	p.MenuTitle = ""
}

func (p *MenuPopup) HandleKey(key string) (tea.Cmd, bool) {
	if !p.IsOpen() {
		return nil, false
	}

	if p.MenuType == types.MenuConfirm {
		switch key {
		case "y", "Y", "enter":
			cmd := p.ConfirmAction()
			p.Close()
			return cmd, true
		case "n", "N", "esc", "q":
			p.Close()
		}
		return nil, true
	}

	if p.MenuType == types.MenuLabels {
		switch key {
		case "j", "down":
			if p.MenuSelected < len(p.MenuItems)-1 {
				p.MenuSelected++
			}
		case "k", "up":
			if p.MenuSelected > 0 {
				p.MenuSelected--
			}
		case " ":
			if p.MenuSelected < len(p.MenuItems) {
				item := &p.MenuItems[p.MenuSelected]
				item.Selected = !item.Selected
				p.State.SelectedLabels[item.ID] = item.Selected
				if !item.Selected {
					delete(p.State.SelectedLabels, item.ID)
				}
			}
		case "enter":
			issue := p.State.GetSelectedIssue()
			if issue != nil && p.UpdateIssueLabels != nil {
				labelIDs := make([]string, 0, len(p.State.SelectedLabels))
				for id := range p.State.SelectedLabels {
					labelIDs = append(labelIDs, id)
				}
				cmd := p.UpdateIssueLabels(issue.ID, labelIDs)
				p.Close()
				return cmd, true
			}
			p.Close()
		case "esc", "q":
			p.Close()
		}
		return nil, true
	}

	switch key {
	case "j", "down":
		if p.MenuSelected < len(p.MenuItems)-1 {
			p.MenuSelected++
		}
	case "k", "up":
		if p.MenuSelected > 0 {
			p.MenuSelected--
		}
	case "enter":
		return p.selectMenuItem(), true
	case "esc", "q":
		p.Close()
	}
	return nil, true
}

func (p *MenuPopup) selectMenuItem() tea.Cmd {
	if p.MenuSelected >= len(p.MenuItems) {
		p.Close()
		return nil
	}

	item := p.MenuItems[p.MenuSelected]
	issue := p.State.GetSelectedIssue()
	if issue == nil {
		p.Close()
		return nil
	}

	var cmd tea.Cmd

	switch p.MenuType {
	case types.MenuState:
		if p.UpdateIssueState != nil {
			cmd = p.UpdateIssueState(issue.ID, item.ID)
		}
	case types.MenuAssign:
		if p.UpdateIssueAssignee != nil {
			cmd = p.UpdateIssueAssignee(issue.ID, item.ID)
		}
	case types.MenuPriority:
		if p.UpdateIssuePriority != nil {
			cmd = p.UpdateIssuePriority(issue.ID, item.ID)
		}
	case types.MenuCycle:
		if p.UpdateIssueCycle != nil {
			var cycleID *string
			if item.ID != "" {
				cycleID = &item.ID
			}
			cmd = p.UpdateIssueCycle(issue.ID, cycleID)
		}
	case types.MenuRelationType:
		p.Close()
		return func() tea.Msg { return types.StatusMsg("Relation type selected: " + item.ID) }
	}

	p.Close()
	return cmd
}

func (p *MenuPopup) RenderMenu() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render(p.MenuTitle))
	b.WriteString("\n\n")

	for i, item := range p.MenuItems {
		itemStyle := lipgloss.NewStyle().Foreground(item.Color)

		prefix := "  "
		if p.MenuType == types.MenuLabels {
			if item.Selected {
				prefix = "[✓] "
			} else {
				prefix = "[ ] "
			}
		}

		if i == p.MenuSelected {
			line := fmt.Sprintf("> %s%s", prefix[2:], item.Label)
			b.WriteString(styles.Selected.Render(line))
		} else {
			line := fmt.Sprintf("%s%s", prefix, itemStyle.Render(item.Label))
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if p.MenuType == types.MenuLabels {
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

func (p *MenuPopup) RenderConfirm() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render(p.ConfirmTitle))
	b.WriteString("\n\n")
	b.WriteString(p.ConfirmMessage)
	b.WriteString("\n\n")
	b.WriteString(styles.Subtle.Render("y: confirm  n/esc: cancel"))

	confirmStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Warning).
		Padding(1, 2).
		Width(50)

	return confirmStyle.Render(b.String())
}
