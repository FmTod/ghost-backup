package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/FmTod/ghost-backup/internal/git"
	"github.com/spf13/cobra"
)

var (
	workflowCron      string
	workflowRetention int
)

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Generate GitHub Actions workflow for pruning old backups",
	Long: `Generate a GitHub Actions workflow file that automatically prunes old backup refs.
Must be run from within a git repository.

The workflow will delete backup refs older than the specified retention period.`,
	RunE: runWorkflow,
}

func init() {
	rootCmd.AddCommand(workflowCmd)

	workflowCmd.Flags().StringVarP(&workflowCron, "cron", "c", "0 2 * * 0", "Cron schedule for workflow (default: weekly at 2am Sunday)")
	workflowCmd.Flags().IntVarP(&workflowRetention, "retention", "r", 30, "Number of days to keep backups (default: 30)")
}

func runWorkflow(*cobra.Command, []string) error {
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

	// Create .github/workflows directory
	workflowDir := filepath.Join(cwd, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		return fmt.Errorf("failed to create workflow directory: %w", err)
	}

	workflowPath := filepath.Join(workflowDir, "ghost-backup-prune.yml")

	// Generate workflow content
	workflowContent := generateWorkflowYAML(workflowCron, workflowRetention)

	// Write a workflow file
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		return fmt.Errorf("failed to write workflow file: %w", err)
	}

	fmt.Printf("✓ Created GitHub Actions workflow: %s\n", workflowPath)
	fmt.Printf("\nWorkflow Configuration:\n")
	fmt.Printf("  - Schedule: %s\n", workflowCron)
	fmt.Printf("  - Retention: %d days\n", workflowRetention)
	fmt.Printf("\nThe workflow will:\n")
	fmt.Printf("  1. Run on schedule: %s\n", workflowCron)
	fmt.Printf("  2. Delete backup refs older than %d days\n", workflowRetention)
	fmt.Printf("  3. Can be manually triggered from the Actions tab\n")
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Review the workflow file: %s\n", workflowPath)
	fmt.Printf("  2. Commit and push the workflow to your repository\n")
	fmt.Printf("  3. Enable GitHub Actions in your repository settings\n")

	return nil
}

func generateWorkflowYAML(cron string, retentionDays int) string {
	return fmt.Sprintf(`name: Prune Ghost Backup Refs

on:
  schedule:
    # %s
    - cron: '%s'
  workflow_dispatch:
    inputs:
      retention_days:
        description: 'Number of days to keep backups'
        required: false
        default: '%d'
        type: number

jobs:
  prune-backups:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
          
      - name: Configure Git
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          
      - name: Prune old backup refs
        env:
          RETENTION_DAYS: ${{ inputs.retention_days || %d }}
        run: |
          set -e
          
          echo "Pruning backup refs older than $RETENTION_DAYS days..."
          
          # Calculate cutoff timestamp
          CUTOFF_DATE=$(date -d "$RETENTION_DAYS days ago" +%%s)
          echo "Cutoff date: $(date -d "@$CUTOFF_DATE")"
          
          # Fetch all backup refs
          git fetch origin 'refs/backups/*:refs/backups/*' || true
          
          # Count refs
          TOTAL_REFS=$(git for-each-ref --format='%%(refname)' refs/backups/ | wc -l)
          echo "Total backup refs: $TOTAL_REFS"
          
          if [ "$TOTAL_REFS" -eq 0 ]; then
            echo "No backup refs found"
            exit 0
          fi
          
          # Find and delete old refs
          DELETED_COUNT=0
          
          git for-each-ref --format='%%(refname) %%(creatordate:unix)' refs/backups/ | while read ref timestamp; do
            if [ "$timestamp" -lt "$CUTOFF_DATE" ]; then
              echo "Deleting old ref: $ref ($(date -d "@$timestamp"))"
              
              # Delete locally
              git update-ref -d "$ref" || true
              
              # Delete from remote
              REMOTE_REF="${ref#refs/backups/}"
              git push origin --delete "refs/backups/$REMOTE_REF" 2>/dev/null || echo "Warning: Failed to delete $ref from remote"
              
              DELETED_COUNT=$((DELETED_COUNT + 1))
            else
              echo "Keeping ref: $ref ($(date -d "@$timestamp"))"
            fi
          done
          
          echo ""
          echo "Summary:"
          echo "  Total refs: $TOTAL_REFS"
          echo "  Deleted: $DELETED_COUNT"
          echo "  Remaining: $((TOTAL_REFS - DELETED_COUNT))"
          
      - name: Summary
        run: |
          echo "### Ghost Backup Cleanup Complete" >> $GITHUB_STEP_SUMMARY
          echo "" >> $GITHUB_STEP_SUMMARY
          echo "- **Retention Period**: ${{ inputs.retention_days || %d }} days" >> $GITHUB_STEP_SUMMARY
          echo "- **Status**: ✅ Successfully pruned old backup refs" >> $GITHUB_STEP_SUMMARY
`, describeCron(cron), cron, retentionDays, retentionDays, retentionDays)
}

func describeCron(cron string) string {
	// Provide human-readable descriptions for common cron patterns
	descriptions := map[string]string{
		"0 2 * * 0":   "Weekly at 2am on Sunday",
		"0 2 * * *":   "Daily at 2am",
		"0 */6 * * *": "Every 6 hours",
		"0 0 1 * *":   "Monthly on the 1st at midnight",
		"0 2 * * 1":   "Weekly at 2am on Monday",
	}

	if desc, ok := descriptions[cron]; ok {
		return desc
	}
	return "Custom schedule"
}
