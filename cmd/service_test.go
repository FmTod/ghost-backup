package cmd

import (
	"testing"

	"github.com/kardianos/service"
)

func TestServiceCmd_Configuration(t *testing.T) {
	if serviceCmd == nil {
		t.Fatal("serviceCmd is nil")
	}

	if serviceCmd.Use != "service" {
		t.Errorf("serviceCmd.Use = %s, want service", serviceCmd.Use)
	}

	if serviceCmd.Short == "" {
		t.Error("serviceCmd.Short should not be empty")
	}

	if serviceCmd.Long == "" {
		t.Error("serviceCmd.Long should not be empty")
	}
}

func TestServiceCmd_Subcommands(t *testing.T) {
	// serviceCmd should have subcommands
	commands := serviceCmd.Commands()

	if len(commands) == 0 {
		t.Error("serviceCmd should have subcommands")
	}

	// Expected subcommands
	expectedCmds := []string{"install", "uninstall", "start", "stop", "restart", "status", "run"}

	foundCmds := make(map[string]bool)
	for _, cmd := range commands {
		foundCmds[cmd.Name()] = true
	}

	for _, expected := range expectedCmds {
		if !foundCmds[expected] {
			t.Errorf("Expected subcommand %s not found", expected)
		}
	}
}

func TestServiceInstallCmd_Configuration(t *testing.T) {
	if serviceInstallCmd == nil {
		t.Fatal("serviceInstallCmd is nil")
	}

	if serviceInstallCmd.Use != "install" {
		t.Errorf("serviceInstallCmd.Use = %s, want install", serviceInstallCmd.Use)
	}

	if serviceInstallCmd.RunE == nil {
		t.Error("serviceInstallCmd.RunE should not be nil")
	}
}

func TestServiceUninstallCmd_Configuration(t *testing.T) {
	if serviceUninstallCmd == nil {
		t.Fatal("serviceUninstallCmd is nil")
	}

	if serviceUninstallCmd.Use != "uninstall" {
		t.Errorf("serviceUninstallCmd.Use = %s, want uninstall", serviceUninstallCmd.Use)
	}

	if serviceUninstallCmd.RunE == nil {
		t.Error("serviceUninstallCmd.RunE should not be nil")
	}
}

func TestServiceStartCmd_Configuration(t *testing.T) {
	if serviceStartCmd == nil {
		t.Fatal("serviceStartCmd is nil")
	}

	if serviceStartCmd.Use != "start" {
		t.Errorf("serviceStartCmd.Use = %s, want start", serviceStartCmd.Use)
	}

	if serviceStartCmd.RunE == nil {
		t.Error("serviceStartCmd.RunE should not be nil")
	}
}

func TestServiceStopCmd_Configuration(t *testing.T) {
	if serviceStopCmd == nil {
		t.Fatal("serviceStopCmd is nil")
	}

	if serviceStopCmd.Use != "stop" {
		t.Errorf("serviceStopCmd.Use = %s, want stop", serviceStopCmd.Use)
	}

	if serviceStopCmd.RunE == nil {
		t.Error("serviceStopCmd.RunE should not be nil")
	}
}

func TestServiceRestartCmd_Configuration(t *testing.T) {
	if serviceRestartCmd == nil {
		t.Fatal("serviceRestartCmd is nil")
	}

	if serviceRestartCmd.Use != "restart" {
		t.Errorf("serviceRestartCmd.Use = %s, want restart", serviceRestartCmd.Use)
	}

	if serviceRestartCmd.RunE == nil {
		t.Error("serviceRestartCmd.RunE should not be nil")
	}
}

func TestServiceStatusCmd_Configuration(t *testing.T) {
	if serviceStatusCmd == nil {
		t.Fatal("serviceStatusCmd is nil")
	}

	if serviceStatusCmd.Use != "status" {
		t.Errorf("serviceStatusCmd.Use = %s, want status", serviceStatusCmd.Use)
	}

	if serviceStatusCmd.RunE == nil {
		t.Error("serviceStatusCmd.RunE should not be nil")
	}
}

func TestServiceRunCmd_Configuration(t *testing.T) {
	if serviceRunCmd == nil {
		t.Fatal("serviceRunCmd is nil")
	}

	if serviceRunCmd.Use != "run" {
		t.Errorf("serviceRunCmd.Use = %s, want run", serviceRunCmd.Use)
	}

	if serviceRunCmd.RunE == nil {
		t.Error("serviceRunCmd.RunE should not be nil")
	}
}

func TestGetStatusString(t *testing.T) {
	tests := []struct {
		name     string
		status   service.Status
		expected string
	}{
		{"Running", service.StatusRunning, "Running"},
		{"Stopped", service.StatusStopped, "Stopped"},
		{"Unknown", service.StatusUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStatusString(tt.status)
			if result != tt.expected {
				t.Errorf("getStatusString(%d) = %s, want %s", tt.status, result, tt.expected)
			}
		})
	}
}

func TestGetStatusString_InvalidStatus(t *testing.T) {
	result := getStatusString(service.Status(99))

	// Should return a formatted string for unknown status
	if result == "" {
		t.Error("getStatusString(99) should return non-empty string")
	}

	// Should contain the status number
	expected := "Status(99)"
	if result != expected {
		t.Errorf("getStatusString(99) = %s, want %s", result, expected)
	}
}
