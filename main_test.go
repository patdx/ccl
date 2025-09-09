package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFlagsCorrectly(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedConfig    string
		expectedYolo      bool
		expectedRemaining []string
	}{
		{
			name:              "basic command with claude flags",
			args:              []string{"chat", "--model", "claude-3-5-sonnet-20241022"},
			expectedConfig:    "",
			expectedYolo:      false,
			expectedRemaining: []string{"chat", "--model", "claude-3-5-sonnet-20241022"},
		},
		{
			name:              "yolo flag with command",
			args:              []string{"--yolo", "chat", "--verbose"},
			expectedConfig:    "",
			expectedYolo:      true,
			expectedRemaining: []string{"chat", "--verbose"},
		},
		{
			name:              "config flag with remaining",
			args:              []string{"--config", "zai", "chat", "--model", "claude-3-5-sonnet-20241022"},
			expectedConfig:    "zai",
			expectedYolo:      false,
			expectedRemaining: []string{"chat", "--model", "claude-3-5-sonnet-20241022"},
		},
		{
			name:              "config as positional",
			args:              []string{"zai", "chat", "--verbose"},
			expectedConfig:    "zai",
			expectedYolo:      false,
			expectedRemaining: []string{"chat", "--verbose"},
		},
		{
			name:              "mixed ccl and claude flags",
			args:              []string{"--yolo", "--config", "zai", "chat", "--model", "claude-3-5-sonnet-20241022", "--verbose"},
			expectedConfig:    "zai",
			expectedYolo:      true,
			expectedRemaining: []string{"chat", "--model", "claude-3-5-sonnet-20241022", "--verbose"},
		},
		{
			name:              "unknown config treated as command",
			args:              []string{"unknown-config", "--some-flag"},
			expectedConfig:    "",
			expectedYolo:      false,
			expectedRemaining: []string{"unknown-config", "--some-flag"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock configs for testing
			configs := &Configs{
				Configs: map[string]Config{
					"zai":     {},
					"default": {},
				},
			}
			config, yolo, _, remaining, err := parseFlags(tt.args, configs)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if config != tt.expectedConfig {
				t.Errorf("Config: expected %q, got %q", tt.expectedConfig, config)
			}

			if yolo != tt.expectedYolo {
				t.Errorf("Yolo: expected %t, got %t", tt.expectedYolo, yolo)
			}

			if len(remaining) != len(tt.expectedRemaining) {
				t.Errorf("Remaining args length: expected %d, got %d", len(tt.expectedRemaining), len(remaining))
				t.Errorf("Expected: %v", tt.expectedRemaining)
				t.Errorf("Got: %v", remaining)
				return
			}

			for i, expected := range tt.expectedRemaining {
				if remaining[i] != expected {
					t.Errorf("Remaining arg %d: expected %q, got %q", i, expected, remaining[i])
				}
			}
		})
	}
}

func TestParseFlagsErrorHandling(t *testing.T) {
	configs := &Configs{
		Configs: map[string]Config{
			"zai": {},
		},
	}

	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "config flag without value",
			args:        []string{"--config"},
			expectError: true,
			errorMsg:    "--config requires a value",
		},
		{
			name:        "config flag -c without value",
			args:        []string{"-c"},
			expectError: true,
			errorMsg:    "--config requires a value",
		},
		{
			name:        "valid config flag",
			args:        []string{"--config", "zai"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, _, err := parseFlags(tt.args, configs)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && err != nil && err.Error() != tt.errorMsg {
				t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
			}
		})
	}
}

