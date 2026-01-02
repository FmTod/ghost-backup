package worker

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/FmTod/ghost-backup/internal/config"
	"github.com/FmTod/ghost-backup/internal/git"
	"github.com/FmTod/ghost-backup/internal/security"
)

// Worker manages backup operations for a single repository
type Worker struct {
	repoPath    string
	stopCh      chan struct{}
	stoppedCh   chan struct{}
	logger      *log.Logger
	lastModTime time.Time
	ticker      *time.Ticker
	mu          sync.RWMutex
}

// NewWorker creates a new worker for a repository
func NewWorker(repoPath string, logger *log.Logger) *Worker {
	return &Worker{
		repoPath:  repoPath,
		stopCh:    make(chan struct{}),
		stoppedCh: make(chan struct{}),
		logger:    logger,
	}
}

// Start begins the worker's backup loop
func (w *Worker) Start() {
	defer close(w.stoppedCh)

	w.logger.Printf("[%s] Worker started\n", w.repoPath)

	// Initial backup
	w.performBackup()

	// Start ticker with the initial config
	cfg, err := w.loadConfig()
	if err != nil {
		w.logger.Printf("[%s] Failed to load config: %v\n", w.repoPath, err)
		return
	}

	w.updateTicker(time.Duration(cfg.Interval) * time.Second)

	for {
		// Get ticker channel under mutex protection.
		// Note: Even if the ticker is replaced after we read the channel,
		// the cached channel remains valid (it just stops receiving ticks).
		// The new ticker's channel will be picked up on the next iteration.
		w.mu.RLock()
		tickerCh := w.ticker.C
		w.mu.RUnlock()

		select {
		case <-w.stopCh:
			w.logger.Printf("[%s] Worker stopped\n", w.repoPath)
			return
		case <-tickerCh:
			w.performBackup()
			w.checkConfigReload()
		}
	}
}

// Stop signals the worker to stop
func (w *Worker) Stop() {
	select {
	case <-w.stopCh:
		// Already stopped
		return
	default:
		close(w.stopCh)
	}

	w.mu.Lock()
	if w.ticker != nil {
		w.ticker.Stop()
	}
	w.mu.Unlock()

	// Only wait for stoppedCh if the worker was actually started
	// Use a timeout to prevent hanging on workers that were never started
	select {
	case <-w.stoppedCh:
		// Worker stopped normally
	case <-time.After(100 * time.Millisecond):
		// Worker was never started, or already stopped
	}
}

// updateTicker updates the ticker with a new interval
func (w *Worker) updateTicker(interval time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.ticker != nil {
		w.ticker.Stop()
	}
	w.ticker = time.NewTicker(interval)
}

// checkConfigReload checks if the config file has been modified and reloads if necessary
func (w *Worker) checkConfigReload() {
	configPath := config.GetLocalConfigPath(w.repoPath)
	info, err := os.Stat(configPath)
	if err != nil {
		// Config file doesn't exist, use defaults
		return
	}

	w.mu.Lock()
	lastMod := w.lastModTime
	shouldReload := info.ModTime().After(lastMod)
	if shouldReload {
		w.lastModTime = info.ModTime()
	}
	w.mu.Unlock()

	if shouldReload {
		cfg, err := w.loadConfig()
		if err != nil {
			w.logger.Printf("[%s] Failed to reload config: %v\n", w.repoPath, err)
			return
		}

		w.logger.Printf("[%s] Config reloaded: interval=%ds, scan_secrets=%v\n",
			w.repoPath, cfg.Interval, cfg.ScanSecrets)
		w.updateTicker(time.Duration(cfg.Interval) * time.Second)
	}
}

// loadConfig loads the local configuration for this repository
func (w *Worker) loadConfig() (*config.LocalConfig, error) {
	cfg, err := config.LoadLocalConfig(w.repoPath)
	if err != nil {
		return nil, err
	}

	// Update last mod time
	configPath := config.GetLocalConfigPath(w.repoPath)
	if info, err := os.Stat(configPath); err == nil {
		w.mu.Lock()
		w.lastModTime = info.ModTime()
		w.mu.Unlock()
	}

	return cfg, nil
}

