package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckCredentialsConfigured(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	t.Run("NoCredentials", func(t *testing.T) {
		// No config file exists
		if CheckCredentialsConfigured() {
			t.Error("Expected false when no credentials configured, got true")
		}
	})

	t.Run("OnlyUserConfigured", func(t *testing.T) {
		globalConfig := &GlobalConfig{
			GitUser: "testuser",
		}
		if err := SaveGlobalConfig(globalConfig); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		if !CheckCredentialsConfigured() {
			t.Error("Expected true when user configured, got false")
		}
	})

	t.Run("OnlyTokenConfigured", func(t *testing.T) {
		globalConfig := &GlobalConfig{
			GitToken: "test_token",
		}
		if err := SaveGlobalConfig(globalConfig); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		if !CheckCredentialsConfigured() {
			t.Error("Expected true when token configured, got false")
		}
	})

	t.Run("BothConfigured", func(t *testing.T) {
		globalConfig := &GlobalConfig{
			GitUser:  "testuser",
			GitToken: "test_token",
		}
		if err := SaveGlobalConfig(globalConfig); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		if !CheckCredentialsConfigured() {
			t.Error("Expected true when both configured, got false")
		}
	})

	t.Run("EmptyStrings", func(t *testing.T) {
		globalConfig := &GlobalConfig{
			GitUser:  "",
			GitToken: "",
		}
		if err := SaveGlobalConfig(globalConfig); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		if CheckCredentialsConfigured() {
			t.Error("Expected false when credentials are empty strings, got true")
		}
	})
}

func TestCheckCredentialsConfigured_InvalidConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create config directory
	configDir := filepath.Join(tmpDir, ".config", "ghost-backup")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Write invalid JSON
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, []byte("{invalid json}"), 0600); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Should return false when config is invalid
	if CheckCredentialsConfigured() {
		t.Error("Expected false when config is invalid, got true")
	}
}

func TestCheckCredentialsConfigured_NoConfigFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// No config file exists
	if CheckCredentialsConfigured() {
		t.Error("Expected false when no config file exists, got true")
	}
}

func TestPromptForMissingCredentials_BothAlreadySet(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Set both credentials
	globalConfig := &GlobalConfig{
		GitUser:  "existinguser",
		GitToken: "existing_token",
	}
	if err := SaveGlobalConfig(globalConfig); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Should return false when both are already set (no prompting needed)
	// Note: This test doesn't actually call PromptForMissingCredentials with stdin
	// because that would require mocking stdin, which is complex.
	// Instead we just verify the config loading works
	loadedConfig, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedConfig.GitUser != "existinguser" || loadedConfig.GitToken != "existing_token" {
		t.Error("Config not loaded correctly")
	}
}

func TestEnsureCredentialsConfigured_AlreadyConfigured(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Set both credentials
	globalConfig := &GlobalConfig{
		GitUser:  "existinguser",
		GitToken: "existing_token",
	}
	if err := SaveGlobalConfig(globalConfig); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// EnsureCredentialsConfigured should not prompt or modify when already configured
	// We can't test the full function without mocking stdin, but we can verify
	// that the config remains unchanged
	if err := EnsureCredentialsConfigured(); err != nil {
		t.Fatalf("EnsureCredentialsConfigured failed: %v", err)
	}

	// Verify config is still the same
	loadedConfig, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedConfig.GitUser != "existinguser" {
		t.Errorf("Expected user 'existinguser', got '%s'", loadedConfig.GitUser)
	}

	if loadedConfig.GitToken != "existing_token" {
		t.Errorf("Expected token 'existing_token', got '%s'", loadedConfig.GitToken)
	}
}

func TestCredentialsConfigFile_Permissions(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Save config
	globalConfig := &GlobalConfig{
		GitUser:  "testuser",
		GitToken: "test_token",
	}
	if err := SaveGlobalConfig(globalConfig); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file permissions are 0600
	configPath := filepath.Join(tmpDir, ".config", "ghost-backup", "config.json")
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	expectedPerms := os.FileMode(0600)
	if info.Mode().Perm() != expectedPerms {
		t.Errorf("Expected permissions %v, got %v", expectedPerms, info.Mode().Perm())
	}
}

func TestCredentialsConfig_EmptyConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Load config when no file exists - should return empty config
	globalConfig, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("Failed to load empty config: %v", err)
	}

	if globalConfig.GitUser != "" {
		t.Errorf("Expected empty user, got '%s'", globalConfig.GitUser)
	}

	if globalConfig.GitToken != "" {
		t.Errorf("Expected empty token, got '%s'", globalConfig.GitToken)
	}
}

func TestCredentialsConfig_PartialConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	t.Run("OnlyUser", func(t *testing.T) {
		globalConfig := &GlobalConfig{
			GitUser: "onlyuser",
		}
		if err := SaveGlobalConfig(globalConfig); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		loadedConfig, err := LoadGlobalConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if loadedConfig.GitUser != "onlyuser" {
			t.Errorf("Expected user 'onlyuser', got '%s'", loadedConfig.GitUser)
		}

		if loadedConfig.GitToken != "" {
			t.Errorf("Expected empty token, got '%s'", loadedConfig.GitToken)
		}
	})

	t.Run("OnlyToken", func(t *testing.T) {
		globalConfig := &GlobalConfig{
			GitToken: "onlytoken",
		}
		if err := SaveGlobalConfig(globalConfig); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		loadedConfig, err := LoadGlobalConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if loadedConfig.GitUser != "" {
			t.Errorf("Expected empty user, got '%s'", loadedConfig.GitUser)
		}

		if loadedConfig.GitToken != "onlytoken" {
			t.Errorf("Expected token 'onlytoken', got '%s'", loadedConfig.GitToken)
		}
	})
}

func TestCredentialsConfig_UpdateExisting(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create initial config
	initialConfig := &GlobalConfig{
		GitUser:  "initialuser",
		GitToken: "initialtoken",
	}
	if err := SaveGlobalConfig(initialConfig); err != nil {
		t.Fatalf("Failed to save initial config: %v", err)
	}

	// Update config
	updatedConfig := &GlobalConfig{
		GitUser:  "updateduser",
		GitToken: "updatedtoken",
	}
	if err := SaveGlobalConfig(updatedConfig); err != nil {
		t.Fatalf("Failed to save updated config: %v", err)
	}

	// Verify updated values
	loadedConfig, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedConfig.GitUser != "updateduser" {
		t.Errorf("Expected user 'updateduser', got '%s'", loadedConfig.GitUser)
	}

	if loadedConfig.GitToken != "updatedtoken" {
		t.Errorf("Expected token 'updatedtoken', got '%s'", loadedConfig.GitToken)
	}
}

func TestCredentialsConfig_DirectoryCreation(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Config directory should not exist yet
	configDir := filepath.Join(tmpDir, ".config", "ghost-backup")
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Fatal("Config directory should not exist before saving")
	}

	// Save config - should create directory
	globalConfig := &GlobalConfig{
		GitUser:  "testuser",
		GitToken: "test_token",
	}
	if err := SaveGlobalConfig(globalConfig); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("Config directory should exist after saving")
	}
}
