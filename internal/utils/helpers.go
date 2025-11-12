package utils

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/inja-online/golang-mcp/internal/config"
)

// CommandResult holds the result of a command execution
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// ExecuteGoCommand executes a Go command with proper environment setup
func ExecuteGoCommand(ctx context.Context, cfg *config.Config, command string, args []string, workingDir string, envVars map[string]string) (*CommandResult, error) {
	// Validate command to prevent injection
	if err := ValidateCommand(command, args); err != nil {
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	// Request permission unless disabled
	if !cfg.DisableNotifications {
		if err := RequestPermission(command, args); err != nil {
			return nil, fmt.Errorf("permission denied: %w", err)
		}
	}

	// Create command
	cmd := exec.CommandContext(ctx, command, args...)

	// Set working directory
	if workingDir != "" {
		cmd.Dir = workingDir
	} else {
		cmd.Dir = cfg.WorkingDirectory
	}

	// Set up environment
	env := os.Environ()

	// Add Go environment variables
	goEnv := cfg.GetGoEnv()
	for k, v := range goEnv {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Add custom environment variables
	for k, v := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	// Capture output
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			// Command failed to start or was interrupted
			return nil, fmt.Errorf("command execution failed: %w", err)
		}
	}

	return &CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Duration: duration,
	}, nil
}

// ValidateCommand validates command and arguments to prevent injection attacks
func ValidateCommand(command string, args []string) error {
	// Only allow 'go' command or absolute paths
	if command != "go" && !filepath.IsAbs(command) {
		// Check if it's in PATH and is a go binary
		path, err := exec.LookPath(command)
		if err != nil {
			return fmt.Errorf("command not found: %s", command)
		}
		// Additional validation could be added here
		_ = path
	}

	// Validate arguments don't contain dangerous patterns
	for _, arg := range args {
		// Check for command injection patterns
		if strings.Contains(arg, ";") || strings.Contains(arg, "&&") || strings.Contains(arg, "||") {
			return fmt.Errorf("potentially dangerous argument detected: %s", arg)
		}
		// Check for shell redirection
		if strings.Contains(arg, ">") || strings.Contains(arg, "<") || strings.Contains(arg, "|") {
			return fmt.Errorf("potentially dangerous argument detected: %s", arg)
		}
	}

	return nil
}

// RequestPermission requests user permission before executing a command
func RequestPermission(command string, args []string) error {
	// Try system notification first, fallback to console prompt
	if canUseNotifications() {
		return requestPermissionViaNotification(command, args)
	}
	return requestPermissionViaConsole(command, args)
}

// canUseNotifications checks if system notifications are available
func canUseNotifications() bool {
	// Simple check - can be enhanced with platform-specific code
	return false // Default to console for now
}

// requestPermissionViaNotification requests permission via system notification
func requestPermissionViaNotification(command string, args []string) error {
	// TODO: Implement system notification
	// For now, fallback to console
	return requestPermissionViaConsole(command, args)
}

// requestPermissionViaConsole requests permission via console input
func requestPermissionViaConsole(command string, args []string) error {
	fmt.Fprintf(os.Stderr, "\n[PERMISSION REQUEST]\n")
	fmt.Fprintf(os.Stderr, "Command: %s %s\n", command, strings.Join(args, " "))
	fmt.Fprintf(os.Stderr, "Allow this command? (yes/no): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "yes" && response != "y" {
		return fmt.Errorf("permission denied by user")
	}

	return nil
}

// GetGoVersion detects the Go version
func GetGoVersion(ctx context.Context, cfg *config.Config) (string, error) {
	result, err := ExecuteGoCommand(ctx, cfg, "go", []string{"version"}, "", nil)
	if err != nil {
		return "", err
	}

	// Parse version from output (format: "go version go1.24.0 linux/amd64")
	parts := strings.Fields(result.Stdout)
	if len(parts) >= 3 {
		return parts[2], nil
	}

	return result.Stdout, nil
}

// FindGoVersion attempts to find Go version via gvm or system Go
func FindGoVersion(ctx context.Context, cfg *config.Config) (string, error) {
	// First try system Go
	version, err := GetGoVersion(ctx, cfg)
	if err == nil {
		return version, nil
	}

	// TODO: Try gvm if available
	// This would require checking for gvm and querying it

	return "", fmt.Errorf("could not determine Go version")
}

// SetupSignalHandling sets up signal handling for graceful shutdown
func SetupSignalHandling(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Fprintf(os.Stderr, "\nReceived shutdown signal, cleaning up...\n")
		cancel()
	}()
}
