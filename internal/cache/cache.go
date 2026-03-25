package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Cache struct {
	dir string
	ttl time.Duration
}

type entry struct {
	Data      json.RawMessage `json:"data"`
	ExpiresAt time.Time       `json:"expires_at"`
}

func New(cacheDir string, ttl time.Duration) *Cache {
	return &Cache{dir: cacheDir, ttl: ttl}
}

// Get retrieves a cached value. Returns nil if not found or expired.
func (c *Cache) Get(key string) json.RawMessage {
	path := c.path(key)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var e entry
	if err := json.Unmarshal(data, &e); err != nil {
		return nil
	}

	if time.Now().After(e.ExpiresAt) {
		os.Remove(path)
		return nil
	}

	return e.Data
}

// Set stores a value in the cache.
func (c *Cache) Set(key string, data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	e := entry{
		Data:      jsonData,
		ExpiresAt: time.Now().Add(c.ttl),
	}

	entryData, err := json.Marshal(e)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}

	return os.WriteFile(c.path(key), entryData, 0o644)
}

// Delete removes a cached value.
func (c *Cache) Delete(key string) {
	os.Remove(c.path(key))
}

// GetInto retrieves a cached value and unmarshals it into v.
// Returns false if cache miss.
func (c *Cache) GetInto(key string, v any) bool {
	data := c.Get(key)
	if data == nil {
		return false
	}
	return json.Unmarshal(data, v) == nil
}

func (c *Cache) path(key string) string {
	hash := sha256.Sum256([]byte(key))
	return filepath.Join(c.dir, fmt.Sprintf("%x.json", hash[:8]))
}
