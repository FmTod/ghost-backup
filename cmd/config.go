package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/FmTod/ghost-backup/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage global configuration",
	Long:  `Configure global settings for ghost-backup, such as Git authentication tokens.`,
}

var setTokenCmd = &cobra.Command{
	Use:   "set-token",
	Short: "Set Git credentials for authentication",
	Long: `Set Git username and personal access token to use for non-interactive authentication.
This is required when running as a service to avoid interactive prompts for credentials.

The credentials will be stored in ~/.config/ghost-backup/config.json with restricted permissions.

Example:
  ghost-backup config set-token
  ghost-backup config set-token --username myuser --token ghp_xxxxxxxxxxxx
  ghost-backup config set-token --token ghp_xxxxxxxxxxxx`,
	RunE: runSetToken,
}

var getTokenCmd = &cobra.Command{
	Use:   "get-token",
	Short: "Display the configured Git credentials (token masked)",
	Long:  `Display the currently configured Git username and token in masked form.`,
	RunE:  runGetToken,
}

var clearTokenCmd = &cobra.Command{
	Use:   "clear-token",
	Short: "Clear the configured Git credentials",
	Long:  `Remove the Git username and token from the configuration.`,
	RunE:  runClearToken,
}

var (
	username string
	token    string
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(setTokenCmd)
	configCmd.AddCommand(getTokenCmd)
	configCmd.AddCommand(clearTokenCmd)

	setTokenCmd.Flags().StringVarP(&username, "username", "u", "", "Git username")
	setTokenCmd.Flags().StringVarP(&token, "token", "t", "", "Git personal access token")
}

func runSetToken(cmd *cobra.Command, args []string) error {
	var usernameValue, tokenValue string

	// Handle username
	if username != "" {
		usernameValue = username
	} else {
		// Prompt for username (optional)
		fmt.Print("Enter Git username (optional, press Enter to skip): ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read username: %w", err)
		}
		usernameValue = strings.TrimSpace(input)
	}

	// Handle token
	if token != "" {
		tokenValue = token
	} else {
		// Prompt for token (hidden input)
		fmt.Print("Enter Git personal access token: ")
		tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read token: %w", err)
		}
		fmt.Println() // New line after hidden input
		tokenValue = strings.TrimSpace(string(tokenBytes))
	}

	if tokenValue == "" {
		return fmt.Errorf("token cannot be empty")
	}

	// Load current config
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}

	// Set credentials
	globalConfig.GitUser = usernameValue
	globalConfig.GitToken = tokenValue

	// Save config
	if err := config.SaveGlobalConfig(globalConfig); err != nil {
		return fmt.Errorf("failed to save global config: %w", err)
	}

	configPath, _ := config.GetGlobalConfigPath()
	fmt.Printf("✓ Git credentials saved to %s\n", configPath)
	if usernameValue != "" {
		fmt.Printf("  Username: %s\n", usernameValue)
	}
	fmt.Println("✓ Credentials will be used for non-interactive Git authentication")
	fmt.Println()
	fmt.Println("Note: Restart the service for changes to take effect:")
	fmt.Println("  ghost-backup service restart")

	return nil
}

func runGetToken(cmd *cobra.Command, args []string) error {
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}

	if globalConfig.GitToken == "" {
		fmt.Println("No Git credentials configured")
		fmt.Println()
		fmt.Println("To set credentials, run:")
		fmt.Println("  ghost-backup config set-token")
		return nil
	}

	// Display username if set
	if globalConfig.GitUser != "" {
		fmt.Printf("Git username: %s\n", globalConfig.GitUser)
	}

	// Mask the token, showing only first and last 4 characters
	masked := maskToken(globalConfig.GitToken)
	fmt.Printf("Git token: %s\n", masked)

	return nil
}

func runClearToken(cmd *cobra.Command, args []string) error {
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}

	if globalConfig.GitToken == "" {
		fmt.Println("No Git credentials configured")
		return nil
	}

	// Confirm deletion
	fmt.Print("Are you sure you want to clear the Git credentials? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		fmt.Println("Cancelled")
		return nil
	}

	// Clear credentials
	globalConfig.GitUser = ""
	globalConfig.GitToken = ""

	// Save config
	if err := config.SaveGlobalConfig(globalConfig); err != nil {
		return fmt.Errorf("failed to save global config: %w", err)
	}

	fmt.Println("✓ Git credentials cleared")
	fmt.Println()
	fmt.Println("Note: Restart the service for changes to take effect:")
	fmt.Println("  ghost-backup service restart")

	return nil
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}

	prefix := token[:4]
	suffix := token[len(token)-4:]
	middle := strings.Repeat("*", len(token)-8)

	return prefix + middle + suffix
}
