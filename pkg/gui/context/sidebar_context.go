package context

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mane-pal/lazylinear/pkg/gui/state"
	"github.com/mane-pal/lazylinear/pkg/gui/styles"
	"github.com/mane-pal/lazylinear/pkg/gui/types"
)

// FilterOption represents a sidebar filter option.
type FilterOption struct {
	Key   string
	Label string
}

// FilterOptions defines the available sidebar filters.
var FilterOptions = []FilterOption{
	{"all", "All"},
	{"my_issues", "My Issues"},
}

// SidebarContext is the sidebar panel (teams, filters, states, priority).
type SidebarContext struct {
	BaseContext
	State *state.AppState

	// Closures for triggering actions without circular deps.
	LoadIssues     func() tea.Cmd
	LoadTeamStates func() tea.Cmd
}

func NewSidebarContext(st *state.AppState) *SidebarContext {
	return &SidebarContext{
		BaseContext: NewBaseContext(types.TeamsContext, types.SideContext),
		State:      st,
	}
}

func (c *SidebarContext) HandleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch key {
	case "j", "down":
		switch c.State.SidebarSection {
		case 0:
			if c.State.SelectedTeam < len(c.State.Teams)-1 {
				c.State.SelectedTeam++
			} else {
				c.State.SidebarSection = 1
				c.State.SelectedFilter = 0
			}
		case 1:
			if c.State.SelectedFilter < len(FilterOptions)-1 {
				c.State.SelectedFilter++
			} else {
				c.State.SidebarSection = 2
				c.State.SelectedStateIdx = 0
			}
		case 2:
			if c.State.SelectedStateIdx < len(c.State.TeamStates)-1 {
				c.State.SelectedStateIdx++
			} else {
				c.State.SidebarSection = 3
				c.State.SelectedPriority = 0
			}
		case 3:
			if c.State.SelectedPriority < 3 {
				c.State.SelectedPriority++
			}
		}
	case "k", "up":
		switch c.State.SidebarSection {
		case 0:
			if c.State.SelectedTeam > 0 {
				c.State.SelectedTeam--
			}
		case 1:
			if c.State.SelectedFilter > 0 {
				c.State.SelectedFilter--
			} else {
				c.State.SidebarSection = 0
			}
		case 2:
			if c.State.SelectedStateIdx > 0 {
				c.State.SelectedStateIdx--
			} else {
				c.State.SidebarSection = 1
				c.State.SelectedFilter = len(FilterOptions) - 1
			}
		case 3:
			if c.State.SelectedPriority > 0 {
				c.State.SelectedPriority--
			} else {
				c.State.SidebarSection = 2
				if len(c.State.TeamStates) > 0 {
					c.State.SelectedStateIdx = len(c.State.TeamStates) - 1
				}
			}
		}
	case " ":
		switch c.State.SidebarSection {
		case 1:
			c.State.ActiveFilter = FilterOptions[c.State.SelectedFilter].Key
			c.State.SelectedIssue = 0
			c.State.Loading = true
			return c.LoadIssues()
		case 2:
			if c.State.SelectedStateIdx < len(c.State.TeamStates) {
				stateID := c.State.TeamStates[c.State.SelectedStateIdx].ID
				if c.State.ActiveStateFilters[stateID] {
					delete(c.State.ActiveStateFilters, stateID)
				} else {
					c.State.ActiveStateFilters[stateID] = true
				}
				c.State.SelectedIssue = 0
				c.State.Loading = true
				return c.LoadIssues()
			}
		case 3:
			priority := c.State.SelectedPriority + 1
			if c.State.ActivePriorityFilter == priority {
				c.State.ActivePriorityFilter = -1
			} else {
				c.State.ActivePriorityFilter = priority
			}
			c.State.SelectedIssue = 0
			c.State.Loading = true
			return c.LoadIssues()
		}
	case "enter", "l":
		if c.State.SidebarSection == 0 {
			c.State.SelectedIssue = 0
			c.State.ActiveStateFilters = make(map[string]bool)
			c.State.Loading = true
			c.State.FocusedPanel = types.IssuesContext
			return tea.Batch(c.LoadIssues(), c.LoadTeamStates())
		}
		c.State.FocusedPanel = types.IssuesContext
	case "m":
		if c.State.ActiveFilter == "my_issues" {
			c.State.ActiveFilter = "all"
		} else {
			c.State.ActiveFilter = "my_issues"
		}
		c.State.SelectedIssue = 0
		c.State.SelectedFilter = GetFilterIndex(c.State.ActiveFilter)
		c.State.Loading = true
		return c.LoadIssues()
	case "x":
		c.State.ActiveFilter = "all"
		c.State.ActiveStateFilters = make(map[string]bool)
		for _, st := range c.State.TeamStates {
			if st.Type == "backlog" || st.Type == "unstarted" || st.Type == "started" {
				c.State.ActiveStateFilters[st.ID] = true
			}
		}
		c.State.ActivePriorityFilter = -1
		c.State.SelectedIssue = 0
		c.State.SelectedFilter = 0
		c.State.SelectedStateIdx = 0
		c.State.SelectedPriority = 0
		c.State.Loading = true
		return c.LoadIssues()
	}
	return nil
}

