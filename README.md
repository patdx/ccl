# CCL - Claude Code Launcher

A Go wrapper tool that simplifies using the Claude CLI by managing multiple API configurations and providing convenient shortcuts.

## What CCL Does

CCL acts as a smart wrapper around the Claude CLI that:

- Manages multiple API configurations (Anthropic, Z.ai, etc.)
- Handles environment variables for keys, base URLs, and options
- Provides shortcuts: `--yolo` → `--dangerously-skip-permissions`
- Gives visual feedback by setting terminal/tmux window title
- Preserves all Claude CLI functionality by passing args through

## Installation

### Prerequisites

- Go 1.21 or later
- Claude CLI installed and accessible in your PATH (or set `bin` in config)

### Build and Install

```bash
# Clone the repository
git clone <repository-url>
cd ccl

# Build the binary
make build

# Install to ~/.local/bin (add to PATH if needed)
make install
```

Alternatively, build manually:
```bash
go build -o ccl main.go
```

## Quick Start

1. Create the config file
   - Run any command (e.g., `ccl help`) once to generate `~/.config/ccl/ccl.json` (or `$XDG_CONFIG_HOME/ccl/ccl.json`)
   - You will see: "Created default config ... Please edit the config file to add your API keys"
2. Edit the config
   - Add your API keys and endpoints to the generated file
3. List available configurations
   ```bash
   ccl list
   # output includes: default and any named configs in your file
   ```
4. Use CCL with a specific configuration
   ```bash
   # Use the default config
   ccl default chat "Hello"

   # Use a named config (e.g., zai)
   ccl zai chat "Hello from Z.ai!"

   # Enable yolo mode (converts to --dangerously-skip-permissions)
   ccl zai --yolo chat "Hello"
   # or short form:
   ccl zai -y chat "Hello"

   # Verbose debug output
   ccl kimi --verbose chat "Debug this conversation"
   ```

## Usage

### Synopsis

```bash
ccl <config-name|list|help> [options] <claude-subcommand> [args...]
```

- Subcommands:
  - `list` — list available configurations (always includes `default`)
  - `help` — show detailed help
- `config-name` — ALWAYS required when running Claude (e.g., `default`, `zai`, `kimi`)
- Options (CCL flags):
  - `--yolo`, `-y` — enable yolo mode (adds `--dangerously-skip-permissions`)
  - `--verbose` — enable verbose logging
  - `--help`, `-h` — show help

Notes:
- The configuration name is ALWAYS required as the first argument (except for special subcommands `list` or `help`). Even for the default configuration, you must specify `ccl default`.
- CCL parses its own flags immediately after the config name; the first non-flag token (e.g., `chat`) stops parsing and all remaining args are passed through to Claude.
- To pass flags that might be mistaken for CCL flags, use `--` to terminate CCL flag parsing, e.g.:
  ```bash
  ccl default -- --version
  ```

### Examples

```bash
# List configs
ccl list

# Use default config
ccl default chat "Hello, Claude!"

# Use a specific config
ccl zai chat "Hello from Z.ai!"

# Yolo shortcut → --dangerously-skip-permissions
ccl kimi -y chat "Skip permissions"

# Verbose mode
ccl default --verbose chat "Inspect environment setup"
```

## Configuration

### Default Configuration Location

CCL looks for configuration in this order:
- `$XDG_CONFIG_HOME/ccl/ccl.json` (if `XDG_CONFIG_HOME` is set)
- `~/.config/ccl/ccl.json` (default)

### Configuration Format

```json
{
  "$schema": "https://raw.githubusercontent.com/patdx/ccl/refs/heads/main/ccl.schema.json",
  "bin": "/path/to/claude",
  "default": {
    "env": {}
  },
  "configs": {
    "zai": {
      "env": {
        "ANTHROPIC_BASE_URL": "https://api.z.ai/api/anthropic",
        "ANTHROPIC_AUTH_TOKEN": "YOUR_API_KEY"
      }
    },
    "deepseek": {
      "env": {
        "ANTHROPIC_BASE_URL": "https://api.deepseek.com/anthropic",
        "ANTHROPIC_AUTH_TOKEN": "YOUR_API_KEY",
        "API_TIMEOUT_MS": "600000",
        "ANTHROPIC_MODEL": "deepseek-chat",
        "ANTHROPIC_SMALL_FAST_MODEL": "deepseek-chat",
        "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
      }
    },
    "kimi": {
      "env": {
        "ANTHROPIC_BASE_URL": "https://api.moonshot.ai/anthropic",
        "ANTHROPIC_AUTH_TOKEN": "YOUR_API_KEY"
      }
    }
  }
}
```

### Configuration Properties

- `bin` (optional): Path to the Claude CLI binary. If not specified, CCL uses your `PATH` to find `claude`.
- `default`: Base configuration applied to all runs.
- `configs`: Named configurations; their `env` values override `default.env` keys when selected.

### Environment Variables

Each configuration can set environment variables that will be passed to the Claude CLI:
- `ANTHROPIC_BASE_URL`: API endpoint URL
- `ANTHROPIC_AUTH_TOKEN`: API authentication token
- Any other environment variables your Claude CLI setup requires

### Claude Binary Path

If you installed the Claude CLI in a non-standard location or via an alias, specify the binary path in your config:

```json
{
  "bin": "/home/user/.claude/local/claude",
  "default": { "env": {} }
}
```

Find your Claude binary path with:
```bash
which claude  # Linux/macOS
where claude  # Windows
```

## Features

### Terminal Integration

CCL updates your terminal window title to show the active configuration:
- Regular terminal: sets the title to `claude (config-name)`
- Tmux: renames the current window to show the active config (intentional behavior)

### Security

- Sensitive environment variables (keys including `TOKEN`, `KEY`, `SECRET`, `PASSWORD`) are masked in verbose logs
- Configuration files are created with restrictive permissions (0600)

### Error Handling

- Validates configuration names before execution and lists available options
- Clear guidance when the Claude CLI binary cannot be found (configure `bin` or ensure it’s on `PATH`)

## Development

### Running Tests

```bash
# Run all tests
make test
# or
go test ./...

# Run specific test
go test -run TestParseArgs

# Run with coverage
go test -cover ./...
```

### Code Style

```bash
# Format code
gofmt -w .

# Vet code
go vet ./...
```

### Build Commands

```bash
make build    # Build binary
make install  # Build and install to ~/.local/bin
make clean    # Remove binary
make test     # Run tests
make all      # Clean and build
```

## Troubleshooting

### Claude CLI Not Found

If CCL can't find the Claude CLI:
1. Run `which claude` (or `where claude` on Windows) to find the binary path
2. Add the path to your config file: `"bin": "/path/to/claude"`
3. Or ensure the Claude CLI is in your `PATH`

Common installation hints:
- Shell alias: check your shell rc files (e.g., `.bashrc`, `.zshrc`) for `alias claude="..."`
- Direct installs are often at `/usr/local/bin/claude` or `~/.claude/local/claude`

### Configuration Issues

- Use `ccl list` to see available configurations
- Use `ccl <config> --verbose ...` to debug environment variable setup (sensitive values are masked)
- Check file permissions on the config file (should be 0600)
