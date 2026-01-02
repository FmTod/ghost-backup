package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// EnsureCredentialsConfigured checks if git_user and git_token are configured
// and prompts the user to set them if missing
func EnsureCredentialsConfigured() error {
	globalConfig, err := LoadGlobalConfig()
	if err != nil {
		// Non-fatal, create empty config
		globalConfig = &GlobalConfig{}
	}

	needsUpdate := false

	// Check if git_user is set
	if globalConfig.GitUser == "" {
		fmt.Println("\n⚠ Git username not configured")
		fmt.Print("Enter your username (or press Enter to skip): ")

		reader := bufio.NewReader(os.Stdin)
		username, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read username: %w", err)
		}

		username = strings.TrimSpace(username)
		if username != "" {
			globalConfig.GitUser = username
			needsUpdate = true
			fmt.Printf("✓ Username set to: %s\n", username)
		}
	}

	// Check if git_token is set
	if globalConfig.GitToken == "" {
		fmt.Println("\n⚠ Git personal access token not configured")
		fmt.Println("A token is required for non-interactive authentication when running as a service.")
		fmt.Print("Enter your GitHub personal access token (or press Enter to skip): ")

		tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read token: %w", err)
		}
		fmt.Println() // New line after hidden input

		token := strings.TrimSpace(string(tokenBytes))
		if token != "" {
			globalConfig.GitToken = token
			needsUpdate = true
			fmt.Println("✓ Token configured")
		}
	}

	// Save if anything was updated
	if needsUpdate {
		if err := SaveGlobalConfig(globalConfig); err != nil {
			return fmt.Errorf("failed to save global config: %w", err)
		}

		configPath, _ := GetGlobalConfigPath()
		fmt.Printf("\n✓ Configuration saved to %s\n", configPath)

		if globalConfig.GitUser != "" {
			fmt.Printf("  Your backups will be identified as: %s\n", globalConfig.GitUser)
		}
	}

	return nil
}

// CheckCredentialsConfigured checks if credentials are configured and warns if not
// Returns true if credentials are configured, false otherwise
func CheckCredentialsConfigured() bool {
	globalConfig, err := LoadGlobalConfig()
	if err != nil {
		return false
	}

	return globalConfig.GitUser != "" || globalConfig.GitToken != ""
}

// PromptForMissingCredentials prompts for any missing credentials
// Returns true if the user provided credentials, false if they skipped
func PromptForMissingCredentials() (bool, error) {
	globalConfig, err := LoadGlobalConfig()
	if err != nil {
		globalConfig = &GlobalConfig{}
	}

	// If both are already set, no need to prompt
	if globalConfig.GitUser != "" && globalConfig.GitToken != "" {
		return false, nil
	}

	fmt.Println("\n" + strings.Repeat("─", 60))
	fmt.Println("Git Credentials Configuration")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println("For the service to work properly, you should configure:")
	fmt.Println("  1. Username - Helps identify your backups in the team")
	fmt.Println("  2. Token - Required for non-interactive git authentication")
	fmt.Println()

	needsUpdate := false

	// Prompt for username if not set
	if globalConfig.GitUser == "" {
		fmt.Print("Git username (press Enter to skip): ")
		reader := bufio.NewReader(os.Stdin)
		username, err := reader.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("failed to read username: %w", err)
		}

		username = strings.TrimSpace(username)
		if username != "" {
			globalConfig.GitUser = username
			needsUpdate = true
		}
	}

	// Prompt for token if not set
	if globalConfig.GitToken == "" {
		fmt.Print("Git personal access token (hidden, press Enter to skip): ")
		tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return false, fmt.Errorf("failed to read token: %w", err)
		}
		fmt.Println()

		token := strings.TrimSpace(string(tokenBytes))
		if token != "" {
			globalConfig.GitToken = token
			needsUpdate = true
		}
	}

	// Save if anything was updated
	if needsUpdate {
		if err := SaveGlobalConfig(globalConfig); err != nil {
			return false, fmt.Errorf("failed to save global config: %w", err)
		}

		fmt.Println("✓ Configuration saved")
		if globalConfig.GitUser != "" {
			fmt.Printf("✓ Backups will be identified as: %s\n", globalConfig.GitUser)
		}
		fmt.Println()
		return true, nil
	}

	return false, nil
}
