package git

import (
"os"
"os/exec"
"path/filepath"
"testing"
)

// setupWorktreeTestRepo creates a main repository and a worktree for testing
// Returns the paths to the main repo and the worktree
func setupWorktreeTestRepo(t *testing.T) (string, string) {
// Create main repository
mainRepo := t.TempDir()

cmd := exec.Command("git", "init")
cmd.Dir = mainRepo
if err := cmd.Run(); err != nil {
t.Fatalf("Failed to initialize main repo: %v", err)
}

// Configure git
configCmds := [][]string{
{"git", "config", "user.name", "Test User"},
{"git", "config", "user.email", "test@example.com"},
}

for _, cmdArgs := range configCmds {
cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
cmd.Dir = mainRepo
if err := cmd.Run(); err != nil {
t.Fatalf("Failed to configure git: %v", err)
}
}

// Create initial commit
testFile := filepath.Join(mainRepo, "test.txt")
if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
t.Fatalf("Failed to create test file: %v", err)
}

cmd = exec.Command("git", "add", ".")
cmd.Dir = mainRepo
if err := cmd.Run(); err != nil {
t.Fatalf("Failed to git add: %v", err)
}

cmd = exec.Command("git", "commit", "-m", "Initial commit")
cmd.Dir = mainRepo
if err := cmd.Run(); err != nil {
t.Fatalf("Failed to git commit: %v", err)
}

// Create worktree
worktreePath := filepath.Join(t.TempDir(), "worktree-branch")
cmd = exec.Command("git", "worktree", "add", worktreePath, "-b", "feature-branch")
cmd.Dir = mainRepo
if err := cmd.Run(); err != nil {
t.Fatalf("Failed to create worktree: %v", err)
}

return mainRepo, worktreePath
}

func TestGitRepo_WorktreeSupport(t *testing.T) {
_, worktreePath := setupWorktreeTestRepo(t)

// Test worktree with GitRepo
repo := NewGitRepo(worktreePath)

// Test IsGitRepo
if !repo.IsGitRepo() {
t.Error("IsGitRepo() = false for worktree, expected true")
}

// Test GetCurrentBranch
branch, err := repo.GetCurrentBranch()
if err != nil {
t.Errorf("GetCurrentBranch() error = %v", err)
}
if branch != "feature-branch" {
t.Errorf("GetCurrentBranch() = %s, want feature-branch", branch)
}

// Test GetUserEmail
email, err := repo.GetUserEmail()
if err != nil {
t.Errorf("GetUserEmail() error = %v", err)
}
if email != "test@example.com" {
t.Errorf("GetUserEmail() = %s, want test@example.com", email)
}

// Test HasChanges in worktree
hasChanges, err := repo.HasChanges()
if err != nil {
t.Errorf("HasChanges() error = %v", err)
}
if hasChanges {
t.Error("HasChanges() = true for clean worktree")
}

// Create a change in worktree
worktreeFile := filepath.Join(worktreePath, "worktree.txt")
if err := os.WriteFile(worktreeFile, []byte("worktree content"), 0644); err != nil {
t.Fatalf("Failed to create worktree file: %v", err)
}

// Test HasChanges with uncommitted changes
hasChanges, err = repo.HasChanges()
if err != nil {
t.Errorf("HasChanges() error after creating file = %v", err)
}
if !hasChanges {
t.Error("HasChanges() = false after creating file in worktree")
}
}

func TestGitRepo_WorktreeCreateStash(t *testing.T) {
_, worktreePath := setupWorktreeTestRepo(t)

// Modify file in worktree
worktreeFile := filepath.Join(worktreePath, "test.txt")
if err := os.WriteFile(worktreeFile, []byte("modified in worktree"), 0644); err != nil {
t.Fatalf("Failed to modify file in worktree: %v", err)
}

// Test CreateStash in worktree
repo := NewGitRepo(worktreePath)
hash, err := repo.CreateStash()
if err != nil {
t.Fatalf("CreateStash() error in worktree = %v", err)
}

if hash == "" {
t.Error("CreateStash() returned empty hash in worktree")
}

if len(hash) != 40 {
t.Errorf("CreateStash() hash length = %d, want 40", len(hash))
}

// Verify stash exists
if !repo.ObjectExists(hash) {
t.Error("Created stash object does not exist")
}
}

func TestGitRepo_WorktreeGetDiff(t *testing.T) {
_, worktreePath := setupWorktreeTestRepo(t)

// Create a commit in worktree
worktreeFile := filepath.Join(worktreePath, "worktree.txt")
if err := os.WriteFile(worktreeFile, []byte("worktree content"), 0644); err != nil {
t.Fatalf("Failed to create file in worktree: %v", err)
}

cmd := exec.Command("git", "add", ".")
cmd.Dir = worktreePath
if err := cmd.Run(); err != nil {
t.Fatalf("Failed to git add in worktree: %v", err)
}

cmd = exec.Command("git", "commit", "-m", "Worktree commit")
cmd.Dir = worktreePath
if err := cmd.Run(); err != nil {
t.Fatalf("Failed to git commit in worktree: %v", err)
}

// Get commit hash
cmd = exec.Command("git", "rev-parse", "HEAD")
cmd.Dir = worktreePath
output, err := cmd.Output()
if err != nil {
t.Fatalf("Failed to get HEAD hash: %v", err)
}
hash := string(output)[:40] // Get first 40 characters (full SHA)

// Test GetDiff in worktree
repo := NewGitRepo(worktreePath)
diff, err := repo.GetDiff(hash)
if err != nil {
t.Fatalf("GetDiff() error in worktree = %v", err)
}

if diff == "" {
t.Error("GetDiff() returned empty diff in worktree")
}
}

func TestGitRepo_WorktreeGetRemote(t *testing.T) {
mainRepo, worktreePath := setupWorktreeTestRepo(t)

// Add a remote to main repo
cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/test/repo.git")
cmd.Dir = mainRepo
if err := cmd.Run(); err != nil {
t.Fatalf("Failed to add remote: %v", err)
}

// Test GetRemote in worktree
repo := NewGitRepo(worktreePath)
remote, err := repo.GetRemote()
if err != nil {
t.Fatalf("GetRemote() error in worktree = %v", err)
}

if remote != "origin" {
t.Errorf("GetRemote() in worktree = %s, want origin", remote)
}
}

func TestGitRepo_WorktreeGetUserName(t *testing.T) {
_, worktreePath := setupWorktreeTestRepo(t)

// Test GetUserName in worktree
repo := NewGitRepo(worktreePath)
userName, err := repo.GetUserName()
if err != nil {
t.Fatalf("GetUserName() error in worktree = %v", err)
}

if userName != "Test User" {
t.Errorf("GetUserName() in worktree = %s, want Test User", userName)
}
}

func TestGitRepo_WorktreeObjectExists(t *testing.T) {
mainRepo, worktreePath := setupWorktreeTestRepo(t)

// Get commit hash from main repo
cmd := exec.Command("git", "rev-parse", "HEAD")
cmd.Dir = mainRepo
output, err := cmd.Output()
if err != nil {
t.Fatalf("Failed to get HEAD hash: %v", err)
}
commitHash := string(output)[:40]

// Test ObjectExists in worktree for commit from main repo
repo := NewGitRepo(worktreePath)
if !repo.ObjectExists(commitHash) {
t.Error("ObjectExists() = false for valid commit hash in worktree")
}

if repo.ObjectExists("0000000000000000000000000000000000000000") {
t.Error("ObjectExists() = true for invalid hash in worktree")
}
}
