# Contributing to MCP Go Server

Thank you for your interest in contributing to MCP Go Server! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project adheres to a Code of Conduct that all contributors are expected to follow. Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

## How to Contribute

### Reporting Bugs

If you find a bug, please open an issue using the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md). Include:

- A clear description of the bug
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Go version, etc.)
- Relevant logs or error messages

### Suggesting Features

Feature requests can be submitted using the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md). Include:

- A clear description of the feature
- Use cases and motivation
- Proposed implementation (if you have ideas)
- Alternatives considered

### Pull Requests

1. **Fork the repository** and create a branch from `main`
2. **Make your changes** following the coding standards below
3. **Write or update tests** for your changes
4. **Ensure all tests pass** locally
5. **Update documentation** if needed
6. **Submit a pull request** using the PR template

#### PR Guidelines

- Use descriptive commit messages following [Conventional Commits](https://www.conventionalcommits.org/)
- Keep PRs focused and reasonably sized
- Reference related issues in your PR description
- Ensure CI checks pass
- Request review from maintainers

## Development Setup

### Prerequisites

- Go 1.24 or higher
- Git
- Make (optional, for using Makefile)

### Getting Started

1. Clone the repository:
```bash
git clone https://github.com/inja-online/golang-mcp.git
cd golang-mcp
```

2. Install dependencies:
```bash
go mod download
```

3. Build the project:
```bash
make build
# or
go build -o bin/mcp-go ./cmd/mcp-go
```

4. Run tests:
```bash
make test
# or
go test -v -race -coverprofile=coverage.out ./...
```

### Using Docker

```bash
docker-compose up --build
```

## Coding Standards

### Go Style

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting (run `make fmt` or `go fmt ./...`)
- Follow the project's existing code style
- Write clear, self-documenting code

### Testing

- Write tests for new features and bug fixes
- Aim for good test coverage
- Use table-driven tests where appropriate
- Run tests with race detector: `go test -race ./...`

### Code Review Checklist

- [ ] Code follows Go conventions
- [ ] Tests are included and passing
- [ ] Documentation is updated
- [ ] No linter warnings
- [ ] Commit messages follow Conventional Commits

## Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage
go tool cover -html=coverage.out
```

## Linting

```bash
# Run linter
make lint

# Or manually
golangci-lint run
```

## Project Structure

```
.
├── cmd/              # Application entry points
│   ├── mcp-go/      # Main server
│   └── test-mcp/    # Test client
├── internal/         # Internal packages
│   ├── config/      # Configuration
│   ├── resources/   # MCP resources
│   ├── tools/       # MCP tools
│   ├── utils/       # Utilities
│   └── prompts/     # MCP prompts
├── .github/         # GitHub workflows and templates
└── docs/            # Documentation (if any)
```

## Questions?

If you have questions, please:
- Open a discussion on GitHub
- Check existing issues and PRs
- Review the [README.md](README.md)

Thank you for contributing!
