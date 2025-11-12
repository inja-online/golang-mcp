package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/inja-online/golang-mcp/internal/config"
)

// PackageDoc represents package documentation
type PackageDoc struct {
	Path        string            `json:"path"`
	Version     string            `json:"version,omitempty"`
	Overview    string            `json:"overview"`
	Description string            `json:"description"`
	Functions   []FunctionDoc     `json:"functions,omitempty"`
	Types       []TypeDoc         `json:"types,omitempty"`
	Constants   []ConstantDoc     `json:"constants,omitempty"`
	Examples    []ExampleDoc      `json:"examples,omitempty"`
	Imports     []string          `json:"imports,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// FunctionDoc represents a function documentation
type FunctionDoc struct {
	Name        string `json:"name"`
	Signature   string `json:"signature"`
	Description string `json:"description"`
}

// TypeDoc represents a type documentation
type TypeDoc struct {
	Name        string      `json:"name"`
	Kind        string      `json:"kind"` // "struct", "interface", "type"
	Description string      `json:"description"`
	Fields      []FieldDoc  `json:"fields,omitempty"`
	Methods     []MethodDoc `json:"methods,omitempty"`
}

// FieldDoc represents a struct field
type FieldDoc struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// MethodDoc represents a method
type MethodDoc struct {
	Name        string `json:"name"`
	Signature   string `json:"signature"`
	Description string `json:"description"`
}

// ConstantDoc represents a constant
type ConstantDoc struct {
	Name        string `json:"name"`
	Value       string `json:"value,omitempty"`
	Description string `json:"description,omitempty"`
}

// ExampleDoc represents an example
type ExampleDoc struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
}

// FetchPackageDocs fetches package documentation from go.dev
func FetchPackageDocs(ctx context.Context, pkgPath, version string, cache *Cache) (*PackageDoc, error) {
	// Check cache first
	cacheKey := GetCacheKey(pkgPath, version)
	if cached, found := cache.Get(cacheKey); found {
		if doc, ok := cached.(*PackageDoc); ok {
			return doc, nil
		}
	}

	// Try to fetch from pkg.go.dev API
	doc, err := fetchFromAPI(ctx, pkgPath, version)
	if err == nil {
		// Cache the result
		_ = cache.Set(cacheKey, doc)
		return doc, nil
	}

	// Fallback to scraping HTML
	doc, err = fetchFromHTML(ctx, pkgPath, version)
	if err == nil {
		// Cache the result
		_ = cache.Set(cacheKey, doc)
		return doc, nil
	}

	// Last resort: use local go doc
	return fetchFromLocal(ctx, pkgPath)
}

// fetchFromAPI fetches documentation from pkg.go.dev API
func fetchFromAPI(ctx context.Context, pkgPath, version string) (*PackageDoc, error) {
	url := fmt.Sprintf("https://pkg.go.dev/%s", pkgPath)
	if version != "" {
		url = fmt.Sprintf("https://pkg.go.dev/%s@%s", pkgPath, version)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Try to parse as JSON (if API supports it)
	var doc PackageDoc
	if err := json.NewDecoder(resp.Body).Decode(&doc); err == nil {
		doc.Path = pkgPath
		doc.Version = version
		return &doc, nil
	}

	return nil, fmt.Errorf("API response not in expected format")
}

// fetchFromHTML scrapes documentation from pkg.go.dev HTML page
func fetchFromHTML(ctx context.Context, pkgPath, version string) (*PackageDoc, error) {
	url := fmt.Sprintf("https://pkg.go.dev/%s", pkgPath)
	if version != "" {
		url = fmt.Sprintf("https://pkg.go.dev/%s@%s", pkgPath, version)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch HTML: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse HTML (simplified - would need proper HTML parsing in production)
	doc := &PackageDoc{
		Path:     pkgPath,
		Version:  version,
		Overview: extractOverview(string(body)),
		Metadata: make(map[string]string),
	}

	return doc, nil
}

// fetchFromLocal fetches documentation using local go doc command
func fetchFromLocal(ctx context.Context, pkgPath string) (*PackageDoc, error) {
	cfg := config.Load()

	// Use ExecuteGoCommand to run go doc
	result, err := ExecuteGoCommand(ctx, cfg, "go", []string{"doc", "-all", pkgPath}, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to run go doc: %w", err)
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("go doc failed: %s", result.Stderr)
	}

	doc := &PackageDoc{
		Path:     pkgPath,
		Overview: result.Stdout,
		Metadata: map[string]string{
			"source": "local",
		},
	}

	return doc, nil
}

// extractOverview extracts overview text from HTML (simplified)
func extractOverview(html string) string {
	// This is a simplified extraction - in production, use proper HTML parsing
	// Look for common patterns in pkg.go.dev HTML

	// Try to find the package documentation section
	start := strings.Index(html, "<!-- Package")
	if start == -1 {
		start = strings.Index(html, "<div class=\"Documentation-overview\"")
	}

	if start == -1 {
		return "Package documentation not found in HTML"
	}

	// Extract a reasonable amount of text
	end := start + 1000
	if end > len(html) {
		end = len(html)
	}

	text := html[start:end]
	// Remove HTML tags (simplified)
	text = strings.ReplaceAll(text, "<", " ")
	text = strings.ReplaceAll(text, ">", " ")

	return strings.TrimSpace(text)
}

// SearchPackages searches for packages on pkg.go.dev based on the given query.
// It returns a list of package paths that match the search query.
// The function makes an HTTP request to pkg.go.dev search endpoint and parses
// the HTML response to extract package paths.
func SearchPackages(ctx context.Context, query string) ([]string, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	log.Printf("Searching packages for query: %s", query)

	searchURL := fmt.Sprintf("https://pkg.go.dev/search?q=%s", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		log.Printf("Failed to create search request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to fetch search results: %v", err)
		return nil, fmt.Errorf("failed to fetch search results: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Search returned status code: %d", resp.StatusCode)
		return nil, fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read search response: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	packages := extractPackagePaths(string(body))
	log.Printf("Found %d packages for query: %s", len(packages), query)

	return packages, nil
}

// extractPackagePaths extracts package paths from pkg.go.dev search results HTML.
// It looks for links that point to package pages (e.g., /github.com/user/repo, /net/http).
func extractPackagePaths(html string) []string {
	var packages []string
	seen := make(map[string]bool)

	re := regexp.MustCompile(`href="/([a-zA-Z0-9._/-]+)"`)
	matches := re.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		pkgPath := match[1]
		if pkgPath == "" || pkgPath == "search" || pkgPath == "about" {
			continue
		}

		if strings.HasPrefix(pkgPath, "search") || strings.HasPrefix(pkgPath, "about") {
			continue
		}

		if strings.Contains(pkgPath, "?") || strings.Contains(pkgPath, "#") {
			continue
		}

		if !isValidPackagePath(pkgPath) {
			continue
		}

		if !seen[pkgPath] {
			packages = append(packages, pkgPath)
			seen[pkgPath] = true
		}
	}

	return packages
}

// isValidPackagePath checks if a path looks like a valid Go package path.
func isValidPackagePath(path string) bool {
	if path == "" {
		return false
	}

	if strings.Contains(path, "?") || strings.Contains(path, "#") {
		return false
	}

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return false
	}

	if strings.HasPrefix(path, "/") {
		return false
	}

	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return false
	}

	firstPart := parts[0]
	if firstPart == "" {
		return false
	}

	if strings.Contains(firstPart, ".") {
		return true
	}

	if firstPart == "net" || firstPart == "fmt" || firstPart == "os" || firstPart == "io" ||
		firstPart == "strings" || firstPart == "bytes" || firstPart == "time" ||
		firstPart == "encoding" || firstPart == "crypto" || firstPart == "database" ||
		firstPart == "html" || firstPart == "image" || firstPart == "math" ||
		firstPart == "path" || firstPart == "reflect" || firstPart == "runtime" ||
		firstPart == "sort" || firstPart == "sync" || firstPart == "testing" ||
		firstPart == "text" || firstPart == "unicode" || firstPart == "unsafe" {
		return true
	}

	return false
}
