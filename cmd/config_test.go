package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/FmTod/ghost-backup/internal/config"
)

func TestConfigCommands(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Test setting token
	t.Run("SetToken", func(t *testing.T) {
		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		testToken := "ghp_test_token_123456789"
		globalConfig.GitToken = testToken

		err = config.SaveGlobalConfig(globalConfig)
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Verify it was saved
		loadedConfig, err := config.LoadGlobalConfig()
		if err != nil {
			t.Fatalf("Failed to load saved config: %v", err)
		}

		if loadedConfig.GitToken != testToken {
			t.Errorf("Expected token %s, got %s", testToken, loadedConfig.GitToken)
		}

		// Verify file permissions (should be 0600)
		configPath := filepath.Join(tmpDir, ".config", "ghost-backup", "config.json")
		info, err := os.Stat(configPath)
		if err != nil {
			t.Fatalf("Failed to stat config file: %v", err)
		}

		expectedPerms := os.FileMode(0600)
		if info.Mode().Perm() != expectedPerms {
			t.Errorf("Expected permissions %v, got %v", expectedPerms, info.Mode().Perm())
		}
	})

	t.Run("SetUserAndToken", func(t *testing.T) {
		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		testUser := "testuser"
		testToken := "ghp_test_token_with_user"
		globalConfig.GitUser = testUser
		globalConfig.GitToken = testToken

		err = config.SaveGlobalConfig(globalConfig)
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Verify both were saved
		loadedConfig, err := config.LoadGlobalConfig()
		if err != nil {
			t.Fatalf("Failed to load saved config: %v", err)
		}

		if loadedConfig.GitUser != testUser {
			t.Errorf("Expected user %s, got %s", testUser, loadedConfig.GitUser)
		}

		if loadedConfig.GitToken != testToken {
			t.Errorf("Expected token %s, got %s", testToken, loadedConfig.GitToken)
		}
	})

	t.Run("SetUserOnly", func(t *testing.T) {
		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Clear previous values
		globalConfig.GitUser = "onlyuser"
		globalConfig.GitToken = "ghp_token_123"

		err = config.SaveGlobalConfig(globalConfig)
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Verify user was saved
		loadedConfig, err := config.LoadGlobalConfig()
		if err != nil {
			t.Fatalf("Failed to load saved config: %v", err)
		}

		if loadedConfig.GitUser != "onlyuser" {
			t.Errorf("Expected user 'onlyuser', got %s", loadedConfig.GitUser)
		}
	})

	t.Run("ClearToken", func(t *testing.T) {
		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		globalConfig.GitToken = ""

		err = config.SaveGlobalConfig(globalConfig)
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Verify it was cleared
		loadedConfig, err := config.LoadGlobalConfig()
		if err != nil {
			t.Fatalf("Failed to load saved config: %v", err)
		}

		if loadedConfig.GitToken != "" {
			t.Errorf("Expected empty token, got %s", loadedConfig.GitToken)
		}
	})

	t.Run("ClearBothUserAndToken", func(t *testing.T) {
		// First set both
		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		globalConfig.GitUser = "clearme"
		globalConfig.GitToken = "ghp_clearme_token"

		err = config.SaveGlobalConfig(globalConfig)
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Now clear both
		globalConfig.GitUser = ""
		globalConfig.GitToken = ""

		err = config.SaveGlobalConfig(globalConfig)
		if err != nil {
			t.Fatalf("Failed to save config after clearing: %v", err)
		}

		// Verify both were cleared
		loadedConfig, err := config.LoadGlobalConfig()
		if err != nil {
			t.Fatalf("Failed to load saved config: %v", err)
		}

		if loadedConfig.GitUser != "" {
			t.Errorf("Expected empty user, got %s", loadedConfig.GitUser)
		}

		if loadedConfig.GitToken != "" {
			t.Errorf("Expected empty token, got %s", loadedConfig.GitToken)
		}
	})
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "Standard GitHub token",
			token:    "ghp_1234567890abcdefghij",
			expected: "ghp_****************ghij",
		},
		{
			name:     "Short token",
			token:    "abc123",
			expected: "******",
		},
		{
			name:     "Empty token",
			token:    "",
			expected: "",
		},
		{
			name:     "Long token",
			token:    "ghp_verylongtokenstring1234567890",
			expected: "ghp_*************************7890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskToken(tt.token)
			if result != tt.expected {
				t.Errorf("maskToken(%s) = %s, want %s", tt.token, result, tt.expected)
			}
		})
	}
}
