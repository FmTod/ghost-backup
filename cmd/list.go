package cmd

import (
	"fmt"
	"os"

	"github.com/FmTod/ghost-backup/internal/git"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available backups for the current repository",
	Long: `List all available backup snapshots for the current repository.
Must be run from within a git repository.`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(*cobra.Command, []string) error {
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

	// Get user email
	userEmail, err := repo.GetUserEmail()
	if err != nil {
		return fmt.Errorf("failed to get user email: %w", err)
	}

	// Get current branch
	branch, err := repo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Get remote
	remote, err := repo.GetRemote()
	if err != nil {
		return fmt.Errorf("failed to get remote: %w", err)
	}

	fmt.Printf("Fetching backups for %s on branch %s...\n\n", userEmail, branch)

	// List backup refs
	refs, err := repo.ListBackupRefs(remote, userEmail, branch)
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	if len(refs) == 0 {
		fmt.Printf("No backups found for this branch.\n")
		return nil
	}

	fmt.Printf("Available backups:\n\n")
	for i, ref := range refs {
		fmt.Printf("%d. %s\n", i+1, ref.Hash[:12])
		fmt.Printf("   Full hash: %s\n", ref.Hash)
		fmt.Printf("   Ref: %s\n\n", ref.Ref)
	}

	fmt.Printf("To restore a backup, run: ghost-backup restore <hash>\n")

	return nil
}
