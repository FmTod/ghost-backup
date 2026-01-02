package worker

import (
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/FmTod/ghost-backup/internal/config"
)

func TestNewWorker(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repoPath := "/test/repo"

	worker := NewWorker(repoPath, logger)

	if worker == nil {
		t.Fatal("NewWorker() returned nil")
	}

	if worker.repoPath != repoPath {
		t.Errorf("repoPath = %s, want %s", worker.repoPath, repoPath)
	}

	if worker.logger != logger {
		t.Error("logger not set correctly")
	}

	if worker.stopCh == nil {
		t.Error("stopCh not initialized")
	}

	if worker.stoppedCh == nil {
		t.Error("stoppedCh not initialized")
	}
}

func TestNewManager(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	manager := NewManager(logger)

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if manager.logger != logger {
		t.Error("logger not set correctly")
	}

	if manager.workers == nil {
		t.Error("workers map not initialized")
	}

	if len(manager.workers) != 0 {
		t.Errorf("workers map should be empty, got %d entries", len(manager.workers))
	}
}

func TestManager_GetWorkerCount(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	manager := NewManager(logger)

	count := manager.GetWorkerCount()
	if count != 0 {
		t.Errorf("GetWorkerCount() = %d, want 0", count)
	}

	// Add a worker manually (not started)
	manager.workers["/test/repo1"] = NewWorker("/test/repo1", logger)

	count = manager.GetWorkerCount()
	if count != 1 {
		t.Errorf("GetWorkerCount() = %d, want 1", count)
	}
}

func TestManager_StartWorkers_EmptyRegistry(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	manager := NewManager(logger)

	registry := &config.Registry{
		Repositories: []string{},
	}

	err := manager.StartWorkers(registry)
	if err != nil {
		t.Errorf("StartWorkers() error = %v", err)
	}

	if manager.GetWorkerCount() != 0 {
		t.Errorf("GetWorkerCount() = %d, want 0 for empty registry", manager.GetWorkerCount())
	}
}

func TestManager_StartWorkers_NonexistentPath(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	manager := NewManager(logger)

	registry := &config.Registry{
		Repositories: []string{"/nonexistent/path/to/repo"},
	}

	err := manager.StartWorkers(registry)
	if err != nil {
		t.Errorf("StartWorkers() error = %v (should not error on nonexistent paths)", err)
	}

	// Worker should not be started for nonexistent path
	if manager.GetWorkerCount() != 0 {
		t.Errorf("GetWorkerCount() = %d, want 0 for nonexistent paths", manager.GetWorkerCount())
	}
}

func TestManager_StopWorkers(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	manager := NewManager(logger)

	// Add some workers manually
	manager.workers["/test/repo1"] = NewWorker("/test/repo1", logger)
	manager.workers["/test/repo2"] = NewWorker("/test/repo2", logger)

	if manager.GetWorkerCount() != 2 {
		t.Fatalf("Setup failed: expected 2 workers, got %d", manager.GetWorkerCount())
	}

	manager.StopWorkers()

	if manager.GetWorkerCount() != 0 {
		t.Errorf("GetWorkerCount() = %d after StopWorkers(), want 0", manager.GetWorkerCount())
	}
}

func TestManager_GetWorkerStatus(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	manager := NewManager(logger)

	// Add workers
	manager.workers["/test/repo1"] = NewWorker("/test/repo1", logger)
	manager.workers["/test/repo2"] = NewWorker("/test/repo2", logger)

	statuses := manager.GetWorkerStatus()

	if len(statuses) != 2 {
		t.Errorf("GetWorkerStatus() returned %d statuses, want 2", len(statuses))
	}

	for _, status := range statuses {
		if !status.Running {
			t.Errorf("Worker status should be Running=true, got %v", status)
		}
	}
}

func TestWorkerStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   WorkerStatus
		contains string
	}{
		{
			name:     "running worker",
			status:   WorkerStatus{RepoPath: "/test/repo", Running: true},
			contains: "running",
		},
		{
			name:     "stopped worker",
			status:   WorkerStatus{RepoPath: "/test/repo", Running: false},
			contains: "stopped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := tt.status.String()
			if str == "" {
				t.Error("String() returned empty string")
			}

			if !contains(str, tt.contains) {
				t.Errorf("String() = %q, should contain %q", str, tt.contains)
			}

			if !contains(str, tt.status.RepoPath) {
				t.Errorf("String() = %q, should contain repo path %q", str, tt.status.RepoPath)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestWorker_LoadConfig_DefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.New(os.Stdout, "", log.LstdFlags)

	worker := NewWorker(tmpDir, logger)

	cfg, err := worker.loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if cfg.Interval != config.DefaultInterval {
		t.Errorf("Interval = %d, want %d", cfg.Interval, config.DefaultInterval)
	}

	if cfg.ScanSecrets != config.DefaultScanSecrets {
		t.Errorf("ScanSecrets = %v, want %v", cfg.ScanSecrets, config.DefaultScanSecrets)
	}
}

func TestWorker_LoadConfig_CustomConfig(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Create custom config
	customConfig := &config.LocalConfig{
		Interval:    120,
		ScanSecrets: false,
	}

	if err := config.SaveLocalConfig(tmpDir, customConfig); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	worker := NewWorker(tmpDir, logger)

	cfg, err := worker.loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if cfg.Interval != 120 {
		t.Errorf("Interval = %d, want 120", cfg.Interval)
	}

	if cfg.ScanSecrets != false {
		t.Errorf("ScanSecrets = %v, want false", cfg.ScanSecrets)
	}
}

func TestWorker_CheckConfigReload(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.New(os.Stdout, "", log.LstdFlags)

	worker := NewWorker(tmpDir, logger)

	// Create initial config
	initialConfig := &config.LocalConfig{
		Interval:    60,
		ScanSecrets: true,
	}

	if err := config.SaveLocalConfig(tmpDir, initialConfig); err != nil {
		t.Fatalf("Failed to save initial config: %v", err)
	}

	// Load config to set lastModTime
	_, err := worker.loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	// Wait a bit to ensure different timestamp
	time.Sleep(10 * time.Millisecond)

	// Modify config
	modifiedConfig := &config.LocalConfig{
		Interval:    90,
		ScanSecrets: false,
	}

	if err := config.SaveLocalConfig(tmpDir, modifiedConfig); err != nil {
		t.Fatalf("Failed to save modified config: %v", err)
	}

	// Check for reload
	worker.checkConfigReload()

	// Verify config was reloaded
	cfg, err := worker.loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() error after reload = %v", err)
	}

	if cfg.Interval != 90 {
		t.Errorf("Interval after reload = %d, want 90", cfg.Interval)
	}
}

func TestManager_ReloadWorkers(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	manager := NewManager(logger)

	// Start with initial registry
	initialRegistry := &config.Registry{
		Repositories: []string{},
	}

	err := manager.StartWorkers(initialRegistry)
	if err != nil {
		t.Fatalf("StartWorkers() error = %v", err)
	}

	// Reload with new registry
	newRegistry := &config.Registry{
		Repositories: []string{},
	}

	err = manager.ReloadWorkers(newRegistry)
	if err != nil {
		t.Errorf("ReloadWorkers() error = %v", err)
	}
}

func TestWorker_Stop_BeforeStart(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	worker := NewWorker("/test/repo", logger)

	// Stopping a worker that hasn't been started should not hang
	// This is a race condition test
	done := make(chan bool, 1)
	go func() {
		worker.Stop()
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("Stop() hung when called on unstarted worker")
	}
}

func TestManager_StartWorkers_ExistingRepo(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.New(os.Stdout, "", log.LstdFlags)
	manager := NewManager(logger)

	// Create a directory that exists
	repoDir := filepath.Join(tmpDir, "test-repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("Failed to create test repo dir: %v", err)
	}

	registry := &config.Registry{
		Repositories: []string{repoDir},
	}

	err := manager.StartWorkers(registry)
	if err != nil {
		t.Fatalf("StartWorkers() error = %v", err)
	}

	// Worker should be created for existing path
	if manager.GetWorkerCount() != 1 {
		t.Errorf("GetWorkerCount() = %d, want 1 for existing path", manager.GetWorkerCount())
	}

	// Clean up
	manager.StopWorkers()
}

func TestManager_ConcurrentAccess(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	manager := NewManager(logger)

	// Test concurrent GetWorkerCount calls
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = manager.GetWorkerCount()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
