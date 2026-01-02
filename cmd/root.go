package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ghost-backup",
	Short: "Ghost Backup - Automated git backup service",
	Long: `Ghost Backup is a background safety net that pushes "invisible" git snapshots
(work-in-progress) to a backup server. It supports monitoring multiple repositories
simultaneously, each with its own configuration.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
