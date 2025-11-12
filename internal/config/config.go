package config

import (
	"os"
	"os/exec"
	"path/filepath"
)

// Config holds the server configuration
type Config struct {
	DisableNotifications bool
	DebugMCP             bool
	EnableLSP            bool
	GoRoot               string
	GoPath               string
	GoOS                 string
	GoArch               string
	GoProxy              string
	WorkingDirectory     string
}

// Load loads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		DisableNotifications: os.Getenv("DISABLE_NOTIFICATIONS") == "true",
		DebugMCP:             os.Getenv("DEBUG_MCP") == "true",
		EnableLSP:            os.Getenv("ENABLE_LSP") == "true",
		GoRoot:               os.Getenv("GOROOT"),
		GoPath:               os.Getenv("GOPATH"),
		GoOS:                 getEnvOrDefault("GOOS", ""),
		GoArch:               getEnvOrDefault("GOARCH", ""),
		GoProxy:              os.Getenv("GOPROXY"),
	}

	// Get working directory
	wd, err := os.Getwd()
	if err == nil {
		cfg.WorkingDirectory = wd
	} else {
		cfg.WorkingDirectory = "."
	}

	// Detect Go environment if not set
	if cfg.GoRoot == "" {
		cfg.GoRoot = detectGoRoot()
	}

	return cfg
}

// GetGoEnv returns a map of Go environment variables
func (c *Config) GetGoEnv() map[string]string {
	env := make(map[string]string)

	if c.GoRoot != "" {
		env["GOROOT"] = c.GoRoot
	}
	if c.GoPath != "" {
		env["GOPATH"] = c.GoPath
	}
	if c.GoOS != "" {
		env["GOOS"] = c.GoOS
	}
	if c.GoArch != "" {
		env["GOARCH"] = c.GoArch
	}
	if c.GoProxy != "" {
		env["GOPROXY"] = c.GoProxy
	}

	return env
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// detectGoRoot attempts to detect GOROOT by finding the go binary
func detectGoRoot() string {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return ""
	}

	// GOROOT is typically the parent of the bin directory
	binDir := filepath.Dir(goBin)
	goroot := filepath.Dir(binDir)

	// Verify it's a valid GOROOT by checking for src directory
	srcPath := filepath.Join(goroot, "src")
	if _, err := os.Stat(srcPath); err == nil {
		return goroot
	}

	return ""
}
