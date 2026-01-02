package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/FmTod/ghost-backup/internal/git"
	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:    "users",
	Hidden: true,
	Short:  "List all users who have backups (hidden)",
	Long: `List all users who have backup snapshots in the remote repository.
This is a hidden command that shows all user identifiers that have backups.
Must be run from within a git repository.`,
	RunE: runUsers,
}

func init() {
	rootCmd.AddCommand(usersCmd)
}

func runUsers(*cobra.Command, []string) error {
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

	// Get remote
	remote, err := repo.GetRemote()
	if err != nil {
		return fmt.Errorf("failed to get remote: %w", err)
	}

	fmt.Printf("Fetching all backup users from remote...\n\n")

	// List all backup users
	users, err := repo.ListAllBackupUsers(remote)
	if err != nil {
		return fmt.Errorf("failed to list backup users: %w", err)
	}

	if len(users) == 0 {
		fmt.Printf("No backup users found.\n")
		return nil
	}

	// Sort users for consistent output
	sort.Strings(users)

	fmt.Printf("Users with backups:\n\n")
	for i, user := range users {
		fmt.Printf("%d. %s\n", i+1, user)
	}

	fmt.Printf("\nTo view a user's backups, run: ghost-backup list --user <username>\n")

	return nil
}
