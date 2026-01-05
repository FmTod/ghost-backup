package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/FmTod/ghost-backup/internal/config"
	"github.com/FmTod/ghost-backup/internal/git"
	"github.com/spf13/cobra"
)

var (
	branchesUser string
)

var branchesCmd = &cobra.Command{
	Use:   "branches",
	Short: "List branches with available backups",
	Long: `List all branches that have backup snapshots in the remote repository.
Must be run from within a git repository.`,
	RunE: runBranches,
}

func init() {
	rootCmd.AddCommand(branchesCmd)

	// Hidden flag to view branches for a specific user
	branchesCmd.Flags().StringVar(&branchesUser, "user", "", "List branches for a specific user (hidden)")
	branchesCmd.Flags().MarkHidden("user")
}

func runBranches(*cobra.Command, []string) error {
	// Get the current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create git repo instance
	repo := git.NewGitRepo(cwd)

	// Verify it's a git repository
	isGitRepo, err := repo.IsGitRepo()
	if err != nil {
		return fmt.Errorf("git validation error: %w", err)
	}
	if !isGitRepo {
		return fmt.Errorf("not a git repository: %s", cwd)
	}

	// Get remote
	remote, err := repo.GetRemote()
	if err != nil {
		return fmt.Errorf("failed to get remote: %w", err)
	}

	var branches []string
	var userIdentifier string

	// If --user flag is provided, list branches for that specific user
	if branchesUser != "" {
		userIdentifier = git.SanitizeRefName(branchesUser)
		fmt.Printf("Fetching branches for user %s...\n\n", userIdentifier)
		branches, err = repo.ListBackupBranchesForUser(remote, userIdentifier)
		if err != nil {
			return fmt.Errorf("failed to list branches for user: %w", err)
		}
	} else {
		// Get user identifier for current user
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
		userIdentifier = git.GenerateUserIdentifier(globalConfig.GitUser, userName, userEmail)

		fmt.Printf("Fetching branches for %s...\n\n", userIdentifier)
		branches, err = repo.ListBackupBranchesForUser(remote, userIdentifier)
		if err != nil {
			return fmt.Errorf("failed to list branches: %w", err)
		}
	}

	if len(branches) == 0 {
		fmt.Printf("No branches with backups found.\n")
		return nil
	}

	// Sort branches for consistent output
	sort.Strings(branches)

	fmt.Printf("Branches with backups:\n\n")
	for i, branch := range branches {
		fmt.Printf("%d. %s\n", i+1, branch)
	}

	fmt.Printf("\nTo view backups for a branch, run: ghost-backup list --branch <branch>\n")

	return nil
}
