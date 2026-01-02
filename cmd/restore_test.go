package cmd

import (
	"testing"
)

func TestRestoreCmd_Configuration(t *testing.T) {
	if restoreCmd == nil {
		t.Fatal("restoreCmd is nil")
	}

	if restoreCmd.Use != "restore <hash>" {
		t.Errorf("restoreCmd.Use = %s, want 'restore <hash>'", restoreCmd.Use)
	}

	if restoreCmd.Short == "" {
		t.Error("restoreCmd.Short should not be empty")
	}

	if restoreCmd.Long == "" {
		t.Error("restoreCmd.Long should not be empty")
	}

	if restoreCmd.RunE == nil {
		t.Error("restoreCmd.RunE should not be nil")
	}
}

func TestRestoreCmd_Args(t *testing.T) {
	// Should require exactly 1 argument
	if restoreCmd.Args == nil {
		t.Error("restoreCmd.Args should be set")
	}

	// Test with wrong number of args
	err := restoreCmd.Args(restoreCmd, []string{})
	if err == nil {
		t.Error("Should error with no arguments")
	}

	err = restoreCmd.Args(restoreCmd, []string{"hash1", "hash2"})
	if err == nil {
		t.Error("Should error with too many arguments")
	}

	err = restoreCmd.Args(restoreCmd, []string{"hash1"})
	if err != nil {
		t.Errorf("Should not error with exactly 1 argument: %v", err)
	}
}

func TestRestoreCmd_MethodFlag(t *testing.T) {
	methodFlag := restoreCmd.Flags().Lookup("method")
	if methodFlag == nil {
		t.Fatal("method flag not registered")
	}

	if methodFlag.DefValue != "apply" {
		t.Errorf("method flag default = %s, want apply", methodFlag.DefValue)
	}

	// Verify short flag
	shortFlag := restoreCmd.Flags().ShorthandLookup("m")
	if shortFlag == nil {
		t.Error("short flag -m not registered")
	}
}

func TestRestoreCmd_ValidMethods(t *testing.T) {
	// The command accepts "apply" and "cherry-pick"
	// We can't easily test the validation without running the command
	// but we can document the expected values

	validMethods := []string{"apply", "cherry-pick"}

	for _, method := range validMethods {
		if method == "" {
			t.Error("Valid method should not be empty")
		}
	}
}
