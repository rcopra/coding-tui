package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rcopra/coding-tui/internal/cache"
)

const (
	websiteAPI = "https://exercism.org/api/v2"
	cliAPI     = "https://api.exercism.org/v1"
)

type Client struct {
	token string
	http  *http.Client
	cache *cache.Cache
}

func NewClient(token string, c *cache.Cache) *Client {
	return &Client{
		token: token,
		cache: c,
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) get(url string, needsAuth bool) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if needsAuth && c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// getCached tries cache first, falls back to HTTP GET, then caches the result.
func (c *Client) getCached(cacheKey, url string, needsAuth bool) ([]byte, error) {
	if c.cache != nil {
		if data := c.cache.Get(cacheKey); data != nil {
			return data, nil
		}
	}

	data, err := c.get(url, needsAuth)
	if err != nil {
		return nil, err
	}

	if c.cache != nil {
		_ = c.cache.Set(cacheKey, json.RawMessage(data))
	}

	return data, nil
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func decode[T any](data []byte) (T, error) {
	var result T
	err := json.Unmarshal(data, &result)
	return result, err
}

func decodeInto(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
