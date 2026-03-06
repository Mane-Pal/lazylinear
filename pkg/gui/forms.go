package gui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mane-pal/lazylinear/pkg/gui/types"
	"github.com/mane-pal/lazylinear/pkg/linear"
	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

// Form functions

func (g *Gui) openCreateForm() {
	formWidth := (g.width * 70) / 100
	if formWidth > 100 {
		formWidth = 100
	}
	if formWidth < 70 {
		formWidth = 70
	}
	inputWidth := formWidth - 10

	ti := textinput.New()
	ti.Placeholder = "Issue title..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = inputWidth

	ta := textarea.New()
	ta.Placeholder = "Description (optional, supports markdown)..."
	ta.SetWidth(inputWidth)
	ta.SetHeight(12)

	search := textinput.New()
	search.Placeholder = "Type to filter..."
	search.CharLimit = 50
	search.Width = 40

	g.formMode = types.FormCreate
	g.formField = types.FieldTitle
	g.formTitle = ti
	g.formDescription = ta
	g.formPriority = 0
	g.formTeam = g.State.SelectedTeam
	g.formAssignee = -1
	g.formState = -1
	g.formProject = -1
	g.formEditIssueID = ""
	g.formListSearch = search
	g.formListSelected = 0
}

func (g *Gui) openCreateFormCmd() tea.Cmd {
	g.openCreateForm()
	return tea.Batch(g.loadTeamMembers(), g.loadProjects())
}

func (g *Gui) openEditForm(issue *models.Issue) {
	ti := textinput.New()
	ti.SetValue(issue.Title)
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 50

	ta := textarea.New()
	ta.SetValue(issue.Description)
	ta.SetWidth(50)
	ta.SetHeight(8)

	g.formMode = types.FormEdit
	g.formField = types.FieldTitle
	g.formTitle = ti
	g.formDescription = ta
	g.formPriority = issue.Priority
	g.formEditIssueID = issue.ID
}

func (g *Gui) closeForm() {
	g.formMode = types.FormNone
	g.formEditIssueID = ""
	g.editingCommentID = ""
}

func (g *Gui) openCommentForm() {
	ta := textarea.New()
	ta.Placeholder = "Write a comment..."
	ta.SetWidth(50)
	ta.SetHeight(6)
	ta.Focus()

	g.formMode = types.FormComment
	g.commentBody = ta
}

func (g *Gui) openEditCommentForm(comment *models.Comment) {
	ta := textarea.New()
	ta.SetValue(comment.Body)
	ta.SetWidth(50)
	ta.SetHeight(6)
	ta.Focus()

	g.formMode = types.FormEditComment
	g.commentBody = ta
	g.editingCommentID = comment.ID
}

func (g *Gui) openDeleteCommentConfirm(comment *models.Comment) {
	g.MenuPopup.MenuType = types.MenuConfirm
	g.MenuPopup.ConfirmTitle = "Delete Comment?"
	g.MenuPopup.ConfirmMessage = fmt.Sprintf("Are you sure you want to delete this comment?\n\n%s", truncate(comment.Body, 50))
	commentID := comment.ID
	g.MenuPopup.ConfirmAction = func() tea.Cmd {
		return g.deleteComment(commentID)
	}
}

func (g *Gui) deleteComment(commentID string) tea.Cmd {
	issueID := ""
	if g.State.DetailedIssue != nil {
		issueID = g.State.DetailedIssue.ID
	}
	return func() tea.Msg {
		if err := g.client.Comments.Delete(context.Background(), commentID); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to delete comment: %v", err))
		}
		return types.CommentDeletedMsg{IssueID: issueID}
	}
}

