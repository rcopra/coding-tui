package api

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// GetLatestSolution fetches the latest solution for an exercise in a track.
func (c *Client) GetLatestSolution(trackSlug, exerciseSlug string) (*Solution, error) {
	url := fmt.Sprintf("%s/solutions/latest?track_id=%s&exercise_id=%s", cliAPI, trackSlug, exerciseSlug)
	data, err := c.get(url, true)
	if err != nil {
		return nil, err
	}

	resp, err := decode[SolutionResponse](data)
	if err != nil {
		return nil, err
	}

	return &resp.Solution, nil
}

// GetSolutionFile downloads a single file from a solution.
func (c *Client) GetSolutionFile(baseURL, path string) ([]byte, error) {
	url := baseURL + path
	return c.get(url, true)
}

// SubmitSolution uploads solution files to Exercism.
// filepaths are absolute paths on disk; relativePaths are how they should appear in the submission.
func (c *Client) SubmitSolution(solutionID string, files map[string]string) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for relativePath, absPath := range files {
		f, err := os.Open(absPath)
		if err != nil {
			return fmt.Errorf("opening %s: %w", absPath, err)
		}

		part, err := writer.CreateFormFile("files[]", relativePath)
		if err != nil {
			f.Close()
			return fmt.Errorf("creating form field: %w", err)
		}

		if _, err := io.Copy(part, f); err != nil {
			f.Close()
			return fmt.Errorf("copying file content: %w", err)
		}
		f.Close()
	}

	if err := writer.Close(); err != nil {
		return err
	}

	url := fmt.Sprintf("%s/solutions/%s", cliAPI, solutionID)
	req, err := http.NewRequest("PATCH", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	_, err = c.do(req)
	return err
}

// CompleteSolution marks an exercise as complete on Exercism.
func (c *Client) CompleteSolution(solutionUUID string) error {
	url := fmt.Sprintf("%s/solutions/%s/complete", websiteAPI, solutionUUID)
	req, err := http.NewRequest("PATCH", url, nil)
	if err != nil {
		return err
	}

	_, err = c.do(req)
	return err
}

// InvalidateExercises clears the cached exercise list for a track
// so fresh recommendations are fetched on next load.
func (c *Client) InvalidateExercises(trackSlug string) {
	if c.cache != nil {
		c.cache.Delete(fmt.Sprintf("exercises:%s", trackSlug))
	}
}

// SolutionFilePaths returns a map of relative path → absolute path for solution files.
func SolutionFilePaths(exerciseDir string, solutionFiles []string) map[string]string {
	files := make(map[string]string)
	for _, rel := range solutionFiles {
		abs := filepath.Join(exerciseDir, rel)
		if _, err := os.Stat(abs); err == nil {
			files[rel] = abs
		}
	}
	return files
}
