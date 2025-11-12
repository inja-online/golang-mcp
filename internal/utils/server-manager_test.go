package utils

import (
	"context"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/inja-online/golang-mcp/internal/config"
)

func TestRingBuffer(t *testing.T) {
	t.Run("add and get all", func(t *testing.T) {
		rb := NewRingBuffer(5)
		rb.Add("line1")
		rb.Add("line2")
		rb.Add("line3")

		all := rb.GetAll()
		if len(all) != 3 {
			t.Errorf("Expected 3 lines, got %d", len(all))
		}
		if all[0] != "line1" || all[1] != "line2" || all[2] != "line3" {
			t.Errorf("Unexpected content: %v", all)
		}
	})

	t.Run("ring buffer overflow", func(t *testing.T) {
		rb := NewRingBuffer(3)
		rb.Add("line1")
		rb.Add("line2")
		rb.Add("line3")
		rb.Add("line4") // Should overwrite line1

		all := rb.GetAll()
		if len(all) != 3 {
			t.Errorf("Expected 3 lines, got %d", len(all))
		}
		if all[0] != "line2" {
			t.Errorf("Expected line2 to be first after overflow, got %q", all[0])
		}
	})

	t.Run("get recent", func(t *testing.T) {
		rb := NewRingBuffer(10)
		for i := 1; i <= 10; i++ {
			rb.Add(fmt.Sprintf("line%d", i))
		}

		recent := rb.GetRecent(3)
		if len(recent) != 3 {
			t.Errorf("Expected 3 recent lines, got %d", len(recent))
		}
		if recent[0] != "line8" || recent[1] != "line9" || recent[2] != "line10" {
			t.Errorf("Unexpected recent lines: %v", recent)
		}
	})

	t.Run("get recent more than available", func(t *testing.T) {
		rb := NewRingBuffer(10)
		rb.Add("line1")
		rb.Add("line2")

		recent := rb.GetRecent(5)
		if len(recent) != 2 {
			t.Errorf("Expected 2 lines, got %d", len(recent))
		}
	})
}

func TestServerManager_StartServer(t *testing.T) {
	cfg := &config.Config{
		DisableNotifications: true,
		WorkingDirectory:     t.TempDir(),
	}

	sm := NewServerManager()
	ctx := context.Background()

	// Test starting a simple command (echo)
	// Note: This test may need adjustment based on OS
	serverInfo, err := sm.StartServer(ctx, cfg, "test-1", "test-server", "echo", []string{"hello"}, "", nil, 100)
	if err != nil {
		// On some systems, echo might not be available, so we'll skip
		t.Skipf("Could not start test server: %v", err)
	}

	if serverInfo == nil {
		t.Fatal("Expected server info but got nil")
	}

	if serverInfo.ID != "test-1" {
		t.Errorf("Expected ID 'test-1', got %q", serverInfo.ID)
	}

	serverInfo.logMutex.RLock()
	status := serverInfo.Status
	serverInfo.logMutex.RUnlock()

	if status != "running" {
		t.Errorf("Expected status 'running', got %q", status)
	}

	// Clean up
	time.Sleep(100 * time.Millisecond) // Give it time to start
	_ = sm.StopServer("test-1", true)
}

func TestServerManager_ListServers(t *testing.T) {
	cfg := &config.Config{
		DisableNotifications: true,
		WorkingDirectory:     t.TempDir(),
	}

	sm := NewServerManager()
	ctx := context.Background()

	// Start a few servers
	_, err1 := sm.StartServer(ctx, cfg, "server-1", "test-1", "echo", []string{"test1"}, "", nil, 100)
	_, err2 := sm.StartServer(ctx, cfg, "server-2", "test-2", "echo", []string{"test2"}, "", nil, 100)

	if err1 != nil || err2 != nil {
		t.Skip("Could not start test servers")
	}

	time.Sleep(100 * time.Millisecond)

	servers := sm.ListServers()
	if len(servers) < 2 {
		t.Errorf("Expected at least 2 servers, got %d", len(servers))
	}

	// Clean up
	_ = sm.StopServer("server-1", true)
	_ = sm.StopServer("server-2", true)
}

func TestServerManager_GetServer(t *testing.T) {
	cfg := &config.Config{
		DisableNotifications: true,
		WorkingDirectory:     t.TempDir(),
	}

	sm := NewServerManager()
	ctx := context.Background()

	_, err := sm.StartServer(ctx, cfg, "test-get", "test", "echo", []string{"test"}, "", nil, 100)
	if err != nil {
		t.Skip("Could not start test server")
	}

	time.Sleep(100 * time.Millisecond)

	server, err := sm.GetServer("test-get")
	if err != nil {
		t.Fatalf("Expected to get server, got error: %v", err)
	}

	if server.ID != "test-get" {
		t.Errorf("Expected ID 'test-get', got %q", server.ID)
	}

	// Test getting non-existent server
	_, err = sm.GetServer("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent server")
	}

	// Clean up
	_ = sm.StopServer("test-get", true)
}

func TestServerManager_StopServer(t *testing.T) {
	cfg := &config.Config{
		DisableNotifications: true,
		WorkingDirectory:     t.TempDir(),
	}

	sm := NewServerManager()
	ctx := context.Background()

	// Start a long-running process (sleep)
	// Use a command that will run for a bit
	var cmd string
	var args []string
	if exec.Command("sleep", "1").Start() == nil {
		cmd = "sleep"
		args = []string{"5"}
	} else if exec.Command("timeout", "5").Start() == nil {
		cmd = "timeout"
		args = []string{"5"}
	} else {
		t.Skip("No suitable command available for testing stop")
	}

	_, err := sm.StartServer(ctx, cfg, "test-stop", "test", cmd, args, "", nil, 100)
	if err != nil {
		t.Skipf("Could not start test server: %v", err)
	}

	time.Sleep(200 * time.Millisecond) // Give it time to start

	// Stop gracefully
	err = sm.StopServer("test-stop", false)
	if err != nil {
		t.Errorf("Failed to stop server gracefully: %v", err)
	}

	// Test stopping non-existent server
	err = sm.StopServer("nonexistent", false)
	if err == nil {
		t.Error("Expected error when stopping non-existent server")
	}
}

func TestServerManager_GetServerLogs(t *testing.T) {
	cfg := &config.Config{
		DisableNotifications: true,
		WorkingDirectory:     t.TempDir(),
	}

	sm := NewServerManager()
	ctx := context.Background()

	_, err := sm.StartServer(ctx, cfg, "test-logs", "test", "echo", []string{"test", "output"}, "", nil, 100)
	if err != nil {
		t.Skip("Could not start test server")
	}

	time.Sleep(200 * time.Millisecond) // Give it time to produce logs

	logs, err := sm.GetServerLogs("test-logs", 0)
	if err != nil {
		t.Fatalf("Failed to get server logs: %v", err)
	}

	// Logs might be empty depending on timing, but should not error
	_ = logs

	// Test getting recent logs
	recentLogs, err := sm.GetServerLogs("test-logs", 5)
	if err != nil {
		t.Fatalf("Failed to get recent logs: %v", err)
	}

	if len(recentLogs) > 5 {
		t.Errorf("Expected at most 5 recent logs, got %d", len(recentLogs))
	}

	// Clean up
	_ = sm.StopServer("test-logs", true)
}