func (g *Gui) handleFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if g.formListOpen {
		return g.handleFormListKey(key, msg)
	}

	switch key {
	case "esc":
		g.closeForm()
		return g, nil
	case "ctrl+s", "ctrl+enter":
		return g.submitForm()
	}

	if g.formMode == types.FormDueDate {
		if key == "enter" {
			return g.submitDueDateForm()
		}
		var cmd tea.Cmd
		g.dueDateInput, cmd = g.dueDateInput.Update(msg)
		return g, cmd
	}

	if g.formMode == types.FormEstimate {
		if key == "enter" {
			return g.submitEstimateForm()
		}
		var cmd tea.Cmd
		g.estimateInput, cmd = g.estimateInput.Update(msg)
		return g, cmd
	}

	if g.formMode == types.FormComment || g.formMode == types.FormEditComment {
		var cmd tea.Cmd
		g.commentBody, cmd = g.commentBody.Update(msg)
		return g, cmd
	}

	switch key {
	case "tab":
		g.nextFormField()
		return g, nil
	case "shift+tab":
		g.prevFormField()
		return g, nil
	}

	var cmd tea.Cmd
	switch g.formField {
	case types.FieldTitle:
		g.formTitle, cmd = g.formTitle.Update(msg)
	case types.FieldDescription:
		g.formDescription, cmd = g.formDescription.Update(msg)
	case types.FieldTeam, types.FieldAssignee, types.FieldProject:
		if key == "enter" {
			g.openFormListPopup(g.formField)
		}
	case types.FieldState:
		maxIdx := len(g.State.TeamStates)
		switch key {
		case "j", "down":
			if g.formState < maxIdx-1 {
				g.formState++
			}
		case "k", "up":
			if g.formState >= 0 {
				g.formState--
			}
		}
	case types.FieldPriority:
		switch key {
		case "j", "down":
			if g.formPriority < 4 {
				g.formPriority++
			}
		case "k", "up":
			if g.formPriority > 0 {
				g.formPriority--
			}
		}
	}

	return g, cmd
}

func (g *Gui) openFormListPopup(field types.FormField) {
	g.formListOpen = true
	g.formListField = field
	g.formListSearch.SetValue("")
	g.formListSearch.Focus()
	g.formListSelected = 0
}

func (g *Gui) closeFormListPopup() {
	g.formListOpen = false
	g.formListSearch.SetValue("")
	g.formListSearch.Blur()
	g.formListSelected = 0
}

func (g *Gui) handleFormListKey(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		g.closeFormListPopup()
		return g, nil
	case "enter":
		switch g.formListField {
		case types.FieldTeam:
			filtered := g.getFilteredTeams()
			if g.formListSelected < len(filtered) {
				g.formTeam = filtered[g.formListSelected]
				g.closeFormListPopup()
				return g, tea.Batch(g.loadTeamMembers(), g.loadFormTeamStates())
			}
		case types.FieldAssignee:
			filtered := g.getFilteredAssignees()
			if g.formListSelected < len(filtered) {
				g.formAssignee = filtered[g.formListSelected] - 1
				g.closeFormListPopup()
			}
		case types.FieldProject:
			filtered := g.getFilteredProjects()
			if g.formListSelected < len(filtered) {
				g.formProject = filtered[g.formListSelected] - 1
				g.closeFormListPopup()
			}
		}
		return g, nil
	case "j", "down":
		maxLen := g.getFilteredListLen()
		if g.formListSelected < maxLen-1 {
			g.formListSelected++
		}
		return g, nil
	case "k", "up":
		if g.formListSelected > 0 {
			g.formListSelected--
		}
		return g, nil
	case "ctrl+n":
		maxLen := g.getFilteredListLen()
		if g.formListSelected < maxLen-1 {
			g.formListSelected++
		}
		return g, nil
	case "ctrl+p":
		if g.formListSelected > 0 {
			g.formListSelected--
		}
		return g, nil
	default:
		var cmd tea.Cmd
		g.formListSearch, cmd = g.formListSearch.Update(msg)
		g.formListSelected = 0
		return g, cmd
	}
}

func (g *Gui) getFilteredListLen() int {
	switch g.formListField {
	case types.FieldTeam:
		return len(g.getFilteredTeams())
	case types.FieldAssignee:
		return len(g.getFilteredAssignees())
	case types.FieldProject:
		return len(g.getFilteredProjects())
	}
	return 0
}

func (g *Gui) getFilteredTeams() []int {
	query := strings.ToLower(g.formListSearch.Value())
	var filtered []int
	for i, t := range g.State.Teams {
		if query == "" || strings.Contains(strings.ToLower(t.Name), query) {
			filtered = append(filtered, i)
		}
	}
	return filtered
}

func (g *Gui) getFilteredAssignees() []int {
	query := strings.ToLower(g.formListSearch.Value())
	var filtered []int
	if query == "" || strings.Contains("unassigned", query) {
		filtered = append(filtered, 0)
	}
	for i, m := range g.State.TeamMembers {
		if query == "" || strings.Contains(strings.ToLower(m.DisplayName), query) || strings.Contains(strings.ToLower(m.Name), query) {
			filtered = append(filtered, i+1)
		}
	}
	return filtered
}

