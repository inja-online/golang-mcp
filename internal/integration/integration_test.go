//go:build integration

package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/inja-online/golang-mcp/internal/config"
	"github.com/inja-online/golang-mcp/internal/resources"
	"github.com/inja-online/golang-mcp/internal/tools"
	"github.com/inja-online/golang-mcp/internal/utils"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestEndToEnd_GoRun(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("Go command not available")
	}

	dir := t.TempDir()
	cfg := &config.Config{
		DisableNotifications: true,
		WorkingDirectory:     dir,
	}

	// Create a simple Go program
	goFile := filepath.Join(dir, "main.go")
	goContent := `package main
import "fmt"
func main() {
	fmt.Println("Hello from integration test!")
}
`
	if err := os.WriteFile(goFile, []byte(goContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create go.mod
	modFile := filepath.Join(dir, "go.mod")
	modContent := "module test\n\ngo 1.24\n"
	if err := os.WriteFile(modFile, []byte(modContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute go run
	result, err := utils.ExecuteGoCommand(ctx, cfg, "go", []string{"run", goFile}, dir, nil)
	if err != nil {
		t.Fatalf("Failed to execute go run: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", result.ExitCode, result.Stderr)
	}

	if !contains(result.Stdout, "Hello from integration test!") {
		t.Errorf("Expected output to contain 'Hello from integration test!', got: %s", result.Stdout)
	}
}

func TestEndToEnd_GoBuild(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("Go command not available")
	}

	dir := t.TempDir()
	cfg := &config.Config{
		DisableNotifications: true,
		WorkingDirectory:     dir,
	}

	// Create a simple Go program
	goFile := filepath.Join(dir, "main.go")
	goContent := `package main
import "fmt"
func main() {
	fmt.Println("Built successfully!")
}
`
	if err := os.WriteFile(goFile, []byte(goContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create go.mod
	modFile := filepath.Join(dir, "go.mod")
	modContent := "module test\n\ngo 1.24\n"
	if err := os.WriteFile(modFile, []byte(modContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	outputFile := filepath.Join(dir, "test-binary")
	if os.Getenv("GOOS") == "windows" {
		outputFile += ".exe"
	}

	// Execute go build
	result, err := utils.ExecuteGoCommand(ctx, cfg, "go", []string{"build", "-o", outputFile, goFile}, dir, nil)
	if err != nil {
		t.Fatalf("Failed to execute go build: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", result.ExitCode, result.Stderr)
	}

	// Verify binary was created
	if _, err := os.Stat(outputFile); err != nil {
		t.Errorf("Expected binary to be created at %s, got error: %v", outputFile, err)
	}
}

func TestEndToEnd_ServerManagement(t *testing.T) {
	cfg := &config.Config{
		DisableNotifications: true,
		WorkingDirectory:     t.TempDir(),
	}

	sm := utils.NewServerManager()
	ctx := context.Background()

	// Start a simple server (echo with sleep to keep it running)
	var cmd string
	var args []string
	if exec.Command("sleep", "1").Start() == nil {
		cmd = "sleep"
		args = []string{"2"}
	} else {
		t.Skip("No suitable command available for server test")
	}

	serverInfo, err := sm.StartServer(ctx, cfg, "test-server", "test", cmd, args, "", nil, 100)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	if serverInfo.Status != "running" {
		t.Errorf("Expected server status 'running', got %q", serverInfo.Status)
	}

	// List servers
	servers := sm.ListServers()
	if len(servers) == 0 {
		t.Error("Expected at least one server in list")
	}

	// Get server logs
	logs, err := sm.GetServerLogs("test-server", 10)
	if err != nil {
		t.Errorf("Failed to get server logs: %v", err)
	}
	_ = logs // Logs might be empty, which is okay

	// Stop server
	time.Sleep(500 * time.Millisecond) // Give it time to run
	err = sm.StopServer("test-server", false)
	if err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}
}

func TestEndToEnd_ResourceDiscovery(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		DisableNotifications: true,
		WorkingDirectory:     dir,
	}

	// Create go.mod
	modFile := filepath.Join(dir, "go.mod")
	modContent := "module test\n\ngo 1.24\n\nrequire github.com/example/dep v1.0.0\n"
	if err := os.WriteFile(modFile, []byte(modContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a test file with build tags
	goFile := filepath.Join(dir, "main.go")
	goContent := `// +build linux darwin

package main
`
	if err := os.WriteFile(goFile, []byte(goContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(dir, "main_test.go")
	testContent := `package main

import "testing"

func TestMain(t *testing.T) {
	// test
}
`
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test that resources can be registered
	// The actual resource functions are internal, so we test registration
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "test",
			Version: "1.0.0",
		},
		nil,
	)
	resources.RegisterGoResources(server, cfg)
	// If registration succeeds, resources work
}

func TestEndToEnd_FullServerSetup(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("Go command not available")
	}

	cfg := &config.Config{
		DisableNotifications: true,
		WorkingDirectory:     t.TempDir(),
	}

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "test-server",
			Version: "1.0.0",
		},
		nil,
	)

	// Initialize subsystems
	tools.InitServerManager()
	tools.InitPackageDocsCache()

	// Register all tools
	tools.RegisterRunTools(server, cfg)
	tools.RegisterGoTools(server, cfg)
	tools.RegisterOptimizationTools(server, cfg)
	tools.RegisterServerTools(server, cfg)
	tools.RegisterPackageDocsTools(server, cfg)

	// Register resources
	resources.RegisterGoResources(server, cfg)

	// If we get here without panicking, the full setup works
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
