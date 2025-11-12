package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save original env vars
	origDisable := os.Getenv("DISABLE_NOTIFICATIONS")
	origGoRoot := os.Getenv("GOROOT")
	origGoPath := os.Getenv("GOPATH")
	origGoOS := os.Getenv("GOOS")
	origGoArch := os.Getenv("GOARCH")
	origGoProxy := os.Getenv("GOPROXY")

	// Clean up after test
	defer func() {
		if origDisable != "" {
			os.Setenv("DISABLE_NOTIFICATIONS", origDisable)
		} else {
			os.Unsetenv("DISABLE_NOTIFICATIONS")
		}
		if origGoRoot != "" {
			os.Setenv("GOROOT", origGoRoot)
		} else {
			os.Unsetenv("GOROOT")
		}
		if origGoPath != "" {
			os.Setenv("GOPATH", origGoPath)
		} else {
			os.Unsetenv("GOPATH")
		}
		if origGoOS != "" {
			os.Setenv("GOOS", origGoOS)
		} else {
			os.Unsetenv("GOOS")
		}
		if origGoArch != "" {
			os.Setenv("GOARCH", origGoArch)
		} else {
			os.Unsetenv("GOARCH")
		}
		if origGoProxy != "" {
			os.Setenv("GOPROXY", origGoProxy)
		} else {
			os.Unsetenv("GOPROXY")
		}
	}()

	t.Run("load with DISABLE_NOTIFICATIONS=true", func(t *testing.T) {
		os.Setenv("DISABLE_NOTIFICATIONS", "true")
		cfg := Load()
		if !cfg.DisableNotifications {
			t.Error("Expected DisableNotifications to be true")
		}
	})

	t.Run("load with DISABLE_NOTIFICATIONS=false", func(t *testing.T) {
		os.Setenv("DISABLE_NOTIFICATIONS", "false")
		cfg := Load()
		if cfg.DisableNotifications {
			t.Error("Expected DisableNotifications to be false")
		}
	})

	t.Run("load with GOROOT set", func(t *testing.T) {
		os.Setenv("GOROOT", "/test/goroot")
		cfg := Load()
		if cfg.GoRoot != "/test/goroot" {
			t.Errorf("Expected GOROOT to be /test/goroot, got %q", cfg.GoRoot)
		}
	})

	t.Run("load with GOPATH set", func(t *testing.T) {
		os.Setenv("GOPATH", "/test/gopath")
		cfg := Load()
		if cfg.GoPath != "/test/gopath" {
			t.Errorf("Expected GOPATH to be /test/gopath, got %q", cfg.GoPath)
		}
	})

	t.Run("load with GOOS set", func(t *testing.T) {
		os.Setenv("GOOS", "linux")
		cfg := Load()
		if cfg.GoOS != "linux" {
			t.Errorf("Expected GOOS to be linux, got %q", cfg.GoOS)
		}
	})

	t.Run("load with GOARCH set", func(t *testing.T) {
		os.Setenv("GOARCH", "amd64")
		cfg := Load()
		if cfg.GoArch != "amd64" {
			t.Errorf("Expected GOARCH to be amd64, got %q", cfg.GoArch)
		}
	})

	t.Run("load with GOPROXY set", func(t *testing.T) {
		os.Setenv("GOPROXY", "https://proxy.golang.org")
		cfg := Load()
		if cfg.GoProxy != "https://proxy.golang.org" {
			t.Errorf("Expected GOPROXY to be https://proxy.golang.org, got %q", cfg.GoProxy)
		}
	})
}

func TestGetGoEnv(t *testing.T) {
	cfg := &Config{
		GoRoot:  "/test/goroot",
		GoPath:  "/test/gopath",
		GoOS:    "linux",
		GoArch:  "amd64",
		GoProxy: "https://proxy.golang.org",
	}

	env := cfg.GetGoEnv()

	if env["GOROOT"] != "/test/goroot" {
		t.Errorf("Expected GOROOT to be /test/goroot, got %q", env["GOROOT"])
	}
	if env["GOPATH"] != "/test/gopath" {
		t.Errorf("Expected GOPATH to be /test/gopath, got %q", env["GOPATH"])
	}
	if env["GOOS"] != "linux" {
		t.Errorf("Expected GOOS to be linux, got %q", env["GOOS"])
	}
	if env["GOARCH"] != "amd64" {
		t.Errorf("Expected GOARCH to be amd64, got %q", env["GOARCH"])
	}
	if env["GOPROXY"] != "https://proxy.golang.org" {
		t.Errorf("Expected GOPROXY to be https://proxy.golang.org, got %q", env["GOPROXY"])
	}
}

func TestGetGoEnv_EmptyValues(t *testing.T) {
	cfg := &Config{
		GoRoot:  "",
		GoPath:  "",
		GoOS:    "",
		GoArch:  "",
		GoProxy: "",
	}

	env := cfg.GetGoEnv()

	// Empty values should not be in the map
	if _, exists := env["GOROOT"]; exists {
		t.Error("GOROOT should not be in env map when empty")
	}
	if _, exists := env["GOPATH"]; exists {
		t.Error("GOPATH should not be in env map when empty")
	}
}

func TestDetectGoRoot(t *testing.T) {
	// This test depends on having Go installed
	// We'll just verify it doesn't panic and returns something reasonable
	goroot := detectGoRoot()
	// If Go is installed, goroot should be non-empty
	// If not, it will be empty, which is fine
	_ = goroot
}