func TestParseFlagsSentinelSupport(t *testing.T) {
	configs := &Configs{
		Configs: map[string]Config{
			"zai": {},
		},
	}

	tests := []struct {
		name              string
		args              []string
		expectedRemaining []string
	}{
		{
			name:              "double dash sentinel",
			args:              []string{"--", "--config", "something"},
			expectedRemaining: []string{"--config", "something"},
		},
		{
			name:              "double dash with yolo before",
			args:              []string{"--yolo", "--", "--config", "value"},
			expectedRemaining: []string{"--config", "value"},
		},
		{
			name:              "double dash with config before",
			args:              []string{"--config", "zai", "--", "--some-flag"},
			expectedRemaining: []string{"--some-flag"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, remaining, err := parseFlags(tt.args, configs)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(remaining) != len(tt.expectedRemaining) {
				t.Errorf("Remaining args length: expected %d, got %d", len(tt.expectedRemaining), len(remaining))
				return
			}

			for i, expected := range tt.expectedRemaining {
				if remaining[i] != expected {
					t.Errorf("Remaining arg %d: expected %q, got %q", i, expected, remaining[i])
				}
			}
		})
	}
}

func TestParseFlagsShorthandYolo(t *testing.T) {
	configs := &Configs{
		Configs: map[string]Config{},
	}

	tests := []struct {
		name         string
		args         []string
		expectedYolo bool
	}{
		{
			name:         "short -y flag",
			args:         []string{"-y"},
			expectedYolo: true,
		},
		{
			name:         "long --yolo flag",
			args:         []string{"--yolo"},
			expectedYolo: true,
		},
		{
			name:         "both flags present (-y wins)",
			args:         []string{"-y", "--yolo"},
			expectedYolo: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, yolo, _, _, err := parseFlags(tt.args, configs)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if yolo != tt.expectedYolo {
				t.Errorf("Yolo: expected %t, got %t", tt.expectedYolo, yolo)
			}
		})
	}
}

func TestGetConfigPathXDG(t *testing.T) {
	// Save original env vars
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
	}()

	tests := []struct {
		name        string
		xdgValue    string
		expectedDir string
	}{
		{
			name:        "XDG_CONFIG_HOME set",
			xdgValue:    "/custom/config",
			expectedDir: "/custom/config/ccl",
		},
		{
			name:        "XDG_CONFIG_HOME empty",
			xdgValue:    "",
			expectedDir: ".config/ccl", // Will contain user home prefix
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.xdgValue == "" {
				os.Unsetenv("XDG_CONFIG_HOME")
			} else {
				os.Setenv("XDG_CONFIG_HOME", tt.xdgValue)
			}

			path := getConfigPath()

			expectedFilename := "ccl.json"
			if filepath.Base(path) != expectedFilename {
				t.Errorf("Expected filename %q, got %q", expectedFilename, filepath.Base(path))
			}

			if tt.xdgValue != "" {
				expectedPath := filepath.Join(tt.xdgValue, "ccl", "ccl.json")
				if path != expectedPath {
					t.Errorf("Expected path %q, got %q", expectedPath, path)
				}
			} else {
				// Should contain .config/ccl/ccl.json
				if filepath.Base(filepath.Dir(path)) != "ccl" {
					t.Errorf("Expected parent directory 'ccl', got %q", filepath.Base(filepath.Dir(path)))
				}
				if filepath.Base(filepath.Dir(filepath.Dir(path))) != ".config" {
					t.Errorf("Expected grandparent directory '.config', got %q", filepath.Base(filepath.Dir(filepath.Dir(path))))
				}
			}
		})
	}
}

func TestIsSensitiveKey(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		sensitive bool
	}{
		{"TOKEN key", "API_TOKEN", true},
		{"token lowercase", "api_token", true},
		{"KEY key", "PRIVATE_KEY", true},
		{"key lowercase", "private_key", true},
		{"SECRET key", "DATABASE_SECRET", true},
		{"secret lowercase", "database_secret", true},
		{"PASSWORD key", "DB_PASSWORD", true},
		{"password lowercase", "db_password", true},
		{"Non-sensitive key", "API_URL", false},
		{"Empty key", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSensitiveKey(tt.key)
			if result != tt.sensitive {
				t.Errorf("isSensitiveKey(%q) = %v, want %v", tt.key, result, tt.sensitive)
			}
		})
	}
}
