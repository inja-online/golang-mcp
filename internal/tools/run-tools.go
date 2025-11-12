package tools

import (
	"context"
	"fmt"

	"github.com/inja-online/golang-mcp/internal/config"
	"github.com/inja-online/golang-mcp/internal/resources"
	"github.com/inja-online/golang-mcp/internal/utils"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterRunTools registers code execution tools
func RegisterRunTools(server *mcp.Server, cfg *config.Config) int {
	count := 0
	// go_run tool
	resources.RegisterTool("go_run", "Execute a Go file directly using 'go run'. Runs the specified Go file with optional arguments and environment variables.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_run",
		Description: "Execute a Go file directly using 'go run'. Runs the specified Go file with optional arguments and environment variables.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		File       string            `json:"file" jsonschema:"required"`
		Args       []string          `json:"args,omitempty"`
		WorkingDir string            `json:"working_dir,omitempty"`
		EnvVars    map[string]string `json:"env_vars,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		// Build go run command
		goArgs := []string{"run", args.File}
		if len(args.Args) > 0 {
			goArgs = append(goArgs, "--")
			goArgs = append(goArgs, args.Args...)
		}

		// Execute command
		result, err := utils.ExecuteGoCommand(ctx, cfg, "go", goArgs, args.WorkingDir, args.EnvVars)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		// Format result
		output := fmt.Sprintf("Exit Code: %d\nDuration: %v\n\n", result.ExitCode, result.Duration)
		if result.Stdout != "" {
			output += fmt.Sprintf("STDOUT:\n%s\n", result.Stdout)
		}
		if result.Stderr != "" {
			output += fmt.Sprintf("STDERR:\n%s\n", result.Stderr)
		}

		// Create result data
		resultData := map[string]interface{}{
			"exit_code": result.ExitCode,
			"stdout":    result.Stdout,
			"stderr":    result.Stderr,
			"duration":  result.Duration.String(),
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, resultData, nil
	})
	count++
	return count
}
