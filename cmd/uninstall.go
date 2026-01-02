package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/neoscode/ghost-backup/internal/config"
	"github.com/neoscode/ghost-backup/internal/service"
	"github.com/spf13/cobra"
)

var (
	uninstallPath string
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall ghost-backup from a repository",
	Long: `Remove a repository from ghost-backup monitoring.
This will remove the repository from the global registry and restart the service.`,
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	uninstallCmd.Flags().StringVarP(&uninstallPath, "path", "p", ".", "Path to the repository")
}

func runUninstall(*cobra.Command, []string) error {
	// Get an absolute path
	absPath, err := filepath.Abs(uninstallPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	fmt.Printf("Uninstalling ghost-backup for repository: %s\n", absPath)

	// Load registry
	registry, err := config.LoadRegistry()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Remove repository from registry
	if err := registry.RemoveRepository(absPath); err != nil {
		return fmt.Errorf("failed to remove repository from registry: %w", err)
	}

	if err := registry.Save(); err != nil {
		return fmt.Errorf("failed to save registry: %w", err)
	}

	fmt.Printf("✓ Removed repository from global registry\n")

	// Remove local config
	configPath := config.GetLocalConfigPath(absPath)
	if _, err := os.Stat(configPath); err == nil {
		if err := os.Remove(configPath); err != nil {
			fmt.Printf("Warning: Failed to remove local config: %v\n", err)
		} else {
			fmt.Printf("✓ Removed local config\n")
		}
	}

	// Restart service to reload configuration
	fmt.Printf("Restarting service to reload configuration...\n")
	if err := service.RestartService(); err != nil {
		fmt.Printf("Warning: Failed to restart service: %v\n", err)
		fmt.Printf("You may need to restart the service manually.\n")
	} else {
		fmt.Printf("✓ Service restarted\n")
	}

	fmt.Printf("\n✓ Uninstallation complete!\n")

	return nil
}
