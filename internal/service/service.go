package service

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/neoscode/ghost-backup/internal/config"
	"github.com/neoscode/ghost-backup/internal/worker"
)

// Program implements the service.Interface
type Program struct {
	logger  service.Logger
	manager *worker.Manager
	logFile *os.File
}

// Start is called when the service starts
func (p *Program) Start(service.Service) error {
	_ = p.logger.Info("Starting ghost-backup service...")

	// Setup file logging
	logPath, err := getLogFilePath()
	if err != nil {
		return fmt.Errorf("failed to get log file path: %w", err)
	}

	// Ensure log directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open the log file
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	p.logFile = logFile

	// Create a logger that writes to a file
	fileLogger := log.New(logFile, "", log.LstdFlags)

	// Create a worker manager
	p.manager = worker.NewManager(fileLogger)

	// Load registry
	registry, err := config.LoadRegistry()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Start workers
	if err := p.manager.StartWorkers(registry); err != nil {
		return fmt.Errorf("failed to start workers: %w", err)
	}

	_ = p.logger.Infof("Service started with %d workers", p.manager.GetWorkerCount())

	go p.run()
	return nil
}

// run is the main service loop
func (p *Program) run() {
	// The workers run in their own goroutines
	// This method just needs to keep running
	select {}
}

// Stop is called when the service stops
func (p *Program) Stop(service.Service) error {
	_ = p.logger.Info("Stopping ghost-backup service...")

	if p.manager != nil {
		p.manager.StopWorkers()
	}

	if p.logFile != nil {
		_ = p.logFile.Close()
	}

	_ = p.logger.Info("Service stopped")
	return nil
}

// getLogFilePath returns the path to the log file
func getLogFilePath() (string, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "ghost-backup.log"), nil
}

// GetLogFilePath returns the path to the log file (exported for CLI)
func GetLogFilePath() (string, error) {
	return getLogFilePath()
}

// NewService creates a new system service
func NewService() (service.Service, error) {
	svcConfig := &service.Config{
		Name:        "ghost-backup",
		DisplayName: "Ghost Backup Service",
		Description: "Automated git backup service that monitors multiple repositories",
	}

	prg := &Program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	// Set logger
	logger, err := s.Logger(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	prg.logger = logger

	return s, nil
}

// InstallService installs the system service
func InstallService() error {
	s, err := NewService()
	if err != nil {
		return err
	}

	err = s.Install()
	if err != nil {
		return fmt.Errorf("failed to install service: %w", err)
	}

	return nil
}

// UninstallService uninstalls the system service
func UninstallService() error {
	s, err := NewService()
	if err != nil {
		return err
	}

	err = s.Uninstall()
	if err != nil {
		return fmt.Errorf("failed to uninstall service: %w", err)
	}

	return nil
}

// StartService starts the system service
func StartService() error {
	s, err := NewService()
	if err != nil {
		return err
	}

	err = s.Start()
	if err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}

// StopService stops the system service
func StopService() error {
	s, err := NewService()
	if err != nil {
		return err
	}

	err = s.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	return nil
}

// RestartService restarts the system service
func RestartService() error {
	s, err := NewService()
	if err != nil {
		return err
	}

	err = s.Restart()
	if err != nil {
		return fmt.Errorf("failed to restart service: %w", err)
	}

	return nil
}

// RunService runs the service (blocking)
func RunService() error {
	s, err := NewService()
	if err != nil {
		return err
	}

	err = s.Run()
	if err != nil {
		return fmt.Errorf("failed to run service: %w", err)
	}

	return nil
}

// GetServiceStatus returns the status of the service
func GetServiceStatus() (service.Status, error) {
	s, err := NewService()
	if err != nil {
		return service.StatusUnknown, err
	}

	status, err := s.Status()
	if err != nil {
		return service.StatusUnknown, fmt.Errorf("failed to get service status: %w", err)
	}

	return status, nil
}

// EnsureServiceRunning ensures the service is installed and running
func EnsureServiceRunning() error {
	status, err := GetServiceStatus()
	if err != nil {
		// Service might not be installed
		if err := InstallService(); err != nil {
			return fmt.Errorf("failed to install service: %w", err)
		}
		status = service.StatusStopped
	}

	if status != service.StatusRunning {
		if err := StartService(); err != nil {
			return fmt.Errorf("failed to start service: %w", err)
		}
	}

	return nil
}
