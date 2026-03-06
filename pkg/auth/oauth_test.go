package auth

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoginFailsWithoutCredentials(t *testing.T) {
	origID, origSecret := ClientID, ClientSecret
	ClientID, ClientSecret = "", ""
	defer func() { ClientID, ClientSecret = origID, origSecret }()

	err := Login()
	if err == nil {
		t.Fatal("expected error when ClientID/ClientSecret are empty")
	}
	if !strings.Contains(err.Error(), "OAuth credentials not configured") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLogoutDeletesToken(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	// Save a token first
	token := &Token{
		AccessToken:  "test_access",
		RefreshToken: "test_refresh",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}
	if err := SaveToken(token); err != nil {
		t.Fatalf("save token: %v", err)
	}

	// Verify it exists
	if _, err := LoadToken(); err != nil {
		t.Fatalf("token should exist: %v", err)
	}

	// Logout will try to POST to RevokeURL (will fail since it's the real URL),
	// but should still delete the local token
	_ = Logout()

	// Token file should be gone
	if _, err := LoadToken(); err == nil {
		t.Error("token should be deleted after logout")
	}
}

func TestLogoutWhenNotAuthenticated(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	// No token saved — Logout should succeed without error
	if err := Logout(); err != nil {
		t.Fatalf("logout with no token should not error: %v", err)
	}
}

func TestCredentialsRefreshWithExpiredToken(t *testing.T) {
	// Test that Refresh() attempts refresh when token is expired.
	// Without a real/mock token endpoint it will fail, but we verify it tries.
	creds := &Credentials{
		HeaderValue: "Bearer old_token",
		token: &Token{
			AccessToken:  "old_token",
			RefreshToken: "refresh_token",
			ExpiresAt:    time.Now().Add(-1 * time.Hour), // expired
		},
	}

	err := creds.Refresh()
	// Will fail because it tries to hit the real TokenURL, but should return an error
	// (not panic or silently ignore the expired token)
	if err == nil {
		t.Error("expected error when refreshing against unreachable token endpoint")
	}
}

func TestCredentialsRefreshSkipsValidToken(t *testing.T) {
	creds := &Credentials{
		HeaderValue: "Bearer valid_token",
		token: &Token{
			AccessToken: "valid_token",
			ExpiresAt:   time.Now().Add(1 * time.Hour), // still valid
		},
	}

	if err := creds.Refresh(); err != nil {
		t.Fatalf("Refresh should be no-op for valid token: %v", err)
	}
	// Header should remain unchanged
	if creds.HeaderValue != "Bearer valid_token" {
		t.Errorf("header changed unexpectedly: %s", creds.HeaderValue)
	}
}

func TestTokenResponseParsing(t *testing.T) {
	// Verify the tokenResponse struct maps correctly from JSON
	raw := `{"access_token":"acc","refresh_token":"ref","expires_in":7200,"token_type":"Bearer","scope":"read,write"}`
	var resp tokenResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.AccessToken != "acc" {
		t.Errorf("access_token: got %s", resp.AccessToken)
	}
	if resp.RefreshToken != "ref" {
		t.Errorf("refresh_token: got %s", resp.RefreshToken)
	}
	if resp.ExpiresIn != 7200 {
		t.Errorf("expires_in: got %d", resp.ExpiresIn)
	}
	if resp.TokenType != "Bearer" {
		t.Errorf("token_type: got %s", resp.TokenType)
	}
	if resp.Scope != "read,write" {
		t.Errorf("scope: got %s", resp.Scope)
	}
}
