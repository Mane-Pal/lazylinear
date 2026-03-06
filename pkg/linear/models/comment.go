package models

import "time"

// Comment represents a comment on an issue
type Comment struct {
	ID        string    `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	User      *User     `json:"user"`
}
