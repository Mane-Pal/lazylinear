package config

// Config holds all application configuration
type Config struct {
	API      APIConfig  `yaml:"api"`
	Defaults Defaults   `yaml:"defaults"`
	UI       UIConfig   `yaml:"ui"`
	Keys     KeysConfig `yaml:"keys"`
}

// APIConfig holds API-related settings
type APIConfig struct {
	Timeout int `yaml:"timeout"` // milliseconds
}

// Defaults holds default filter/view settings
type Defaults struct {
	Team string `yaml:"team"`
	View string `yaml:"view"` // all, my_issues, unassigned
}

// UIConfig holds UI appearance settings
type UIConfig struct {
	Theme      string `yaml:"theme"`
	ShowIcons  bool   `yaml:"show_icons"`
	DateFormat string `yaml:"date_format"` // relative, absolute
}

// KeysConfig holds keybinding overrides
type KeysConfig struct {
	Quit    []string `yaml:"quit"`
	Help    []string `yaml:"help"`
	Refresh []string `yaml:"refresh"`
}

// DefaultConfig returns configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		API: APIConfig{
			Timeout: 30000,
		},
		Defaults: Defaults{
			View: "my_issues",
		},
		UI: UIConfig{
			Theme:      "default",
			ShowIcons:  true,
			DateFormat: "relative",
		},
		Keys: KeysConfig{
			Quit:    []string{"q", "ctrl+c"},
			Help:    []string{"?"},
			Refresh: []string{"r"},
		},
	}
}
