package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/inja-online/golang-mcp/internal/config"
	"github.com/inja-online/golang-mcp/internal/resources"
	"github.com/inja-online/golang-mcp/internal/utils"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var serverManager *utils.ServerManager

// InitServerManager initializes the server manager (called from main)
func InitServerManager() {
	serverManager = utils.NewServerManager()
}

// RegisterServerTools registers server management tools
func RegisterServerTools(server *mcp.Server, cfg *config.Config) int {
	count := 0
	// go_server_start tool
	resources.RegisterTool("go_server_start", "Start a long-running Go server in the background. Returns a server ID for management.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_server_start",
		Description: "Start a long-running Go server in the background. Returns a server ID for management.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		ID         string            `json:"id" jsonschema:"required"`
		Name       string            `json:"name" jsonschema:"required"`
		Command    string            `json:"command" jsonschema:"required"`
		Args       []string          `json:"args" jsonschema:"required"`
		WorkingDir string            `json:"working_dir,omitempty"`
		EnvVars    map[string]string `json:"env_vars,omitempty"`
		LogSize    int               `json:"log_size,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		if serverManager == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Server manager not initialized"},
				},
				IsError: true,
			}, nil, nil
		}

		serverInfo, err := serverManager.StartServer(ctx, cfg, args.ID, args.Name, args.Command, args.Args, args.WorkingDir, args.EnvVars, args.LogSize)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error starting server: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		output := fmt.Sprintf("Server started successfully\nID: %s\nPID: %d\nStatus: %s\n", serverInfo.ID, serverInfo.PID, serverInfo.Status)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, serverInfo, nil
	})
	count++

	// go_server_stop tool
	resources.RegisterTool("go_server_stop", "Stop a running server. Uses graceful shutdown (SIGTERM) by default, or force kill if specified.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_server_stop",
		Description: "Stop a running server. Uses graceful shutdown (SIGTERM) by default, or force kill if specified.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		ID    string `json:"id" jsonschema:"required"`
		Force bool   `json:"force,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		if serverManager == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Server manager not initialized"},
				},
				IsError: true,
			}, nil, nil
		}

		err := serverManager.StopServer(args.ID, args.Force)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error stopping server: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		output := fmt.Sprintf("Server %s stopped successfully", args.ID)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, nil, nil
	})
	count++

	// go_server_list tool
	resources.RegisterTool("go_server_list", "List all running servers with their status and metadata.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_server_list",
		Description: "List all running servers with their status and metadata.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
	}) (*mcp.CallToolResult, any, error) {
		if serverManager == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Server manager not initialized"},
				},
				IsError: true,
			}, nil, nil
		}

		servers := serverManager.ListServers()
		if len(servers) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "No servers running"},
				},
			}, []interface{}{}, nil
		}

		var output strings.Builder
		output.WriteString(fmt.Sprintf("Running Servers (%d):\n\n", len(servers)))

		serverList := make([]map[string]interface{}, 0, len(servers))
		for _, server := range servers {
			serverData := map[string]interface{}{
				"id":         server.ID,
				"name":       server.Name,
				"pid":        server.PID,
				"status":     server.Status,
				"start_time": server.StartTime.Format(time.RFC3339),
				"command":    server.Command,
				"args":       server.Args,
			}
			if server.ExitCode != nil {
				serverData["exit_code"] = *server.ExitCode
			}
			serverList = append(serverList, serverData)

			output.WriteString(fmt.Sprintf("ID: %s\n", server.ID))
			output.WriteString(fmt.Sprintf("Name: %s\n", server.Name))
			output.WriteString(fmt.Sprintf("PID: %d\n", server.PID))
			output.WriteString(fmt.Sprintf("Status: %s\n", server.Status))
			output.WriteString(fmt.Sprintf("Started: %s\n", server.StartTime.Format(time.RFC3339)))
			output.WriteString("\n")
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output.String()},
			},
		}, serverList, nil
	})
	count++

	// go_server_logs tool
	resources.RegisterTool("go_server_logs", "Get logs from a running server. Returns recent logs or all logs if count is 0.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_server_logs",
		Description: "Get logs from a running server. Returns recent logs or all logs if count is 0.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		ID    string `json:"id" jsonschema:"required"`
		Count int    `json:"count,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		if serverManager == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Server manager not initialized"},
				},
				IsError: true,
			}, nil, nil
		}

		logs, err := serverManager.GetServerLogs(args.ID, args.Count)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error getting logs: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		output := fmt.Sprintf("Logs for server %s (%d lines):\n\n%s", args.ID, len(logs), strings.Join(logs, "\n"))
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, logs, nil
	})
	count++

	// go_server_status tool
	resources.RegisterTool("go_server_status", "Get detailed status of a server including PID, uptime, and logs.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_server_status",
		Description: "Get detailed status of a server including PID, uptime, and logs.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		ID string `json:"id" jsonschema:"required"`
	}) (*mcp.CallToolResult, any, error) {
		if serverManager == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Server manager not initialized"},
				},
				IsError: true,
			}, nil, nil
		}

		serverInfo, err := serverManager.GetServer(args.ID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}

		uptime := time.Since(serverInfo.StartTime)
		output := "Server Status:\n"
		output += fmt.Sprintf("ID: %s\n", serverInfo.ID)
		output += fmt.Sprintf("Name: %s\n", serverInfo.Name)
		output += fmt.Sprintf("PID: %d\n", serverInfo.PID)
		output += fmt.Sprintf("Status: %s\n", serverInfo.Status)
		output += fmt.Sprintf("Uptime: %v\n", uptime)
		output += fmt.Sprintf("Command: %s %v\n", serverInfo.Command, serverInfo.Args)
		if serverInfo.ExitCode != nil {
			output += fmt.Sprintf("Exit Code: %d\n", *serverInfo.ExitCode)
		}

		statusData := map[string]interface{}{
			"id":         serverInfo.ID,
			"name":       serverInfo.Name,
			"pid":        serverInfo.PID,
			"status":     serverInfo.Status,
			"uptime":     uptime.String(),
			"start_time": serverInfo.StartTime.Format(time.RFC3339),
			"command":    serverInfo.Command,
			"args":       serverInfo.Args,
		}
		if serverInfo.ExitCode != nil {
			statusData["exit_code"] = *serverInfo.ExitCode
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, statusData, nil
	})
	count++
	return count
}
