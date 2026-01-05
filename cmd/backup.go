package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/FmTod/ghost-backup/internal/config"
	"github.com/FmTod/ghost-backup/internal/git"
	"github.com/FmTod/ghost-backup/internal/security"
	"github.com/spf13/cobra"
)

var (
	backupPath string
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup immediately",
	Long: `Create a backup of the repository right now, without waiting for the scheduled interval.
This will create a backup if there are uncommitted changes in the repository.`,
	RunE: runBackup,
}

func init() {
	rootCmd.AddCommand(backupCmd)

	backupCmd.Flags().StringVarP(&backupPath, "path", "p", ".", "Path to the repository")
}

func runBackup(*cobra.Command, []string) error {
	// Get an absolute path
	absPath, err := filepath.Abs(backupPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	fmt.Printf("Creating backup for repository: %s\n", absPath)

	// Verify it's a git repository
	repo := git.NewGitRepo(absPath)
	isGitRepo, err := repo.IsGitRepo()
	if err != nil {
		return fmt.Errorf("git validation error: %w", err)
	}
	if !isGitRepo {
		return fmt.Errorf("not a git repository: %s", absPath)
	}

	// Load local config
	localConfig, err := config.LoadLocalConfig(absPath)
	if err != nil {
		fmt.Printf("Warning: Failed to load config, using defaults: %v\n", err)
		localConfig = &config.LocalConfig{
			Interval:    config.DefaultInterval,
			ScanSecrets: config.DefaultScanSecrets,
			OnlyStaged:  config.DefaultOnlyStaged,
		}
	}

	// Check if there are changes
	hasChanges, err := repo.HasChanges()
	if err != nil {
		return fmt.Errorf("failed to check for changes: %w", err)
	}

	if !hasChanges {
		fmt.Println("✓ No uncommitted changes to backup")
		return nil
	}

	fmt.Println("Found uncommitted changes, creating backup...")

	// Create stash
	hash, err := repo.CreateStash(localConfig.OnlyStaged)
	if err != nil {
		return fmt.Errorf("failed to create stash: %w", err)
	}

	fmt.Printf("✓ Created stash: %s\n", hash)

	// If secret scanning is enabled, scan the diff
	if localConfig.ScanSecrets {
		if !security.IsGitleaksAvailable() {
			fmt.Println("⚠ gitleaks not available, skipping secret scan")
		} else {
			fmt.Println("Scanning for secrets...")
			diff, err := repo.GetDiff(hash)
			if err != nil {
				return fmt.Errorf("failed to get diff: %w", err)
			}

			result, err := security.ScanDiff(diff)
			if err != nil {
				return fmt.Errorf("secret scan failed: %w", err)
			}

			if result.HasSecrets {
				fmt.Printf("⚠ WARNING: Secrets detected in uncommitted changes!\n")
				fmt.Printf("Gitleaks output:\n%s\n", result.Output)
				return fmt.Errorf("backup aborted due to detected secrets")
			}
			fmt.Println("✓ No secrets detected")
		}
	}

	// Get user email and branch
	userEmail, err := repo.GetUserEmail()
	if err != nil {
		return fmt.Errorf("failed to get user email: %w", err)
	}

	// Get user name for identifier generation
	userName, _ := repo.GetUserName()

	// Load global config to get git_user if configured
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		// Non-fatal, use empty config
		globalConfig = &config.GlobalConfig{}
	}

	// Generate user identifier
	userIdentifier := git.GenerateUserIdentifier(globalConfig.GitUser, userName, userEmail)

	branch, err := repo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Get remote
	remote, err := repo.GetRemote()
	if err != nil {
		return fmt.Errorf("failed to get remote: %w", err)
	}

	// Push to backup ref
	fmt.Println("Pushing backup to remote...")
	err = repo.PushToBackupRef(hash, userIdentifier, branch, remote)
	if err != nil {
		return fmt.Errorf("failed to push backup: %w", err)
	}

	fmt.Printf("✓ Backup completed successfully!\n")
	fmt.Printf("  Hash: %s\n", hash)
	fmt.Printf("  Ref: refs/backups/%s/%s\n", userIdentifier, branch)

	// Show how to restore
	fmt.Printf("\nTo restore this backup:\n")
	fmt.Printf("  ghost-backup restore %s\n", hash)

	return nil
}
