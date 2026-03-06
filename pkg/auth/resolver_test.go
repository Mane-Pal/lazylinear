package auth

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestResolveOAuthToken(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	token := &Token{
		AccessToken: "oauth_test_token",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}
	if err := SaveToken(token); err != nil {
		t.Fatalf("save token: %v", err)
	}

	creds, err := Resolve()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.HeaderValue != "Bearer oauth_test_token" {
		t.Errorf("expected Bearer header, got %s", creds.HeaderValue)
	}
}

func TestResolveNoCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	_, err := Resolve()
	if err == nil {
		t.Fatal("expected error when no credentials available")
	}
	if !strings.Contains(err.Error(), "not authenticated") {
		t.Errorf("expected descriptive error, got: %v", err)
	}
}

func TestCredentialsRefreshNoToken(t *testing.T) {
	creds := &Credentials{HeaderValue: "Bearer test"}
	// No token set — Refresh should be a no-op
	if err := creds.Refresh(); err != nil {
		t.Fatalf("Refresh with no token should be no-op: %v", err)
	}
}
