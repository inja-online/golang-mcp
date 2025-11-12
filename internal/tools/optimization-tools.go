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

// RegisterOptimizationTools registers optimization and profiling tools
func RegisterOptimizationTools(server *mcp.Server, cfg *config.Config) int {
	count := 0
	// go_profile tool
	resources.RegisterTool("go_profile", "Generate performance profile using pprof. Creates CPU or memory profiles for analysis.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_profile",
		Description: "Generate performance profile using pprof. Creates CPU or memory profiles for analysis.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Type       string `json:"type" jsonschema:"required"`
		Output     string `json:"output" jsonschema:"required"`
		Duration   string `json:"duration,omitempty"`
		Package    string `json:"package,omitempty"`
		WorkingDir string `json:"working_dir,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		var goArgs []string

		switch args.Type {
		case "cpu":
			goArgs = []string{"test", "-cpuprofile", args.Output}
		case "mem":
			goArgs = []string{"test", "-memprofile", args.Output}
		default:
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Unknown profile type: %s. Supported: cpu, mem", args.Type)},
				},
				IsError: true,
			}, nil, nil
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

		output := fmt.Sprintf("Profile generated: %s\n\n%s", args.Output, formatCommandResult(result))
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, result, nil
	})
	count++

	// go_trace tool
	resources.RegisterTool("go_trace", "Generate execution trace for Go programs.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_trace",
		Description: "Generate execution trace for Go programs.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Output     string `json:"output" jsonschema:"required"`
		Package    string `json:"package,omitempty"`
		WorkingDir string `json:"working_dir,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		goArgs := []string{"test", "-trace", args.Output}
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

		output := fmt.Sprintf("Trace generated: %s\n\n%s", args.Output, formatCommandResult(result))
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, result, nil
	})
	count++

	// go_benchmark tool
	resources.RegisterTool("go_benchmark", "Run benchmarks and analyze results. Supports filtering and custom benchmark settings.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_benchmark",
		Description: "Run benchmarks and analyze results. Supports filtering and custom benchmark settings.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Pattern    string `json:"pattern,omitempty"`
		Count      int    `json:"count,omitempty"`
		Timeout    string `json:"timeout,omitempty"`
		Package    string `json:"package,omitempty"`
		WorkingDir string `json:"working_dir,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		goArgs := []string{"test", "-bench", "."}
		if args.Pattern != "" {
			goArgs = append(goArgs, "-bench", args.Pattern)
		}
		if args.Count > 0 {
			goArgs = append(goArgs, "-count", fmt.Sprintf("%d", args.Count))
		}
		if args.Timeout != "" {
			goArgs = append(goArgs, "-timeout", args.Timeout)
		}
		goArgs = append(goArgs, "-benchmem")

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

	// go_race_detect tool
	resources.RegisterTool("go_race_detect", "Detect race conditions in Go code using the race detector.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_race_detect",
		Description: "Detect race conditions in Go code using the race detector.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Package    string `json:"package,omitempty"`
		WorkingDir string `json:"working_dir,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		goArgs := []string{"test", "-race", "-v"}
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
		if result.ExitCode == 0 {
			output = "No race conditions detected.\n\n" + output
		} else {
			output = "Race conditions detected!\n\n" + output
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, result, nil
	})
	count++

	// go_memory_profile tool
	resources.RegisterTool("go_memory_profile", "Generate memory profile for memory usage analysis.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_memory_profile",
		Description: "Generate memory profile for memory usage analysis.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Output     string `json:"output" jsonschema:"required"`
		Package    string `json:"package,omitempty"`
		WorkingDir string `json:"working_dir,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		goArgs := []string{"test", "-memprofile", args.Output}
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

		output := fmt.Sprintf("Memory profile generated: %s\n\n%s", args.Output, formatCommandResult(result))
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, result, nil
	})
	count++

	// go_optimize_suggest tool
	resources.RegisterTool("go_optimize_suggest", "Analyze code and provide optimization suggestions based on profiling and benchmarking.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_optimize_suggest",
		Description: "Analyze code and provide optimization suggestions based on profiling and benchmarking.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Package    string `json:"package,omitempty"`
		WorkingDir string `json:"working_dir,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		var suggestions []string

		// Run benchmarks to get baseline
		benchResult, err := utils.ExecuteGoCommand(ctx, cfg, "go", []string{"test", "-bench", ".", "-benchmem"}, args.WorkingDir, nil)
		if err == nil && benchResult.Stdout != "" {
			suggestions = append(suggestions, "Benchmark Results:\n"+benchResult.Stdout)
		}

		// Check for common issues
		goArgs := []string{"vet", "-all"}
		if args.Package != "" {
			goArgs = append(goArgs, args.Package)
		}
		vetResult, err := utils.ExecuteGoCommand(ctx, cfg, "go", goArgs, args.WorkingDir, nil)
		if err == nil && vetResult.Stderr != "" {
			suggestions = append(suggestions, "Static Analysis Issues:\n"+vetResult.Stderr)
		}

		// General suggestions
		suggestions = append(suggestions, "Optimization Suggestions:")
		suggestions = append(suggestions, "1. Use -race flag during testing to detect race conditions")
		suggestions = append(suggestions, "2. Profile CPU and memory usage with go tool pprof")
		suggestions = append(suggestions, "3. Use benchmarks to measure performance improvements")
		suggestions = append(suggestions, "4. Consider using sync.Pool for frequently allocated objects")
		suggestions = append(suggestions, "5. Review use of channels and goroutines for concurrency")

		output := strings.Join(suggestions, "\n\n")
		return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: output},
				},
			}, map[string]interface{}{
				"suggestions": suggestions,
			}, nil
	})
	count++
	return count
}
