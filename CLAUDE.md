# Parka Development Guidelines

## Build & Test Commands
- Build: `make build` (runs go generate ./... then go build ./...)
- Test: `make test` (runs go test ./...)
- Run single test: `go test -run TestName ./path/to/package`
- Lint: `make lint` (runs golangci-lint run -v)
- Install: `make install` (builds and copies binary to PATH)
- Docs: `make docs` (runs godoc server)
- Release: `make goreleaser` (for snapshot releases)

## Code Style Guidelines
- Go version: Go 1.23 with go1.23.3 toolchain
- Formatting: Use gofmt for all code
- Error handling: Use pkg/errors for wrapping errors with context
- Function style: Options pattern with WithX functions for configuration
- Middleware ordering: Pre-middlewares run first, parameter filters run next, then post-middlewares
- Imports order: Standard library, third-party packages, then internal packages
- Naming: Use descriptive names, avoid abbreviations except for common ones
- Security: Never expose secrets, use environment variables for sensitive configuration
- Tests: Write comprehensive tests with clear setup, execution, and verification
- Documentation: Document all exported types and functions