package cmd

import (
	"fmt"
	"os"

	"github.com/FmTod/ghost-backup/internal/config"
	"github.com/FmTod/ghost-backup/internal/git"
	"github.com/spf13/cobra"
)

var (
	inspectShowDiff bool
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <hash>",
	Short: "Inspect a backup snapshot",
	Long: `Inspect a backup snapshot to see detailed information including
commit details, files changed, and optionally the full diff.
Must be run from within a git repository.`,
	Args: cobra.ExactArgs(1),
	RunE: runInspect,
}

func init() {
	rootCmd.AddCommand(inspectCmd)

	inspectCmd.Flags().BoolVarP(&inspectShowDiff, "diff", "d", false, "Show the full diff")
}

func runInspect(_ *cobra.Command, args []string) error {
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

	// Get user email and branch for fetching
	userEmail, err := repo.GetUserEmail()
	if err != nil {
		return fmt.Errorf("failed to get user email: %w", err)
	}

	// Get user name for identifier generation
	// Error is non-fatal as GenerateUserIdentifier has fallback logic
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

	// Check if the object exists locally, fetch if not
	if !repo.ObjectExists(hash) {
		// Fetch the backup ref to ensure we have the object
		refName := fmt.Sprintf("refs/backups/%s/%s", git.SanitizeRefName(userIdentifier), git.SanitizeRefName(branch))
		fmt.Printf("Fetching backup from %s...\n", refName)

		if err := repo.FetchBackupRef(remote, refName); err != nil {
			return fmt.Errorf("failed to fetch backup ref: %w", err)
		}
	}

	// Get commit info
	fmt.Printf("\n=== Backup Information ===\n\n")
	commitInfo, err := repo.GetCommitInfo(hash)
	if err != nil {
		return fmt.Errorf("failed to get commit info: %w", err)
	}
	fmt.Print(commitInfo)

	// Get files changed
	fmt.Printf("\n=== Files Changed ===\n\n")
	filesChanged, err := repo.GetFilesChanged(hash)
	if err != nil {
		return fmt.Errorf("failed to get files changed: %w", err)
	}
	fmt.Print(filesChanged)

	// Show diff if requested
	if inspectShowDiff {
		fmt.Printf("\n=== Diff ===\n\n")
		diff, err := repo.GetDiff(hash)
		if err != nil {
			return fmt.Errorf("failed to get diff: %w", err)
		}
		fmt.Print(diff)
	} else {
		fmt.Printf("\nTo see the full diff, use: ghost-backup inspect %s --diff\n", hash)
	}

	return nil
}
