package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestCache_GetSet(t *testing.T) {
	cache := NewCache(1*time.Hour, false, "")
	defer cache.Clear()

	key := "test-key"
	value := "test-value"

	// Set value
	if err := cache.Set(key, value); err != nil {
		t.Fatalf("Failed to set cache value: %v", err)
	}

	// Get value
	cached, found := cache.Get(key)
	if !found {
		t.Error("Expected to find cached value")
	}

	if cached != value {
		t.Errorf("Expected %q, got %q", value, cached)
	}
}

func TestCache_TTLExpiration(t *testing.T) {
	cache := NewCache(100*time.Millisecond, false, "")
	defer cache.Clear()

	key := "test-key"
	value := "test-value"

	// Set value
	if err := cache.Set(key, value); err != nil {
		t.Fatalf("Failed to set cache value: %v", err)
	}

	// Should be available immediately
	_, found := cache.Get(key)
	if !found {
		t.Error("Expected to find cached value immediately")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, found = cache.Get(key)
	if found {
		t.Error("Expected cached value to be expired")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache(1*time.Hour, false, "")
	defer cache.Clear()

	key := "test-key"
	value := "test-value"

	// Set value
	if err := cache.Set(key, value); err != nil {
		t.Fatalf("Failed to set cache value: %v", err)
	}

	// Delete value
	cache.Delete(key)

	// Should not be found
	_, found := cache.Get(key)
	if found {
		t.Error("Expected cached value to be deleted")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(1*time.Hour, false, "")
	defer cache.Clear()

	_ = cache.Set("key1", "value1")
	_ = cache.Set("key2", "value2")
	_ = cache.Set("key3", "value3")

	// Clear all
	cache.Clear()

	// None should be found
	if _, found := cache.Get("key1"); found {
		t.Error("Expected key1 to be cleared")
	}
	if _, found := cache.Get("key2"); found {
		t.Error("Expected key2 to be cleared")
	}
	if _, found := cache.Get("key3"); found {
		t.Error("Expected key3 to be cleared")
	}
}

func TestCache_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache.json")

	cache := NewCache(1*time.Hour, true, cachePath)
	defer cache.Clear()

	key := "test-key"
	value := "test-value"

	// Set value
	if err := cache.Set(key, value); err != nil {
		t.Fatalf("Failed to set cache value: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("Cache file should exist: %v", err)
	}

	// Create new cache instance and load from disk
	cache2 := NewCache(1*time.Hour, true, cachePath)
	defer cache2.Clear()

	// Should be able to retrieve the value
	cached, found := cache2.Get(key)
	if !found {
		t.Error("Expected to find cached value after reload")
	}

	if cached != value {
		t.Errorf("Expected %q, got %q", value, cached)
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	cache := NewCache(1*time.Hour, false, "")
	defer cache.Clear()

	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				value := fmt.Sprintf("value-%d-%d", id, j)
				_ = cache.Set(key, value)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				_, _ = cache.Get(key)
			}
		}(i)
	}

	wg.Wait()
	// If we get here without panicking, concurrent access works
}

func TestGetCacheKey(t *testing.T) {
	tests := []struct {
		name    string
		pkgPath string
		version string
		want    string
	}{
		{
			name:    "package without version",
			pkgPath: "encoding/json",
			version: "",
			want:    "pkg:encoding/json",
		},
		{
			name:    "package with version",
			pkgPath: "encoding/json",
			version: "v1.0.0",
			want:    "pkg:encoding/json@v1.0.0",
		},
		{
			name:    "package with latest version",
			pkgPath: "github.com/example/pkg",
			version: "v2.1.3",
			want:    "pkg:github.com/example/pkg@v2.1.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCacheKey(tt.pkgPath, tt.version)
			if got != tt.want {
				t.Errorf("GetCacheKey() = %q, want %q", got, tt.want)
			}
		})
	}
}