func (g *Gui) getFilteredProjects() []int {
	query := strings.ToLower(g.formListSearch.Value())
	var filtered []int
	if query == "" || strings.Contains("none", query) {
		filtered = append(filtered, 0)
	}
	for i, p := range g.State.Projects {
		if query == "" || strings.Contains(strings.ToLower(p.Name), query) {
			filtered = append(filtered, i+1)
		}
	}
	return filtered
}

func (g *Gui) nextFormField() {
	switch g.formField {
	case types.FieldTitle:
		g.formTitle.Blur()
		g.formDescription.Focus()
		g.formField = types.FieldDescription
	case types.FieldDescription:
		g.formDescription.Blur()
		g.formField = types.FieldTeam
	case types.FieldTeam:
		g.formField = types.FieldAssignee
	case types.FieldAssignee:
		g.formField = types.FieldState
	case types.FieldState:
		g.formField = types.FieldPriority
	case types.FieldPriority:
		g.formField = types.FieldProject
	case types.FieldProject:
		g.formField = types.FieldTitle
		g.formTitle.Focus()
	}
}

func (g *Gui) prevFormField() {
	switch g.formField {
	case types.FieldTitle:
		g.formTitle.Blur()
		g.formField = types.FieldProject
	case types.FieldDescription:
		g.formDescription.Blur()
		g.formTitle.Focus()
		g.formField = types.FieldTitle
	case types.FieldTeam:
		g.formDescription.Focus()
		g.formField = types.FieldDescription
	case types.FieldAssignee:
		g.formField = types.FieldTeam
	case types.FieldState:
		g.formField = types.FieldAssignee
	case types.FieldPriority:
		g.formField = types.FieldState
	case types.FieldProject:
		g.formField = types.FieldPriority
	}
}

func (g *Gui) submitForm() (tea.Model, tea.Cmd) {
	if g.formMode == types.FormComment {
		body := strings.TrimSpace(g.commentBody.Value())
		if body == "" {
			return g, func() tea.Msg { return types.StatusMsg("Comment cannot be empty") }
		}
		if g.State.DetailedIssue == nil {
			return g, func() tea.Msg { return types.StatusMsg("No issue selected") }
		}
		return g, g.createComment(g.State.DetailedIssue.ID, body)
	} else if g.formMode == types.FormEditComment {
		body := strings.TrimSpace(g.commentBody.Value())
		if body == "" {
			return g, func() tea.Msg { return types.StatusMsg("Comment cannot be empty") }
		}
		return g, g.updateComment(g.editingCommentID, body)
	}

	title := strings.TrimSpace(g.formTitle.Value())
	if title == "" {
		return g, func() tea.Msg { return types.StatusMsg("Title is required") }
	}

	if g.formMode == types.FormCreate {
		return g, g.createIssue()
	} else if g.formMode == types.FormEdit {
		description := g.formDescription.Value()
		return g, g.editIssue(g.formEditIssueID, title, description)
	}

	g.closeForm()
	return g, nil
}

func (g *Gui) createComment(issueID, body string) tea.Cmd {
	return func() tea.Msg {
		comment, err := g.client.Comments.Create(context.Background(), issueID, body)
		if err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to create comment: %v", err))
		}
		return types.CommentCreatedMsg{Comment: comment}
	}
}

func (g *Gui) updateComment(commentID, body string) tea.Cmd {
	return func() tea.Msg {
		if err := g.client.Comments.Update(context.Background(), commentID, body); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update comment: %v", err))
		}
		return types.CommentCreatedMsg{}
	}
}

