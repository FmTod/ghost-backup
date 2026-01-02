package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Registry holds the global list of repositories being monitored
type Registry struct {
	Repositories []string `json:"repositories"`
	mu           sync.RWMutex
}

// LocalConfig represents the per-repository configuration
type LocalConfig struct {
	Interval    int  `json:"interval"`     // Backup interval in seconds
	ScanSecrets bool `json:"scan_secrets"` // Whether to scan for secrets using gitleaks
}

const (
	DefaultInterval    = 60
	DefaultScanSecrets = true
)

// GetConfigDir returns the global config directory path
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "ghost-backup"), nil
}

// GetRegistryPath returns the path to the global registry file
func GetRegistryPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "registry.json"), nil
}

// LoadRegistry loads the global registry from disk
func LoadRegistry() (*Registry, error) {
	registryPath, err := GetRegistryPath()
	if err != nil {
		return nil, err
	}

	// Create a config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(registryPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	reg := &Registry{Repositories: []string{}}

	// If a registry file doesn't exist, return an empty registry
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		return reg, nil
	}

	data, err := os.ReadFile(registryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry: %w", err)
	}

	if len(data) == 0 {
		return reg, nil
	}

	if err := json.Unmarshal(data, reg); err != nil {
		return nil, fmt.Errorf("failed to parse registry: %w", err)
	}

	return reg, nil
}

// Save writes the registry to disk
func (r *Registry) Save() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	registryPath, err := GetRegistryPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	if err := os.WriteFile(registryPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write registry: %w", err)
	}

	return nil
}

// AddRepository adds a repository to the registry if it doesn't exist
func (r *Registry) AddRepository(path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already exists
	for _, repo := range r.Repositories {
		if repo == path {
			return nil // Already exists
		}
	}

	r.Repositories = append(r.Repositories, path)
	return nil
}

// RemoveRepository removes a repository from the registry
func (r *Registry) RemoveRepository(path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, repo := range r.Repositories {
		if repo == path {
			r.Repositories = append(r.Repositories[:i], r.Repositories[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("repository not found in registry: %s", path)
}

// GetRepositories returns a copy of the repository list
func (r *Registry) GetRepositories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	repos := make([]string, len(r.Repositories))
	copy(repos, r.Repositories)
	return repos
}

// GetLocalConfigPath returns the path to the local config file for a repository
func GetLocalConfigPath(repoPath string) string {
	return filepath.Join(repoPath, ".ghost-backup.json")
}

// LoadLocalConfig loads the local configuration from a repository
func LoadLocalConfig(repoPath string) (*LocalConfig, error) {
	configPath := GetLocalConfigPath(repoPath)

	// Default config
	config := &LocalConfig{
		Interval:    DefaultInterval,
		ScanSecrets: DefaultScanSecrets,
	}

	// If a config file doesn't exist, return default
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read local config: %w", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse local config: %w", err)
	}

	return config, nil
}

// SaveLocalConfig saves the local configuration to a repository
func SaveLocalConfig(repoPath string, config *LocalConfig) error {
	configPath := GetLocalConfigPath(repoPath)

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal local config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write local config: %w", err)
	}

	return nil
}
