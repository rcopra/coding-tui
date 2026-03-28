package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the exercism API configuration (from exercism CLI's user.json).
type Config struct {
	APIBaseURL string `json:"apibaseurl"`
	Token      string `json:"token"`
	Workspace  string `json:"workspace"`
}

// GymConfig holds gym-specific settings (from ~/.config/gym/config.json).
type GymConfig struct {
	// Style is a glamour style name ("dark", "light", "dracula", "tokyo-night", "pink")
	// or a path to a custom JSON style file.
	Style string `json:"style"`
}

func Load() (*Config, error) {
	path := filepath.Join(configDir(), "exercism", "user.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading exercism config at %s: %w\nRun 'exercism configure' first", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing exercism config: %w", err)
	}

	if cfg.Token == "" {
		return nil, fmt.Errorf("no token found in exercism config\nRun 'exercism configure --token=<token>' first")
	}

	return &cfg, nil
}

// LoadGym loads gym-specific config from ~/.config/gym/config.json.
// Returns defaults if the file doesn't exist.
func LoadGym() *GymConfig {
	path := filepath.Join(configDir(), "gym", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return &GymConfig{Style: "dark"}
	}
	var gc GymConfig
	if err := json.Unmarshal(data, &gc); err != nil {
		return &GymConfig{Style: "dark"}
	}
	if gc.Style == "" {
		gc.Style = "dark"
	}
	return &gc
}

func configDir() string {
	if dir := os.Getenv("EXERCISM_CONFIG_HOME"); dir != "" {
		return dir
	}
	// Match exercism CLI behavior: always use ~/.config on all platforms
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config")
}
