package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/neoscode/ghost-backup/internal/config"
	"github.com/neoscode/ghost-backup/internal/git"
	"github.com/neoscode/ghost-backup/internal/security"
	"github.com/spf13/cobra"
)

var (
	checkPath string
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check configuration and validate setup",
	Long: `Check the ghost-backup configuration for a repository and validate the setup.
This command verifies:
  - Repository is a valid git repository
  - Configuration file exists and is valid
  - Remote is configured
  - Service is running
  - Optional: gitleaks availability`,
	RunE: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)

	checkCmd.Flags().StringVarP(&checkPath, "path", "p", ".", "Path to the repository")
}

func runCheck(cmd *cobra.Command, args []string) error {
	// Get absolute path
	absPath, err := filepath.Abs(checkPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	fmt.Printf("Checking ghost-backup configuration for: %s\n\n", absPath)

	hasErrors := false
	warnings := []string{}

	// Check 1: Git repository
	fmt.Printf("🔍 Checking git repository...\n")
	repo := git.NewGitRepo(absPath)
	if !repo.IsGitRepo() {
		fmt.Printf("   ❌ Not a valid git repository\n")
		hasErrors = true
	} else {
		fmt.Printf("   ✓ Valid git repository\n")
	}

	// Check 2: Configuration file
	fmt.Printf("\n🔍 Checking configuration file...\n")
	configPath := config.GetLocalConfigPath(absPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("   ⚠️  Configuration file not found: %s\n", configPath)
		fmt.Printf("   💡 Run 'ghost-backup init' to create it\n")
		warnings = append(warnings, "Configuration file missing")
	} else {
		fmt.Printf("   ✓ Configuration file exists: %s\n", configPath)

		// Validate configuration
		cfg, err := config.LoadLocalConfig(absPath)
		if err != nil {
			fmt.Printf("   ❌ Failed to load configuration: %v\n", err)
			hasErrors = true
		} else {
			fmt.Printf("   ✓ Configuration is valid\n")
			fmt.Printf("     - Interval: %d seconds\n", cfg.Interval)
			fmt.Printf("     - Scan secrets: %v\n", cfg.ScanSecrets)

			// Validate interval
			if cfg.Interval < 10 {
				fmt.Printf("   ⚠️  Interval is very short (%d seconds)\n", cfg.Interval)
				warnings = append(warnings, "Very short backup interval may cause performance issues")
			}
		}
	}

	// Check 3: Registry
	fmt.Printf("\n🔍 Checking global registry...\n")
	registry, err := config.LoadRegistry()
	if err != nil {
		fmt.Printf("   ❌ Failed to load registry: %v\n", err)
		hasErrors = true
	} else {
		isInRegistry := false
		for _, repoPath := range registry.GetRepositories() {
			if repoPath == absPath {
				isInRegistry = true
				break
			}
		}

		if !isInRegistry {
			fmt.Printf("   ⚠️  Repository not found in global registry\n")
			fmt.Printf("   💡 Run 'ghost-backup init' to add it\n")
			warnings = append(warnings, "Repository not in registry")
		} else {
			fmt.Printf("   ✓ Repository is registered\n")
		}
	}

	// Check 4: Git configuration
	if repo.IsGitRepo() {
		fmt.Printf("\n🔍 Checking git configuration...\n")

		// Check user email
		userEmail, err := repo.GetUserEmail()
		if err != nil {
			fmt.Printf("   ❌ Failed to get user email: %v\n", err)
			hasErrors = true
		} else if userEmail == "" {
			fmt.Printf("   ❌ User email not configured\n")
			fmt.Printf("   💡 Run 'git config user.email \"your@email.com\"'\n")
			hasErrors = true
		} else {
			fmt.Printf("   ✓ User email: %s\n", userEmail)
		}

		// Check remote
		remote, err := repo.GetRemote()
		if err != nil {
			fmt.Printf("   ❌ No remote configured: %v\n", err)
			fmt.Printf("   💡 Run 'git remote add origin <url>'\n")
			hasErrors = true
		} else {
			fmt.Printf("   ✓ Remote configured: %s\n", remote)
		}

		// Check current branch
		branch, err := repo.GetCurrentBranch()
		if err != nil {
			fmt.Printf("   ⚠️  Failed to get current branch: %v\n", err)
			warnings = append(warnings, "Could not determine current branch")
		} else {
			fmt.Printf("   ✓ Current branch: %s\n", branch)
		}
	}

	// Check 5: Service status
	fmt.Printf("\n🔍 Checking service status...\n")
	status, err := getServiceStatus()
	if err != nil {
		fmt.Printf("   ⚠️  Could not determine service status: %v\n", err)
		warnings = append(warnings, "Service status unknown")
	} else {
		fmt.Printf("   ✓ Service status: %s\n", status)
		if status != "Running" {
			fmt.Printf("   ⚠️  Service is not running\n")
			fmt.Printf("   💡 Run 'ghost-backup service start' to start it\n")
			warnings = append(warnings, "Service not running")
		}
	}

	// Check 6: Gitleaks (optional)
	fmt.Printf("\n🔍 Checking gitleaks availability...\n")
	if security.IsGitleaksAvailable() {
		fmt.Printf("   ✓ gitleaks is installed and available\n")
	} else {
		fmt.Printf("   ⚠️  gitleaks not found in PATH\n")
		fmt.Printf("   💡 Install from: https://github.com/gitleaks/gitleaks\n")
		fmt.Printf("   Note: Secret scanning will be disabled without gitleaks\n")
		warnings = append(warnings, "gitleaks not available for secret scanning")
	}

	// Summary
	fmt.Printf("\n" + "═"*50 + "\n")
	fmt.Printf("SUMMARY\n")
	fmt.Printf("═"*50 + "\n")

	if hasErrors {
		fmt.Printf("❌ Configuration has ERRORS that need to be fixed\n")
	} else if len(warnings) > 0 {
		fmt.Printf("⚠️  Configuration has warnings\n")
	} else {
		fmt.Printf("✅ Configuration is valid and ready to use\n")
	}

	if len(warnings) > 0 {
		fmt.Printf("\n⚠️  Warnings (%d):\n", len(warnings))
		for i, warning := range warnings {
			fmt.Printf("   %d. %s\n", i+1, warning)
		}
	}

	if hasErrors {
		fmt.Printf("\n💡 Fix the errors above and run 'ghost-backup check' again\n")
		return fmt.Errorf("configuration check failed")
	}

	return nil
}

func getServiceStatus() (string, error) {
	status, err := getServiceStatusHelper()
	if err != nil {
		return "Unknown", err
	}

	switch status {
	case 0: // StatusUnknown
		return "Unknown", nil
	case 1: // StatusRunning
		return "Running", nil
	case 2: // StatusStopped
		return "Stopped", nil
	default:
		return "Unknown", nil
	}
}

// Helper function to allow testing
func getServiceStatusHelper() (int, error) {
	status, err := svc.GetServiceStatus()
	if err != nil {
		return 0, err
	}
	return int(status), nil
}
