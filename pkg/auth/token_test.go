package auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTokenIsExpired(t *testing.T) {
	t.Run("future token is not expired", func(t *testing.T) {
		token := &Token{ExpiresAt: time.Now().Add(1 * time.Hour)}
		if token.IsExpired() {
			t.Error("token with 1h remaining should not be expired")
		}
	})

	t.Run("past token is expired", func(t *testing.T) {
		token := &Token{ExpiresAt: time.Now().Add(-1 * time.Hour)}
		if !token.IsExpired() {
			t.Error("token expired 1h ago should be expired")
		}
	})

	t.Run("token expiring within 5 minutes is considered expired", func(t *testing.T) {
		token := &Token{ExpiresAt: time.Now().Add(3 * time.Minute)}
		if !token.IsExpired() {
			t.Error("token expiring in 3 minutes should be considered expired (5 min margin)")
		}
	})

	t.Run("token expiring in 6 minutes is not expired", func(t *testing.T) {
		token := &Token{ExpiresAt: time.Now().Add(6 * time.Minute)}
		if token.IsExpired() {
			t.Error("token expiring in 6 minutes should not be expired")
		}
	})
}

func TestSaveLoadRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	original := &Token{
		AccessToken:  "acc_test_123",
		RefreshToken: "ref_test_456",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Truncate(time.Millisecond),
		Scope:        "read,write",
		TokenType:    "Bearer",
	}

	if err := SaveToken(original); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadToken()
	if err != nil {
		t.Fatalf("load error: %v", err)
	}

	if loaded.AccessToken != original.AccessToken {
		t.Errorf("access token: got %s, want %s", loaded.AccessToken, original.AccessToken)
	}
	if loaded.RefreshToken != original.RefreshToken {
		t.Errorf("refresh token: got %s, want %s", loaded.RefreshToken, original.RefreshToken)
	}
	if loaded.Scope != original.Scope {
		t.Errorf("scope: got %s, want %s", loaded.Scope, original.Scope)
	}
}

func TestSaveTokenPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	token := &Token{
		AccessToken: "test",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}

	if err := SaveToken(token); err != nil {
		t.Fatalf("save error: %v", err)
	}

	path := filepath.Join(tmpDir, "lazylinear", "credentials.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("expected permissions 0600, got %04o", perm)
	}
}

func TestDeleteToken(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	token := &Token{AccessToken: "test", ExpiresAt: time.Now().Add(1 * time.Hour)}
	SaveToken(token)

	if err := DeleteToken(); err != nil {
		t.Fatalf("delete error: %v", err)
	}

	if _, err := LoadToken(); err == nil {
		t.Error("expected error loading deleted token")
	}
}

func TestDeleteTokenNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	// Should not error when file doesn't exist
	if err := DeleteToken(); err != nil {
		t.Fatalf("delete non-existent should not error: %v", err)
	}
}
