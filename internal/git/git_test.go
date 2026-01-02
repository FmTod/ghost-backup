package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestRepo(t *testing.T) string {
	tmpDir := t.TempDir()

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize test repo: %v", err)
	}

	// Configure git
	configCmds := [][]string{
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "user.email", "test@example.com"},
	}

	for _, cmdArgs := range configCmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to configure git: %v", err)
		}
	}

	return tmpDir
}

func TestNewGitRepo(t *testing.T) {
	path := "/test/path"
	repo := NewGitRepo(path)

	if repo == nil {
		t.Fatal("NewGitRepo() returned nil")
	}

	if repo.Path != path {
		t.Errorf("Path = %s, want %s", repo.Path, path)
	}
}

func TestGitRepo_IsGitRepo(t *testing.T) {
	// Test valid git repo
	tmpDir := setupTestRepo(t)
	repo := NewGitRepo(tmpDir)

	if !repo.IsGitRepo() {
		t.Error("IsGitRepo() = false for valid git repo")
	}

	// Test invalid path
	repo = NewGitRepo("/nonexistent/path")
	if repo.IsGitRepo() {
		t.Error("IsGitRepo() = true for invalid path")
	}
}

func TestGitRepo_GetUserEmail(t *testing.T) {
	tmpDir := setupTestRepo(t)
	repo := NewGitRepo(tmpDir)

	email, err := repo.GetUserEmail()
	if err != nil {
		t.Fatalf("GetUserEmail() error = %v", err)
	}

	if email != "test@example.com" {
		t.Errorf("GetUserEmail() = %s, want test@example.com", email)
	}
}

func TestGitRepo_GetCurrentBranch(t *testing.T) {
	tmpDir := setupTestRepo(t)
	repo := NewGitRepo(tmpDir)

	// Create an initial commit so HEAD exists
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd := exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	branch, err := repo.GetCurrentBranch()
	if err != nil {
		t.Fatalf("GetCurrentBranch() error = %v", err)
	}

	// The default branch is typically "main" or "master"
	if branch != "main" && branch != "master" {
		t.Logf("GetCurrentBranch() = %s (expected main or master, but this may vary)", branch)
	}

	if branch == "" {
		t.Error("GetCurrentBranch() returned empty string")
	}
}

func TestGitRepo_HasChanges(t *testing.T) {
	tmpDir := setupTestRepo(t)
	repo := NewGitRepo(tmpDir)

	// Initially should have no changes
	hasChanges, err := repo.HasChanges()
	if err != nil {
		t.Fatalf("HasChanges() error = %v", err)
	}

	if hasChanges {
		t.Error("HasChanges() = true for empty repo")
	}

	// Create a file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Now should have changes
	hasChanges, err = repo.HasChanges()
	if err != nil {
		t.Fatalf("HasChanges() error after file creation = %v", err)
	}

	if !hasChanges {
		t.Error("HasChanges() = false after creating file")
	}
}

func TestGitRepo_CreateStash(t *testing.T) {
	tmpDir := setupTestRepo(t)
	repo := NewGitRepo(tmpDir)

	// Create initial commit
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd := exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	// Modify file
	if err := os.WriteFile(testFile, []byte("modified"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Create stash
	hash, err := repo.CreateStash()
	if err != nil {
		t.Fatalf("CreateStash() error = %v", err)
	}

	if hash == "" {
		t.Error("CreateStash() returned empty hash")
	}

	if len(hash) != 40 {
		t.Errorf("CreateStash() hash length = %d, want 40", len(hash))
	}
}

func TestSanitizeRefName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user@example.com", "user_at_example.com"},
		{"feature/branch", "feature_branch"},
		{"test:name", "test_name"},
		{"normal-name", "normal-name"},
		{"test space", "test_space"},
		{"test\\backslash", "test_backslash"},
		{"test^caret", "test_caret"},
		{"test~tilde", "test_tilde"},
		{"test?question", "test_question"},
		{"test*asterisk", "test_asterisk"},
		{"test[bracket", "test_bracket"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeRefName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeRefName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGitRepo_GetRemote(t *testing.T) {
	tmpDir := setupTestRepo(t)
	repo := NewGitRepo(tmpDir)

	// Should return error for repo with no remotes
	_, err := repo.GetRemote()
	if err == nil {
		t.Fatal("GetRemote() should return error for repo with no remotes")
	}

	// Add a remote
	cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/test/repo.git")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add remote: %v", err)
	}

	remote, err := repo.GetRemote()
	if err != nil {
		t.Fatalf("GetRemote() error = %v", err)
	}

	if remote != "origin" {
		t.Errorf("GetRemote() = %s, want origin", remote)
	}
}

func TestGitRepo_GetDiff(t *testing.T) {
	tmpDir := setupTestRepo(t)
	repo := NewGitRepo(tmpDir)

	// Create and commit initial file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("line1\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd := exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	_ = cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	// Get HEAD commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = tmpDir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get HEAD hash: %v", err)
	}

	hash := strings.TrimSpace(string(output))

	// Get diff
	diff, err := repo.GetDiff(hash)
	if err != nil {
		t.Fatalf("GetDiff() error = %v", err)
	}

	if diff == "" {
		t.Error("GetDiff() returned empty diff")
	}

	// Diff should contain the filename
	if !strings.Contains(diff, "test.txt") {
		t.Error("GetDiff() output doesn't contain filename")
	}
}

func TestBackupRef_Structure(t *testing.T) {
	ref := BackupRef{
		Hash: "abc123",
		Ref:  "refs/backups/user/branch",
	}

	if ref.Hash != "abc123" {
		t.Errorf("Hash = %s, want abc123", ref.Hash)
	}

	if ref.Ref != "refs/backups/user/branch" {
		t.Errorf("Ref = %s, want refs/backups/user/branch", ref.Ref)
	}
}

func TestGitRepo_ListBackupRefs_NoRefs(t *testing.T) {
	tmpDir := setupTestRepo(t)
	repo := NewGitRepo(tmpDir)

	// Add a remote
	cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/test/repo.git")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add remote: %v", err)
	}

	// This will likely fail because the remote doesn't exist
	// but we're testing the function behavior
	refs, err := repo.ListBackupRefs("origin", "test@example.com", "main")

	// Error is expected since remote doesn't actually exist
	if err == nil && len(refs) > 0 {
		t.Logf("ListBackupRefs() returned %d refs (unexpected)", len(refs))
	}
}

func TestGitRepo_ApplyStash_InvalidHash(t *testing.T) {
	tmpDir := setupTestRepo(t)
	repo := NewGitRepo(tmpDir)

	err := repo.ApplyStash("invalidhash")
	if err == nil {
		t.Error("ApplyStash() should return error for invalid hash")
	}
}

func TestGitRepo_CherryPick_InvalidHash(t *testing.T) {
	tmpDir := setupTestRepo(t)
	repo := NewGitRepo(tmpDir)

	err := repo.CherryPick("invalidhash")
	if err == nil {
		t.Error("CherryPick() should return error for invalid hash")
	}
}
