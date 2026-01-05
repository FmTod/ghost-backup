package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/FmTod/ghost-backup/internal/config"
	"github.com/FmTod/ghost-backup/internal/git"
	"github.com/FmTod/ghost-backup/internal/service"
	"github.com/spf13/cobra"
)

var (
	initPath        string
	initInterval    int
	initScanSecrets bool
	initOnlyStaged  bool
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
	initCmd.Flags().IntVarP(&initInterval, "interval", "i", config.DefaultInterval, "Backup interval in minutes")
	initCmd.Flags().BoolVarP(&initScanSecrets, "scan-secrets", "s", config.DefaultScanSecrets, "Enable secret scanning with gitleaks")
	initCmd.Flags().BoolVarP(&initOnlyStaged, "only-staged", "o", config.DefaultOnlyStaged, "Backup only staged changes (exclude unstaged)")
}

func runInit(*cobra.Command, []string) error {
	// Get an absolute path
	absPath, err := filepath.Abs(initPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Verify it's a git repository
	repo := git.NewGitRepo(absPath)
	isGitRepo, err := repo.IsGitRepo()
	if err != nil {
		return fmt.Errorf("git validation error: %w", err)
	}
	if !isGitRepo {
		return fmt.Errorf("not a git repository: %s", absPath)
	}

	fmt.Printf("Initializing ghost-backup for repository: %s\n", absPath)

	// Create/update local config
	localConfig := &config.LocalConfig{
		Interval:    initInterval,
		ScanSecrets: initScanSecrets,
		OnlyStaged:  initOnlyStaged,
	}

	if err := config.SaveLocalConfig(absPath, localConfig); err != nil {
		return fmt.Errorf("failed to save local config: %w", err)
	}

	fmt.Printf("✓ Created local config: .ghost-backup.json\n")
	fmt.Printf("  - Interval: %d minutes\n", initInterval)
	fmt.Printf("  - Scan secrets: %v\n", initScanSecrets)
	fmt.Printf("  - Only staged: %v\n", initOnlyStaged)

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

	// Check and prompt for credentials before starting service
	if !config.CheckCredentialsConfigured() {
		fmt.Println("\n" + strings.Repeat("─", 60))
		fmt.Println("⚠ Git credentials not configured")
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println("For the service to work properly, you should configure:")
		fmt.Println("  • Username - Identifies your backups in the team")
		fmt.Println("  • Token - Required for non-interactive git push")
		fmt.Print("\nWould you like to configure them now? (Y/n): ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))

		if response == "" || response == "y" || response == "yes" {
			if _, err := config.PromptForMissingCredentials(); err != nil {
				fmt.Printf("Warning: Failed to configure credentials: %v\n", err)
				fmt.Println("You can configure them later with:")
				fmt.Println("  ghost-backup config set-token --username <user> --token <token>")
			}
		} else {
			fmt.Println("\nSkipped. You can configure credentials later with:")
			fmt.Println("  ghost-backup config set-token --username <user> --token <token>")
		}
		fmt.Println()
	}

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
