package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// TestCase represents a single test result.
type TestCase struct {
	Name    string
	Status  string // "passed", "failed", "pending"
	Message string // failure message if any
}

// TestResult holds structured test output.
type TestResult struct {
	Passed    bool
	Total     int
	PassCount int
	FailCount int
	Cases     []TestCase
	RawOutput string // fallback when structured parsing isn't available
}

// track slug → test command parts
var testCommands = map[string][]string{
	"go":         {"go", "test", "-v", "./..."},
	"rust":       {"cargo", "test"},
	"python":     {"python3", "-m", "pytest", "-v"},
	"javascript": {"npx", "jest", "--json", "--no-coverage"},
	"typescript": {"npx", "jest", "--json", "--no-coverage"},
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
func (w *Workspace) RunTests(trackSlug, exerciseSlug string) (*TestResult, error) {
	dir := w.ExerciseDir(trackSlug, exerciseSlug)
	if _, err := os.Stat(dir); err != nil {
		return nil, fmt.Errorf("exercise not downloaded: %s", dir)
	}

	// Install shared dependencies at track root if needed
	if installCmd, ok := installCommands[trackSlug]; ok {
		trackRoot := filepath.Dir(dir)
		if _, err := os.Stat(filepath.Join(trackRoot, "node_modules")); err != nil {
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
					Passed:    false,
					RawOutput: fmt.Sprintf("Failed to install dependencies:\n%s", string(out)),
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
	output, runErr := cmd.CombinedOutput()

	// Try structured parsing for supported runners
	switch trackSlug {
	case "javascript", "typescript":
		if result := parseJestJSON(output); result != nil {
			return result, nil
		}
	case "go":
		if result := parseGoTestVerbose(string(output), runErr == nil); result != nil {
			return result, nil
		}
	}

	// Fallback: raw output
	return &TestResult{
		Passed:    runErr == nil,
		RawOutput: string(output),
	}, nil
}

// Jest JSON output parser
func parseJestJSON(output []byte) *TestResult {
	// Jest might prefix the JSON with npm/pnpm output lines
	// Find the JSON object in the output
	raw := string(output)
	jsonStart := strings.Index(raw, "{")
	if jsonStart < 0 {
		return nil
	}

	var jest struct {
		NumPassedTests int  `json:"numPassedTests"`
		NumFailedTests int  `json:"numFailedTests"`
		NumTotalTests  int  `json:"numTotalTests"`
		Success        bool `json:"success"`
		TestResults    []struct {
			AssertionResults []struct {
				FullName        string   `json:"fullName"`
				Status          string   `json:"status"`
				FailureMessages []string `json:"failureMessages"`
			} `json:"assertionResults"`
		} `json:"testResults"`
	}

	if err := json.Unmarshal([]byte(raw[jsonStart:]), &jest); err != nil {
		// Try the full output
		if err := json.Unmarshal(output, &jest); err != nil {
			return nil
		}
	}

	result := &TestResult{
		Passed:    jest.Success,
		Total:     jest.NumTotalTests,
		PassCount: jest.NumPassedTests,
		FailCount: jest.NumFailedTests,
	}

	for _, suite := range jest.TestResults {
		for _, a := range suite.AssertionResults {
			tc := TestCase{
				Name:   a.FullName,
				Status: a.Status,
			}
			if len(a.FailureMessages) > 0 {
				tc.Message = cleanFailureMessage(a.FailureMessages[0])
			}
			result.Cases = append(result.Cases, tc)
		}
	}

	return result
}

// Parse verbose go test output
func parseGoTestVerbose(output string, passed bool) *TestResult {
	result := &TestResult{Passed: passed}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--- PASS:") {
			name := strings.TrimPrefix(trimmed, "--- PASS: ")
			if idx := strings.Index(name, " ("); idx > 0 {
				name = name[:idx]
			}
			result.Cases = append(result.Cases, TestCase{Name: name, Status: "passed"})
			result.PassCount++
			result.Total++
		} else if strings.HasPrefix(trimmed, "--- FAIL:") {
			name := strings.TrimPrefix(trimmed, "--- FAIL: ")
			if idx := strings.Index(name, " ("); idx > 0 {
				name = name[:idx]
			}
			result.Cases = append(result.Cases, TestCase{Name: name, Status: "failed"})
			result.FailCount++
			result.Total++
		}
	}

	if len(result.Cases) == 0 {
		return nil // couldn't parse, fall back to raw
	}

	return result
}

// Clean up jest failure messages — strip ANSI, trim paths, keep the useful part
func cleanFailureMessage(msg string) string {
	lines := strings.Split(msg, "\n")
	var cleaned []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Skip lines that are just stack traces to node_modules
		if strings.Contains(trimmed, "node_modules/") {
			continue
		}
		// Skip lines that are just "at Object.<anonymous>"
		if strings.HasPrefix(trimmed, "at Object.") || strings.HasPrefix(trimmed, "at new") {
			continue
		}
		cleaned = append(cleaned, trimmed)
	}
	if len(cleaned) > 6 {
		cleaned = cleaned[:6]
	}
	return strings.Join(cleaned, "\n")
}

func resolveTestCommand(trackSlug, dir string) ([]string, error) {
	if trackSlug == "ruby" {
		testFile, err := findRubyTestFile(dir)
		if err != nil {
			return nil, err
		}
		return []string{"ruby", testFile}, nil
	}

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