func (c *SidebarContext) View(width, height int) string {
	style := styles.Panel
	if c.State.FocusedPanel == types.TeamsContext {
		style = styles.PanelFocused
	}

	var b strings.Builder

	b.WriteString(styles.PanelTitle.Render("─ Teams ─"))
	b.WriteString("\n\n")

	if len(c.State.Teams) == 0 {
		b.WriteString(styles.Subtle.Render("  Loading..."))
	} else {
		for i, team := range c.State.Teams {
			countStr := ""
			if i == c.State.SelectedTeam {
				count := len(c.State.GetDisplayIssues())
				countStr = fmt.Sprintf(" (%d)", count)
			}

			isSelected := i == c.State.SelectedTeam && c.State.SidebarSection == 0
			if isSelected {
				line := fmt.Sprintf("> %s%s", team.Key, countStr)
				b.WriteString(styles.Selected.Render(line))
			} else if i == c.State.SelectedTeam {
				b.WriteString(fmt.Sprintf("  %s%s", styles.Title.Render(team.Key), countStr))
			} else {
				b.WriteString(fmt.Sprintf("  %s", team.Key))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(styles.PanelTitle.Render("─ Filters ─"))
	b.WriteString("\n\n")

	for i, filter := range FilterOptions {
		isActive := c.State.ActiveFilter == filter.Key
		isSelected := c.State.SidebarSection == 1 && c.State.SelectedFilter == i && c.State.FocusedPanel == types.TeamsContext

		prefix := "  "
		if isActive {
			prefix = "● "
		}

		if isSelected {
			b.WriteString(styles.Selected.Render(fmt.Sprintf("> %s", filter.Label)))
		} else if isActive {
			b.WriteString(styles.Title.Render(prefix + filter.Label))
		} else {
			b.WriteString(styles.Subtle.Render(prefix + filter.Label))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.PanelTitle.Render("─ States ─"))
	b.WriteString("\n\n")

	if len(c.State.TeamStates) == 0 {
		b.WriteString(styles.Subtle.Render("  Loading..."))
		b.WriteString("\n")
	} else {
		for i, st := range c.State.TeamStates {
			count := c.State.CountIssuesByState(st.ID)
			isActive := c.State.ActiveStateFilters[st.ID]
			isSelected := c.State.SidebarSection == 2 && c.State.SelectedStateIdx == i && c.State.FocusedPanel == types.TeamsContext

			stateStyle := lipgloss.NewStyle().Foreground(styles.StateColor(st.Type))
			icon := styles.StateIcon(st.Type)
			label := fmt.Sprintf("%s %s (%d)", icon, st.Name, count)

			if isSelected {
				b.WriteString(styles.Selected.Render(fmt.Sprintf("> %s", label)))
			} else if isActive {
				b.WriteString(stateStyle.Bold(true).Render("● " + label))
			} else {
				b.WriteString(stateStyle.Render("  " + label))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(styles.PanelTitle.Render("─ Priority ─"))
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
		count := c.State.CountIssuesByPriority(prio.level)
		isActive := c.State.ActivePriorityFilter == prio.level
		isSelected := c.State.SidebarSection == 3 && c.State.SelectedPriority == i && c.State.FocusedPanel == types.TeamsContext

		prioStyle := lipgloss.NewStyle().Foreground(prio.color)
		label := fmt.Sprintf("%s (%d)", prio.label, count)

		if isSelected {
			b.WriteString(styles.Selected.Render(fmt.Sprintf("> %s", label)))
		} else if isActive {
			b.WriteString(prioStyle.Bold(true).Render("● " + label))
		} else {
			b.WriteString(prioStyle.Render("  " + label))
		}
		b.WriteString("\n")
	}

	return style.Width(width).Height(height).Render(b.String())
}

// CycleSidebarSection cycles through sidebar sections.
func CycleSidebarSection(st *state.AppState, direction int) {
	numSections := 4
	st.SidebarSection = (st.SidebarSection + direction + numSections) % numSections

	switch st.SidebarSection {
	case 0:
		// Keep selectedTeam as is
	case 1:
		st.SelectedFilter = 0
	case 2:
		st.SelectedStateIdx = 0
	case 3:
		st.SelectedPriority = 0
	}
}

// GetFilterIndex returns the index of a filter option by key.
func GetFilterIndex(filter string) int {
	for i, f := range FilterOptions {
		if f.Key == filter {
			return i
		}
	}
	return 0
}
