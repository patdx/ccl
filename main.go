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

// Global flag variables
var (
	configFlag  = flag.String("config", "", "configuration name to use")
	yoloFlag    = flag.Bool("yolo", false, "enable yolo mode")
	verboseFlag = flag.Bool("verbose", false, "enable verbose logging")
	shellFlag   = flag.Bool("shell", false, "force execution via shell (for testing)")
	listFlag    = flag.Bool("list", false, "list available configurations")
	helpFlag    = flag.Bool("help", false, "show help message")
)

func init() {
	// Add aliases for flags
	flag.StringVar(configFlag, "c", *configFlag, "alias for -config")
	flag.BoolVar(yoloFlag, "y", *yoloFlag, "alias for -yolo")
	flag.BoolVar(helpFlag, "h", *helpFlag, "alias for -help")
}

func main() {
	startTime := time.Now()
	configPath := getConfigPath()

	// Parse flags
	flag.Parse()

	if *verboseFlag {
		fmt.Printf("Using config: %s\n", configPath)
		fmt.Printf("Initial args: %v\n", os.Args[1:])
		fmt.Printf("Parsed flags: config=%s, yolo=%v, verbose=%v, shell=%v, list=%v, help=%v\n",
			*configFlag, *yoloFlag, *verboseFlag, *shellFlag, *listFlag, *helpFlag)
		fmt.Printf("Remaining args: %v\n", flag.Args())
	}

	// Load configurations
	configs, err := loadConfigs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle help flag
	if *helpFlag {
		fmt.Fprintf(os.Stderr, "Usage: %s [config-name] [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -config, -c string   configuration name to use\n")
		fmt.Fprintf(os.Stderr, "  -yolo, -y            enable yolo mode\n")
		fmt.Fprintf(os.Stderr, "  -list                list available configurations\n")
		fmt.Fprintf(os.Stderr, "  -verbose             enable verbose logging\n")
		fmt.Fprintf(os.Stderr, "  -shell               force execution via shell (for testing)\n")
		fmt.Fprintf(os.Stderr, "  -help, -h            show help message\n")
		fmt.Fprintf(os.Stderr, "\nOther options are passed through to claude command\n")
		os.Exit(0)
	}

	// Handle list flag
	if *listFlag {
		fmt.Fprintf(os.Stderr, "Available configurations:\n")
		for name := range configs.Configs {
			fmt.Fprintf(os.Stderr, "  %s\n", name)
		}
		os.Exit(0)
	}

	// Determine config name - check if first remaining arg is a config name
	configName := *configFlag
	if configName == "" && len(flag.Args()) > 0 {
		firstArg := flag.Args()[0]
		if _, exists := configs.Configs[firstArg]; exists {
			configName = firstArg
			if *verboseFlag {
				fmt.Printf("Using first argument as config name: %s\n", configName)
			}
		}
	}

	// Select config
	var selectedConfig Config = configs.Default
	if configName != "" {
		if config, exists := configs.Configs[configName]; exists {
			selectedConfig = config
			if *verboseFlag {
				fmt.Printf("Selected config: %+v\n", selectedConfig)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Config '%s' not found\n", configName)
			fmt.Fprintf(os.Stderr, "Available configurations:\n")
			for name := range configs.Configs {
				fmt.Fprintf(os.Stderr, "  %s\n", name)
			}
			fmt.Fprintf(os.Stderr, "Use -list to see all available configurations\n")
			os.Exit(1)
		}
	}

	// Build transformed args
	var transformedArgs []string
	if *yoloFlag {
		if *verboseFlag {
			fmt.Println("Transforming --yolo to --dangerously-skip-permissions")
		}
		transformedArgs = append(transformedArgs, "--dangerously-skip-permissions")
	}

	// Add remaining arguments (skip first if it was used as config name)
	remainingArgs := flag.Args()
	if configName != "" && len(remainingArgs) > 0 && remainingArgs[0] == configName {
		remainingArgs = remainingArgs[1:]
	}
	transformedArgs = append(transformedArgs, remainingArgs...)

	if *verboseFlag {
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
	if *verboseFlag {
		fmt.Printf("Original env count: %d\n", originalCount)
	}

	// Merge default.Env first
	if configs.Default.Env != nil {
		for key, value := range configs.Default.Env {
			envMap[key] = value
			if *verboseFlag {
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
			if *verboseFlag {
				if isSensitiveKey(key) {
					fmt.Printf("Added selected env var: %s=***masked***\n", key)
				} else {
					fmt.Printf("Added selected env var: %s=%s\n", key, value)
				}
			}
		}
	} else {
		if *verboseFlag {
			fmt.Println("No environment variables configured in selected config")
		}
	}

	// Convert map back to []string
	env := make([]string, 0, len(envMap))
	for key, value := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	if *verboseFlag {
		fmt.Printf("Final env count: %d (added %d)\n", len(env), len(env)-originalCount)
	}

	// Set terminal title based on selected config
	setTerminalTitle(configName)

	// Check if claude is available
	lookPathStart := time.Now()
	var claudePath string
	var useShell bool

	// Try exec.LookPath first (fast for actual executables)
	claudePath, lookPathErr := exec.LookPath("claude")
	if lookPathErr != nil || *shellFlag {
		// If not found or --shell flag is used, execute via user's shell to handle aliases
		useShell = true
		userShell := os.Getenv("SHELL")
		if userShell == "" {
			userShell = "/bin/sh" // fallback
		}
		claudePath = userShell
		if *shellFlag && *verboseFlag {
			fmt.Printf("Forcing shell mode due to --shell flag\n")
		}
	}

	lookPathTime := time.Since(lookPathStart)
	if *verboseFlag {
		fmt.Printf("LookPath time: %v\n", lookPathTime)
	}

	execTime := time.Since(startTime)

	var finalArgs []string
	if useShell {
		// Execute via shell: shell -c "exec claude args..."
		shellArgs := strings.Join(append([]string{"exec", "claude"}, transformedArgs...), " ")
		finalArgs = []string{filepath.Base(claudePath), "-c", shellArgs}
		if *verboseFlag {
			fmt.Printf("Executing via shell: %s -c \"%s\"\n", claudePath, shellArgs)
		}
	} else {
		// Execute directly: claude args...
		finalArgs = append([]string{"claude"}, transformedArgs...)
		if *verboseFlag {
			fmt.Printf("Executing directly: %s with args %v\n", claudePath, finalArgs)
		}
	}

	if *verboseFlag {
		fmt.Printf("ccl startup time: %v\n", execTime)
	}

	err = syscall.Exec(claudePath, finalArgs, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing claude: %v\n", err)
		os.Exit(1)
	}
}
