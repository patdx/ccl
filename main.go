package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type Config struct {
	Env map[string]string `json:"env,omitempty"`
}

//go:embed ccl.example.json
var defaultConfigJSON []byte

type Configs struct {
	Default Config            `json:"default"`
	Configs map[string]Config `json:"configs"`
}

func getConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "ccl", "ccl.json")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}
	return filepath.Join(home, ".config", "ccl", "ccl.json")
}

func loadConfigs() (*Configs, error) {
	configPath := getConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configDir := filepath.Dir(configPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return nil, fmt.Errorf("error creating config directory %s: %v", configDir, err)
		}

		if err := os.WriteFile(configPath, defaultConfigJSON, 0600); err != nil {
			return nil, fmt.Errorf("error writing config file %s: %v", configPath, err)
		}
		fmt.Printf("Created default config at %s\n", configPath)
		fmt.Println("Please edit the config file to add your API keys")
		return nil, fmt.Errorf("config file created at %s, please edit it and run again", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config: %v", err)
	}

	var configs Configs
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("error parsing config: %v", err)
	}

	return &configs, nil
}

// setTerminalTitle sets the terminal window title to show the current config.
// When running inside tmux, this will rename the tmux window to show which
// config is being used, which may be surprising to users who expect their
// window names to remain unchanged. This behavior is intentional to provide
// visual feedback about which claude config is active.
func setTerminalTitle(configName string) {
	var title string
	if configName != "" {
		title = fmt.Sprintf("claude (%s)", configName)
	} else {
		title = "claude"
	}

	// Check if we're in tmux
	if os.Getenv("TMUX") != "" {
		// Use tmux command to set window name
		// Note: This will rename the current tmux window, which may be unexpected
		// but provides useful visual feedback about the active config
		cmd := exec.Command("tmux", "rename-window", title)
		cmd.Run() // Ignore errors, as tmux might not be available
	} else {
		// Use ANSI escape sequence to set terminal title
		fmt.Printf("\033]0;%s\007", title)
	}
}

func isSensitiveKey(key string) bool {
	upper := strings.ToUpper(key)
	return strings.Contains(upper, "TOKEN") || strings.Contains(upper, "KEY") ||
		strings.Contains(upper, "SECRET") || strings.Contains(upper, "PASSWORD")
}

func parseFlags(args []string, configs *Configs) (configName string, yolo bool, verbose bool, remainingArgs []string, err error) {
	var i int
	for i = 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--config" || arg == "-c" {
			if i+1 < len(args) {
				configName = args[i+1]
				i++ // Skip the next arg (config value)
			} else {
				err = fmt.Errorf("--config requires a value")
				return
			}
		} else if arg == "--yolo" || arg == "-y" {
			yolo = true
		} else if arg == "--verbose" {
			verbose = true
		} else if arg == "--" {
			// Everything after -- is passed through verbatim
			remainingArgs = append(remainingArgs, args[i+1:]...)
			break
		} else if arg[0] != '-' && configName == "" {
			// First non-flag argument could be config name
			if _, exists := configs.Configs[arg]; exists {
				configName = arg
			} else {
				// Not a config name, add to remaining args
				remainingArgs = append(remainingArgs, args[i:]...)
				break
			}
		} else {
			// All other args go to remaining
			remainingArgs = append(remainingArgs, args[i:]...)
			break
		}
	}
	return
}

