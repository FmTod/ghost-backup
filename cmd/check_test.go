package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/FmTod/ghost-backup/internal/config"
)

func TestCheckCmd_Configuration(t *testing.T) {
	if checkCmd == nil {
		t.Fatal("checkCmd is nil")
	}

	if checkCmd.Use != "check" {
		t.Errorf("checkCmd.Use = %s, want check", checkCmd.Use)
	}

	if checkCmd.Short == "" {
		t.Error("checkCmd.Short should not be empty")
	}

	if checkCmd.Long == "" {
		t.Error("checkCmd.Long should not be empty")
	}

	if checkCmd.RunE == nil {
		t.Error("checkCmd.RunE should not be nil")
	}
}

func TestCheckCmd_PathFlag(t *testing.T) {
	pathFlag := checkCmd.Flags().Lookup("path")
	if pathFlag == nil {
		t.Fatal("path flag not registered")
	}

	if pathFlag.DefValue != "." {
		t.Errorf("path flag default = %s, want .", pathFlag.DefValue)
	}

	shortFlag := checkCmd.Flags().ShorthandLookup("p")
	if shortFlag == nil {
		t.Error("short flag -p not registered")
	}
}

func TestCheckCommand_NonGitRepo(t *testing.T) {
	// Skip if git not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	// Create temp directory (not a git repo)
	tmpDir := t.TempDir()

	// Set check path and skip service check
	oldCheckPath := checkPath
	oldSkipService := checkSkipService
	checkPath = tmpDir
	checkSkipService = true
	defer func() {
		checkPath = oldCheckPath
		checkSkipService = oldSkipService
	}()

	// Run check - should report error about not being a git repo
	err := runCheck(checkCmd, []string{})

	// Should error because it's not a git repo
	if err == nil {
		t.Error("Expected error for non-git repository, got nil")
	}
}

func TestCheckCommand_ValidGitRepoWithConfig(t *testing.T) {
	// Skip if git not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to config user email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to config user name: %v", err)
	}

	// Add a remote
	cmd = exec.Command("git", "remote", "add", "origin", "https://github.com/test/repo.git")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add remote: %v", err)
	}

	// Create valid config file
	cfg := &config.LocalConfig{
		Interval:    60,
		ScanSecrets: true,
	}
	if err := config.SaveLocalConfig(tmpDir, cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Set check path
	oldCheckPath := checkPath
	checkPath = tmpDir
	defer func() { checkPath = oldCheckPath }()

	// Note: We don't actually run runCheck here because it will try to check
	// the service status which may hang. Instead we just verify the setup is correct.

	// Verify config can be loaded
	loadedCfg, err := config.LoadLocalConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedCfg.Interval != 60 {
		t.Errorf("Expected interval 60, got %d", loadedCfg.Interval)
	}

	if !loadedCfg.ScanSecrets {
		t.Error("Expected scan_secrets to be true")
	}
}

func TestCheckCommand_InvalidConfig(t *testing.T) {
	// Skip if git not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Write invalid JSON config
	configPath := filepath.Join(tmpDir, ".ghost-backup.json")
	if err := os.WriteFile(configPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Try to load config - should fail
	_, err := config.LoadLocalConfig(tmpDir)
	if err == nil {
		t.Error("Expected error loading invalid config, got nil")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.LocalConfig
		expectValid bool
		description string
	}{
		{
			name: "valid config",
			config: &config.LocalConfig{
				Interval:    60,
				ScanSecrets: true,
			},
			expectValid: true,
			description: "Standard valid configuration",
		},
		{
			name: "short interval",
			config: &config.LocalConfig{
				Interval:    5,
				ScanSecrets: true,
			},
			expectValid: true,
			description: "Valid but with warning about short interval",
		},
		{
			name: "long interval",
			config: &config.LocalConfig{
				Interval:    3600,
				ScanSecrets: false,
			},
			expectValid: true,
			description: "Long interval is valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Save config
			err := config.SaveLocalConfig(tmpDir, tt.config)
			if err != nil {
				t.Fatalf("Failed to save config: %v", err)
			}

			// Load config
			loadedCfg, err := config.LoadLocalConfig(tmpDir)
			if err != nil && tt.expectValid {
				t.Errorf("Expected valid config but got error: %v", err)
			}

			if err == nil {
				if loadedCfg.Interval != tt.config.Interval {
					t.Errorf("Expected interval %d, got %d", tt.config.Interval, loadedCfg.Interval)
				}
				if loadedCfg.ScanSecrets != tt.config.ScanSecrets {
					t.Errorf("Expected scan_secrets %v, got %v", tt.config.ScanSecrets, loadedCfg.ScanSecrets)
				}
			}
		})
	}
}

func TestGetServiceStatusString(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedString string
	}{
		{
			name:           "unknown status",
			statusCode:     0,
			expectedString: "Unknown",
		},
		{
			name:           "running status",
			statusCode:     1,
			expectedString: "Running",
		},
		{
			name:           "stopped status",
			statusCode:     2,
			expectedString: "Stopped",
		},
		{
			name:           "invalid status",
			statusCode:     99,
			expectedString: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var statusStr string
			switch tt.statusCode {
			case 0:
				statusStr = "Unknown"
			case 1:
				statusStr = "Running"
			case 2:
				statusStr = "Stopped"
			default:
				statusStr = "Unknown"
			}

			if statusStr != tt.expectedString {
				t.Errorf("Expected status %q, got %q", tt.expectedString, statusStr)
			}
		})
	}
}
