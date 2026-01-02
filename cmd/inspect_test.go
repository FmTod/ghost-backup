package cmd

import (
	"testing"
)

func TestInspectCmd_Configuration(t *testing.T) {
	if inspectCmd == nil {
		t.Fatal("inspectCmd is nil")
	}

	if inspectCmd.Use != "inspect <hash>" {
		t.Errorf("inspectCmd.Use = %s, want 'inspect <hash>'", inspectCmd.Use)
	}

	if inspectCmd.Short == "" {
		t.Error("inspectCmd.Short should not be empty")
	}

	if inspectCmd.Long == "" {
		t.Error("inspectCmd.Long should not be empty")
	}

	if inspectCmd.RunE == nil {
		t.Error("inspectCmd.RunE should not be nil")
	}
}

func TestInspectCmd_DiffFlag(t *testing.T) {
	// inspect command should have a --diff flag
	diffFlag := inspectCmd.Flags().Lookup("diff")
	if diffFlag == nil {
		t.Fatal("inspect command should have a --diff flag")
	}

	if diffFlag.Shorthand != "d" {
		t.Errorf("--diff flag shorthand should be 'd', got %s", diffFlag.Shorthand)
	}

	if diffFlag.DefValue != "false" {
		t.Errorf("--diff flag default value should be 'false', got %s", diffFlag.DefValue)
	}
}

func TestInspectCmd_UserFlag(t *testing.T) {
	// inspect command should have a hidden --user flag
	userFlag := inspectCmd.Flags().Lookup("user")
	if userFlag == nil {
		t.Fatal("inspect command should have a --user flag")
	}

	if userFlag.Hidden != true {
		t.Error("--user flag should be hidden")
	}

	if userFlag.DefValue != "" {
		t.Errorf("--user flag default value should be empty, got %s", userFlag.DefValue)
	}
}
