package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type Config struct {
	Env map[string]string `json:"env,omitempty"`
}

//go:embed ccl.example.json
var defaultConfigJSON []byte

type Configs struct {
	Bin     string            `json:"bin,omitempty"`
	Default Config            `json:"default"`
	Configs map[string]Config `json:"configs"`
}

type FlagOptions struct {
	Yolo    bool
	Verbose bool
	Help    bool
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

func parseArgs(args []string) (*FlagOptions, []string, error) {
	// Create a new FlagSet for this parsing operation
	fs := flag.NewFlagSet("ccl", flag.ContinueOnError)
	
	// Define flags on this specific FlagSet
	opts := &FlagOptions{}
	fs.BoolVar(&opts.Yolo, "yolo", false, "enable yolo mode")
	fs.BoolVar(&opts.Yolo, "y", false, "alias for -yolo")
	fs.BoolVar(&opts.Verbose, "verbose", false, "enable verbose logging")
	fs.BoolVar(&opts.Help, "help", false, "show help message")
	fs.BoolVar(&opts.Help, "h", false, "alias for -help")
	
	// Parse the provided args instead of os.Args
	err := fs.Parse(args)
	if err != nil {
		return nil, nil, err
	}
	
	// Return parsed options and remaining args
	return opts, fs.Args(), nil
}


func main() {
	startTime := time.Now()
	configPath := getConfigPath()

	// Parse arguments - handle special subcommands first
	args := os.Args[1:]
	
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: config name or subcommand required as first argument\n")
		fmt.Fprintf(os.Stderr, "Usage: %s <config-name|list> [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s list                    list available configurations\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s <config-name> [options]  run claude with specified config\n", os.Args[0])
		os.Exit(1)
	}
	
	// Handle special subcommands
	if args[0] == "list" {
		// Load configurations for listing
		configs, err := loadConfigs()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Fprintf(os.Stderr, "Available configurations:\n")
		fmt.Fprintf(os.Stderr, "  default\n")
		for name := range configs.Configs {
			fmt.Fprintf(os.Stderr, "  %s\n", name)
		}
		os.Exit(0)
	}
	
	// Parse config name from first argument
	configName := args[0]
	argsForParsing := args[1:] // remaining args after config name
	
	// Parse flags from remaining args
	opts, remainingArgs, err := parseArgs(argsForParsing)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if opts.Verbose {
		fmt.Printf("Using config: %s\n", configPath)
		fmt.Printf("Initial args: %v\n", os.Args[1:])
		fmt.Printf("Config name: %s\n", configName)
		fmt.Printf("Args for parsing: %v\n", argsForParsing)
		fmt.Printf("Parsed flags: yolo=%v, verbose=%v, help=%v\n",
			opts.Yolo, opts.Verbose, opts.Help)
		fmt.Printf("Remaining args: %v\n", remainingArgs)
	}

	// Load configurations
	configs, err := loadConfigs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle help flag
	if opts.Help {
		fmt.Fprintf(os.Stderr, "Usage: %s <config-name> [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s list\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nSubcommands:\n")
		fmt.Fprintf(os.Stderr, "  list                 list available configurations\n")
		fmt.Fprintf(os.Stderr, "\nArguments:\n")
		fmt.Fprintf(os.Stderr, "  <config-name>        configuration name to use (required)\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -yolo, -y            enable yolo mode\n")
		fmt.Fprintf(os.Stderr, "  -verbose             enable verbose logging\n")
		fmt.Fprintf(os.Stderr, "  -help, -h            show help message\n")
		fmt.Fprintf(os.Stderr, "\nOther options are passed through to claude command\n")
		os.Exit(0)
	}


	// Select config
	var selectedConfig Config
	if configName == "default" {
		selectedConfig = configs.Default
		if opts.Verbose {
			fmt.Printf("Using default config: %+v\n", selectedConfig)
		}
	} else {
		if config, exists := configs.Configs[configName]; exists {
			selectedConfig = config
			if opts.Verbose {
				fmt.Printf("Selected config: %+v\n", selectedConfig)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Config '%s' not found\n", configName)
			fmt.Fprintf(os.Stderr, "Available configurations:\n")
			fmt.Fprintf(os.Stderr, "  default\n")
			for name := range configs.Configs {
				fmt.Fprintf(os.Stderr, "  %s\n", name)
			}
			fmt.Fprintf(os.Stderr, "Use -list to see all available configurations\n")
			os.Exit(1)
		}
	}

	// Build transformed args
	var transformedArgs []string
	if opts.Yolo {
		if opts.Verbose {
			fmt.Println("Transforming --yolo to --dangerously-skip-permissions")
		}
		transformedArgs = append(transformedArgs, "--dangerously-skip-permissions")
	}

	// Add remaining arguments (config name was already consumed)
	transformedArgs = append(transformedArgs, remainingArgs...)

	if opts.Verbose {
		fmt.Printf("Final selected config: %+v\n", selectedConfig)
		fmt.Printf("Config name: '%s'\n", configName)
		fmt.Printf("Transformed args: %v\n", transformedArgs)
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
	if opts.Verbose {
		fmt.Printf("Original env count: %d\n", originalCount)
	}

	// Merge default.Env first
	if configs.Default.Env != nil {
		for key, value := range configs.Default.Env {
			envMap[key] = value
			if opts.Verbose {
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
			if opts.Verbose {
				if isSensitiveKey(key) {
					fmt.Printf("Added selected env var: %s=***masked***\n", key)
				} else {
					fmt.Printf("Added selected env var: %s=%s\n", key, value)
				}
			}
		}
	} else {
		if opts.Verbose {
			fmt.Println("No environment variables configured in selected config")
		}
	}

	// Convert map back to []string
	env := make([]string, 0, len(envMap))
	for key, value := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	if opts.Verbose {
		fmt.Printf("Final env count: %d (added %d)\n", len(env), len(env)-originalCount)
	}

	// Set terminal title based on selected config
	setTerminalTitle(configName)

	// Determine claude binary path
	lookPathStart := time.Now()
	var claudePath string

	if configs.Bin != "" {
		// Use configured binary path
		claudePath = configs.Bin
		if opts.Verbose {
			fmt.Printf("Using configured binary path: %s\n", claudePath)
		}
	} else {
		// Fall back to exec.LookPath
		var err error
		claudePath, err = exec.LookPath("claude")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: claude binary not found in PATH and no bin configured.\n")
			fmt.Fprintf(os.Stderr, "To fix this, run `which claude` (or `where claude` on Windows) and add the path to your config:\n")
			fmt.Fprintf(os.Stderr, "Config file: %s\n", configPath)
			fmt.Fprintf(os.Stderr, "Add: \"bin\": \"/path/to/claude\"\n")
			os.Exit(1)
		}
		if opts.Verbose {
			fmt.Printf("Found claude in PATH: %s\n", claudePath)
		}
	}

	lookPathTime := time.Since(lookPathStart)
	if opts.Verbose {
		fmt.Printf("Binary resolution time: %v\n", lookPathTime)
	}

	execTime := time.Since(startTime)

	// Execute directly: claude args...
	finalArgs := append([]string{"claude"}, transformedArgs...)
	if opts.Verbose {
		fmt.Printf("Executing: %s with args %v\n", claudePath, finalArgs)
	}

	if opts.Verbose {
		fmt.Printf("ccl startup time: %v\n", execTime)
	}

	err = syscall.Exec(claudePath, finalArgs, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing claude: %v\n", err)
		os.Exit(1)
	}
}
