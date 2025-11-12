package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/inja-online/golang-mcp/internal/config"
	"github.com/inja-online/golang-mcp/internal/resources"
	"github.com/inja-online/golang-mcp/internal/utils"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterGoTools registers core Go operation tools
func RegisterGoTools(server *mcp.Server, cfg *config.Config) int {
	count := 0
	// go_build tool
	resources.RegisterTool("go_build", "Build Go packages and dependencies. Supports various build flags like -race, -tags, -ldflags, etc.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_build",
		Description: "Build Go packages and dependencies. Supports various build flags like -race, -tags, -ldflags, etc.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Package    string   `json:"package,omitempty"`
		Output     string   `json:"output,omitempty"`
		Race       bool     `json:"race,omitempty"`
		Tags       []string `json:"tags,omitempty"`
		LDFlags    string   `json:"ldflags,omitempty"`
		TrimPath   bool     `json:"trimpath,omitempty"`
		WorkingDir string   `json:"working_dir,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		goArgs := []string{"build"}

		if args.Race {
			goArgs = append(goArgs, "-race")
		}
		if len(args.Tags) > 0 {
			goArgs = append(goArgs, "-tags", strings.Join(args.Tags, ","))
		}
		if args.LDFlags != "" {
			goArgs = append(goArgs, "-ldflags", args.LDFlags)
		}
		if args.TrimPath {
			goArgs = append(goArgs, "-trimpath")
		}
		if args.Output != "" {
			goArgs = append(goArgs, "-o", args.Output)
		}
		if args.Package != "" {
			goArgs = append(goArgs, args.Package)
		}

		result, err := utils.ExecuteGoCommand(ctx, cfg, "go", goArgs, args.WorkingDir, nil)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v\n%s", err, result.Stderr)},
				},
				IsError: true,
			}, nil, nil
		}

		output := formatCommandResult(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, result, nil
	})
	count++

	// go_test tool
	resources.RegisterTool("go_test", "Run Go tests with coverage, benchmarks, and race detection. Supports various test flags.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_test",
		Description: "Run Go tests with coverage, benchmarks, and race detection. Supports various test flags.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Package    string `json:"package,omitempty"`
		Cover      bool   `json:"cover,omitempty"`
		CoverPkg   string `json:"cover_pkg,omitempty"`
		Bench      bool   `json:"bench,omitempty"`
		Race       bool   `json:"race,omitempty"`
		Verbose    bool   `json:"verbose,omitempty"`
		Timeout    string `json:"timeout,omitempty"`
		WorkingDir string `json:"working_dir,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		goArgs := []string{"test"}

		if args.Cover {
			goArgs = append(goArgs, "-cover")
		}
		if args.CoverPkg != "" {
			goArgs = append(goArgs, "-coverpkg", args.CoverPkg)
		}
		if args.Bench {
			goArgs = append(goArgs, "-bench", ".")
		}
		if args.Race {
			goArgs = append(goArgs, "-race")
		}
		if args.Verbose {
			goArgs = append(goArgs, "-v")
		}
		if args.Timeout != "" {
			goArgs = append(goArgs, "-timeout", args.Timeout)
		}
		if args.Package != "" {
			goArgs = append(goArgs, args.Package)
		}

		result, err := utils.ExecuteGoCommand(ctx, cfg, "go", goArgs, args.WorkingDir, nil)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v\n%s", err, result.Stderr)},
				},
				IsError: true,
			}, nil, nil
		}

		output := formatCommandResult(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, result, nil
	})
	count++

	// go_fmt tool
	resources.RegisterTool("go_fmt", "Format Go code using 'go fmt'. Formats the specified package or files.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_fmt",
		Description: "Format Go code using 'go fmt'. Formats the specified package or files.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Paths      []string `json:"paths,omitempty"`
		WorkingDir string   `json:"working_dir,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		goArgs := []string{"fmt"}
		if len(args.Paths) > 0 {
			goArgs = append(goArgs, args.Paths...)
		}

		result, err := utils.ExecuteGoCommand(ctx, cfg, "go", goArgs, args.WorkingDir, nil)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v\n%s", err, result.Stderr)},
				},
				IsError: true,
			}, nil, nil
		}

		output := formatCommandResult(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, result, nil
	})
	count++

	// go_mod tool
	resources.RegisterTool("go_mod", "Manage Go modules. Supports init, tidy, download, vendor, and get operations.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_mod",
		Description: "Manage Go modules. Supports init, tidy, download, vendor, and get operations.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Operation  string   `json:"operation" jsonschema:"required"`
		ModulePath string   `json:"module_path,omitempty"`
		Packages   []string `json:"packages,omitempty"`
		WorkingDir string   `json:"working_dir,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		var goArgs []string

		switch args.Operation {
		case "init":
			goArgs = []string{"mod", "init"}
			if args.ModulePath != "" {
				goArgs = append(goArgs, args.ModulePath)
			}
		case "tidy":
			goArgs = []string{"mod", "tidy"}
		case "download":
			goArgs = []string{"mod", "download"}
		case "vendor":
			goArgs = []string{"mod", "vendor"}
		case "get":
			goArgs = []string{"mod", "get"}
			if len(args.Packages) > 0 {
				goArgs = append(goArgs, args.Packages...)
			}
		default:
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Unknown operation: %s. Supported: init, tidy, download, vendor, get", args.Operation)},
				},
				IsError: true,
			}, nil, nil
		}

		result, err := utils.ExecuteGoCommand(ctx, cfg, "go", goArgs, args.WorkingDir, nil)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v\n%s", err, result.Stderr)},
				},
				IsError: true,
			}, nil, nil
		}

		output := formatCommandResult(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, result, nil
	})
	count++

	// go_doc tool
	resources.RegisterTool("go_doc", "Generate documentation for Go packages using 'go doc'.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_doc",
		Description: "Generate documentation for Go packages using 'go doc'.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Package    string `json:"package" jsonschema:"required"`
		All        bool   `json:"all,omitempty"`
		WorkingDir string `json:"working_dir,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		goArgs := []string{"doc"}
		if args.All {
			goArgs = append(goArgs, "-all")
		}
		goArgs = append(goArgs, args.Package)

		result, err := utils.ExecuteGoCommand(ctx, cfg, "go", goArgs, args.WorkingDir, nil)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v\n%s", err, result.Stderr)},
				},
				IsError: true,
			}, nil, nil
		}

		output := formatCommandResult(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, result, nil
	})
	count++

	// go_lint tool
	resources.RegisterTool("go_lint", "Lint Go code using golangci-lint (if available) or go vet.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_lint",
		Description: "Lint Go code using golangci-lint (if available) or go vet.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Package    string `json:"package,omitempty"`
		Linter     string `json:"linter,omitempty"`
		WorkingDir string `json:"working_dir,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		var command string
		var cmdArgs []string

		if args.Linter == "golangci-lint" {
			command = "golangci-lint"
			cmdArgs = []string{"run"}
			if args.Package != "" {
				cmdArgs = append(cmdArgs, args.Package)
			}
		} else {
			command = "go"
			cmdArgs = []string{"vet"}
			if args.Package != "" {
				cmdArgs = append(cmdArgs, args.Package)
			}
		}

		result, err := utils.ExecuteGoCommand(ctx, cfg, command, cmdArgs, args.WorkingDir, nil)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v\n%s", err, result.Stderr)},
				},
				IsError: true,
			}, nil, nil
		}

		output := formatCommandResult(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, result, nil
	})
	count++

	// go_cross_compile tool
	resources.RegisterTool("go_cross_compile", "Cross-compile Go code for different platforms and architectures.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_cross_compile",
		Description: "Cross-compile Go code for different platforms and architectures.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Package    string `json:"package,omitempty"`
		Output     string `json:"output" jsonschema:"required"`
		GOOS       string `json:"goos" jsonschema:"required"`
		GOARCH     string `json:"goarch" jsonschema:"required"`
		WorkingDir string `json:"working_dir,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		goArgs := []string{"build", "-o", args.Output}
		if args.Package != "" {
			goArgs = append(goArgs, args.Package)
		}

		envVars := map[string]string{
			"GOOS":   args.GOOS,
			"GOARCH": args.GOARCH,
		}

		result, err := utils.ExecuteGoCommand(ctx, cfg, "go", goArgs, args.WorkingDir, envVars)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v\n%s", err, result.Stderr)},
				},
				IsError: true,
			}, nil, nil
		}

		output := formatCommandResult(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, result, nil
	})
	count++
	return count
}

// formatCommandResult formats a command result for display
func formatCommandResult(result *utils.CommandResult) string {
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Exit Code: %d\n", result.ExitCode))
	output.WriteString(fmt.Sprintf("Duration: %v\n\n", result.Duration))

	if result.Stdout != "" {
		output.WriteString(fmt.Sprintf("STDOUT:\n%s\n", result.Stdout))
	}
	if result.Stderr != "" {
		output.WriteString(fmt.Sprintf("STDERR:\n%s\n", result.Stderr))
	}

	return output.String()
}
