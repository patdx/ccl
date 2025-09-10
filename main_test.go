package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedYolo   bool
		expectedVerbose bool
		expectedHelp   bool
		expectedRemaining []string
		expectError    bool
	}{
		{
			name:           "yolo flag",
			args:           []string{"--yolo", "chat", "--verbose"},
			expectedYolo:   true,
			expectedVerbose: false,
			expectedHelp:   false,
			expectedRemaining: []string{"chat", "--verbose"},
			expectError:    false,
		},
		{
			name:           "short yolo flag",
			args:           []string{"-y", "chat"},
			expectedYolo:   true,
			expectedVerbose: false,
			expectedHelp:   false,
			expectedRemaining: []string{"chat"},
			expectError:    false,
		},
		{
			name:           "verbose flag",
			args:           []string{"--verbose", "chat"},
			expectedYolo:   false,
			expectedVerbose: true,
			expectedHelp:   false,
			expectedRemaining: []string{"chat"},
			expectError:    false,
		},
		{
			name:           "help flag",
			args:           []string{"--help"},
			expectedYolo:   false,
			expectedVerbose: false,
			expectedHelp:   true,
			expectedRemaining: []string{},
			expectError:    false,
		},
		{
			name:           "short help flag",
			args:           []string{"-h"},
			expectedYolo:   false,
			expectedVerbose: false,
			expectedHelp:   true,
			expectedRemaining: []string{},
			expectError:    false,
		},
		{
			name:           "mixed flags",
			args:           []string{"--yolo", "--verbose", "chat", "--model", "claude-3-5-sonnet-20241022"},
			expectedYolo:   true,
			expectedVerbose: true,
			expectedHelp:   false,
			expectedRemaining: []string{"chat", "--model", "claude-3-5-sonnet-20241022"},
			expectError:    false,
		},
		{
			name:           "no flags",
			args:           []string{"chat", "--model", "claude-3-5-sonnet-20241022"},
			expectedYolo:   false,
			expectedVerbose: false,
			expectedHelp:   false,
			expectedRemaining: []string{"chat", "--model", "claude-3-5-sonnet-20241022"},
			expectError:    false,
		},
		{
			name:           "empty args",
			args:           []string{},
			expectedYolo:   false,
			expectedVerbose: false,
			expectedHelp:   false,
			expectedRemaining: []string{},
			expectError:    false,
		},
		{
			name:           "flags with claude args",
			args:           []string{"-y", "--verbose", "chat", "--model", "claude-3-5-sonnet-20241022"},
			expectedYolo:   true,
			expectedVerbose: true,
			expectedHelp:   false,
			expectedRemaining: []string{"chat", "--model", "claude-3-5-sonnet-20241022"},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, remaining, err := parseArgs(tt.args)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if tt.expectError {
				return // Don't check other fields if we expected an error
			}


			if opts.Yolo != tt.expectedYolo {
				t.Errorf("Yolo: expected %t, got %t", tt.expectedYolo, opts.Yolo)
			}

			if opts.Verbose != tt.expectedVerbose {
				t.Errorf("Verbose: expected %t, got %t", tt.expectedVerbose, opts.Verbose)
			}


			if opts.Help != tt.expectedHelp {
				t.Errorf("Help: expected %t, got %t", tt.expectedHelp, opts.Help)
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

func TestParseArgsErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "unknown flag",
			args:        []string{"--unknown-flag"},
			expectError: true,
			errorMsg:    "flag provided but not defined: -unknown-flag",
		},
		{
			name:        "valid flags",
			args:        []string{"--yolo", "--verbose"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parseArgs(tt.args)

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