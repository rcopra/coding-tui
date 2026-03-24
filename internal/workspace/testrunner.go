package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// TestResult holds the output of a local test run.
type TestResult struct {
	Passed bool
	Output string
}

// track slug → test command parts
var testCommands = map[string][]string{
	"go":         {"go", "test", "-v", "./..."},
	"rust":       {"cargo", "test"},
	"python":     {"python3", "-m", "pytest", "-v"},
	"javascript": {"npm", "test"},
	"typescript": {"npm", "test"},
	"ruby":       {"ruby", "-r", "minitest/autorun"},
	"elixir":     {"mix", "test"},
	"java":       {"gradle", "test"},
	"c":          {"make", "test"},
	"cpp":        {"make", "test"},
	"csharp":     {"dotnet", "test"},
	"swift":      {"swift", "test"},
	"kotlin":     {"gradle", "test"},
	"haskell":    {"stack", "test"},
	"clojure":    {"lein", "test"},
	"scala":      {"sbt", "test"},
	"lua":        {"busted"},
	"r":          {"Rscript", "test"},
	"bash":       {"bats"},
	"zig":        {"zig", "test"},
	"php":        {"phpunit"},
}

// Tracks that need dependency installation before tests can run.
var installCommands = map[string][]string{
	"javascript": {"npm", "install"},
	"typescript": {"npm", "install"},
}

// RunTests runs the test suite for an exercise locally.
// Automatically installs dependencies first for tracks that need it.
func (w *Workspace) RunTests(trackSlug, exerciseSlug string) (*TestResult, error) {
	dir := w.ExerciseDir(trackSlug, exerciseSlug)
	if _, err := os.Stat(dir); err != nil {
		return nil, fmt.Errorf("exercise not downloaded: %s", dir)
	}

	// Install shared dependencies at track root if needed
	if installCmd, ok := installCommands[trackSlug]; ok {
		trackRoot := filepath.Dir(dir)
		if _, err := os.Stat(filepath.Join(trackRoot, "node_modules")); err != nil {
			// Copy package.json to track root if missing
			pkgSrc := filepath.Join(dir, "package.json")
			pkgDst := filepath.Join(trackRoot, "package.json")
			if _, err := os.Stat(pkgDst); err != nil {
				if data, readErr := os.ReadFile(pkgSrc); readErr == nil {
					_ = os.WriteFile(pkgDst, data, 0o644)
				}
			}
			cmd := exec.Command(installCmd[0], installCmd[1:]...)
			cmd.Dir = trackRoot
			if out, err := cmd.CombinedOutput(); err != nil {
				return &TestResult{
					Passed: false,
					Output: fmt.Sprintf("Failed to install dependencies:\n%s", string(out)),
				}, nil
			}
		}
	}

	cmdParts, err := resolveTestCommand(trackSlug, dir)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	passed := err == nil

	return &TestResult{
		Passed: passed,
		Output: string(output),
	}, nil
}

func resolveTestCommand(trackSlug, dir string) ([]string, error) {
	// Special case: ruby needs the test file name
	if trackSlug == "ruby" {
		testFile, err := findRubyTestFile(dir)
		if err != nil {
			return nil, err
		}
		return []string{"ruby", testFile}, nil
	}

	// Special case: bash needs the test file
	if trackSlug == "bash" {
		testFile, err := findTestFileByExtension(dir, ".bats")
		if err != nil {
			return nil, err
		}
		return []string{"bats", testFile}, nil
	}

	parts, ok := testCommands[trackSlug]
	if !ok {
		return nil, fmt.Errorf("no test command configured for track %q\nRun tests manually in: %s", trackSlug, dir)
	}

	return parts, nil
}

func findRubyTestFile(dir string) (string, error) {
	// Check .exercism/config.json for test files
	configPath := filepath.Join(dir, ".exercism", "config.json")
	if data, err := os.ReadFile(configPath); err == nil {
		var config struct {
			Files struct {
				Test []string `json:"test"`
			} `json:"files"`
		}
		if err := json.Unmarshal(data, &config); err == nil && len(config.Files.Test) > 0 {
			return config.Files.Test[0], nil
		}
	}

	// Fallback: find *_test.rb
	return findTestFileByPattern(dir, "*_test.rb")
}

func findTestFileByExtension(dir, ext string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ext) {
			return e.Name(), nil
		}
	}
	return "", fmt.Errorf("no %s file found in %s", ext, dir)
}

func findTestFileByPattern(dir, pattern string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no test file matching %s found in %s", pattern, dir)
	}
	return filepath.Base(matches[0]), nil
}
