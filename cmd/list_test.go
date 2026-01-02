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

func TestListCmd_UserFlag(t *testing.T) {
	// list command should have a hidden --user flag
	userFlag := listCmd.Flags().Lookup("user")
	if userFlag == nil {
		t.Fatal("list command should have a --user flag")
	}

	if userFlag.Hidden != true {
		t.Error("--user flag should be hidden")
	}

	if userFlag.DefValue != "" {
		t.Errorf("--user flag default value should be empty, got %s", userFlag.DefValue)
	}
}
