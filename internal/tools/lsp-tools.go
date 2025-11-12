package tools

import (
	"context"
	"fmt"

	"github.com/inja-online/golang-mcp/internal/config"
	"github.com/inja-online/golang-mcp/internal/lsp"
	"github.com/inja-online/golang-mcp/internal/resources"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterLSPTools registers LSP-related tools. Returns number of tools registered.
func RegisterLSPTools(server *mcp.Server, cfg *config.Config) int {
	count := 0

	// Initialize manager for sessions
	manager := lsp.NewManager()
	_ = manager // TODO: wire a shared manager instance into server lifecycle if needed

	// lsp_start_session
	resources.RegisterTool("lsp_start_session", "Start an LSP session for a workspace root URI.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "lsp_start_session",
		Description: "Start an LSP session for a workspace root URI.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		RootURI   string            `json:"root_uri" jsonschema:"required"`
		GoplsPath string            `json:"gopls_path,omitempty"`
		Args      []string          `json:"args,omitempty"`
		Env       map[string]string `json:"env,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		// TODO: call manager.StartSession and return session handle info
		_, err := manager.StartSession(ctx, args.RootURI, lsp.SessionOptions{
			GoplsPath: args.GoplsPath,
			Args:      args.Args,
			Env:       args.Env,
		})
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("error starting session: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "session started"},
			},
		}, map[string]string{"root_uri": args.RootURI}, nil
	})
	count++

	// lsp_shutdown_session
	resources.RegisterTool("lsp_shutdown_session", "Shutdown an LSP session for a workspace root URI.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "lsp_shutdown_session",
		Description: "Shutdown an LSP session for a workspace root URI.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		RootURI string `json:"root_uri" jsonschema:"required"`
	}) (*mcp.CallToolResult, any, error) {
		// TODO: call manager.ShutdownSession
		if err := manager.ShutdownSession(ctx, args.RootURI); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("error shutting down session: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "session shutdown"},
			},
		}, map[string]string{"root_uri": args.RootURI}, nil
	})
	count++

	// lsp_request
	resources.RegisterTool("lsp_request", "Send a request to an LSP session and wait for a result.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "lsp_request",
		Description: "Send a request to an LSP session and wait for a result.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		RootURI string      `json:"root_uri" jsonschema:"required"`
		Method  string      `json:"method" jsonschema:"required"`
		Params  interface{} `json:"params,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		// TODO: locate session and proxy Request to it
		if _, ok := manager.GetSession(args.RootURI); !ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "session not found"},
				},
				IsError: true,
			}, nil, nil
		}

		// Placeholder: return minimal result
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "request dispatched (stub)"},
			},
		}, map[string]interface{}{"method": args.Method}, nil
	})
	count++

	// lsp_notify
	resources.RegisterTool("lsp_notify", "Send a notification to an LSP session.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "lsp_notify",
		Description: "Send a notification to an LSP session.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		RootURI string      `json:"root_uri" jsonschema:"required"`
		Method  string      `json:"method" jsonschema:"required"`
		Params  interface{} `json:"params,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		// TODO: locate session and proxy Notify to it
		if _, ok := manager.GetSession(args.RootURI); !ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "session not found"},
				},
				IsError: true,
			}, nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "notification dispatched (stub)"},
			},
		}, nil, nil
	})
	count++

	// lsp_subscribe_diagnostics
	resources.RegisterTool("lsp_subscribe_diagnostics", "Subscribe to diagnostics published by an LSP session.", nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "lsp_subscribe_diagnostics",
		Description: "Subscribe to diagnostics published by an LSP session.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		RootURI string `json:"root_uri" jsonschema:"required"`
	}) (*mcp.CallToolResult, any, error) {
		// TODO: wire up diagnostics subscription using manager and SessionHandle.SubscribeDiagnostics
		if _, ok := manager.GetSession(args.RootURI); !ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "session not found"},
				},
				IsError: true,
			}, nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "subscribed to diagnostics (stub)"},
			},
		}, nil, nil
	})
	count++

	return count
}
