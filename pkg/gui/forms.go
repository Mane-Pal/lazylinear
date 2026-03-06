package gui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mane-pal/lazylinear/pkg/linear"
	"github.com/mane-pal/lazylinear/pkg/linear/models"
)

// Form functions

func (g *Gui) openCreateForm() {
	// Calculate dynamic widths based on window size (70% of width, max 100)
	formWidth := (g.width * 70) / 100
	if formWidth > 100 {
		formWidth = 100
	}
	if formWidth < 70 {
		formWidth = 70
	}
	inputWidth := formWidth - 10

	// Initialize text input for title
	ti := textinput.New()
	ti.Placeholder = "Issue title..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = inputWidth

	// Initialize textarea for description (tall for comfortable editing)
	ta := textarea.New()
	ta.Placeholder = "Description (optional, supports markdown)..."
	ta.SetWidth(inputWidth)
	ta.SetHeight(12)

	// Initialize search input for list fields
	search := textinput.New()
	search.Placeholder = "Type to filter..."
	search.CharLimit = 50
	search.Width = 40

	g.formMode = FormCreate
	g.formField = FieldTitle
	g.formTitle = ti
	g.formDescription = ta
	g.formPriority = 0
	g.formTeam = g.selectedTeam      // Default to currently selected team
	g.formAssignee = -1              // Unassigned by default
	g.formState = -1                 // Default state
	g.formProject = -1               // No project by default
	g.formEditIssueID = ""
	g.formListSearch = search
	g.formListSelected = 0
}

// openCreateFormCmd returns a command to load data needed for the create form
func (g *Gui) openCreateFormCmd() tea.Cmd {
	g.openCreateForm()
	// Load team members and projects for the form
	return tea.Batch(g.loadTeamMembers(), g.loadProjects())
}

func (g *Gui) openEditForm(issue *models.Issue) {
	// Initialize text input with current title
	ti := textinput.New()
	ti.SetValue(issue.Title)
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 50

	// Initialize textarea with current description
	ta := textarea.New()
	ta.SetValue(issue.Description)
	ta.SetWidth(50)
	ta.SetHeight(8)

	g.formMode = FormEdit
	g.formField = FieldTitle
	g.formTitle = ti
	g.formDescription = ta
	g.formPriority = issue.Priority
	g.formEditIssueID = issue.ID
}

func (g *Gui) closeForm() {
	g.formMode = FormNone
	g.formEditIssueID = ""
	g.editingCommentID = ""
}

// Comment form functions

func (g *Gui) openCommentForm() {
	ta := textarea.New()
	ta.Placeholder = "Write a comment..."
	ta.SetWidth(50)
	ta.SetHeight(6)
	ta.Focus()

	g.formMode = FormComment
	g.commentBody = ta
}

func (g *Gui) openEditCommentForm(comment *models.Comment) {
	ta := textarea.New()
	ta.SetValue(comment.Body)
	ta.SetWidth(50)
	ta.SetHeight(6)
	ta.Focus()

	g.formMode = FormEditComment
	g.commentBody = ta
	g.editingCommentID = comment.ID
}

func (g *Gui) openDeleteCommentConfirm(comment *models.Comment) {
	g.menuType = MenuConfirm
	g.confirmTitle = "Delete Comment?"
	g.confirmMessage = fmt.Sprintf("Are you sure you want to delete this comment?\n\n%s", truncate(comment.Body, 50))
	commentID := comment.ID
	g.confirmAction = func() tea.Cmd {
		return g.deleteComment(commentID)
	}
}

func (g *Gui) deleteComment(commentID string) tea.Cmd {
	issueID := ""
	if g.detailedIssue != nil {
		issueID = g.detailedIssue.ID
	}
	return func() tea.Msg {
		if err := g.client.Comments.Delete(context.Background(), commentID); err != nil {
			return statusMsg(fmt.Sprintf("Failed to delete comment: %v", err))
		}
		return commentDeletedMsg{issueID: issueID}
	}
}

