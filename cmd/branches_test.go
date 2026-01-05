package cmd

import (
	"testing"
)

func TestBranchesCmd_Configuration(t *testing.T) {
	if branchesCmd == nil {
		t.Fatal("branchesCmd is nil")
	}

	if branchesCmd.Use != "branches" {
		t.Errorf("branchesCmd.Use = %s, want branches", branchesCmd.Use)
	}

	if branchesCmd.Short == "" {
		t.Error("branchesCmd.Short should not be empty")
	}

	if branchesCmd.Long == "" {
		t.Error("branchesCmd.Long should not be empty")
	}

	if branchesCmd.RunE == nil {
		t.Error("branchesCmd.RunE should not be nil")
	}
}

func TestBranchesCmd_UserFlag(t *testing.T) {
	// branches command should have a hidden --user flag
	userFlag := branchesCmd.Flags().Lookup("user")
	if userFlag == nil {
		t.Fatal("branches command should have a --user flag")
	}

	if userFlag.Hidden != true {
		t.Error("--user flag should be hidden")
	}

	if userFlag.DefValue != "" {
		t.Errorf("--user flag default value should be empty, got %s", userFlag.DefValue)
	}
}
