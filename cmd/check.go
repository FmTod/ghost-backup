package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/FmTod/ghost-backup/internal/config"
	"github.com/FmTod/ghost-backup/internal/git"
	"github.com/FmTod/ghost-backup/internal/security"
	"github.com/FmTod/ghost-backup/internal/service"
	"github.com/spf13/cobra"
)

var (
	checkPath        string
	checkSkipService bool
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
	checkCmd.Flags().BoolVar(&checkSkipService, "skip-service", false, "Skip service status check (for testing)")
	checkCmd.Flags().MarkHidden("skip-service")
}

func runCheck(cmd *cobra.Command, args []string) error {
	// Get absolute path
	absPath, err := filepath.Abs(checkPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	fmt.Printf("Checking ghost-backup configuration for: %s\n\n", absPath)

	// Check and prompt for credentials if not configured (skip in test mode)
	if !checkSkipService && !config.CheckCredentialsConfigured() {
		fmt.Println("[!] Git credentials not configured")
		fmt.Println("    Credentials are needed for the service to push backups automatically.")
		fmt.Print("\nWould you like to configure them now? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			if _, err := config.PromptForMissingCredentials(); err != nil {
				fmt.Printf("Warning: Failed to configure credentials: %v\n", err)
			}
		} else {
			fmt.Println("Skipped. You can configure credentials later with:")
			fmt.Println("  ghost-backup config set-token --username <user> --token <token>")
			fmt.Println()
		}
	}

	hasErrors := false
	warnings := []string{}

	// Check 1: Git repository
	fmt.Printf("[*] Checking git repository...\n")
	repo := git.NewGitRepo(absPath)
	isGitRepo, err := repo.IsGitRepo()
	if err != nil {
		fmt.Printf("   [FAIL] Git validation error: %v\n", err)
		hasErrors = true
	} else if !isGitRepo {
		fmt.Printf("   [FAIL] Not a valid git repository\n")
		hasErrors = true
	} else {
		fmt.Printf("   [PASS] Valid git repository\n")
	}

	// Check 2: Configuration file
	fmt.Printf("\n[*] Checking configuration file...\n")
	configPath := config.GetLocalConfigPath(absPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("   [WARN] Configuration file not found: %s\n", configPath)
		fmt.Printf("   [INFO] Run 'ghost-backup init' to create it\n")
		warnings = append(warnings, "Configuration file missing")
	} else {
		fmt.Printf("   [PASS] Configuration file exists: %s\n", configPath)

		// Validate configuration
		cfg, err := config.LoadLocalConfig(absPath)
		if err != nil {
			fmt.Printf("   [FAIL] Failed to load configuration: %v\n", err)
			hasErrors = true
		} else {
			fmt.Printf("   [PASS] Configuration is valid\n")
			fmt.Printf("     - Interval: %d minutes\n", cfg.Interval)
			fmt.Printf("     - Scan secrets: %v\n", cfg.ScanSecrets)
		}
	}

	// Check 3: Registry
	fmt.Printf("\n[*] Checking global registry...\n")
	registry, err := config.LoadRegistry()
	if err != nil {
		fmt.Printf("   [FAIL] Failed to load registry: %v\n", err)
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
			fmt.Printf("   [WARN] Repository not found in global registry\n")
			fmt.Printf("   [INFO] Run 'ghost-backup init' to add it\n")
			warnings = append(warnings, "Repository not in registry")
		} else {
			fmt.Printf("   [PASS] Repository is registered\n")
		}
	}

	// Check 4: Git configuration
	isGitRepo2, _ := repo.IsGitRepo()
	if isGitRepo2 {
		fmt.Printf("\n[*] Checking git configuration...\n")

		// Check user email
		userEmail, err := repo.GetUserEmail()
		if err != nil {
			fmt.Printf("   [FAIL] Failed to get user email: %v\n", err)
			hasErrors = true
		} else if userEmail == "" {
			fmt.Printf("   [FAIL] User email not configured\n")
			fmt.Printf("   [INFO] Run 'git config user.email \"your@email.com\"'\n")
			hasErrors = true
		} else {
			fmt.Printf("   [PASS] User email: %s\n", userEmail)
		}

		// Check user name
		userName, _ := repo.GetUserName()
		if userName != "" {
			fmt.Printf("   [PASS] User name: %s\n", userName)
		}

		// Check remote
		remote, err := repo.GetRemote()
		if err != nil {
			fmt.Printf("   [FAIL] No remote configured: %v\n", err)
			fmt.Printf("   [INFO] Run 'git remote add origin <url>'\n")
			hasErrors = true
		} else {
			fmt.Printf("   [PASS] Remote configured: %s\n", remote)
		}

		// Check current branch
		branch, err := repo.GetCurrentBranch()
		if err != nil {
			fmt.Printf("   [WARN] Failed to get current branch: %v\n", err)
			warnings = append(warnings, "Could not determine current branch")
		} else {
			fmt.Printf("   [PASS] Current branch: %s\n", branch)
		}

		// Check user identifier
		fmt.Printf("\n[*] Checking user identifier for backups...\n")
		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			fmt.Printf("   [WARN] Failed to load global config: %v\n", err)
			globalConfig = &config.GlobalConfig{} // Use empty config
		}

		userIdentifier := git.GenerateUserIdentifier(globalConfig.GitUser, userName, userEmail)
		fmt.Printf("   [INFO] Backup identifier: %s\n", userIdentifier)

		if globalConfig.GitUser != "" {
			fmt.Printf("     Source: global config (git_user)\n")
		} else if userName != "" {
			fmt.Printf("     Source: git username (sanitized from: %s)\n", userName)
		} else {
			fmt.Printf("     Source: sanitized email\n")
			fmt.Printf("   [TIP] Set a custom identifier for better team visibility:\n")
			fmt.Printf("         ghost-backup config set-token --username yourname\n")
		}

		if branch != "" {
			fmt.Printf("   [INFO] Backup ref: refs/backups/%s/%s\n", userIdentifier, branch)
		}
	}

	// Check 5: Service status (skip in tests to avoid hangs)
	if !checkSkipService {
		fmt.Printf("\n[*] Checking service status...\n")
		status, err := getServiceStatus()
		if err != nil {
			fmt.Printf("   [WARN] Could not determine service status: %v\n", err)
			warnings = append(warnings, "Service status unknown")
		} else {
			fmt.Printf("   [PASS] Service status: %s\n", status)
			if status != "Running" {
				fmt.Printf("   [WARN] Service is not running\n")
				fmt.Printf("   [INFO] Run 'ghost-backup service start' to start it\n")
				warnings = append(warnings, "Service not running")
			}
		}
	}

	// Check 6: Gitleaks (optional)
	fmt.Printf("\n[*] Checking gitleaks availability...\n")
	if security.IsGitleaksAvailable() {
		fmt.Printf("   [PASS] gitleaks is installed and available\n")
	} else {
		fmt.Printf("   [WARN] gitleaks not found in PATH\n")
		fmt.Printf("   [INFO] Install from: https://github.com/gitleaks/gitleaks\n")
		fmt.Printf("   Note: Secret scanning will be disabled without gitleaks\n")
		warnings = append(warnings, "gitleaks not available for secret scanning")
	}

	// Summary
	fmt.Printf("\n%s\n", strings.Repeat("═", 50))
	fmt.Printf("SUMMARY\n")
	fmt.Printf("%s\n", strings.Repeat("═", 50))

	if hasErrors {
		fmt.Printf("[FAIL] Configuration has ERRORS that need to be fixed\n")
	} else if len(warnings) > 0 {
		fmt.Printf("[WARN] Configuration has warnings\n")
	} else {
		fmt.Printf("[PASS] Configuration is valid and ready to use\n")
	}

	if len(warnings) > 0 {
		fmt.Printf("\nWarnings (%d):\n", len(warnings))
		for i, warning := range warnings {
			fmt.Printf("   %d. %s\n", i+1, warning)
		}
	}

	if hasErrors {
		fmt.Printf("\n[INFO] Fix the errors above and run 'ghost-backup check' again\n")
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
	status, err := service.GetServiceStatus()
	if err != nil {
		return 0, err
	}
	return int(status), nil
}
