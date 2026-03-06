package types

// ContextKey identifies a context/panel
type ContextKey string

const (
	TeamsContext   ContextKey = "teams"
	FiltersContext ContextKey = "filters"
	IssuesContext  ContextKey = "issues"
	DetailContext  ContextKey = "detail"
	MenuContext    ContextKey = "menu"
	ConfirmContext ContextKey = "confirm"
	HelpContext    ContextKey = "help"
)

// ContextKind determines how contexts interact
type ContextKind int

const (
	SideContext ContextKind = iota
	MainContext
	PopupContext
)
