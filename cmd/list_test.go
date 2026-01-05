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

func TestListCmd_BranchFlag(t *testing.T) {
	// list command should have a visible --branch flag
	branchFlag := listCmd.Flags().Lookup("branch")
	if branchFlag == nil {
		t.Fatal("list command should have a --branch flag")
	}

	if branchFlag.Hidden {
		t.Error("--branch flag should be visible (not hidden)")
	}

	if branchFlag.DefValue != "" {
		t.Errorf("--branch flag default value should be empty, got %s", branchFlag.DefValue)
	}
}

func TestListCmd_AllFlag(t *testing.T) {
	// list command should have a hidden --all flag
	allFlag := listCmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Fatal("list command should have an --all flag")
	}

	if allFlag.Hidden != true {
		t.Error("--all flag should be hidden")
	}

	if allFlag.DefValue != "false" {
		t.Errorf("--all flag default value should be false, got %s", allFlag.DefValue)
	}
}