func (g *Gui) createIssue() tea.Cmd {
	title := strings.TrimSpace(g.formTitle.Value())
	description := g.formDescription.Value()
	priority := g.formPriority

	var teamID string
	if g.formTeam >= 0 && g.formTeam < len(g.State.Teams) {
		teamID = g.State.Teams[g.formTeam].ID
	} else if g.State.SelectedTeam < len(g.State.Teams) {
		teamID = g.State.Teams[g.State.SelectedTeam].ID
	} else {
		return func() tea.Msg { return types.StatusMsg("No team selected") }
	}

	var assigneeID, stateID, projectID *string

	if g.formAssignee >= 0 && g.formAssignee < len(g.State.TeamMembers) {
		assigneeID = &g.State.TeamMembers[g.formAssignee].ID
	}

	if g.formState >= 0 && g.formState < len(g.State.TeamStates) {
		stateID = &g.State.TeamStates[g.formState].ID
	}

	if g.formProject >= 0 && g.formProject < len(g.State.Projects) {
		projectID = &g.State.Projects[g.formProject].ID
	}

	return func() tea.Msg {
		input := linear.IssueCreateInput{
			TeamID:      teamID,
			Title:       title,
			Description: description,
			Priority:    priority,
			AssigneeID:  assigneeID,
			StateID:     stateID,
			ProjectID:   projectID,
		}

		issue, err := g.client.Issues.Create(context.Background(), input)
		if err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to create issue: %v", err))
		}

		g.closeForm()
		return types.StatusMsg(fmt.Sprintf("Created %s", issue.Identifier))
	}
}

func (g *Gui) editIssue(issueID, title, description string) tea.Cmd {
	return func() tea.Msg {
		if err := g.client.Issues.Update(context.Background(), issueID, &title, &description); err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to update issue: %v", err))
		}

		g.closeForm()
		return types.ReloadIssuesMsg{}
	}
}

func (g *Gui) openDueDateInput(issue *models.Issue) {
	ti := textinput.New()
	if issue.DueDate != nil {
		ti.SetValue(*issue.DueDate)
	}
	ti.Placeholder = "YYYY-MM-DD (or empty to clear)"
	ti.Focus()
	ti.CharLimit = 20
	ti.Width = 30

	g.formMode = types.FormDueDate
	g.dueDateInput = ti
	g.inputIssueID = issue.ID
}

func (g *Gui) openEstimateInput(issue *models.Issue) {
	ti := textinput.New()
	if issue.Estimate != nil {
		ti.SetValue(fmt.Sprintf("%.0f", *issue.Estimate))
	}
	ti.Placeholder = "Story points (or empty to clear)"
	ti.Focus()
	ti.CharLimit = 10
	ti.Width = 30

	g.formMode = types.FormEstimate
	g.estimateInput = ti
	g.inputIssueID = issue.ID
}

func (g *Gui) openParentPicker(issue *models.Issue) tea.Cmd {
	if issue.Team == nil {
		return func() tea.Msg {
			return types.StatusMsg("No team for issue")
		}
	}

	issueCopy := issue
	return func() tea.Msg {
		filter := models.IssueFilter{TeamID: issueCopy.Team.ID}
		issues, err := g.client.Issues.List(context.Background(), filter)
		if err != nil {
			return types.StatusMsg(fmt.Sprintf("Failed to load issues: %v", err))
		}
		return types.ParentPickerMsg{Issues: issues, Issue: issueCopy}
	}
}

func (g *Gui) openCreateSubIssueFormCmd(parentIssue *models.Issue) tea.Cmd {
	g.openCreateForm()
	g.formEditIssueID = parentIssue.ID
	return tea.Batch(g.loadTeamMembers(), g.loadProjects())
}

func (g *Gui) submitDueDateForm() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(g.dueDateInput.Value())
	issueID := g.inputIssueID

	var dueDate *string
	if value != "" {
		if value == "tomorrow" {
			tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
			dueDate = &tomorrow
		} else if strings.HasPrefix(value, "+") {
			days := 0
			fmt.Sscanf(value, "+%dd", &days)
			if days > 0 {
				future := time.Now().AddDate(0, 0, days).Format("2006-01-02")
				dueDate = &future
			} else {
				dueDate = &value
			}
		} else {
			dueDate = &value
		}
	}

	g.closeForm()
	return g, g.updateIssueDueDate(issueID, dueDate)
}

func (g *Gui) submitEstimateForm() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(g.estimateInput.Value())
	issueID := g.inputIssueID

	var estimate *float64
	if value != "" {
		var e float64
		if _, err := fmt.Sscanf(value, "%f", &e); err == nil {
			estimate = &e
		} else {
			g.closeForm()
			return g, func() tea.Msg { return types.StatusMsg("Invalid estimate value") }
		}
	}

	g.closeForm()
	return g, g.updateIssueEstimate(issueID, estimate)
}
