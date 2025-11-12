package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CacheEntry represents a cached item
type CacheEntry struct {
	Key       string
	Value     interface{}
	ExpiresAt time.Time
}

// Cache provides TTL-based caching
type Cache struct {
	entries map[string]*CacheEntry
	mu      sync.RWMutex
	ttl     time.Duration
	persist bool
	path    string
}

// NewCache creates a new cache with the specified TTL
func NewCache(ttl time.Duration, persist bool, cachePath string) *Cache {
	cache := &Cache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
		persist: persist,
		path:    cachePath,
	}

	if persist && cachePath != "" {
		cache.loadFromDisk()
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Value, true
}

// Set stores a value in the cache
func (c *Cache) Set(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &CacheEntry{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}

	if c.persist && c.path != "" {
		return c.saveToDisk()
	}

	return nil
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)

	if c.persist && c.path != "" {
		_ = c.saveToDisk()
	}
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)

	if c.persist && c.path != "" {
		_ = c.saveToDisk()
	}
}

// cleanup periodically removes expired entries
func (c *Cache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.entries {
			if now.After(entry.ExpiresAt) {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

// saveToDisk saves the cache to disk
func (c *Cache) saveToDisk() error {
	if c.path == "" {
		return nil
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Serialize cache entries
	data, err := json.Marshal(c.entries)
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	// Write to file
	if err := os.WriteFile(c.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// loadFromDisk loads the cache from disk
func (c *Cache) loadFromDisk() {
	if c.path == "" {
		return
	}

	data, err := os.ReadFile(c.path)
	if err != nil {
		// File doesn't exist or can't be read, start with empty cache
		return
	}

	var entries map[string]*CacheEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		// Invalid cache file, start with empty cache
		return
	}

	// Filter out expired entries
	now := time.Now()
	for key, entry := range entries {
		if now.Before(entry.ExpiresAt) {
			c.entries[key] = entry
		}
	}
}

// GetCacheKey generates a cache key for package documentation
func GetCacheKey(pkgPath, version string) string {
	if version != "" {
		return fmt.Sprintf("pkg:%s@%s", pkgPath, version)
	}
	return fmt.Sprintf("pkg:%s", pkgPath)
}
