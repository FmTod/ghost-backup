package cmd

import (
	"fmt"
	"os"

	"github.com/FmTod/ghost-backup/internal/config"
	"github.com/FmTod/ghost-backup/internal/git"
	"github.com/spf13/cobra"
)

var (
	restoreMethod string
)

var restoreCmd = &cobra.Command{
	Use:   "restore <hash>",
	Short: "Restore a backup snapshot",
	Long: `Restore a backup snapshot by hash.
Must be run from within a git repository.

Methods:
  - apply: Apply the stash to the working directory (default)
  - cherry-pick: Cherry-pick the changes as a commit`,
	Args: cobra.ExactArgs(1),
	RunE: runRestore,
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	restoreCmd.Flags().StringVarP(&restoreMethod, "method", "m", "apply", "Restore method (apply, cherry-pick)")
}

func runRestore(_ *cobra.Command, args []string) error {
	hash := args[0]

	// Get the current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create git repo instance
	repo := git.NewGitRepo(cwd)

	// Verify it's a git repository
	if !repo.IsGitRepo() {
		return fmt.Errorf("not a git repository: %s", cwd)
	}

	fmt.Printf("Restoring backup: %s\n", hash)

	// Get user email and branch for fetching
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

	remote, err := repo.GetRemote()
	if err != nil {
		return fmt.Errorf("failed to get remote: %w", err)
	}

	// Fetch the backup ref to ensure we have the object
	refName := fmt.Sprintf("refs/backups/%s/%s", git.SanitizeRefName(userIdentifier), git.SanitizeRefName(branch))
	fmt.Printf("Fetching backup from %s...\n", refName)

	if err := repo.FetchBackupRef(remote, refName); err != nil {
		return fmt.Errorf("failed to fetch backup ref: %w", err)
	}

	// Restore based on method
	switch restoreMethod {
	case "apply":
		fmt.Printf("Applying stash...\n")
		if err := repo.ApplyStash(hash); err != nil {
			return fmt.Errorf("failed to apply stash: %w", err)
		}
		fmt.Printf("✓ Backup applied successfully\n")

	case "cherry-pick":
		fmt.Printf("Cherry-picking changes...\n")
		if err := repo.CherryPick(hash); err != nil {
			return fmt.Errorf("failed to cherry-pick: %w", err)
		}
		fmt.Printf("✓ Changes cherry-picked successfully (not committed)\n")
		fmt.Printf("Review the changes and commit when ready.\n")

	default:
		return fmt.Errorf("invalid restore method: %s", restoreMethod)
	}

	return nil
}
