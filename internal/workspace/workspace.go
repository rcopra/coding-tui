package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rcopra/coding-tui/internal/api"
)

type Workspace struct {
	root   string
	client *api.Client
}

func New(root string, client *api.Client) *Workspace {
	return &Workspace{root: root, client: client}
}

// ExerciseDir returns the path where an exercise's files live.
func (w *Workspace) ExerciseDir(trackSlug, exerciseSlug string) string {
	return filepath.Join(w.root, trackSlug, exerciseSlug)
}

// IsDownloaded checks if an exercise has been downloaded locally.
func (w *Workspace) IsDownloaded(trackSlug, exerciseSlug string) bool {
	dir := w.ExerciseDir(trackSlug, exerciseSlug)
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// Download fetches all exercise files from the API and writes them to the workspace.
// Returns the exercise directory path.
func (w *Workspace) Download(trackSlug, exerciseSlug string) (string, error) {
	sol, err := w.client.GetLatestSolution(trackSlug, exerciseSlug)
	if err != nil {
		return "", fmt.Errorf("getting solution: %w", err)
	}

	dir := w.ExerciseDir(trackSlug, exerciseSlug)

	for _, file := range sol.Files {
		data, err := w.client.GetSolutionFile(sol.FileDownloadBaseURL, file)
		if err != nil {
			return "", fmt.Errorf("downloading %s: %w", file, err)
		}

		dest := filepath.Join(dir, file)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return "", fmt.Errorf("creating directory for %s: %w", file, err)
		}

		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return "", fmt.Errorf("writing %s: %w", file, err)
		}
	}

	return dir, nil
}

// ReadInstructions reads the README.md from a downloaded exercise.
// Falls back to fetching from the API if not downloaded.
func (w *Workspace) ReadInstructions(trackSlug, exerciseSlug string) (string, error) {
	// Try local first
	dir := w.ExerciseDir(trackSlug, exerciseSlug)
	readmePath := filepath.Join(dir, "README.md")
	if data, err := os.ReadFile(readmePath); err == nil {
		return string(data), nil
	}

	// Fetch from API
	sol, err := w.client.GetLatestSolution(trackSlug, exerciseSlug)
	if err != nil {
		return "", fmt.Errorf("getting solution: %w", err)
	}

	for _, file := range sol.Files {
		if strings.EqualFold(filepath.Base(file), "readme.md") {
			data, err := w.client.GetSolutionFile(sol.FileDownloadBaseURL, file)
			if err != nil {
				return "", fmt.Errorf("fetching README: %w", err)
			}
			return string(data), nil
		}
	}

	return "", fmt.Errorf("no README found for %s/%s", trackSlug, exerciseSlug)
}

// ReadHints reads the HINTS.md from a downloaded exercise or the API.
func (w *Workspace) ReadHints(trackSlug, exerciseSlug string) (string, error) {
	// Try local first
	dir := w.ExerciseDir(trackSlug, exerciseSlug)
	hintsPath := filepath.Join(dir, "HINTS.md")
	if data, err := os.ReadFile(hintsPath); err == nil {
		return string(data), nil
	}

	// Fetch from API
	sol, err := w.client.GetLatestSolution(trackSlug, exerciseSlug)
	if err != nil {
		return "", err
	}

	for _, file := range sol.Files {
		if strings.EqualFold(filepath.Base(file), "hints.md") {
			data, err := w.client.GetSolutionFile(sol.FileDownloadBaseURL, file)
			if err != nil {
				return "", err
			}
			return string(data), nil
		}
	}

	return "", nil // No hints is not an error
}

// ExerciseMetadata from .exercism/metadata.json
type ExerciseMetadata struct {
	Track    string `json:"track"`
	Exercise string `json:"exercise"`
	ID       string `json:"id"`
	URL      string `json:"url"`
}

// ExerciseConfig from .exercism/config.json
type ExerciseConfig struct {
	Files struct {
		Solution []string `json:"solution"`
		Test     []string `json:"test"`
	} `json:"files"`
}

// ReadMetadata reads the .exercism/metadata.json for a downloaded exercise.
func (w *Workspace) ReadMetadata(trackSlug, exerciseSlug string) (*ExerciseMetadata, error) {
	path := filepath.Join(w.ExerciseDir(trackSlug, exerciseSlug), ".exercism", "metadata.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading metadata: %w", err)
	}
	var meta ExerciseMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parsing metadata: %w", err)
	}
	return &meta, nil
}

// ReadConfig reads the .exercism/config.json for a downloaded exercise.
func (w *Workspace) ReadConfig(trackSlug, exerciseSlug string) (*ExerciseConfig, error) {
	path := filepath.Join(w.ExerciseDir(trackSlug, exerciseSlug), ".exercism", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg ExerciseConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

// SubmitSolution uploads the solution files for an exercise.
func (w *Workspace) SubmitSolution(trackSlug, exerciseSlug string) error {
	meta, err := w.ReadMetadata(trackSlug, exerciseSlug)
	if err != nil {
		return err
	}

	config, err := w.ReadConfig(trackSlug, exerciseSlug)
	if err != nil {
		return err
	}

	dir := w.ExerciseDir(trackSlug, exerciseSlug)
	files := api.SolutionFilePaths(dir, config.Files.Solution)
	if len(files) == 0 {
		return fmt.Errorf("no solution files found in %s", dir)
	}

	return w.client.SubmitSolution(meta.ID, files)
}
