package prompts

import (
	"context"

	"github.com/inja-online/golang-mcp/internal/config"
	"github.com/inja-online/golang-mcp/internal/resources"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterGoPrompts registers Go-related prompts with the MCP server.
func RegisterGoPrompts(server *mcp.Server, cfg *config.Config) int {
	count := 0
	setupArgs := []*mcp.PromptArgument{
		{
			Name:        "project_name",
			Description: "Name of the Go project to set up",
			Required:    true,
		},
		{
			Name:        "module_path",
			Description: "Go module path (e.g., github.com/username/project)",
			Required:    false,
		},
	}
	resources.RegisterPrompt("setup-go-project", "Guide for setting up a new Go project with module initialization, directory structure, and best practices", setupArgs)
	server.AddPrompt(&mcp.Prompt{
		Name:        "setup-go-project",
		Description: "Guide for setting up a new Go project with module initialization, directory structure, and best practices",
		Arguments:   setupArgs,
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		projectName := req.Params.Arguments["project_name"]
		if projectName == "" {
			projectName = "my-project"
		}
		modulePath := req.Params.Arguments["module_path"]
		if modulePath == "" {
			modulePath = projectName
		}

		prompt := `You are setting up a new Go project: ` + projectName + `

Steps to follow:
1. Initialize Go module: go mod init ` + modulePath + `
2. Create standard directory structure:
   - cmd/ for main applications
   - internal/ for private application code
   - pkg/ for public library code (optional)
   - api/ for API definitions (optional)
   - web/ for web assets (optional)
3. Create a README.md with project description
4. Add .gitignore for Go projects
5. Set up basic main.go in cmd/` + projectName + `/main.go
6. Run go mod tidy to ensure dependencies are clean

Use the go_mod tool with operation "init" to initialize the module.
Use go_fmt to format code after creation.
Use go_build to verify the project builds correctly.`

		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: prompt},
				},
			},
		}, nil
	})
	count++

	writeTestsArgs := []*mcp.PromptArgument{
		{
			Name:        "package_path",
			Description: "Package path to write tests for",
			Required:    true,
		},
		{
			Name:        "test_type",
			Description: "Type of tests to write: unit, benchmark, integration, or all",
			Required:    false,
		},
	}
	resources.RegisterPrompt("write-go-tests", "Template for writing comprehensive Go tests including unit tests, benchmarks, and table-driven tests", writeTestsArgs)
	server.AddPrompt(&mcp.Prompt{
		Name:        "write-go-tests",
		Description: "Template for writing comprehensive Go tests including unit tests, benchmarks, and table-driven tests",
		Arguments:   writeTestsArgs,
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		packagePath := req.Params.Arguments["package_path"]
		if packagePath == "" {
			packagePath = "./..."
		}
		testType := req.Params.Arguments["test_type"]
		if testType == "" {
			testType = "all"
		}

		prompt := `Write comprehensive tests for Go package: ` + packagePath + `

Test requirements:
1. Create ` + packagePath + `_test.go file
2. Use table-driven tests for multiple test cases
3. Test both success and error cases
4. Use subtests with t.Run() for better organization
5. Include test helpers for common setup/teardown
6. Use testify/assert or standard testing package

` + func() string {
			switch testType {
			case "unit":
				return `Focus on unit tests:
- Test individual functions in isolation
- Mock external dependencies
- Test edge cases and error conditions`
			case "benchmark":
				return `Focus on benchmarks:
- Use Benchmark* functions
- Test performance of critical paths
- Compare different implementations`
			case "integration":
				return `Focus on integration tests:
- Test component interactions
- Use real dependencies where possible
- Test end-to-end workflows`
			default:
				return `Include all test types:
- Unit tests for individual functions
- Benchmarks for performance-critical code
- Integration tests for component interactions`
			}
		}() + `

Use go_test tool to run tests with coverage.
Use go_benchmark tool to run and analyze benchmarks.
Use go_race_detect tool to check for race conditions.`

		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: prompt},
				},
			},
		}, nil
	})
	count++

	optimizeArgs := []*mcp.PromptArgument{
		{
			Name:        "package_path",
			Description: "Package to optimize",
			Required:    true,
		},
		{
			Name:        "optimization_goal",
			Description: "Optimization goal: cpu, memory, concurrency, or general",
			Required:    false,
		},
	}
	resources.RegisterPrompt("optimize-go-performance", "Guide for profiling and optimizing Go code performance using pprof, benchmarks, and race detection", optimizeArgs)
	server.AddPrompt(&mcp.Prompt{
		Name:        "optimize-go-performance",
		Description: "Guide for profiling and optimizing Go code performance using pprof, benchmarks, and race detection",
		Arguments:   optimizeArgs,
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		packagePath := req.Params.Arguments["package_path"]
		if packagePath == "" {
			packagePath = "./..."
		}
		goal := req.Params.Arguments["optimization_goal"]
		if goal == "" {
			goal = "general"
		}

		prompt := `Optimize Go package performance: ` + packagePath + `

Optimization workflow:
1. Establish baseline with go_benchmark tool
2. Profile the code:
   - Use go_profile tool with type "cpu" for CPU profiling
   - Use go_memory_profile tool for memory profiling
   - Use go_trace tool for execution tracing
3. Analyze profiles using go tool pprof
4. Identify bottlenecks and optimize
5. Re-run benchmarks to measure improvements
6. Use go_race_detect to ensure no race conditions

` + func() string {
			switch goal {
			case "cpu":
				return `CPU optimization focus:
- Identify hot paths and CPU-intensive operations
- Optimize algorithms and data structures
- Reduce allocations in hot paths
- Consider caching and memoization`
			case "memory":
				return `Memory optimization focus:
- Profile memory allocations with go_memory_profile
- Use sync.Pool for frequently allocated objects
- Reduce garbage collection pressure
- Optimize data structure sizes`
			case "concurrency":
				return `Concurrency optimization focus:
- Use go_race_detect to verify thread safety
- Profile goroutine usage and channel operations
- Optimize lock contention
- Consider worker pools for CPU-bound tasks`
			default:
				return `General optimization:
- Profile both CPU and memory
- Optimize based on actual bottlenecks
- Measure before and after changes
- Consider trade-offs between performance and maintainability`
			}
		}() + `

Use go_optimize_suggest tool for automated optimization suggestions.`

		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: prompt},
				},
			},
		}, nil
	})
	count++

	debugArgs := []*mcp.PromptArgument{
		{
			Name:        "issue_type",
			Description: "Type of issue: panic, race, performance, deadlock, or unknown",
			Required:    false,
		},
		{
			Name:        "package_path",
			Description: "Package where the issue occurs",
			Required:    false,
		},
	}
	resources.RegisterPrompt("debug-go-issue", "Systematic approach to debugging Go programs including race conditions, panics, and performance issues", debugArgs)
	server.AddPrompt(&mcp.Prompt{
		Name:        "debug-go-issue",
		Description: "Systematic approach to debugging Go programs including race conditions, panics, and performance issues",
		Arguments:   debugArgs,
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		issueType := req.Params.Arguments["issue_type"]
		packagePath := req.Params.Arguments["package_path"]
		if issueType == "" {
			issueType = "unknown"
		}

		prompt := `Debug Go issue` + func() string {
			if packagePath != "" {
				return " in package: " + packagePath
			}
			return ""
		}() + `

Debugging steps:
1. Reproduce the issue consistently
2. Add logging to understand execution flow
3. Run go_lint to check for static analysis issues
4. Use go_test with verbose flag to see detailed test output

` + func() string {
			switch issueType {
			case "panic":
				return `Panic debugging:
- Check stack trace for panic location
- Use defer/recover to handle panics gracefully
- Check for nil pointer dereferences
- Verify slice/array bounds
- Use go_test to run tests and catch panics early`
			case "race":
				return `Race condition debugging:
- Use go_race_detect tool to identify race conditions
- Review concurrent access to shared state
- Use proper synchronization (mutexes, channels)
- Check for data races in maps and slices
- Verify goroutine lifecycle management`
			case "performance":
				return `Performance issue debugging:
- Use go_profile to identify CPU bottlenecks
- Use go_memory_profile to find memory issues
- Run go_benchmark to measure performance
- Check for unnecessary allocations
- Profile with realistic workloads`
			case "deadlock":
				return `Deadlock debugging:
- Use go_trace tool to visualize goroutine interactions
- Check for circular channel dependencies
- Verify mutex lock ordering
- Look for goroutines waiting indefinitely
- Use timeout contexts to prevent hangs`
			default:
				return `General debugging:
- Run go_lint for static analysis
- Use go_test with -v flag for verbose output
- Check logs and error messages
- Use go_race_detect for concurrency issues
- Profile with go_profile if performance-related`
			}
		}() + `

After fixing, verify with:
- go_test to ensure tests pass
- go_race_detect to check for races
- go_build to ensure code compiles
- go_lint to verify code quality`

		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: prompt},
				},
			},
		}, nil
	})
	count++

	addDepArgs := []*mcp.PromptArgument{
		{
			Name:        "package_path",
			Description: "Package path to add (e.g., github.com/gin-gonic/gin)",
			Required:    true,
		},
		{
			Name:        "version",
			Description: "Specific version or latest (default: latest)",
			Required:    false,
		},
	}
	resources.RegisterPrompt("add-go-dependency", "Guide for adding and managing Go dependencies using go mod", addDepArgs)
	server.AddPrompt(&mcp.Prompt{
		Name:        "add-go-dependency",
		Description: "Guide for adding and managing Go dependencies using go mod",
		Arguments:   addDepArgs,
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		packagePath := req.Params.Arguments["package_path"]
		if packagePath == "" {
			return &mcp.GetPromptResult{
				Messages: []*mcp.PromptMessage{
					{
						Role:    "user",
						Content: &mcp.TextContent{Text: "Error: package_path is required"},
					},
				},
			}, nil
		}
		version := req.Params.Arguments["version"]

		prompt := `Add Go dependency: ` + packagePath
		if version != "" {
			prompt += `@` + version
		}
		prompt += `

Steps:
1. Use go_pkg_docs tool to review package documentation first
2. Use go_mod tool with operation "get" to add the dependency:
   - If version specified: go mod get ` + packagePath + `@` + func() string {
			if version != "" {
				return version
			}
			return "latest"
		}() + `
   - If latest: go mod get ` + packagePath + `
3. Review go.mod and go.sum changes
4. Use go_mod tool with operation "tidy" to clean up dependencies
5. Use go_build to verify the project still builds
6. Use go_test to ensure tests still pass

Best practices:
- Check package documentation with go_pkg_docs before adding
- Pin to specific versions for production code
- Review dependency licenses
- Keep dependencies minimal and up-to-date
- Use go mod vendor if vendoring is required`

		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: prompt},
				},
			},
		}, nil
	})
	count++

	reviewArgs := []*mcp.PromptArgument{
		{
			Name:        "package_path",
			Description: "Package to review",
			Required:    false,
		},
	}
	resources.RegisterPrompt("go-code-review", "Checklist for reviewing Go code including style, performance, security, and best practices", reviewArgs)
	server.AddPrompt(&mcp.Prompt{
		Name:        "go-code-review",
		Description: "Checklist for reviewing Go code including style, performance, security, and best practices",
		Arguments:   reviewArgs,
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		packagePath := req.Params.Arguments["package_path"]
		pathContext := ""
		if packagePath != "" {
			pathContext = " for package: " + packagePath
		}

		prompt := `Go code review checklist` + pathContext + `:

1. Code Quality:
   - Run go_fmt to ensure code is formatted
   - Run go_lint to check for linting issues
   - Verify error handling is comprehensive
   - Check for proper context usage in long-running operations

2. Testing:
   - Run go_test to verify all tests pass
   - Check test coverage with go_test -cover
   - Ensure edge cases are tested
   - Verify benchmarks exist for performance-critical code

3. Concurrency:
   - Run go_race_detect to check for race conditions
   - Verify proper use of mutexes and channels
   - Check for goroutine leaks
   - Ensure proper context cancellation

4. Performance:
   - Review for unnecessary allocations
   - Check for efficient data structures
   - Verify no obvious performance bottlenecks
   - Consider profiling if performance is critical

5. Documentation:
   - Verify exported functions have doc comments
   - Check that examples exist for public APIs
   - Ensure README is up to date
   - Review package documentation with go_doc

6. Dependencies:
   - Review go.mod for unnecessary dependencies
   - Check for security vulnerabilities
   - Verify dependency versions are appropriate
   - Use go_mod tidy to clean up

7. Build and Deployment:
   - Verify go_build succeeds
   - Check cross-compilation if needed (go_cross_compile)
   - Review build flags and tags
   - Ensure proper error messages and logging

Use the available tools to automate checks:
- go_fmt for formatting
- go_lint for static analysis
- go_test for testing
- go_race_detect for concurrency issues
- go_build to verify compilation`

		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: prompt},
				},
			},
		}, nil
	})
	count++

	deployArgs := []*mcp.PromptArgument{
		{
			Name:        "target_os",
			Description: "Target operating system: linux, darwin, windows",
			Required:    false,
		},
		{
			Name:        "target_arch",
			Description: "Target architecture: amd64, arm64, 386",
			Required:    false,
		},
	}
	resources.RegisterPrompt("go-server-deployment", "Guide for building and deploying Go servers including cross-compilation, optimization, and production best practices", deployArgs)
	server.AddPrompt(&mcp.Prompt{
		Name:        "go-server-deployment",
		Description: "Guide for building and deploying Go servers including cross-compilation, optimization, and production best practices",
		Arguments:   deployArgs,
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		targetOS := req.Params.Arguments["target_os"]
		targetArch := req.Params.Arguments["target_arch"]
		if targetOS == "" {
			targetOS = "linux"
		}
		if targetArch == "" {
			targetArch = "amd64"
		}

		prompt := `Deploy Go server for ` + targetOS + `/` + targetArch + `:

Pre-deployment steps:
1. Run go_test to ensure all tests pass
2. Run go_race_detect to check for race conditions
3. Run go_lint to verify code quality
4. Review and update dependencies with go_mod tidy

Build for production:
1. Use go_cross_compile tool with:
   - GOOS: ` + targetOS + `
   - GOARCH: ` + targetArch + `
   - Output: specify binary name
2. Or use go_build with optimization flags:
   - -trimpath for reproducible builds
   - -ldflags for version embedding
   - Consider -race flag only for testing, not production

Production considerations:
- Use go_server_start tool to test server startup
- Monitor with go_server_status and go_server_logs
- Set appropriate environment variables
- Configure proper logging
- Use graceful shutdown (SIGTERM)
- Consider health check endpoints
- Set up proper error handling and recovery

Security:
- Review dependencies for vulnerabilities
- Use minimal base images if containerizing
- Set appropriate file permissions
- Avoid hardcoded secrets
- Use environment variables for configuration

Performance:
- Profile with go_profile before deployment
- Optimize based on go_benchmark results
- Consider connection pooling
- Review memory usage with go_memory_profile
- Set appropriate resource limits

After deployment:
- Monitor server logs with go_server_logs
- Check server status with go_server_status
- Profile production workloads if needed
- Set up alerts for errors and performance issues`

		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: prompt},
				},
			},
		}, nil
	})
	count++
	return count
}
