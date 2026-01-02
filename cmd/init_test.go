package cmd

import (
	"testing"
)

func TestInitCmd_Configuration(t *testing.T) {
	if initCmd == nil {
		t.Fatal("initCmd is nil")
	}

	if initCmd.Use != "init" {
		t.Errorf("initCmd.Use = %s, want init", initCmd.Use)
	}

	if initCmd.Short == "" {
		t.Error("initCmd.Short should not be empty")
	}

	if initCmd.Long == "" {
		t.Error("initCmd.Long should not be empty")
	}

	if initCmd.RunE == nil {
		t.Error("initCmd.RunE should not be nil")
	}
}

func TestInitCmd_Flags(t *testing.T) {
	pathFlag := initCmd.Flags().Lookup("path")
	if pathFlag == nil {
		t.Error("path flag not registered")
	}

	intervalFlag := initCmd.Flags().Lookup("interval")
	if intervalFlag == nil {
		t.Error("interval flag not registered")
	}

	scanSecretsFlag := initCmd.Flags().Lookup("scan-secrets")
	if scanSecretsFlag == nil {
		t.Error("scan-secrets flag not registered")
	}
}
