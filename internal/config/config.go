package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	APIBaseURL string `json:"apibaseurl"`
	Token      string `json:"token"`
	Workspace  string `json:"workspace"`
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

func configDir() string {
	if dir := os.Getenv("EXERCISM_CONFIG_HOME"); dir != "" {
		return dir
	}
	// Match exercism CLI behavior: always use ~/.config on all platforms
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config")
}
