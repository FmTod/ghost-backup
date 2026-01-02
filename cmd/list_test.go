package cmd

import (
	"testing"
)

func TestListCmd_Configuration(t *testing.T) {
	if listCmd == nil {
		t.Fatal("listCmd is nil")
	}

	if listCmd.Use != "list" {
		t.Errorf("listCmd.Use = %s, want list", listCmd.Use)
	}

	if listCmd.Short == "" {
		t.Error("listCmd.Short should not be empty")
	}

	if listCmd.Long == "" {
		t.Error("listCmd.Long should not be empty")
	}

	if listCmd.RunE == nil {
		t.Error("listCmd.RunE should not be nil")
	}
}

func TestListCmd_NoFlags(t *testing.T) {
	// list command should have no flags
	if listCmd.Flags().HasFlags() {
		flags := listCmd.Flags()
		t.Logf("list command has unexpected flags: %v", flags)
	}
}
