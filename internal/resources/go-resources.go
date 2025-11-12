package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/inja-online/golang-mcp/internal/config"
	"github.com/inja-online/golang-mcp/internal/utils"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	toolsRegistry     = make(map[string]*ToolMetadata)
	promptsRegistry   = make(map[string]*PromptMetadata)
	resourcesRegistry = make(map[string]*ResourceMetadata)
	registryMu        sync.RWMutex
)

// ToolMetadata represents metadata for a registered tool.
type ToolMetadata struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema,omitempty"`
}

// PromptMetadata represents metadata for a registered prompt.
type PromptMetadata struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Arguments   []*mcp.PromptArgument `json:"arguments,omitempty"`
}

// ResourceMetadata represents metadata for a registered resource.
type ResourceMetadata struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MIMEType    string `json:"mime_type,omitempty"`
}

// RegisterTool registers a tool in the registry.
func RegisterTool(name, description string, inputSchema map[string]interface{}) {
	registryMu.Lock()
	defer registryMu.Unlock()
	toolsRegistry[name] = &ToolMetadata{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
	}
}

// RegisterPrompt registers a prompt in the registry.
func RegisterPrompt(name, description string, arguments []*mcp.PromptArgument) {
	registryMu.Lock()
	defer registryMu.Unlock()
	promptsRegistry[name] = &PromptMetadata{
		Name:        name,
		Description: description,
		Arguments:   arguments,
	}
}

// RegisterResourceMetadata registers resource metadata in the registry.
func RegisterResourceMetadata(uri, name, description, mimeType string) {
	registryMu.Lock()
	defer registryMu.Unlock()
	resourcesRegistry[uri] = &ResourceMetadata{
		URI:         uri,
		Name:        name,
		Description: description,
		MIMEType:    mimeType,
	}
}