func (g *Gui) handleFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Handle list popup first if open (so Esc closes popup, not form)
	if g.formListOpen {
		return g.handleFormListKey(key, msg)
	}

	// Global form keys
	switch key {
	case "esc":
		g.closeForm()
		return g, nil
	case "ctrl+s", "ctrl+enter":
		return g.submitForm()
	}

	// Due date form
	if g.formMode == FormDueDate {
		if key == "enter" {
			return g.submitDueDateForm()
		}
		var cmd tea.Cmd
		g.dueDateInput, cmd = g.dueDateInput.Update(msg)
		return g, cmd
	}

	// Estimate form
	if g.formMode == FormEstimate {
		if key == "enter" {
			return g.submitEstimateForm()
		}
		var cmd tea.Cmd
		g.estimateInput, cmd = g.estimateInput.Update(msg)
		return g, cmd
	}

	// Comment forms only have one field
	if g.formMode == FormComment || g.formMode == FormEditComment {
		var cmd tea.Cmd
		g.commentBody, cmd = g.commentBody.Update(msg)
		return g, cmd
	}

	// Issue form navigation
	switch key {
	case "tab":
		g.nextFormField()
		return g, nil
	case "shift+tab":
		g.prevFormField()
		return g, nil
	}

	// Field-specific input handling for issue forms
	var cmd tea.Cmd
	switch g.formField {
	case FieldTitle:
		g.formTitle, cmd = g.formTitle.Update(msg)
	case FieldDescription:
		g.formDescription, cmd = g.formDescription.Update(msg)
	case FieldTeam, FieldAssignee, FieldProject:
		// Press Enter to open list popup
		if key == "enter" {
			g.openFormListPopup(g.formField)
		}
	case FieldState:
		maxIdx := len(g.teamStates)
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
	case FieldPriority:
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

func (g *Gui) openFormListPopup(field FormField) {
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
		// Select the highlighted item
		switch g.formListField {
		case FieldTeam:
			filtered := g.getFilteredTeams()
			if g.formListSelected < len(filtered) {
				g.formTeam = filtered[g.formListSelected]
				g.closeFormListPopup()
				return g, tea.Batch(g.loadTeamMembers(), g.loadFormTeamStates())
			}
		case FieldAssignee:
			filtered := g.getFilteredAssignees()
			if g.formListSelected < len(filtered) {
				g.formAssignee = filtered[g.formListSelected] - 1
				g.closeFormListPopup()
			}
		case FieldProject:
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
		// Pass to search input
		var cmd tea.Cmd
		g.formListSearch, cmd = g.formListSearch.Update(msg)
		g.formListSelected = 0 // Reset selection when search changes
		return g, cmd
	}
}

func (g *Gui) getFilteredListLen() int {
	switch g.formListField {
	case FieldTeam:
		return len(g.getFilteredTeams())
	case FieldAssignee:
		return len(g.getFilteredAssignees())
	case FieldProject:
		return len(g.getFilteredProjects())
	}
	return 0
}

// Helper functions for filtered lists
func (g *Gui) getFilteredTeams() []int {
	query := strings.ToLower(g.formListSearch.Value())
	var filtered []int
	for i, t := range g.teams {
		if query == "" || strings.Contains(strings.ToLower(t.Name), query) {
			filtered = append(filtered, i)
		}
	}
	return filtered
}

func (g *Gui) getFilteredAssignees() []int {
	query := strings.ToLower(g.formListSearch.Value())
	var filtered []int
	// Index 0 is "Unassigned"
	if query == "" || strings.Contains("unassigned", query) {
		filtered = append(filtered, 0)
	}
	for i, m := range g.teamMembers {
		if query == "" || strings.Contains(strings.ToLower(m.DisplayName), query) || strings.Contains(strings.ToLower(m.Name), query) {
			filtered = append(filtered, i+1) // +1 because 0 is "Unassigned"
		}
	}
	return filtered
}

func (g *Gui) getFilteredProjects() []int {
	query := strings.ToLower(g.formListSearch.Value())
	var filtered []int
	// Index 0 is "None"
	if query == "" || strings.Contains("none", query) {
		filtered = append(filtered, 0)
	}
	for i, p := range g.projects {
		if query == "" || strings.Contains(strings.ToLower(p.Name), query) {
			filtered = append(filtered, i+1) // +1 because 0 is "None"
		}
	}
	return filtered
}

func (g *Gui) nextFormField() {
	// Order: Title -> Description -> Team -> Assignee -> State -> Priority -> Project -> Title
	switch g.formField {
	case FieldTitle:
		g.formTitle.Blur()
		g.formDescription.Focus()
		g.formField = FieldDescription
	case FieldDescription:
		g.formDescription.Blur()
		g.formField = FieldTeam
	case FieldTeam:
		g.formField = FieldAssignee
	case FieldAssignee:
		g.formField = FieldState
	case FieldState:
		g.formField = FieldPriority
	case FieldPriority:
		g.formField = FieldProject
	case FieldProject:
		g.formField = FieldTitle
		g.formTitle.Focus()
	}
}

func (g *Gui) prevFormField() {
	switch g.formField {
	case FieldTitle:
		g.formTitle.Blur()
		g.formField = FieldProject
	case FieldDescription:
		g.formDescription.Blur()
		g.formTitle.Focus()
		g.formField = FieldTitle
	case FieldTeam:
		g.formDescription.Focus()
		g.formField = FieldDescription
	case FieldAssignee:
		g.formField = FieldTeam
	case FieldState:
		g.formField = FieldAssignee
	case FieldPriority:
		g.formField = FieldState
	case FieldProject:
		g.formField = FieldPriority
	}
}

func (g *Gui) submitForm() (tea.Model, tea.Cmd) {
	// Handle comment forms
	if g.formMode == FormComment {
		body := strings.TrimSpace(g.commentBody.Value())
		if body == "" {
			return g, func() tea.Msg { return statusMsg("Comment cannot be empty") }
		}
		if g.detailedIssue == nil {
			return g, func() tea.Msg { return statusMsg("No issue selected") }
		}
		return g, g.createComment(g.detailedIssue.ID, body)
	} else if g.formMode == FormEditComment {
		body := strings.TrimSpace(g.commentBody.Value())
		if body == "" {
			return g, func() tea.Msg { return statusMsg("Comment cannot be empty") }
		}
		return g, g.updateComment(g.editingCommentID, body)
	}

	// Handle issue forms
	title := strings.TrimSpace(g.formTitle.Value())
	if title == "" {
		return g, func() tea.Msg { return statusMsg("Title is required") }
	}

	if g.formMode == FormCreate {
		return g, g.createIssue()
	} else if g.formMode == FormEdit {
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
			return statusMsg(fmt.Sprintf("Failed to create comment: %v", err))
		}
		return commentCreatedMsg{comment}
	}
}

