package context

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mane-pal/lazylinear/pkg/gui/state"
	"github.com/mane-pal/lazylinear/pkg/gui/styles"
	"github.com/mane-pal/lazylinear/pkg/gui/types"
)

// DetailContext is the issue detail + comments panel.
type DetailContext struct {
	BaseContext
	State *state.AppState

	// Closures for actions
	OpenInBrowser        func(url string) tea.Cmd
	CopyToClipboard      func(text string) tea.Cmd
	OpenStateMenu        func() tea.Cmd
	OpenAssignMenu       func() tea.Cmd
	OpenPriorityMenu     func()
	OpenLabelsMenu       func() tea.Cmd
	OpenCycleMenu        func() tea.Cmd
	OpenCommentForm      func()
	OpenEditForm         func(issueID string)
	OpenEditCommentForm  func(commentID string)
	OpenDeleteCommentConfirm func(commentID string)
	OpenArchiveConfirm   func(issueID string)
	OpenDueDateInput     func(issueID string)
	OpenEstimateInput    func(issueID string)
	OpenParentPicker     func(issueID string) tea.Cmd
	OpenRelationTypeMenu func()
	OpenCreateSubIssueFormCmd func(issueID string) tea.Cmd
	LoadSelectedIssueDetails func() tea.Cmd
}

func NewDetailContext(st *state.AppState) *DetailContext {
	return &DetailContext{
		BaseContext: NewBaseContext(types.DetailContext, types.MainContext),
		State:      st,
	}
}

func (c *DetailContext) HandleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch key {
	case "h", "esc":
		c.State.FocusedPanel = types.IssuesContext
		c.State.SelectedComment = -1
	case "j", "down":
		if c.State.DetailedIssue != nil {
			comments := c.State.DetailedIssue.GetComments()
			if len(comments) > 0 {
				if c.State.SelectedComment < len(comments)-1 {
					c.State.SelectedComment++
				}
			}
		}
	case "k", "up":
		if c.State.SelectedComment > -1 {
			c.State.SelectedComment--
		}
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
	case "L":
		if c.OpenLabelsMenu != nil {
			return c.OpenLabelsMenu()
		}
	case "c":
		if c.State.SelectedComment == -1 && c.State.DetailedIssue != nil && c.OpenCommentForm != nil {
			c.OpenCommentForm()
		}
	case "D":
		if issue := c.State.GetSelectedIssue(); issue != nil && c.OpenDueDateInput != nil {
			c.OpenDueDateInput(issue.ID)
		}
	case "E":
		if issue := c.State.GetSelectedIssue(); issue != nil && c.OpenEstimateInput != nil {
			c.OpenEstimateInput(issue.ID)
		}
	case "P":
		if issue := c.State.GetSelectedIssue(); issue != nil && c.OpenParentPicker != nil {
			return c.OpenParentPicker(issue.ID)
		}
	case "R":
		if c.OpenRelationTypeMenu != nil {
			c.OpenRelationTypeMenu()
		}
	case "C":
		if c.OpenCycleMenu != nil {
			return c.OpenCycleMenu()
		}
	case "e":
		if c.State.SelectedComment >= 0 && c.State.DetailedIssue != nil {
			comments := c.State.DetailedIssue.GetComments()
			if c.State.SelectedComment < len(comments) {
				comment := comments[c.State.SelectedComment]
				if c.State.CurrentUser != nil && comment.User != nil && comment.User.ID == c.State.CurrentUser.ID {
					if c.OpenEditCommentForm != nil {
						c.OpenEditCommentForm(comment.ID)
					}
				} else {
					return func() tea.Msg { return types.StatusMsg("Can only edit your own comments") }
				}
			}
		} else if issue := c.State.GetSelectedIssue(); issue != nil && c.OpenEditForm != nil {
			c.OpenEditForm(issue.ID)
		}
	case "d":
		if c.State.SelectedComment >= 0 && c.State.DetailedIssue != nil {
			comments := c.State.DetailedIssue.GetComments()
			if c.State.SelectedComment < len(comments) {
				comment := comments[c.State.SelectedComment]
				if c.State.CurrentUser != nil && comment.User != nil && comment.User.ID == c.State.CurrentUser.ID {
					if c.OpenDeleteCommentConfirm != nil {
						c.OpenDeleteCommentConfirm(comment.ID)
					}
				} else {
					return func() tea.Msg { return types.StatusMsg("Can only delete your own comments") }
				}
			}
		} else if issue := c.State.GetSelectedIssue(); issue != nil && c.OpenArchiveConfirm != nil {
			c.OpenArchiveConfirm(issue.ID)
		}
	case "N":
		if issue := c.State.GetSelectedIssue(); issue != nil && c.OpenCreateSubIssueFormCmd != nil {
			return c.OpenCreateSubIssueFormCmd(issue.ID)
		}
	}
	return nil
}

