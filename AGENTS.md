# Claude Code Launcher (CCL) Agent Guidelines

## Project Goal
CCL is a Go wrapper tool that manages configurations and launches the Claude CLI with different API settings. It simplifies using the Claude CLI by managing multiple API configurations, setting environment variables automatically, and providing shortcuts like `--yolo` → `--dangerously-skip-permissions`.

## Build Commands
- Build: `go build -o ccl main.go` or `./build.sh` or `make build`
- Install: `make install` (installs to ~/.local/bin)
- Run: `go run main.go`
- Clean: `make clean` or `rm -f ccl`

## Test Commands
- Run all tests: `go test ./...`
- Run single test: `go test -run TestName`
- Test with coverage: `go test -cover ./...`

## Code Style Guidelines

### Formatting & Linting
- Format code: `gofmt -w .`
- Vet code: `go vet ./...`
- Use `go fmt` for consistent formatting

### Naming Conventions
- Functions: PascalCase for exported, camelCase for unexported
- Variables: camelCase, PascalCase for exported
- Structs: PascalCase
- Constants: PascalCase

### Error Handling
- Check errors immediately after operations
- Use `fmt.Fprintf(os.Stderr, ...)` for error messages
- Exit with `os.Exit(1)` on fatal errors
- Return errors from functions when appropriate

### Imports
- Group standard library imports first
- Use blank lines between import groups
- Keep imports organized alphabetically within groups

### Types & Structs
- Use struct tags for JSON marshaling: `json:"field_name"`
- Define config structs with clear field names
- Use pointer receivers for methods that modify structs