package main

import (
	"context"
	"log"
	"os"

	"github.com/inja-online/golang-mcp/internal/config"
	"github.com/inja-online/golang-mcp/internal/prompts"
	"github.com/inja-online/golang-mcp/internal/resources"
	"github.com/inja-online/golang-mcp/internal/tools"
	"github.com/inja-online/golang-mcp/internal/utils"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	version = "0.0.1"
	name    = "mcp-go-server"
)

var debugLog func(format string, v ...interface{})

func initDebugLogging(cfg *config.Config) {
	if cfg.DebugMCP {
		debugLog = func(format string, v ...interface{}) {
			log.Printf("[DEBUG] "+format, v...)
		}
	} else {
		debugLog = func(format string, v ...interface{}) {}
	}
}

func main() {
	// CRITICAL: Set logging to stderr (MCP requirement)
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Println("Starting MCP Go Server...")

	// Load configuration
	cfg := config.Load()
	initDebugLogging(cfg)

	debugLog("Configuration loaded: DebugMCP=%v, WorkingDirectory=%s", cfg.DebugMCP, cfg.WorkingDirectory)

	// Create MCP server
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    name,
			Version: version,
		},
		nil,
	)

	debugLog("MCP server created: %s v%s", name, version)

	// Initialize subsystems
	tools.InitServerManager()
	tools.InitPackageDocsCache()
	debugLog("Subsystems initialized: ServerManager, PackageDocsCache")

	// Register all tools
	log.Println("Registering tools...")
	toolCount := 0
	runToolsCount := tools.RegisterRunTools(server, cfg)
	debugLog("Registered run tools: %d tools", runToolsCount)
	toolCount += runToolsCount
	goToolsCount := tools.RegisterGoTools(server, cfg)
	debugLog("Registered Go tools: %d tools", goToolsCount)
	toolCount += goToolsCount
	optToolsCount := tools.RegisterOptimizationTools(server, cfg)
	debugLog("Registered optimization tools: %d tools", optToolsCount)
	toolCount += optToolsCount
	serverToolsCount := tools.RegisterServerTools(server, cfg)
	debugLog("Registered server tools: %d tools", serverToolsCount)
	toolCount += serverToolsCount
	pkgDocsToolsCount := tools.RegisterPackageDocsTools(server, cfg)
	debugLog("Registered package docs tools: %d tools", pkgDocsToolsCount)
	toolCount += pkgDocsToolsCount

	// Conditionally register LSP tools
	if cfg.EnableLSP {
		lspToolsCount := tools.RegisterLSPTools(server, cfg)
		debugLog("Registered LSP tools: %d tools", lspToolsCount)
		toolCount += lspToolsCount
	}

	log.Printf("Registered %d tools total", toolCount)

	// Register resources
	log.Println("Registering resources...")
	resourceCount := resources.RegisterGoResources(server, cfg)
	debugLog("Registered %d resources", resourceCount)
	log.Printf("Registered %d resources", resourceCount)

	// Register prompts
	log.Println("Registering prompts...")
	promptCount := prompts.RegisterGoPrompts(server, cfg)
	debugLog("Registered %d prompts", promptCount)
	log.Printf("Registered %d prompts", promptCount)

	debugLog("Server initialization complete: %d tools, %d resources, %d prompts", toolCount, resourceCount, promptCount)

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	utils.SetupSignalHandling(cancel)

	// Run server with stdio transport
	log.Println("Starting server on stdio transport...")
	debugLog("Waiting for MCP client connection...")
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