func (c *DetailContext) View(width, height int) string {
	style := styles.Panel
	if c.State.FocusedPanel == types.DetailContext {
		style = styles.PanelFocused
	}

	var b strings.Builder

	b.WriteString(styles.PanelTitle.Render("─ Detail ─"))
	b.WriteString("\n\n")

	issue := c.State.GetSelectedIssue()
	if issue == nil {
		b.WriteString(styles.Subtle.Render("  Select an issue"))
		return style.Width(width).Height(height).Render(b.String())
	}

	displayIssue := issue
	if c.State.DetailedIssue != nil && c.State.DetailedIssue.ID == issue.ID {
		displayIssue = c.State.DetailedIssue
	}

	b.WriteString(styles.Title.Render(displayIssue.Identifier))
	b.WriteString("\n\n")

	titleWidth := width - 4
	if titleWidth < 20 {
		titleWidth = 20
	}
	wrappedTitle := wrapText(displayIssue.Title, titleWidth)
	b.WriteString(wrappedTitle)
	b.WriteString("\n\n")

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

	if displayIssue.Description != "" {
		b.WriteString("\n")
		b.WriteString(styles.Subtle.Render("─────────────────────────"))
		b.WriteString("\n\n")
		desc := wrapText(displayIssue.Description, titleWidth)
		lines := strings.Split(desc, "\n")
		if len(lines) > 6 {
			desc = strings.Join(lines[:6], "\n") + "..."
		}
		b.WriteString(desc)
	}

	comments := displayIssue.GetComments()
	if len(comments) > 0 || c.State.FocusedPanel == types.DetailContext {
		b.WriteString("\n\n")
		b.WriteString(styles.Subtle.Render("─ Comments ─"))
		if c.State.FocusedPanel == types.DetailContext {
			b.WriteString(styles.Subtle.Render(" (c: add, j/k: navigate)"))
		}
		b.WriteString("\n\n")

		if len(comments) == 0 {
			b.WriteString(styles.Subtle.Render("  No comments yet"))
		} else {
			for i, comment := range comments {
				isSelected := c.State.FocusedPanel == types.DetailContext && i == c.State.SelectedComment

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

func wrapText(s string, width int) string {
	if width <= 0 {
		return s
	}
	var result strings.Builder
	words := strings.Fields(s)
	lineLen := 0

	for _, word := range words {
		wordLen := len(word)

		if lineLen+wordLen+1 > width && lineLen > 0 {
			result.WriteString("\n")
			lineLen = 0
		}

		if lineLen > 0 {
			result.WriteString(" ")
			lineLen++
		}

		if wordLen > width {
			for len(word) > width {
				result.WriteString(word[:width])
				result.WriteString("\n")
				word = word[width:]
			}
			result.WriteString(word)
			lineLen = len(word)
		} else {
			result.WriteString(word)
			lineLen += wordLen
		}
	}

	return result.String()
}

func formatTimeAgo(t time.Time) string {
	diff := time.Since(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 2, 2006")
	}
}