// RegisterGoResources registers Go project discovery resources
func RegisterGoResources(server *mcp.Server, cfg *config.Config) int {
	count := 0
	// go://modules resource
	RegisterResourceMetadata("go://modules", "Go Modules", "List of Go modules and dependencies in the current workspace", "application/json")
	server.AddResource(&mcp.Resource{
		URI:         "go://modules",
		Name:        "Go Modules",
		Description: "List of Go modules and dependencies in the current workspace",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// Read go.mod file
		modPath := filepath.Join(cfg.WorkingDirectory, "go.mod")
		if _, err := os.Stat(modPath); os.IsNotExist(err) {
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{URI: "go://modules", Text: "No go.mod file found in current directory"},
				},
			}, nil
		}

		modData, err := os.ReadFile(modPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read go.mod: %w", err)
		}

		// Parse module information
		moduleInfo := parseGoMod(string(modData))

		jsonData, err := json.MarshalIndent(moduleInfo, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal module info: %w", err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "go://modules", MIMEType: "application/json", Text: string(jsonData)},
			},
		}, nil
	})
	count++

	// go://build-tags resource
	RegisterResourceMetadata("go://build-tags", "Build Tags", "Build tags and constraints found in the project", "application/json")
	server.AddResource(&mcp.Resource{
		URI:         "go://build-tags",
		Name:        "Build Tags",
		Description: "Build tags and constraints found in the project",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// Discover build tags by scanning Go files
		tags := discoverBuildTags(cfg.WorkingDirectory)

		jsonData, err := json.MarshalIndent(tags, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal build tags: %w", err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "go://build-tags", MIMEType: "application/json", Text: string(jsonData)},
			},
		}, nil
	})
	count++

	// go://tests resource
	RegisterResourceMetadata("go://tests", "Test Files", "List of test files and benchmarks in the project", "application/json")
	server.AddResource(&mcp.Resource{
		URI:         "go://tests",
		Name:        "Test Files",
		Description: "List of test files and benchmarks in the project",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		tests := discoverTestFiles(cfg.WorkingDirectory)

		jsonData, err := json.MarshalIndent(tests, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal test files: %w", err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "go://tests", MIMEType: "application/json", Text: string(jsonData)},
			},
		}, nil
	})
	count++

	// go://workspace resource
	RegisterResourceMetadata("go://workspace", "Go Workspace", "Go workspace structure and configuration", "application/json")
	server.AddResource(&mcp.Resource{
		URI:         "go://workspace",
		Name:        "Go Workspace",
		Description: "Go workspace structure and configuration",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		workspace := discoverWorkspace(cfg.WorkingDirectory)

		jsonData, err := json.MarshalIndent(workspace, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal workspace: %w", err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "go://workspace", MIMEType: "application/json", Text: string(jsonData)},
			},
		}, nil
	})
	count++

	// go://pkg-docs/{path} resource
	RegisterResourceMetadata("go://pkg-docs/{path}", "Package Documentation", "Fetch package documentation from go.dev. Supports versioned paths (e.g., go://pkg-docs/encoding/json@v1.0.0)", "application/json")
	server.AddResource(&mcp.Resource{
		URI:         "go://pkg-docs/{path}",
		Name:        "Package Documentation",
		Description: "Fetch package documentation from go.dev. Supports versioned paths (e.g., go://pkg-docs/encoding/json@v1.0.0)",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// Parse path and version from URI
		// URI format: go://pkg-docs/{path} or go://pkg-docs/{path}@version
		uri := req.Params.URI
		if !strings.HasPrefix(uri, "go://pkg-docs/") {
			return nil, fmt.Errorf("invalid URI format: %s", uri)
		}

		pathPart := strings.TrimPrefix(uri, "go://pkg-docs/")
		if pathPart == "" {
			return nil, fmt.Errorf("package path is required")
		}

		// Parse version if present
		var pkgPath, version string
		if idx := strings.LastIndex(pathPart, "@"); idx != -1 {
			pkgPath = pathPart[:idx]
			version = pathPart[idx+1:]
		} else {
			pkgPath = pathPart
		}

		cacheDir := filepath.Join(os.TempDir(), "mcp-golang-cache")
		_ = os.MkdirAll(cacheDir, 0755)
		cachePath := filepath.Join(cacheDir, "pkg-docs.json")
		cache := utils.NewCache(24*time.Hour, true, cachePath)

		// Fetch package documentation
		doc, err := utils.FetchPackageDocs(ctx, pkgPath, version, cache)
		if err != nil {
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{URI: uri, Text: fmt.Sprintf("Error fetching package documentation: %v", err)},
				},
			}, nil
		}

		// Marshal to JSON
		jsonData, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal package documentation: %w", err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: uri, MIMEType: "application/json", Text: string(jsonData)},
			},
		}, nil
	})
	count++

	RegisterResourceMetadata("go://tools", "Available Tools", "List of all available tools with their names, descriptions, and parameter schemas", "application/json")
	server.AddResource(&mcp.Resource{
		URI:         "go://tools",
		Name:        "Available Tools",
		Description: "List of all available tools with their names, descriptions, and parameter schemas",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		registryMu.RLock()
		defer registryMu.RUnlock()

		toolsList := make([]*ToolMetadata, 0, len(toolsRegistry))
		for _, tool := range toolsRegistry {
			toolsList = append(toolsList, tool)
		}

		result := map[string]interface{}{
			"tools": toolsList,
			"count": len(toolsList),
		}

		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tools: %w", err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "go://tools", MIMEType: "application/json", Text: string(jsonData)},
			},
		}, nil
	})
	count++

	RegisterResourceMetadata("go://prompts", "Available Prompts", "List of all available prompts with their names, descriptions, and arguments", "application/json")
	server.AddResource(&mcp.Resource{
		URI:         "go://prompts",
		Name:        "Available Prompts",
		Description: "List of all available prompts with their names, descriptions, and arguments",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		registryMu.RLock()
		defer registryMu.RUnlock()

		promptsList := make([]*PromptMetadata, 0, len(promptsRegistry))
		for _, prompt := range promptsRegistry {
			promptsList = append(promptsList, prompt)
		}

		result := map[string]interface{}{
			"prompts": promptsList,
			"count":   len(promptsList),
		}

		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal prompts: %w", err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "go://prompts", MIMEType: "application/json", Text: string(jsonData)},
			},
		}, nil
	})
	count++

	RegisterResourceMetadata("go://resources", "Available Resources", "List of all available resources with their URIs, names, and descriptions", "application/json")
	server.AddResource(&mcp.Resource{
		URI:         "go://resources",
		Name:        "Available Resources",
		Description: "List of all available resources with their URIs, names, and descriptions",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		registryMu.RLock()
		defer registryMu.RUnlock()

		resourcesList := make([]*ResourceMetadata, 0, len(resourcesRegistry))
		for _, resource := range resourcesRegistry {
			resourcesList = append(resourcesList, resource)
		}

		result := map[string]interface{}{
			"resources": resourcesList,
			"count":     len(resourcesList),
		}

		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal resources: %w", err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: "go://resources", MIMEType: "application/json", Text: string(jsonData)},
			},
		}, nil
	})
	count++
	return count
}

