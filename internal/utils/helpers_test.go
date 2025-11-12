package utils

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/inja-online/golang-mcp/internal/config"
)

// testContext returns a context with timeout for testing
func testContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// createTestConfig creates a test configuration
func createTestConfig(t *testing.T, disableNotifications bool) *config.Config {
	cfg := &config.Config{
		DisableNotifications: disableNotifications,
		WorkingDirectory:     t.TempDir(),
	}
	return cfg
}

func TestExecuteGoCommand(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		args        []string
		expectError bool
		skipIfNoGo  bool
	}{
		{
			name:        "valid go version command",
			command:     "go",
			args:        []string{"version"},
			expectError: false,
			skipIfNoGo:  true,
		},
		{
			name:        "invalid command",
			command:     "nonexistent-command-xyz",
			args:        []string{},
			expectError: true,
			skipIfNoGo:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipIfNoGo {
				// Check if go is available
				if _, err := exec.LookPath("go"); err != nil {
					t.Skip("Go command not available")
				}
			}

			cfg := createTestConfig(t, true)
			ctx := testContext(t)

			result, err := ExecuteGoCommand(ctx, cfg, tt.command, tt.args, "", nil)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("Expected result but got nil")
				}
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		args        []string
		expectError bool
	}{
		{
			name:        "valid go command",
			command:     "go",
			args:        []string{"version"},
			expectError: false,
		},
		{
			name:        "command with injection attempt - semicolon",
			command:     "go",
			args:        []string{"version", ";", "rm", "-rf", "/"},
			expectError: true,
		},
		{
			name:        "command with injection attempt - &&",
			command:     "go",
			args:        []string{"version", "&&", "rm", "-rf", "/"},
			expectError: true,
		},
		{
			name:        "command with injection attempt - ||",
			command:     "go",
			args:        []string{"version", "||", "rm", "-rf", "/"},
			expectError: true,
		},
		{
			name:        "command with redirection - >",
			command:     "go",
			args:        []string{"version", ">", "file.txt"},
			expectError: true,
		},
		{
			name:        "command with redirection - <",
			command:     "go",
			args:        []string{"version", "<", "file.txt"},
			expectError: true,
		},
		{
			name:        "command with pipe",
			command:     "go",
			args:        []string{"version", "|", "grep"},
			expectError: true,
		},
		{
			name:        "valid args without injection",
			command:     "go",
			args:        []string{"build", "-o", "output", "main.go"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommand(tt.command, tt.args)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestRequestPermission(t *testing.T) {
	// This test is tricky because RequestPermission reads from stdin
	// We'll test that it respects DISABLE_NOTIFICATIONS
	// For actual permission requests, we'd need to mock stdin

	t.Run("permission request with notifications disabled", func(t *testing.T) {
		// When notifications are disabled, ExecuteGoCommand should skip permission
		cfg := createTestConfig(t, true)
		ctx := testContext(t)

		// This should not block or error when notifications are disabled
		// We can't easily test the interactive permission request without mocking stdin
		// So we test that ExecuteGoCommand works with notifications disabled
		if _, err := exec.LookPath("go"); err != nil {
			t.Skip("Go command not available")
		}

		_, err := ExecuteGoCommand(ctx, cfg, "go", []string{"version"}, "", nil)
		if err != nil {
			t.Errorf("Unexpected error when notifications disabled: %v", err)
		}
	})
}

func TestGetGoVersion(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("Go command not available")
	}

	cfg := createTestConfig(t, true)
	ctx := testContext(t)

	version, err := GetGoVersion(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to get Go version: %v", err)
	}

	if version == "" {
		t.Error("Expected version string but got empty")
	}

	t.Logf("Detected Go version: %s", version)
}

func TestFindGoVersion(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("Go command not available")
	}

	cfg := createTestConfig(t, true)
	ctx := testContext(t)

	version, err := FindGoVersion(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to find Go version: %v", err)
	}

	if version == "" {
		t.Error("Expected version string but got empty")
	}

	t.Logf("Found Go version: %s", version)
}

func TestSetupSignalHandling(t *testing.T) {
	// Test that signal handling can be set up without panicking
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// This should not panic
	SetupSignalHandling(cancel)

	// Give it a moment to set up
	time.Sleep(100 * time.Millisecond)
}
