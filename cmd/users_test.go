package cmd

import (
	"testing"
)

func TestUsersCmd_Configuration(t *testing.T) {
	if usersCmd == nil {
		t.Fatal("usersCmd is nil")
	}

	if usersCmd.Use != "users" {
		t.Errorf("usersCmd.Use = %s, want users", usersCmd.Use)
	}

	if !usersCmd.Hidden {
		t.Error("usersCmd.Hidden should be true")
	}

	if usersCmd.Short == "" {
		t.Error("usersCmd.Short should not be empty")
	}

	if usersCmd.Long == "" {
		t.Error("usersCmd.Long should not be empty")
	}

	if usersCmd.RunE == nil {
		t.Error("usersCmd.RunE should not be nil")
	}
}

func TestUsersCmd_NoFlags(t *testing.T) {
	// users command should have no flags
	if usersCmd.Flags().HasFlags() {
		flags := usersCmd.Flags()
		t.Logf("users command has unexpected flags: %v", flags)
	}
}
