package cmd

import (
	"testing"
)

func TestUninstallCmd_Configuration(t *testing.T) {
	if uninstallCmd == nil {
		t.Fatal("uninstallCmd is nil")
	}

	if uninstallCmd.Use != "uninstall" {
		t.Errorf("uninstallCmd.Use = %s, want uninstall", uninstallCmd.Use)
	}

	if uninstallCmd.Short == "" {
		t.Error("uninstallCmd.Short should not be empty")
	}

	if uninstallCmd.Long == "" {
		t.Error("uninstallCmd.Long should not be empty")
	}

	if uninstallCmd.RunE == nil {
		t.Error("uninstallCmd.RunE should not be nil")
	}
}

func TestUninstallCmd_PathFlag(t *testing.T) {
	pathFlag := uninstallCmd.Flags().Lookup("path")
	if pathFlag == nil {
		t.Fatal("path flag not registered")
	}

	if pathFlag.DefValue != "." {
		t.Errorf("path flag default = %s, want .", pathFlag.DefValue)
	}

	shortFlag := uninstallCmd.Flags().ShorthandLookup("p")
	if shortFlag == nil {
		t.Error("short flag -p not registered")
	}
}
