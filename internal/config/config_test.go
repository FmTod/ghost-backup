package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() error = %v", err)
	}

	if dir == "" {
		t.Error("GetConfigDir() returned empty string")
	}

	if !filepath.IsAbs(dir) {
		t.Errorf("GetConfigDir() should return absolute path, got %s", dir)
	}
}

func TestGetRegistryPath(t *testing.T) {
	path, err := GetRegistryPath()
	if err != nil {
		t.Fatalf("GetRegistryPath() error = %v", err)
	}

	if path == "" {
		t.Error("GetRegistryPath() returned empty string")
	}

	if filepath.Base(path) != "registry.json" {
		t.Errorf("GetRegistryPath() should end with registry.json, got %s", path)
	}
}

func TestRegistry_AddRepository(t *testing.T) {
	registry := &Registry{Repositories: []string{}}

	err := registry.AddRepository("/path/to/repo1")
	if err != nil {
		t.Errorf("AddRepository() error = %v", err)
	}

	if len(registry.Repositories) != 1 {
		t.Errorf("Expected 1 repository, got %d", len(registry.Repositories))
	}

	if registry.Repositories[0] != "/path/to/repo1" {
		t.Errorf("Expected /path/to/repo1, got %s", registry.Repositories[0])
	}

	err = registry.AddRepository("/path/to/repo1")
	if err != nil {
		t.Errorf("AddRepository() duplicate error = %v", err)
	}

	if len(registry.Repositories) != 1 {
		t.Errorf("Expected 1 repository after duplicate add, got %d", len(registry.Repositories))
	}
}

func TestRegistry_RemoveRepository(t *testing.T) {
	registry := &Registry{
		Repositories: []string{"/path/to/repo1", "/path/to/repo2", "/path/to/repo3"},
	}

	err := registry.RemoveRepository("/path/to/repo2")
	if err != nil {
		t.Errorf("RemoveRepository() error = %v", err)
	}

	if len(registry.Repositories) != 2 {
		t.Errorf("Expected 2 repositories, got %d", len(registry.Repositories))
	}

	err = registry.RemoveRepository("/path/to/nonexistent")
	if err == nil {
		t.Error("RemoveRepository() should return error for non-existent repo")
	}
}

func TestRegistry_GetRepositories(t *testing.T) {
	registry := &Registry{
		Repositories: []string{"/path/to/repo1", "/path/to/repo2"},
	}

	repos := registry.GetRepositories()

	if len(repos) != 2 {
		t.Errorf("Expected 2 repositories, got %d", len(repos))
	}

	repos[0] = "/modified"
	if registry.Repositories[0] == "/modified" {
		t.Error("GetRepositories() should return a copy, not the original slice")
	}
}

func TestGetLocalConfigPath(t *testing.T) {
	repoPath := "/path/to/repo"
	configPath := GetLocalConfigPath(repoPath)

	expected := filepath.Join(repoPath, ".ghost-backup.json")
	if configPath != expected {
		t.Errorf("GetLocalConfigPath() = %s, want %s", configPath, expected)
	}
}

func TestLoadLocalConfig_DefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()

	config, err := LoadLocalConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadLocalConfig() error = %v", err)
	}

	if config.Interval != DefaultInterval {
		t.Errorf("Default interval = %d, want %d", config.Interval, DefaultInterval)
	}

	if config.ScanSecrets != DefaultScanSecrets {
		t.Errorf("Default ScanSecrets = %v, want %v", config.ScanSecrets, DefaultScanSecrets)
	}
}

func TestLoadLocalConfig_ExistingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".ghost-backup.json")

	testConfig := &LocalConfig{
		Interval:    120,
		ScanSecrets: false,
	}

	data, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	config, err := LoadLocalConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadLocalConfig() error = %v", err)
	}

	if config.Interval != 120 {
		t.Errorf("Interval = %d, want 120", config.Interval)
	}

	if config.ScanSecrets != false {
		t.Errorf("ScanSecrets = %v, want false", config.ScanSecrets)
	}
}

func TestSaveLocalConfig(t *testing.T) {
	tmpDir := t.TempDir()

	config := &LocalConfig{
		Interval:    90,
		ScanSecrets: true,
	}

	err := SaveLocalConfig(tmpDir, config)
	if err != nil {
		t.Fatalf("SaveLocalConfig() error = %v", err)
	}

	configPath := filepath.Join(tmpDir, ".ghost-backup.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	loadedConfig, err := LoadLocalConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.Interval != 90 {
		t.Errorf("Saved interval = %d, want 90", loadedConfig.Interval)
	}

	if loadedConfig.ScanSecrets != true {
		t.Errorf("Saved ScanSecrets = %v, want true", loadedConfig.ScanSecrets)
	}
}
