package auth

import (
	"fmt"
	"sync"
)

// Credentials holds the resolved authentication state
type Credentials struct {
	HeaderValue string

	mu    sync.Mutex
	token *Token
}

// Refresh checks and refreshes the OAuth token if expired.
func (c *Credentials) Refresh() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token == nil || !c.token.IsExpired() {
		return nil
	}

	newToken, err := RefreshAccessToken(c.token)
	if err != nil {
		return err
	}
	c.token = newToken
	c.HeaderValue = "Bearer " + newToken.AccessToken
	return nil
}

// Resolve loads the saved OAuth token and returns Credentials.
func Resolve() (*Credentials, error) {
	token, err := LoadToken()
	if err == nil && token.AccessToken != "" {
		return &Credentials{
			HeaderValue: "Bearer " + token.AccessToken,
			token:       token,
		}, nil
	}

	return nil, fmt.Errorf("not authenticated\n\nTo authenticate, run:\n  lazylinear auth login")
}