func (g *Gui) updateComment(commentID, body string) tea.Cmd {
	return func() tea.Msg {
		if err := g.client.Comments.Update(context.Background(), commentID, body); err != nil {
			return statusMsg(fmt.Sprintf("Failed to update comment: %v", err))
		}
		return commentCreatedMsg{} // Reuse to trigger reload
	}
}

func (g *Gui) createIssue() tea.Cmd {
	// Get form values
	title := strings.TrimSpace(g.formTitle.Value())
	description := g.formDescription.Value()
	priority := g.formPriority

	// Get team from form selection
	var teamID string
	if g.formTeam >= 0 && g.formTeam < len(g.teams) {
		teamID = g.teams[g.formTeam].ID
	} else if g.selectedTeam < len(g.teams) {
		teamID = g.teams[g.selectedTeam].ID
	} else {
		return func() tea.Msg { return statusMsg("No team selected") }
	}

	// Optional fields
	var assigneeID, stateID, projectID *string

	if g.formAssignee >= 0 && g.formAssignee < len(g.teamMembers) {
		assigneeID = &g.teamMembers[g.formAssignee].ID
	}

	if g.formState >= 0 && g.formState < len(g.teamStates) {
		stateID = &g.teamStates[g.formState].ID
	}

	if g.formProject >= 0 && g.formProject < len(g.projects) {
		projectID = &g.projects[g.formProject].ID
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
			return statusMsg(fmt.Sprintf("Failed to create issue: %v", err))
		}

		g.closeForm()
		return statusMsg(fmt.Sprintf("Created %s", issue.Identifier))
	}
}