// performBackup executes the backup logic for this repository
func (w *Worker) performBackup() {
	w.logger.Printf("[%s] Starting backup...\n", w.repoPath)

	// Load config
	cfg, err := w.loadConfig()
	if err != nil {
		w.logger.Printf("[%s] Failed to load config: %v\n", w.repoPath, err)
		return
	}

	// Create git repo instance
	repo := git.NewGitRepo(w.repoPath)

	// Verify it's a git repo
	if !repo.IsGitRepo() {
		w.logger.Printf("[%s] Not a valid git repository\n", w.repoPath)
		return
	}

	// Check if there are changes
	hasChanges, err := repo.HasChanges()
	if err != nil {
		w.logger.Printf("[%s] Failed to check for changes: %v\n", w.repoPath, err)
		return
	}

	if !hasChanges {
		w.logger.Printf("[%s] No changes to backup\n", w.repoPath)
		return
	}

	// Create stash
	hash, err := repo.CreateStash()
	if err != nil {
		w.logger.Printf("[%s] Failed to create stash: %v\n", w.repoPath, err)
		return
	}

	w.logger.Printf("[%s] Created stash: %s\n", w.repoPath, hash)

	// If secret scanning is enabled, scan the diff
	if cfg.ScanSecrets {
		if !security.IsGitleaksAvailable() {
			w.logger.Printf("[%s] WARNING: gitleaks not available, skipping secret scan\n", w.repoPath)
		} else {
			diff, err := repo.GetDiff(hash)
			if err != nil {
				w.logger.Printf("[%s] Failed to get diff: %v\n", w.repoPath, err)
				return
			}

			result, err := security.ScanDiff(diff)
			if err != nil {
				w.logger.Printf("[%s] Secret scan failed: %v\n", w.repoPath, err)
				return
			}

			if result.HasSecrets {
				w.logger.Printf("[%s] SECRETS DETECTED! Aborting backup.\n", w.repoPath)
				w.logger.Printf("[%s] Gitleaks output:\n%s\n", w.repoPath, result.Output)
				return
			}

			w.logger.Printf("[%s] Secret scan passed\n", w.repoPath)
		}
	}

	// Get user email and branch
	userEmail, err := repo.GetUserEmail()
	if err != nil {
		w.logger.Printf("[%s] Failed to get user email: %v\n", w.repoPath, err)
		return
	}

	// Get user name for identifier generation
	userName, _ := repo.GetUserName()

	// Load global config to get git_user if configured
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		w.logger.Printf("[%s] Warning: Failed to load global config: %v\n", w.repoPath, err)
		globalConfig = &config.GlobalConfig{} // Use empty config
	}

	// Generate user identifier
	userIdentifier := git.GenerateUserIdentifier(globalConfig.GitUser, userName, userEmail)

	branch, err := repo.GetCurrentBranch()
	if err != nil {
		w.logger.Printf("[%s] Failed to get current branch: %v\n", w.repoPath, err)
		return
	}

	// Get remote
	remote, err := repo.GetRemote()
	if err != nil {
		w.logger.Printf("[%s] Failed to get remote: %v\n", w.repoPath, err)
		return
	}

	// Push to back up ref
	err = repo.PushToBackupRef(hash, userIdentifier, branch, remote)
	if err != nil {
		w.logger.Printf("[%s] Failed to push backup: %v\n", w.repoPath, err)
		return
	}

	w.logger.Printf("[%s] Backup completed successfully: refs/backups/%s/%s\n",
		w.repoPath, userIdentifier, branch)
}

// Manager manages multiple workers
type Manager struct {
	workers map[string]*Worker
	logger  *log.Logger
	mu      sync.Mutex
}

// NewManager creates a new worker manager
func NewManager(logger *log.Logger) *Manager {
	return &Manager{
		workers: make(map[string]*Worker),
		logger:  logger,
	}
}

// StartWorkers starts workers for all repositories in the registry
func (m *Manager) StartWorkers(registry *config.Registry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	repos := registry.GetRepositories()
	m.logger.Printf("Starting workers for %d repositories\n", len(repos))

	for _, repoPath := range repos {
		if _, exists := m.workers[repoPath]; exists {
			continue // Worker already running
		}

		// Verify the repository exists
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			m.logger.Printf("WARNING: Repository path does not exist: %s\n", repoPath)
			continue
		}

		worker := NewWorker(repoPath, m.logger)
		m.workers[repoPath] = worker
		go worker.Start()
	}

	return nil
}

// StopWorkers stops all running workers
func (m *Manager) StopWorkers() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Printf("Stopping all workers...\n")

	for repoPath, worker := range m.workers {
		m.logger.Printf("Stopping worker for %s\n", repoPath)
		worker.Stop()
	}

	m.workers = make(map[string]*Worker)
}

// ReloadWorkers stops current workers and starts new ones based on the registry
func (m *Manager) ReloadWorkers(registry *config.Registry) error {
	m.StopWorkers()
	return m.StartWorkers(registry)
}

// GetWorkerCount returns the number of active workers
func (m *Manager) GetWorkerCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.workers)
}

// GetWorkerStatus returns status information for all workers
func (m *Manager) GetWorkerStatus() []WorkerStatus {
	m.mu.Lock()
	defer m.mu.Unlock()

	var statuses []WorkerStatus
	for repoPath := range m.workers {
		statuses = append(statuses, WorkerStatus{
			RepoPath: repoPath,
			Running:  true,
		})
	}
	return statuses
}

// WorkerStatus represents the status of a worker
//
//goland:noinspection GoNameStartsWithPackageName
type WorkerStatus struct {
	RepoPath string
	Running  bool
}

// String returns a string representation of worker status
func (ws WorkerStatus) String() string {
	status := "stopped"
	if ws.Running {
		status = "running"
	}
	return fmt.Sprintf("%s: %s", ws.RepoPath, status)
}
