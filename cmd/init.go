package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/FmTod/ghost-backup/internal/config"
	"github.com/FmTod/ghost-backup/internal/git"
	"github.com/FmTod/ghost-backup/internal/service"
	"github.com/spf13/cobra"
)

var (
	initPath        string
	initInterval    int
	initScanSecrets bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize ghost-backup for a repository",
	Long: `Initialize ghost-backup for the current repository. This will:
  1. Create/update the local .git/ghost-config.json
  2. Add the repository to the global registry
  3. Ensure the system service is installed and running`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&initPath, "path", "p", ".", "Path to the repository")
	initCmd.Flags().IntVarP(&initInterval, "interval", "i", config.DefaultInterval, "Backup interval in seconds")
	initCmd.Flags().BoolVarP(&initScanSecrets, "scan-secrets", "s", config.DefaultScanSecrets, "Enable secret scanning with gitleaks")
}

func runInit(*cobra.Command, []string) error {
	// Get an absolute path
	absPath, err := filepath.Abs(initPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Verify it's a git repository
	repo := git.NewGitRepo(absPath)
	if !repo.IsGitRepo() {
		return fmt.Errorf("not a git repository: %s", absPath)
	}

	fmt.Printf("Initializing ghost-backup for repository: %s\n", absPath)

	// Create/update local config
	localConfig := &config.LocalConfig{
		Interval:    initInterval,
		ScanSecrets: initScanSecrets,
	}

	if err := config.SaveLocalConfig(absPath, localConfig); err != nil {
		return fmt.Errorf("failed to save local config: %w", err)
	}

	fmt.Printf("✓ Created local config: .ghost-backup.json\n")
	fmt.Printf("  - Interval: %d seconds\n", initInterval)
	fmt.Printf("  - Scan secrets: %v\n", initScanSecrets)

	// Load registry
	registry, err := config.LoadRegistry()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Add repository to registry
	if err := registry.AddRepository(absPath); err != nil {
		return fmt.Errorf("failed to add repository to registry: %w", err)
	}

	if err := registry.Save(); err != nil {
		return fmt.Errorf("failed to save registry: %w", err)
	}

	fmt.Printf("✓ Added repository to global registry\n")

	// Ensure service is running
	fmt.Printf("Ensuring service is installed and running...\n")

	if err := service.EnsureServiceRunning(); err != nil {
		return fmt.Errorf("failed to ensure service is running: %w", err)
	}

	fmt.Printf("✓ Service is running\n")

	// Restart service to reload configuration
	fmt.Printf("Restarting service to reload configuration...\n")
	if err := service.RestartService(); err != nil {
		// Non-fatal error, service might pick up changes anyway
		fmt.Printf("Warning: Failed to restart service: %v\n", err)
		fmt.Printf("You may need to restart the service manually.\n")
	} else {
		fmt.Printf("✓ Service restarted\n")
	}

	fmt.Printf("\n✓ Initialization complete!\n")
	fmt.Printf("\nTo modify settings, edit: %s\n", config.GetLocalConfigPath(absPath))

	logPath, err := service.GetLogFilePath()
	if err == nil {
		fmt.Printf("To view logs: tail -f %s\n", logPath)
	}

	return nil
}
