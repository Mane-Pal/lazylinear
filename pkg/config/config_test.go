package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.API.Timeout != 30000 {
		t.Errorf("expected timeout 30000, got %d", cfg.API.Timeout)
	}
	if cfg.Defaults.View != "my_issues" {
		t.Errorf("expected view 'my_issues', got %s", cfg.Defaults.View)
	}
	if cfg.UI.Theme != "default" {
		t.Errorf("expected theme 'default', got %s", cfg.UI.Theme)
	}
	if !cfg.UI.ShowIcons {
		t.Error("expected ShowIcons true")
	}
	if cfg.UI.DateFormat != "relative" {
		t.Errorf("expected date format 'relative', got %s", cfg.UI.DateFormat)
	}
	if len(cfg.Keys.Quit) != 2 || cfg.Keys.Quit[0] != "q" {
		t.Errorf("expected quit keys [q, ctrl+c], got %v", cfg.Keys.Quit)
	}
	if len(cfg.Keys.Help) != 1 || cfg.Keys.Help[0] != "?" {
		t.Errorf("expected help keys [?], got %v", cfg.Keys.Help)
	}
	if len(cfg.Keys.Refresh) != 1 || cfg.Keys.Refresh[0] != "r" {
		t.Errorf("expected refresh keys [r], got %v", cfg.Keys.Refresh)
	}
}

func TestLoadSucceeds(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should have defaults
	if cfg.API.Timeout != 30000 {
		t.Errorf("expected default timeout, got %d", cfg.API.Timeout)
	}
}

func TestLoadFromConfigFile(t *testing.T) {
	// Create a temp config dir
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "lazylinear")
	os.MkdirAll(configDir, 0o755)
	configPath := filepath.Join(configDir, "config.yaml")

	// Write a custom config
	cfg := &Config{
		API: APIConfig{Timeout: 5000},
		Defaults: Defaults{
			Team: "Engineering",
			View: "all",
		},
		UI: UIConfig{
			Theme:      "dark",
			ShowIcons:  false,
			DateFormat: "absolute",
		},
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}

	// Set XDG_CONFIG_HOME to our temp dir so Load() finds the file
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if origXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	loaded, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loaded.API.Timeout != 5000 {
		t.Errorf("expected timeout 5000, got %d", loaded.API.Timeout)
	}
	if loaded.Defaults.Team != "Engineering" {
		t.Errorf("expected team Engineering, got %s", loaded.Defaults.Team)
	}
	if loaded.Defaults.View != "all" {
		t.Errorf("expected view all, got %s", loaded.Defaults.View)
	}
	if loaded.UI.Theme != "dark" {
		t.Errorf("expected theme dark, got %s", loaded.UI.Theme)
	}
	if loaded.UI.ShowIcons {
		t.Error("expected ShowIcons false")
	}
	if loaded.UI.DateFormat != "absolute" {
		t.Errorf("expected date format absolute, got %s", loaded.UI.DateFormat)
	}
}

func TestConfigYAMLRoundTrip(t *testing.T) {
	original := DefaultConfig()
	original.Defaults.Team = "MyTeam"

	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var loaded Config
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if loaded.Defaults.Team != "MyTeam" {
		t.Errorf("expected MyTeam, got %s", loaded.Defaults.Team)
	}
	if loaded.API.Timeout != original.API.Timeout {
		t.Errorf("timeout mismatch: %d vs %d", loaded.API.Timeout, original.API.Timeout)
	}
}

func TestGetConfigPath(t *testing.T) {
	t.Run("uses XDG_CONFIG_HOME", func(t *testing.T) {
		orig := os.Getenv("XDG_CONFIG_HOME")
		os.Setenv("XDG_CONFIG_HOME", "/custom/config")
		defer func() {
			if orig != "" {
				os.Setenv("XDG_CONFIG_HOME", orig)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}
		}()

		path := getConfigPath()
		expected := "/custom/config/lazylinear/config.yaml"
		if path != expected {
			t.Errorf("expected %s, got %s", expected, path)
		}
	})

	t.Run("falls back to home dir", func(t *testing.T) {
		orig := os.Getenv("XDG_CONFIG_HOME")
		os.Unsetenv("XDG_CONFIG_HOME")
		defer func() {
			if orig != "" {
				os.Setenv("XDG_CONFIG_HOME", orig)
			}
		}()

		path := getConfigPath()
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".config", "lazylinear", "config.yaml")
		if path != expected {
			t.Errorf("expected %s, got %s", expected, path)
		}
	})
}
