package cmd

import (
	"testing"
)

func TestBackupCmd_Configuration(t *testing.T) {
	if backupCmd.Use != "backup" {
		t.Errorf("Use = %s, want backup", backupCmd.Use)
	}

	if backupCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if backupCmd.Long == "" {
		t.Error("Long description is empty")
	}

	if backupCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestBackupCmd_PathFlag(t *testing.T) {
	flag := backupCmd.Flags().Lookup("path")
	if flag == nil {
		t.Fatal("path flag not found")
	}

	if flag.Shorthand != "p" {
		t.Errorf("path flag shorthand = %s, want p", flag.Shorthand)
	}

	if flag.DefValue != "." {
		t.Errorf("path flag default = %s, want .", flag.DefValue)
	}
}
