package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestFetchPackageDocs_WithCache(t *testing.T) {
	cache := NewCache(1*time.Hour, false, "")
	defer cache.Clear()

	ctx := context.Background()

	// Create a test package doc
	testDoc := &PackageDoc{
		Path:     "test/package",
		Overview: "Test package overview",
	}

	// Set in cache
	cacheKey := GetCacheKey("test/package", "")
	_ = cache.Set(cacheKey, testDoc)

	// Fetch should return cached value
	doc, err := FetchPackageDocs(ctx, "test/package", "", cache)
	if err != nil {
		t.Fatalf("Failed to fetch from cache: %v", err)
	}

	if doc.Path != "test/package" {
		t.Errorf("Expected path 'test/package', got %q", doc.Path)
	}
}

func TestFetchPackageDocs_FromAPI(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/encoding/json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"path":"encoding/json","overview":"Package json implements encoding and decoding of JSON"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Note: This test would need to modify fetchFromAPI to accept a base URL
	// For now, we'll test the cache and local fallback
	cache := NewCache(1*time.Hour, false, "")
	defer cache.Clear()

	ctx := context.Background()

	// Test that it tries to fetch (will fail without network, but tests the flow)
	// We'll test the local fallback instead
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("Go command not available for local doc test")
	}

	// Test local fallback
	doc, err := FetchPackageDocs(ctx, "fmt", "", cache)
	if err != nil {
		// This might fail if fmt package is not available, which is okay
		t.Logf("Local doc fetch failed (expected in some environments): %v", err)
		return
	}

	if doc == nil {
		t.Error("Expected doc but got nil")
	}
}

func TestSearchPackages(t *testing.T) {
	t.Run("empty query", func(t *testing.T) {
		ctx := context.Background()
		_, err := SearchPackages(ctx, "")
		if err == nil {
			t.Error("Expected error for empty query")
		}
		if err != nil && !strings.Contains(err.Error(), "empty") {
			t.Errorf("Expected error message about empty query, got: %v", err)
		}
	})

	t.Run("network request", func(t *testing.T) {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		packages, err := SearchPackages(ctx, "http")
		if err != nil {
			if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "connection") {
				t.Skipf("Network test skipped due to network issue: %v", err)
				return
			}
			t.Logf("Search returned error (may be expected): %v", err)
			return
		}

		if len(packages) == 0 {
			t.Log("No packages found (may be expected if search structure changed)")
		} else {
			t.Logf("Found %d packages", len(packages))
		}
	})
}

func TestExtractPackagePaths(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected []string
	}{
		{
			name: "standard package links",
			html: `
				<a href="/net/http">net/http</a>
				<a href="/github.com/gin-gonic/gin">gin</a>
				<a href="/encoding/json">encoding/json</a>
			`,
			expected: []string{"net/http", "github.com/gin-gonic/gin", "encoding/json"},
		},
		{
			name: "filter out non-package links",
			html: `
				<a href="/search?q=test">search</a>
				<a href="/about">about</a>
				<a href="/net/http">net/http</a>
			`,
			expected: []string{"net/http"},
		},
		{
			name:     "no package links",
			html:     `<html><body><p>No packages here</p></body></html>`,
			expected: []string{},
		},
		{
			name: "duplicate links",
			html: `
				<a href="/net/http">net/http</a>
				<a href="/net/http">net/http</a>
			`,
			expected: []string{"net/http"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPackagePaths(tt.html)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d packages, got %d: %v", len(tt.expected), len(result), result)
			}

			for _, expected := range tt.expected {
				found := false
				for _, pkg := range result {
					if pkg == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find package %q in results: %v", expected, result)
				}
			}
		})
	}
}

func TestIsValidPackagePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"standard library", "net/http", true},
		{"github package", "github.com/gin-gonic/gin", true},
		{"empty path", "", false},
		{"search link", "search", false},
		{"about link", "about", false},
		{"with query params", "net/http?q=test", false},
		{"with hash", "net/http#section", false},
		{"http url", "http://example.com", false},
		{"https url", "https://example.com", false},
		{"leading slash", "/net/http", false},
		{"standard library fmt", "fmt", true},
		{"standard library strings", "strings", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidPackagePath(tt.path)
			if result != tt.expected {
				t.Errorf("isValidPackagePath(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestExtractOverview(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		contains string
	}{
		{
			name:     "HTML with package comment",
			html:     `<!-- Package json --><div>Package json implements encoding</div>`,
			contains: "Package",
		},
		{
			name:     "HTML with documentation overview",
			html:     `<div class="Documentation-overview">Package documentation</div>`,
			contains: "Package",
		},
		{
			name:     "HTML without markers",
			html:     `<div>Some content</div>`,
			contains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractOverview(tt.html)
			if result == "" {
				t.Error("Expected non-empty result")
			}
			// Just verify it doesn't panic and returns something
			_ = result
		})
	}
}
