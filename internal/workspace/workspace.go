package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rcopra/coding-tui/internal/api"
)

type Workspace struct {
	root      string
	client    *api.Client
	dismissed map[string]bool // track/slug → true
}

func New(root string, client *api.Client) *Workspace {
	w := &Workspace{root: root, client: client}
	w.loadDismissed()
	return w
}

func (w *Workspace) dismissedPath() string {
	return filepath.Join(w.root, ".coding-tui-dismissed.json")
}

func (w *Workspace) loadDismissed() {
	w.dismissed = make(map[string]bool)
	data, err := os.ReadFile(w.dismissedPath())
	if err != nil {
		return
	}
	var slugs []string
	if json.Unmarshal(data, &slugs) == nil {
		for _, s := range slugs {
			w.dismissed[s] = true
		}
	}
}

func (w *Workspace) saveDismissed() {
	slugs := make([]string, 0, len(w.dismissed))
	for s := range w.dismissed {
		slugs = append(slugs, s)
	}
	data, _ := json.Marshal(slugs)
	_ = os.WriteFile(w.dismissedPath(), data, 0o644)
}

func (w *Workspace) IsDismissed(trackSlug, exerciseSlug string) bool {
	return w.dismissed[trackSlug+"/"+exerciseSlug]
}

func (w *Workspace) ToggleDismiss(trackSlug, exerciseSlug string) bool {
	key := trackSlug + "/" + exerciseSlug
	if w.dismissed[key] {
		delete(w.dismissed, key)
	} else {
		w.dismissed[key] = true
	}
	w.saveDismissed()
	return w.dismissed[key]
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

	// Auto-install dependencies for tracks that need it
	w.installDeps(trackSlug, dir)

	// Silence noisy linter rules on exercise stubs
	w.patchLinterConfig(trackSlug, dir)

	return dir, nil
}

// installDeps installs shared dependencies at the track root level.
// Node's module resolution walks up the directory tree, so installing
// once at ~/exercism/javascript/ serves all exercises under it.
// Only runs if node_modules doesn't already exist at the track root.
func (w *Workspace) installDeps(trackSlug, exerciseDir string) {
	switch trackSlug {
	case "javascript", "typescript":
		trackRoot := filepath.Dir(exerciseDir)
		nodeModules := filepath.Join(trackRoot, "node_modules")

		// Already installed at track root — nothing to do
		if _, err := os.Stat(nodeModules); err == nil {
			return
		}

		// Copy the exercise's package.json to the track root and install there
		pkgSrc := filepath.Join(exerciseDir, "package.json")
		pkgDst := filepath.Join(trackRoot, "package.json")
		if _, err := os.Stat(pkgDst); err != nil {
			if data, err := os.ReadFile(pkgSrc); err == nil {
				_ = os.WriteFile(pkgDst, data, 0o644)
			}
		}

		cmd := exec.Command("npm", "install")
		cmd.Dir = trackRoot
		_ = cmd.Run()
	}
}

// patchLinterConfig silences noisy linter rules in the exercise directory.
// For JS/TS: patches the exercise's own eslint.config.mjs at the <<inject-rules-here>> marker.
// For other languages: writes config files at the track root.
func (w *Workspace) patchLinterConfig(trackSlug, exerciseDir string) {
	switch trackSlug {
	case "javascript", "typescript":
		w.patchESLintConfig(exerciseDir)
	default:
		// Other languages: write at track root (their linters DO cascade)
		trackRoot := filepath.Dir(exerciseDir)
		if overrides, ok := trackRootLinterOverrides[trackSlug]; ok {
			for filename, content := range overrides {
				path := filepath.Join(trackRoot, filename)
				if _, err := os.Stat(path); err == nil {
					continue
				}
				_ = os.MkdirAll(filepath.Dir(path), 0o755)
				_ = os.WriteFile(path, []byte(content), 0o644)
			}
		}
	}
}

const eslintInjectMarker = "// <<inject-rules-here>>"
const eslintRuleOverrides = `// <<inject-rules-here>> (patched by coding-tui)
  {
    rules: {
      "no-unused-vars": "off",
      "@typescript-eslint/no-unused-vars": "off",
      "no-constant-condition": "off",
      "no-empty-function": "off",
      "no-unreachable": "off",
    },
  },`

// patchESLintConfig injects rule overrides into an exercise's eslint.config.mjs
// at the <<inject-rules-here>> marker that exercism includes in every config.
func (w *Workspace) patchESLintConfig(exerciseDir string) {
	configPath := filepath.Join(exerciseDir, "eslint.config.mjs")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return
	}
	content := string(data)
	if !strings.Contains(content, eslintInjectMarker) {
		return // no marker or already patched
	}
	if strings.Contains(content, "patched by coding-tui") {
		return // already done
	}
	content = strings.Replace(content, eslintInjectMarker, eslintRuleOverrides, 1)
	_ = os.WriteFile(configPath, []byte(content), 0o644)
}

// Track-root linter configs for languages where configs cascade upward.
var trackRootLinterOverrides = map[string]map[string]string{
	"python": {
		"setup.cfg": `# Auto-generated by coding-tui
[flake8]
ignore = F841,W503,E501,F401
`,
		".pylintrc": `# Auto-generated by coding-tui
[MESSAGES CONTROL]
disable=unused-variable,unused-argument,unused-import,missing-docstring,too-few-public-methods
`,
	},
	"rust": {
		".cargo/config.toml": `# Auto-generated by coding-tui
[build]
rustflags = ["-A", "unused_variables", "-A", "dead_code", "-A", "unused_imports"]
`,
	},
	"ruby": {
		".rubocop.yml": `# Auto-generated by coding-tui
AllCops:
  DisabledByDefault: true
Lint/UnusedMethodArgument:
  Enabled: false
Lint/UnusedBlockArgument:
  Enabled: false
Lint/UselessAssignment:
  Enabled: false
Style/FrozenStringLiteralComment:
  Enabled: false
Metrics:
  Enabled: false
`,
	},
	"go": {
		"staticcheck.conf": `# Auto-generated by coding-tui
checks = ["all", "-U1000"]
`,
	},
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

// CompleteSolution marks an exercise as complete on Exercism.
func (w *Workspace) CompleteSolution(trackSlug, exerciseSlug string) error {
	meta, err := w.ReadMetadata(trackSlug, exerciseSlug)
	if err != nil {
		return err
	}
	if err := w.client.CompleteSolution(meta.ID); err != nil {
		return err
	}
	w.client.InvalidateExercises(trackSlug)
	return nil
}

// SolutionFilePath returns the absolute path to the primary solution file.
func (w *Workspace) SolutionFilePath(trackSlug, exerciseSlug string) (string, error) {
	cfg, err := w.ReadConfig(trackSlug, exerciseSlug)
	if err != nil {
		return "", err
	}
	if len(cfg.Files.Solution) == 0 {
		return "", fmt.Errorf("no solution files defined for %s/%s", trackSlug, exerciseSlug)
	}
	return filepath.Join(w.ExerciseDir(trackSlug, exerciseSlug), cfg.Files.Solution[0]), nil
}
