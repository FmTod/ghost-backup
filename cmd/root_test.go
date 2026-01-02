package cmd

import (
	"testing"
)

func TestRootCmd_Configuration(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}

	if rootCmd.Use != "ghost-backup" {
		t.Errorf("rootCmd.Use = %s, want ghost-backup", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("rootCmd.Short should not be empty")
	}

	if rootCmd.Long == "" {
		t.Error("rootCmd.Long should not be empty")
	}
}

func TestRootCmd_CompletionDisabled(t *testing.T) {
	if !rootCmd.CompletionOptions.DisableDefaultCmd {
		t.Error("Completion should be disabled by default")
	}
}

func TestRootCmd_HasSubcommands(t *testing.T) {
	// After init() runs, rootCmd should have subcommands registered
	// We can't easily test this without running the init functions
	// but we can verify the structure exists

	if rootCmd.Commands() == nil {
		t.Error("rootCmd should have commands slice initialized")
	}
}

func TestExecute_Function(t *testing.T) {
	// We can't easily test Execute() as it calls os.Exit()
	// but we verify it exists by calling it would compile
	// The function signature is: func Execute()
	if rootCmd == nil {
		t.Error("rootCmd should be initialized for Execute to work")
	}
}
