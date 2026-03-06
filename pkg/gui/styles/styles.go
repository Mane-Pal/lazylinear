package styles

import "github.com/charmbracelet/lipgloss"

// Colors
var (
	Primary   = lipgloss.Color("#7aa2f7")
	Secondary = lipgloss.Color("#565f89")
	Success   = lipgloss.Color("#9ece6a")
	Warning   = lipgloss.Color("#e0af68")
	Error     = lipgloss.Color("#f7768e")
	BgDark    = lipgloss.Color("#1a1b26")
	FgLight   = lipgloss.Color("#c0caf5")
	BgSelect  = lipgloss.Color("#3b4261")

	// State colors
	StateBacklog    = lipgloss.Color("#565f89")
	StateTodo       = lipgloss.Color("#bb9af7")
	StateInProgress = lipgloss.Color("#7aa2f7")
	StateInReview   = lipgloss.Color("#e0af68")
	StateDone       = lipgloss.Color("#9ece6a")
	StateCanceled   = lipgloss.Color("#565f89")

	// Priority colors
	PriorityUrgent = lipgloss.Color("#f7768e")
	PriorityHigh   = lipgloss.Color("#ff9e64")
	PriorityMedium = lipgloss.Color("#e0af68")
	PriorityLow    = lipgloss.Color("#9ece6a")
)

// Base styles
var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary)

	Subtle = lipgloss.NewStyle().
		Foreground(Secondary)

	Selected = lipgloss.NewStyle().
			Background(Primary).
			Foreground(lipgloss.Color("#1a1b26")).
			Bold(true)

	// Panel styles
	Panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Secondary).
		Padding(0, 1)

	PanelFocused = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(0, 1)

	PanelTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Secondary)

	// Status bar
	StatusBar = lipgloss.NewStyle().
			Background(BgDark).
			Foreground(FgLight).
			Padding(0, 1)

	// Help
	HelpKey = lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true)

	HelpDesc = lipgloss.NewStyle().
			Foreground(Secondary)

	// Logo
	Logo = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary)

	// Error
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	// Warning style
	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning).
			Bold(true)
)

// StateColor returns the color for a workflow state type
func StateColor(stateType string) lipgloss.Color {
	switch stateType {
	case "backlog":
		return StateBacklog
	case "unstarted":
		return StateTodo
	case "started":
		return StateInProgress
	case "completed":
		return StateDone
	case "canceled":
		return StateCanceled
	default:
		return Secondary
	}
}

// PriorityColor returns the color for a priority level
func PriorityColor(priority int) lipgloss.Color {
	switch priority {
	case 1:
		return PriorityUrgent
	case 2:
		return PriorityHigh
	case 3:
		return PriorityMedium
	case 4:
		return PriorityLow
	default:
		return Secondary
	}
}

// PriorityLabel returns a label for a priority level
func PriorityLabel(priority int) string {
	switch priority {
	case 1:
		return "Urgent"
	case 2:
		return "High"
	case 3:
		return "Medium"
	case 4:
		return "Low"
	default:
		return "None"
	}
}

// StateIcon returns an icon for a state type
func StateIcon(stateType string) string {
	switch stateType {
	case "backlog":
		return "○"
	case "unstarted":
		return "○"
	case "started":
		return "●"
	case "completed":
		return "✓"
	case "canceled":
		return "✗"
	default:
		return "○"
	}
}
