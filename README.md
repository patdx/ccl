# CCL - Claude Code Launcher

A Go wrapper tool that simplifies using the Claude CLI by managing multiple API configurations and providing convenient shortcuts.

## What CCL Does

CCL acts as a smart wrapper around the Claude CLI that:

- **Manages Multiple API Configurations**: Switch between different Claude API endpoints (Anthropic, Z.ai, etc.) with ease
- **Handles Environment Variables**: Automatically sets API keys, base URLs, and other environment variables
- **Provides Shortcuts**: Converts `--yolo` to `--dangerously-skip-permissions` for convenience
- **Visual Feedback**: Shows the active configuration in your terminal/tmux window title
- **Preserves All Claude CLI Functionality**: Passes through all Claude CLI arguments seamlessly

## Installation

### Prerequisites

- Go 1.25.1 or later
- Claude CLI installed and accessible in your PATH

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

1. **First Run**: CCL will create a default configuration file at `~/.config/ccl/ccl.json`
2. **Edit Configuration**: Add your API keys and endpoints to the config file
3. **Use CCL**: Replace `claude` with `ccl` in your commands

```bash
# Instead of: claude chat "Hello"
ccl chat "Hello"

# Use specific configuration
ccl zai chat "Hello"

# Enable yolo mode (converts to --dangerously-skip-permissions)
ccl --yolo chat "Hello"
```

## Configuration

### Default Configuration Location

CCL looks for configuration in this order:
- `$XDG_CONFIG_HOME/ccl/ccl.json` (if XDG_CONFIG_HOME is set)
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

- `bin` (optional): Path to the Claude CLI binary. If not specified, CCL will use `which claude` to find it in PATH.
- `default`: Default configuration applied when no specific config is selected
- `configs`: Named configurations that can be selected at runtime

### Environment Variables

Each configuration can set environment variables that will be passed to the Claude CLI:
- `ANTHROPIC_BASE_URL`: API endpoint URL
- `ANTHROPIC_AUTH_TOKEN`: API authentication token
- Any other environment variables your Claude CLI setup requires

### Claude Binary Path

If you installed Claude CLI via an alias (e.g., `alias claude="/home/user/.claude/local/claude"`), you'll need to specify the binary path in your config:

```json
{
  "bin": "/home/user/.claude/local/claude",
  "default": { "env": {} }
}
```

To find your Claude binary path, run:
```bash
which claude  # On Linux/macOS
where claude  # On Windows
```

## Usage

### Basic Commands

```bash
# Use default configuration
ccl chat "Hello, Claude!"

# Use specific configuration
ccl zai chat "Hello from Z.ai!"
ccl --config anthropic chat "Hello from Anthropic!"

# List available configurations
ccl --list

# Enable verbose output
ccl --verbose chat "Debug this conversation"

# Use yolo mode (shortcut for --dangerously-skip-permissions)
ccl --yolo chat "Skip all permissions"
```

### Argument Parsing

CCL intelligently parses arguments:
- Configuration names can be positional: `ccl zai chat`
- Or explicit: `ccl --config zai chat`
- CCL flags are processed first, then remaining arguments are passed to Claude CLI
- Use `--` to pass arguments that might conflict: `ccl -- --config value-for-claude`

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--config` | `-c` | Specify configuration to use |
| `--yolo` | `-y` | Enable yolo mode (converts to `--dangerously-skip-permissions`) |
| `--list` | | List available configurations |
| `--verbose` | | Enable verbose logging |
| `--help` | `-h` | Show help message |

## Features

### Terminal Integration

CCL updates your terminal window title to show the active configuration:
- Regular terminal: Shows "claude (config-name)"
- Tmux: Renames the current window to show active config

### Security

- Sensitive environment variables (containing TOKEN, KEY, SECRET, PASSWORD) are masked in verbose output
- Configuration files are created with restrictive permissions (0600)

### Error Handling

- Validates configuration names before execution
- Provides helpful error messages for missing configurations
- Falls back to common Claude CLI installation paths if `which claude` fails

## Development

### Running Tests

```bash
# Run all tests
make test
# or
go test ./...

# Run specific test
go test -run TestParseFlagsCorrectly

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
3. Alternatively, ensure Claude CLI is installed and in your PATH

Common installation locations:
- Shell alias: Check your `.bashrc`/`.zshrc` for `alias claude="..."`
- Direct install: `/usr/local/bin/claude`, `~/.claude/local/claude`

### Configuration Issues

- Use `ccl --list` to see available configurations
- Use `ccl --verbose` to debug environment variable setup
- Check file permissions on config file (should be 0600)

### API Key Management

- Never commit API keys to version control
- Use environment-specific configurations for different projects
- Sensitive values are automatically masked in verbose output

