# AGENTS.md - Claude Usage Widget

This document provides essential information for AI coding agents working on the Claude Usage Widget project.

## Project Overview

Claude Usage Widget is a Go-based CLI tool for monitoring Claude AI usage with progress bars, burn rate predictions, and customizable display formats.

## Build Commands

### Primary Build Commands
- **Standard build**: `make build` (optimized build)
- **Development build**: `make dev` (with debug info)
- **Release build**: `make release` (optimized build)
- **Dependencies**: `make deps`

### Makefile Targets
```bash
make build    # Optimized build with stripped symbols
make dev      # Development build with debug info
make release  # Optimized release build
make clean    # Remove build artifacts
make test     # Run all tests
make deps     # Download and tidy dependencies
make help     # Show available targets
```

## Testing Commands

### Run All Tests
```bash
make test
```

### Run Single Test
```bash
go test -run TestFunctionName
```

### Run Tests with Coverage
```bash
go test -cover ./...
```

### Fuzz Testing
```bash
go test -fuzz=FuzzFunctionName -fuzztime=30s
```

## Linting and Code Quality

### GolangCI-Lint
**Command**: `golangci-lint run`

**Enabled Linters**: errcheck, gofmt, goimports, govet, ineffassign, misspell, revive, staticcheck, unused

### Pre-commit Hooks (Lefthook)
**Pre-commit**: `go fmt ./...`, `go vet ./...`, `go test ./...`
**Pre-push**: `go mod tidy`, `go mod verify`, `go vet ./...`, `go test -v ./...`, formatting checks, `make build`

## Code Style Guidelines

### Go Standards
- Follow `gofmt` and `goimports`
- Maximum line length: 120 characters
- Use tabs for indentation

### Naming Conventions
- **Exported**: PascalCase (`LoadConfig`, `NewAPIClient`)
- **Private**: camelCase (`loadConfig`, `newAPIClient`)
- **Tests**: PascalCase with `Test` prefix
- **Constants**: PascalCase or ALL_CAPS

### Imports
Standard library first, third-party second, local last. Group by blank lines.

### Error Handling
Always check errors, use custom error types, return early with `fmt.Errorf` for context.

### Context Usage
Use `context.Context` for API calls with 45-second timeouts.

### Logging
Use structured logging with `slog` including context fields.

### Testing Patterns
- Table-driven tests for multiple cases
- Test success and error cases
- Use `httptest` for integration tests

### File Organization
One main package with files grouped by functionality:
- `main.go` - Entry point and CLI
- `config.go` - Configuration loading
- `api.go` - API client and HTTP requests
- `cache.go` - Caching logic
- `display.go` - Output formatting
- `errors.go` - Custom error types
- `burn.go` - Burn rate calculations
- `session.go` - Session processing
- `lock.go` - File locking

## Development Workflow

1. **Setup**: `make deps`
2. **Development**: `make dev` for debug builds
3. **Testing**: `make test` and run specific tests as needed
4. **Linting**: Run golangci-lint and fix issues
5. **Pre-commit**: Lefthook runs quality checks automatically
6. **Release**: `make release` for optimized builds

## Common Development Tasks

### Adding New Features
1. Create or modify source files
2. Add comprehensive tests
3. Run `make test` and `golangci-lint run`
4. Build and test manually

### Debugging Issues
1. Enable debug logging: `export CLAUDE_CODE_USAGE_DEBUG=true`
2. Use verbose output: `--verbose` flag
3. Check API responses and error messages

This document should be updated when development practices change.