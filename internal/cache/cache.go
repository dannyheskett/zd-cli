package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	cacheDirName = ".zd"
	cacheSubDir  = "cache"
	// DefaultTTL is the default cache TTL
	DefaultTTL = 10 * time.Minute
)

// Entry represents a cached item with expiration
type Entry struct {
	Data      json.RawMessage `json:"data"`
	ExpiresAt time.Time       `json:"expires_at"`
	CreatedAt time.Time       `json:"created_at"`
}

// Cache handles caching of API responses
type Cache struct {
	dir string
	ttl time.Duration
}

// New creates a new cache instance with the specified TTL
func New(ttl time.Duration) (*Cache, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(home, cacheDirName, cacheSubDir)
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Cache{
		dir: cacheDir,
		ttl: ttl,
	}, nil
}

// Get retrieves a cached item by key
func (c *Cache) Get(key string) ([]byte, bool) {
	path := c.keyToPath(key)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		// Invalid cache entry, remove it
		os.Remove(path)
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		os.Remove(path)
		return nil, false
	}

	return entry.Data, true
}

// Set stores an item in the cache
func (c *Cache) Set(key string, data []byte) error {
	entry := Entry{
		Data:      data,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(c.ttl),
	}

	entryData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	path := c.keyToPath(key)
	if err := os.WriteFile(path, entryData, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) error {
	path := c.keyToPath(key)
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// Clear removes all cached items
func (c *Cache) Clear() error {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			path := filepath.Join(c.dir, entry.Name())
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove cache file %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

// keyToPath converts a cache key to a file path
func (c *Cache) keyToPath(key string) string {
	// Hash the key to create a safe filename
	hash := sha256.Sum256([]byte(key))
	filename := hex.EncodeToString(hash[:]) + ".json"
	return filepath.Join(c.dir, filename)
}

// PruneExpired removes all expired cache entries
func (c *Cache) PruneExpired() error {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(c.dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var cacheEntry Entry
		if err := json.Unmarshal(data, &cacheEntry); err != nil {
			// Invalid entry, remove it
			os.Remove(path)
			continue
		}

		if now.After(cacheEntry.ExpiresAt) {
			os.Remove(path)
		}
	}

	return nil
}