func main() {
	configPath := getConfigPath()
	var verboseFlag bool
	// Parse verbose flag early to control initial logging
	for _, arg := range os.Args[1:] {
		if arg == "--verbose" {
			verboseFlag = true
			break
		}
	}
	if verboseFlag {
		fmt.Printf("Using config: %s\n", configPath)
	}

	configs, err := loadConfigs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle help flag manually before parsing
	for _, arg := range os.Args[1:] {
		if arg == "--help" || arg == "-h" {
			fmt.Fprintf(os.Stderr, "Usage: %s [config-name] [options]\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "Options:\n")
			fmt.Fprintf(os.Stderr, "  -c, --config string   configuration name to use\n")
			fmt.Fprintf(os.Stderr, "  -y, --yolo            enable yolo mode\n")
			fmt.Fprintf(os.Stderr, "  --list                list available configurations\n")
			fmt.Fprintf(os.Stderr, "  --verbose             enable verbose logging\n")
			fmt.Fprintf(os.Stderr, "\nOther options are passed through to claude command\n")
			os.Exit(0)
		}
	}

	// Handle list flag manually before parsing
	for _, arg := range os.Args[1:] {
		if arg == "--list" {
			fmt.Fprintf(os.Stderr, "Available configurations:\n")
			for name := range configs.Configs {
				fmt.Fprintf(os.Stderr, "  %s\n", name)
			}
			os.Exit(0)
		}
	}

	if verboseFlag {
		fmt.Printf("Initial args: %v\n", os.Args[1:])
	}

	// Parse flags manually to avoid consuming claude flags
	configName, yoloFlag, verboseFlag, args, parseErr := parseFlags(os.Args[1:], configs)
	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", parseErr)
		os.Exit(1)
	}

	if verboseFlag {
		fmt.Printf("Parsed: config=%s, yolo=%v, verbose=%v, remaining=%v\n", configName, yoloFlag, verboseFlag, args)
	}

	// Select config
	var selectedConfig Config = configs.Default
	if configName != "" {
		if config, exists := configs.Configs[configName]; exists {
			selectedConfig = config
			if verboseFlag {
				fmt.Printf("Selected config: %+v\n", selectedConfig)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Config '%s' not found\n", configName)
			fmt.Fprintf(os.Stderr, "Available configurations:\n")
			for name := range configs.Configs {
				fmt.Fprintf(os.Stderr, "  %s\n", name)
			}
			fmt.Fprintf(os.Stderr, "Use --list to see all available configurations\n")
			os.Exit(1)
		}
	}

	// Build transformed args
	var transformedArgs []string
	if yoloFlag {
		if verboseFlag {
			fmt.Println("Transforming --yolo to --dangerously-skip-permissions")
		}
		transformedArgs = append(transformedArgs, "--dangerously-skip-permissions")
	}

	// Add remaining arguments
	transformedArgs = append(transformedArgs, args...)

	if verboseFlag {
		fmt.Printf("Final selected config: %+v\n", selectedConfig)
		fmt.Printf("Config name: '%s'\n", configName)
		fmt.Printf("Transformed args: %v\n", transformedArgs)
	}

	// Try to find claude using which (handles aliases in interactive shells)
	// Note: exec.LookPath is insufficient because 'claude' may be an alias
	var claudePath string
	cmd := exec.Command("which", "claude")
	output, err := cmd.Output()
	if err == nil {
		claudePath = strings.TrimSpace(string(output))
		if verboseFlag {
			fmt.Printf("Found claude via 'which': %s\n", claudePath)
		}
	} else {
		if verboseFlag {
			fmt.Printf("'which claude' failed: %v\n", err)
		}
		// Fallback to common locations if which fails
		possiblePaths := []string{
			filepath.Join(os.Getenv("HOME"), ".claude", "local", "claude"),
			"/usr/local/bin/claude",
			"/opt/claude/claude",
		}

		if verboseFlag {
			fmt.Printf("Trying fallback paths: %v\n", possiblePaths)
		}
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				claudePath = path
				if verboseFlag {
					fmt.Printf("Found claude at fallback path: %s\n", path)
				}
				break
			}
		}

		if claudePath == "" {
			fmt.Fprintf(os.Stderr, "claude command not found\n")
			os.Exit(1)
		}
	}

	// Build environment from a map to avoid duplicates
	envMap := make(map[string]string)

	// Start with current environment
	for _, envVar := range os.Environ() {
		if parts := strings.SplitN(envVar, "=", 2); len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	originalCount := len(envMap)
	if verboseFlag {
		fmt.Printf("Original env count: %d\n", originalCount)
	}

	// Merge default.Env first
	if configs.Default.Env != nil {
		for key, value := range configs.Default.Env {
			envMap[key] = value
			if verboseFlag {
				if isSensitiveKey(key) {
					fmt.Printf("Added default env var: %s=***masked***\n", key)
				} else {
					fmt.Printf("Added default env var: %s=%s\n", key, value)
				}
			}
		}
	}

	// Then merge selected config env (overrides default)
	if selectedConfig.Env != nil {
		for key, value := range selectedConfig.Env {
			envMap[key] = value
			if verboseFlag {
				if isSensitiveKey(key) {
					fmt.Printf("Added selected env var: %s=***masked***\n", key)
				} else {
					fmt.Printf("Added selected env var: %s=%s\n", key, value)
				}
			}
		}
	} else {
		if verboseFlag {
			fmt.Println("No environment variables configured in selected config")
		}
	}

	// Convert map back to []string
	env := make([]string, 0, len(envMap))
	for key, value := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	if verboseFlag {
		fmt.Printf("Final env count: %d (added %d)\n", len(env), len(env)-originalCount)
	}

	// Set terminal title based on selected config
	setTerminalTitle(configName)

	claudeArgs := append([]string{"claude"}, transformedArgs...)
	if verboseFlag {
		fmt.Printf("Final claude args: %v\n", claudeArgs)
		fmt.Printf("Executing: %s with args %v\n", claudePath, claudeArgs)
	}

	err = syscall.Exec(claudePath, claudeArgs, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing claude: %v\n", err)
		os.Exit(1)
	}
}
