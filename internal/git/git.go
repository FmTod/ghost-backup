package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GitRepo represents a git repository
//
//goland:noinspection GoNameStartsWithPackageName
type GitRepo struct {
	Path string
}

// NewGitRepo creates a new GitRepo instance
func NewGitRepo(path string) *GitRepo {
	return &GitRepo{Path: path}
}

// IsGitRepo checks if the path is a valid git repository
func (g *GitRepo) IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = g.Path
	err := cmd.Run()
	return err == nil
}

// GetUserEmail retrieves the user's git email
func (g *GitRepo) GetUserEmail() (string, error) {
	cmd := exec.Command("git", "config", "user.email")
	cmd.Dir = g.Path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get user email: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetCurrentBranch retrieves the current branch name
func (g *GitRepo) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = g.Path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// HasChanges checks if there are any uncommitted changes
func (g *GitRepo) HasChanges() (bool, error) {
	// Check both staged and unstaged changes
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = g.Path
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check for changes: %w", err)
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// CreateStash creates a stash and returns the hash
func (g *GitRepo) CreateStash() (string, error) {
	// Use 'git stash create' which creates a stash without modifying the working directory
	cmd := exec.Command("git", "stash", "create")
	cmd.Dir = g.Path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to create stash: %w", err)
	}

	hash := strings.TrimSpace(string(output))
	if hash == "" {
		return "", fmt.Errorf("no changes to stash")
	}

	return hash, nil
}

// GetDiff returns the diff for a specific commit/stash hash
func (g *GitRepo) GetDiff(hash string) (string, error) {
	cmd := exec.Command("git", "show", hash, "-p")
	cmd.Dir = g.Path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}
	return string(output), nil
}

// PushToBackupRef pushes a hash to a backup reference
func (g *GitRepo) PushToBackupRef(hash, userEmail, branch, remote string) error {
	// Create ref name: refs/backups/<user_email>/<branch_name>
	refName := fmt.Sprintf("refs/backups/%s/%s", SanitizeRefName(userEmail), SanitizeRefName(branch))

	// Push the hash to the remote ref
	// Format: git push <remote> <hash>:<ref>
	cmd := exec.Command("git", "push", remote, fmt.Sprintf("%s:%s", hash, refName))
	cmd.Dir = g.Path

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push to backup ref: %w, stderr: %s", err, stderr.String())
	}

	return nil
}

// ListBackupRefs lists all backup references for the current user and branch
func (g *GitRepo) ListBackupRefs(remote, userEmail, branch string) ([]BackupRef, error) {
	refPattern := fmt.Sprintf("refs/backups/%s/%s", SanitizeRefName(userEmail), SanitizeRefName(branch))

	// Fetch refs from remote
	cmd := exec.Command("git", "ls-remote", remote, refPattern)
	cmd.Dir = g.Path
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list backup refs: %w", err)
	}

	var refs []BackupRef
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			refs = append(refs, BackupRef{
				Hash: parts[0],
				Ref:  parts[1],
			})
		}
	}

	return refs, nil
}

// FetchBackupRef fetches a specific backup reference
func (g *GitRepo) FetchBackupRef(remote, refName string) error {
	cmd := exec.Command("git", "fetch", remote, fmt.Sprintf("%s:%s", refName, refName))
	cmd.Dir = g.Path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch backup ref: %w", err)
	}
	return nil
}

// ApplyStash applies a stash by hash
func (g *GitRepo) ApplyStash(hash string) error {
	cmd := exec.Command("git", "stash", "apply", hash)
	cmd.Dir = g.Path

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply stash: %w, stderr: %s", err, stderr.String())
	}

	return nil
}

// CherryPick applies a commit by hash
func (g *GitRepo) CherryPick(hash string) error {
	cmd := exec.Command("git", "cherry-pick", "--no-commit", hash)
	cmd.Dir = g.Path

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to cherry-pick: %w, stderr: %s", err, stderr.String())
	}

	return nil
}

// GetRemote returns the default remote (typically "origin")
func (g *GitRepo) GetRemote() (string, error) {
	cmd := exec.Command("git", "remote")
	cmd.Dir = g.Path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote: %w", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return "", fmt.Errorf("no remotes configured")
	}

	remotes := strings.Split(outputStr, "\n")

	// Prefer "origin" if available
	for _, remote := range remotes {
		if remote == "origin" {
			return remote, nil
		}
	}

	return remotes[0], nil
}

// BackupRef represents a backup reference
type BackupRef struct {
	Hash string
	Ref  string
}

// SanitizeRefName sanitizes a string to be used in a git ref name
func SanitizeRefName(s string) string {
	// Replace characters that are not allowed in git ref names
	replacer := strings.NewReplacer(
		"@", "_at_",
		" ", "_",
		":", "_",
		"/", "_",
		"\\", "_",
		"^", "_",
		"~", "_",
		"?", "_",
		"*", "_",
		"[", "_",
	)
	return replacer.Replace(s)
}