func (g *Gui) editIssue(issueID, title, description string) tea.Cmd {
	return func() tea.Msg {
		if err := g.client.Issues.Update(context.Background(), issueID, &title, &description); err != nil {
			return statusMsg(fmt.Sprintf("Failed to update issue: %v", err))
		}

		g.closeForm()
		return reloadIssuesMsg{}
	}
}

// Due Date input functions

func (g *Gui) openDueDateInput(issue *models.Issue) {
	ti := textinput.New()
	if issue.DueDate != nil {
		ti.SetValue(*issue.DueDate)
	}
	ti.Placeholder = "YYYY-MM-DD (or empty to clear)"
	ti.Focus()
	ti.CharLimit = 20
	ti.Width = 30

	g.formMode = FormDueDate
	g.dueDateInput = ti
	g.inputIssueID = issue.ID
}

// Estimate input functions

func (g *Gui) openEstimateInput(issue *models.Issue) {
	ti := textinput.New()
	if issue.Estimate != nil {
		ti.SetValue(fmt.Sprintf("%.0f", *issue.Estimate))
	}
	ti.Placeholder = "Story points (or empty to clear)"
	ti.Focus()
	ti.CharLimit = 10
	ti.Width = 30

	g.formMode = FormEstimate
	g.estimateInput = ti
	g.inputIssueID = issue.ID
}

// Parent picker functions

type parentPickerMsg struct {
	issues []*models.Issue
	issue  *models.Issue
}

func (g *Gui) openParentPicker(issue *models.Issue) tea.Cmd {
	if issue.Team == nil {
		return func() tea.Msg {
			return statusMsg("No team for issue")
		}
	}

	issueCopy := issue
	return func() tea.Msg {
		// Fetch all issues for the team to pick parent from
		filter := models.IssueFilter{TeamID: issueCopy.Team.ID}
		issues, err := g.client.Issues.List(context.Background(), filter)
		if err != nil {
			return statusMsg(fmt.Sprintf("Failed to load issues: %v", err))
		}
		return parentPickerMsg{issues: issues, issue: issueCopy}
	}
}

// Sub-issue creation

func (g *Gui) openCreateSubIssueFormCmd(parentIssue *models.Issue) tea.Cmd {
	g.openCreateForm()
	// Set the parent ID (we'll need to store this)
	g.formEditIssueID = parentIssue.ID // Temporarily store parent ID
	return tea.Batch(g.loadTeamMembers(), g.loadProjects())
}

func (g *Gui) submitDueDateForm() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(g.dueDateInput.Value())
	issueID := g.inputIssueID

	var dueDate *string
	if value != "" {
		// Parse relative dates
		if value == "tomorrow" {
			tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
			dueDate = &tomorrow
		} else if strings.HasPrefix(value, "+") {
			// Parse +Nd format (e.g., +3d for 3 days)
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
			return g, func() tea.Msg { return statusMsg("Invalid estimate value") }
		}
	}

	g.closeForm()
	return g, g.updateIssueEstimate(issueID, estimate)
}
