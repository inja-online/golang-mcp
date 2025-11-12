# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-11-12

### Added

#### Core Features
- Initial release of MCP Go Server
- Full Model Context Protocol (MCP) implementation
- JSON-RPC 2.0 communication protocol support
- Comprehensive Go development toolset with 22 tools

#### Code Execution Tools
- `go_run` - Execute Go files directly with support for arguments and environment variables

#### Go Operations Tools (7 tools)
- `go_build` - Build Go packages with race detector, build tags, and linker flags
- `go_test` - Run tests with coverage, benchmarks, and race detection
- `go_fmt` - Format Go code using go fmt
- `go_mod` - Manage Go modules (init, tidy, download, vendor, get)
- `go_doc` - Generate documentation for Go packages
- `go_lint` - Lint Go code using golangci-lint or go vet
- `go_cross_compile` - Cross-compile for different platforms and architectures

#### Optimization Tools (6 tools)
- `go_profile` - Generate CPU and memory performance profiles using pprof
- `go_trace` - Generate execution traces for Go programs
- `go_benchmark` - Run benchmarks and analyze results
- `go_race_detect` - Detect race conditions in Go code
- `go_memory_profile` - Generate memory profiles for analysis
- `go_optimize_suggest` - Analyze code and provide optimization suggestions

#### Server Management Tools (5 tools)
- `go_server_start` - Start long-running Go servers in the background
- `go_server_stop` - Stop running servers (graceful or force kill)
- `go_server_list` - List all running servers
- `go_server_logs` - Get logs from running servers
- `go_server_status` - Get detailed status of servers

#### Package Documentation Tools (3 tools)
- `go_pkg_docs` - Fetch package documentation from go.dev (pkg.go.dev)
- `go_pkg_search` - Search for packages on go.dev
- `go_pkg_examples` - Extract examples from package documentation

#### LSP Tools (5 tools, optional)
- `lsp_start_session` - Start LSP session for workspace
- `lsp_shutdown_session` - Shutdown LSP session
- `lsp_request` - Send LSP request and wait for result
- `lsp_notify` - Send LSP notification (non-blocking)
- `lsp_subscribe_diagnostics` - Subscribe to diagnostics from LSP server

#### Resources (8 total)
- `go://modules` - List Go modules and dependencies
- `go://build-tags` - Discover build tags and constraints
- `go://tests` - List test files and benchmarks
- `go://workspace` - Go workspace structure and configuration
- `go://pkg-docs/{path}` - Fetch package documentation (supports versioned paths)
- `go://tools` - List all available tools with schemas
- `go://prompts` - List all available prompts with arguments
- `go://resources` - List all available resources

#### Prompts (7 total)
- `setup-go-project` - Guide for setting up new Go projects
- `write-go-tests` - Template for writing comprehensive Go tests
- `optimize-go-performance` - Guide for profiling and optimizing Go code
- `debug-go-issue` - Systematic approach to debugging Go programs
- `add-go-dependency` - Guide for adding and managing Go dependencies
- `go-code-review` - Checklist for reviewing Go code
- `go-server-deployment` - Guide for building and deploying Go servers

#### Installation Methods
- Go Install support via `go install`
- Build from source with Makefile
- Pre-built binaries for Linux, macOS, and Windows
- Docker container support with GitHub Container Registry
- Docker Compose configuration

#### Platform Support
- Comprehensive guides for 36 platforms including:
  - Claude Desktop
  - Cursor
  - VS Code
  - Cline (VS Code extension)
  - Windsurf
  - Zed
  - JetBrains AI Assistant (all IDEs)
  - Perplexity Desktop
  - Amazon Q Developer CLI
  - Smithery
  - Claude Code CLI
  - Amp
  - Warp Terminal
  - Copilot Coding Agent
  - Copilot CLI
  - LM Studio
  - Visual Studio 2022
  - Roo Code
  - Gemini CLI
  - Qwen Coder
  - Opencode
  - OpenAI Codex
  - Kiro
  - Trae
  - Bun runtime
  - Deno runtime
  - Docker
  - Desktop Extension (.mcpb)
  - Windows (detailed guide)
  - BoltAI
  - Rovo Dev CLI
  - Zencoder
  - Qodo Gen
  - Factory
  - Crush
  - Augment Code

#### Configuration
- Environment variable support for all settings
- `DISABLE_NOTIFICATIONS` - Disable permission prompts
- `DEBUG_MCP` - Enable debug logging
- `ENABLE_LSP` - Enable LSP tools (requires gopls)
- `GOROOT` - Custom Go root directory
- `GOPATH` - Custom Go workspace path
- `GOOS/GOARCH` - Target platform for cross-compilation
- `GOPROXY` - Go module proxy URL

#### Documentation
- Comprehensive README with accessibility features
- Quick Start guide with platform jump links
- Installation methods with verification steps
- Platform-specific guides with step-by-step instructions
- Troubleshooting section with symptoms → solutions format
- FAQ section with common questions
- Quick reference tables for platform compatibility
- Usage examples and advanced patterns
- Configuration guide with examples

#### Developer Experience
- Visual indicators (emojis) for quick scanning
- Structured information with tables and bullet points
- Reduced cognitive load with short paragraphs
- Navigation helpers with jump links
- Success indicators for completed steps
- Platform-specific troubleshooting guides

### Security
- Command validation to prevent injection attacks
- Permission system with user prompts (can be disabled)
- Commands run with same permissions as MCP server process
- Support for system notifications or console prompts

### Technical Details
- Built with Go 1.24+
- Uses Model Context Protocol SDK
- Supports stdio (standard input/output) communication
- Cross-platform support (Linux, macOS, Windows)
- Minimal dependencies
- Docker image based on Alpine Linux

[1.0.0]: https://github.com/inja-online/golang-mcp/releases/tag/v1.0.0

