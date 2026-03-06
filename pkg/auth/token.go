package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Token holds OAuth2 token data
type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
}

// IsExpired returns true if the token expires within 5 minutes
func (t *Token) IsExpired() bool {
	return time.Now().Add(5 * time.Minute).After(t.ExpiresAt)
}

// credentialsPath returns the path to the credentials file
func credentialsPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "lazylinear", "credentials.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "lazylinear", "credentials.json")
}

// SaveToken writes the token to disk with restricted permissions
func SaveToken(token *Token) error {
	path := credentialsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal token: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write credentials: %w", err)
	}
	return nil
}

// LoadToken reads the token from disk
func LoadToken() (*Token, error) {
	data, err := os.ReadFile(credentialsPath())
	if err != nil {
		return nil, err
	}

	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}
	return &token, nil
}

// DeleteToken removes the credentials file
func DeleteToken() error {
	path := credentialsPath()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove credentials: %w", err)
	}
	return nil
}
