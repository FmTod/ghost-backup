package service

import (
	"strings"
	"testing"

	"github.com/kardianos/service"
)

func TestGetLogFilePath(t *testing.T) {
	path, err := GetLogFilePath()
	if err != nil {
		t.Fatalf("GetLogFilePath() error = %v", err)
	}

	if path == "" {
		t.Error("GetLogFilePath() returned empty string")
	}

	if !strings.HasSuffix(path, "ghost-backup.log") {
		t.Errorf("GetLogFilePath() should end with ghost-backup.log, got %s", path)
	}
}

func TestNewService(t *testing.T) {
	svc, err := NewService()
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if svc == nil {
		t.Fatal("NewService() returned nil service")
	}
}

func TestProgram_Structure(t *testing.T) {
	var _ service.Interface = (*Program)(nil)
}

func TestGetServiceStatus_NoService(t *testing.T) {
	status, err := GetServiceStatus()

	if err == nil {
		if status != service.StatusRunning &&
			status != service.StatusStopped &&
			status != service.StatusUnknown {
			t.Errorf("GetServiceStatus() returned invalid status: %v", status)
		}
	} else {
		t.Logf("GetServiceStatus() error (expected if service not installed): %v", err)

		if status != service.StatusUnknown {
			t.Errorf("GetServiceStatus() should return StatusUnknown on error, got %v", status)
		}
	}
}

func TestProgram_Structure_Initialization(t *testing.T) {
	prog := &Program{}

	if prog.logger != nil {
		t.Error("New Program should have nil logger before initialization")
	}

	if prog.manager != nil {
		t.Error("New Program should have nil manager before initialization")
	}

	if prog.logFile != nil {
		t.Error("New Program should have nil logFile before initialization")
	}
}