// parseGoMod parses go.mod file content
func parseGoMod(content string) map[string]interface{} {
	info := map[string]interface{}{
		"module":  "",
		"go":      "",
		"require": []string{},
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				info["module"] = parts[1]
			}
		} else if strings.HasPrefix(line, "go ") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				info["go"] = parts[1]
			}
		} else if strings.HasPrefix(line, "require ") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				requires := info["require"].([]string)
				info["require"] = append(requires, strings.Join(parts[1:], " "))
			}
		}
	}

	return info
}

// discoverBuildTags discovers build tags in Go files
func discoverBuildTags(dir string) map[string]interface{} {
	tags := make(map[string]bool)
	files := []string{}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
			content, err := os.ReadFile(path)
			if err == nil {
				// Simple tag extraction (look for +build comments)
				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					if strings.HasPrefix(strings.TrimSpace(line), "// +build ") {
						tagLine := strings.TrimPrefix(strings.TrimSpace(line), "// +build ")
						for _, tag := range strings.Fields(tagLine) {
							tags[tag] = true
						}
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	tagList := make([]string, 0, len(tags))
	for tag := range tags {
		tagList = append(tagList, tag)
	}

	return map[string]interface{}{
		"tags":  tagList,
		"files": files,
	}
}

// discoverTestFiles discovers test files
func discoverTestFiles(dir string) map[string]interface{} {
	testFiles := []string{}
	benchmarkFiles := []string{}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, "_test.go") {
			content, err := os.ReadFile(path)
			if err == nil {
				contentStr := string(content)
				testFiles = append(testFiles, path)
				if strings.Contains(contentStr, "func Benchmark") {
					benchmarkFiles = append(benchmarkFiles, path)
				}
			}
		}
		return nil
	})

	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	return map[string]interface{}{
		"test_files":      testFiles,
		"benchmark_files": benchmarkFiles,
	}
}

// discoverWorkspace discovers workspace structure
func discoverWorkspace(dir string) map[string]interface{} {
	workspace := map[string]interface{}{
		"working_directory": dir,
		"has_go_mod":        false,
		"has_go_work":       false,
		"modules":           []string{},
	}

	// Check for go.mod
	modPath := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(modPath); err == nil {
		workspace["has_go_mod"] = true
		modData, err := os.ReadFile(modPath)
		if err == nil {
			modInfo := parseGoMod(string(modData))
			if module, ok := modInfo["module"].(string); ok {
				workspace["modules"] = []string{module}
			}
		}
	}

	// Check for go.work
	workPath := filepath.Join(dir, "go.work")
	if _, err := os.Stat(workPath); err == nil {
		workspace["has_go_work"] = true
		// Parse go.work file
		workData, err := os.ReadFile(workPath)
		if err == nil {
			modules := parseGoWork(string(workData))
			workspace["modules"] = modules
		}
	}

	return workspace
}

// parseGoWork parses go.work file
func parseGoWork(content string) []string {
	modules := []string{}
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "use ") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				modules = append(modules, parts[1])
			}
		}
	}
	return modules
}
