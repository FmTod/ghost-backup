package security

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	// GitleaksTimeout is the maximum time to wait for gitleaks to complete
	GitleaksTimeout = 60 * time.Second
)

// ScanResult represents the result of a gitleaks scan
type ScanResult struct {
	HasSecrets bool
	Output     string
	Error      error
}

// ScanDiff scans a git diff for secrets using gitleaks
// Returns true if secrets are found
func ScanDiff(diff string) (*ScanResult, error) {
	result := &ScanResult{}

	// Check if gitleaks is installed
	if !isGitleaksInstalled() {
		result.Error = fmt.Errorf("gitleaks not found in PATH")
		return result, result.Error
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), GitleaksTimeout)
	defer cancel()

	// Run gitleaks with --no-git flag to treat input as raw text
	// Use stdin to pass the diff
	cmd := exec.CommandContext(ctx, "gitleaks", "detect", "--no-git", "--verbose", "--redact")
	cmd.Stdin = strings.NewReader(diff)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// gitleaks returns exit code 1 if secrets are found
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			result.Error = fmt.Errorf("gitleaks scan timed out after %v", GitleaksTimeout)
			return result, result.Error
		}

		// Check if the error is due to secrets being found
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 1 {
				// Secrets found
				result.HasSecrets = true
				result.Output = stderr.String()
				return result, nil
			}
		}

		// Other error
		result.Error = fmt.Errorf("gitleaks scan failed: %w, stderr: %s", err, stderr.String())
		return result, result.Error
	}

	// No secrets found
	result.HasSecrets = false
	result.Output = stderr.String()
	return result, nil
}

// isGitleaksInstalled checks if gitleaks is available in PATH
func isGitleaksInstalled() bool {
	cmd := exec.Command("gitleaks", "version")
	err := cmd.Run()
	return err == nil
}

// IsGitleaksAvailable returns whether gitleaks is installed and available
func IsGitleaksAvailable() bool {
	return isGitleaksInstalled()
}
