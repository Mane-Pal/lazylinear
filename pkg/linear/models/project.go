package models

import "time"

// Project represents a Linear project
type Project struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	State       string     `json:"state"` // planned, started, paused, completed, canceled
	Progress    float64    `json:"progress"`
	StartDate   *string    `json:"startDate"`
	TargetDate  *string    `json:"targetDate"`
	URL         string     `json:"url"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	Lead        *User      `json:"lead"`
	IssueCount  int        `json:"-"` // Computed from issues connection
}

// Cycle represents a Linear cycle (sprint)
type Cycle struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Number    int    `json:"number"`
	StartsAt  string `json:"startsAt"`
	EndsAt    string `json:"endsAt"`
}

// ProjectFilter defines filters for listing projects
type ProjectFilter struct {
	TeamID string
}
