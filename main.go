package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/rcopra/coding-tui/internal/api"
	"github.com/rcopra/coding-tui/internal/cache"
	"github.com/rcopra/coding-tui/internal/config"
	"github.com/rcopra/coding-tui/internal/ui"
	"github.com/rcopra/coding-tui/internal/workspace"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cacheDir := filepath.Join(cacheBaseDir(), "coding-tui")
	c := cache.New(cacheDir, 10*time.Minute)

	client := api.NewClient(cfg.Token, c)
	ws := workspace.New(cfg.Workspace, client)
	tracks := ui.NewTracksScreen(client, ws)
	root := ui.NewRoot(tracks)

	p := tea.NewProgram(root)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func cacheBaseDir() string {
	if dir, err := os.UserCacheDir(); err == nil {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache")
}
