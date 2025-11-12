package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/inja-online/golang-mcp/internal/config"
	"github.com/inja-online/golang-mcp/internal/resources"
	"github.com/inja-online/golang-mcp/internal/utils"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var pkgDocsCache *utils.Cache

// InitPackageDocsCache initializes the package documentation cache (called from main)
func InitPackageDocsCache() {
	cacheDir := filepath.Join(os.TempDir(), "mcp-golang-cache")
	_ = os.MkdirAll(cacheDir, 0755)
	cachePath := filepath.Join(cacheDir, "pkg-docs.json")
	pkgDocsCache = utils.NewCache(24*time.Hour, true, cachePath)
}

// RegisterPackageDocsTools registers package documentation tools
func RegisterPackageDocsTools(server *mcp.Server, cfg *config.Config) int {
	count := 0
	// go_pkg_docs tool
	resources.RegisterTool("go_pkg_docs", "Fetch package documentation from go.dev (pkg.go.dev). Returns package overview, functions, types, and examples.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_pkg_docs",
		Description: "Fetch package documentation from go.dev (pkg.go.dev). Returns package overview, functions, types, and examples.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Package string `json:"package" jsonschema:"required"`
		Version string `json:"version,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		if pkgDocsCache == nil {
			InitPackageDocsCache()
		}

		doc, err := utils.FetchPackageDocs(ctx, args.Package, args.Version, pkgDocsCache)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error fetching documentation: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		// Format documentation for display
		var output strings.Builder
		output.WriteString(fmt.Sprintf("Package: %s\n", doc.Path))
		if doc.Version != "" {
			output.WriteString(fmt.Sprintf("Version: %s\n", doc.Version))
		}
		output.WriteString("\n")
		if doc.Overview != "" {
			output.WriteString(fmt.Sprintf("Overview:\n%s\n\n", doc.Overview))
		}
		if doc.Description != "" {
			output.WriteString(fmt.Sprintf("Description:\n%s\n\n", doc.Description))
		}

		if len(doc.Functions) > 0 {
			output.WriteString("Functions:\n")
			for _, fn := range doc.Functions {
				output.WriteString(fmt.Sprintf("  %s %s\n    %s\n", fn.Name, fn.Signature, fn.Description))
			}
			output.WriteString("\n")
		}

		if len(doc.Types) > 0 {
			output.WriteString("Types:\n")
			for _, typ := range doc.Types {
				output.WriteString(fmt.Sprintf("  %s (%s)\n    %s\n", typ.Name, typ.Kind, typ.Description))
			}
			output.WriteString("\n")
		}

		if len(doc.Examples) > 0 {
			output.WriteString("Examples:\n")
			for _, ex := range doc.Examples {
				output.WriteString(fmt.Sprintf("  %s:\n%s\n", ex.Name, ex.Code))
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output.String()},
			},
		}, doc, nil
	})
	count++

	// go_pkg_search tool
	resources.RegisterTool("go_pkg_search", "Search for packages on go.dev. Returns a list of matching package paths.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_pkg_search",
		Description: "Search for packages on go.dev. Returns a list of matching package paths.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Query string `json:"query" jsonschema:"required"`
	}) (*mcp.CallToolResult, any, error) {
		packages, err := utils.SearchPackages(ctx, args.Query)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error searching packages: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		if len(packages) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("No packages found for query: %s", args.Query)},
				},
			}, map[string]interface{}{"packages": []string{}}, nil
		}

		output := fmt.Sprintf("Found %d packages:\n\n", len(packages))
		for i, pkg := range packages {
			output += fmt.Sprintf("%d. %s\n", i+1, pkg)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, map[string]interface{}{"packages": packages}, nil
	})
	count++

	// go_pkg_examples tool
	resources.RegisterTool("go_pkg_examples", "Extract and return examples from package documentation.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_pkg_examples",
		Description: "Extract and return examples from package documentation.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Package string `json:"package" jsonschema:"required"`
		Version string `json:"version,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		if pkgDocsCache == nil {
			InitPackageDocsCache()
		}

		doc, err := utils.FetchPackageDocs(ctx, args.Package, args.Version, pkgDocsCache)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error fetching documentation: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		if len(doc.Examples) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("No examples found for package: %s", args.Package)},
				},
			}, []interface{}{}, nil
		}

		var output strings.Builder
		output.WriteString(fmt.Sprintf("Examples for %s:\n\n", args.Package))
		for i, ex := range doc.Examples {
			output.WriteString(fmt.Sprintf("Example %d: %s\n", i+1, ex.Name))
			if ex.Description != "" {
				output.WriteString(fmt.Sprintf("Description: %s\n", ex.Description))
			}
			output.WriteString(fmt.Sprintf("Code:\n%s\n\n", ex.Code))
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output.String()},
			},
		}, doc.Examples, nil
	})
	count++
	return count
}
