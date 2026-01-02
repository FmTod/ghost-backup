package cmd

import (
	"strings"
	"testing"
)

func TestWorkflowCmd_Configuration(t *testing.T) {
	if workflowCmd == nil {
		t.Fatal("workflowCmd is nil")
	}

	if workflowCmd.Use != "workflow" {
		t.Errorf("workflowCmd.Use = %s, want workflow", workflowCmd.Use)
	}

	if workflowCmd.Short == "" {
		t.Error("workflowCmd.Short should not be empty")
	}

	if workflowCmd.Long == "" {
		t.Error("workflowCmd.Long should not be empty")
	}

	if workflowCmd.RunE == nil {
		t.Error("workflowCmd.RunE should not be nil")
	}
}

func TestWorkflowCmd_Flags(t *testing.T) {
	cronFlag := workflowCmd.Flags().Lookup("cron")
	if cronFlag == nil {
		t.Fatal("cron flag not registered")
	}

	if cronFlag.DefValue != "0 2 * * 0" {
		t.Errorf("cron flag default = %s, want '0 2 * * 0'", cronFlag.DefValue)
	}

	retentionFlag := workflowCmd.Flags().Lookup("retention")
	if retentionFlag == nil {
		t.Fatal("retention flag not registered")
	}

	if retentionFlag.DefValue != "30" {
		t.Errorf("retention flag default = %s, want 30", retentionFlag.DefValue)
	}

	// Verify short flags
	shortCronFlag := workflowCmd.Flags().ShorthandLookup("c")
	if shortCronFlag == nil {
		t.Error("short flag -c not registered")
	}

	shortRetentionFlag := workflowCmd.Flags().ShorthandLookup("r")
	if shortRetentionFlag == nil {
		t.Error("short flag -r not registered")
	}
}

func TestGenerateWorkflowYAML(t *testing.T) {
	cron := "0 2 * * 0"
	retention := 30

	yaml := generateWorkflowYAML(cron, retention)

	if yaml == "" {
		t.Fatal("generateWorkflowYAML returned empty string")
	}

	// Verify key components are present
	requiredStrings := []string{
		"name: Prune Ghost Backup Refs",
		"on:",
		"schedule:",
		cron,
		"workflow_dispatch:",
		"jobs:",
		"prune-backups:",
		"runs-on: ubuntu-latest",
		"steps:",
		"Checkout repository",
		"Configure Git",
		"Prune old backup refs",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(yaml, required) {
			t.Errorf("generateWorkflowYAML() output missing required string: %s", required)
		}
	}

	// Verify retention days appears in the YAML
	if !strings.Contains(yaml, "30") {
		t.Error("generateWorkflowYAML() should include retention days value")
	}
}

func TestGenerateWorkflowYAML_CustomValues(t *testing.T) {
	cron := "0 0 1 * *"
	retention := 60

	yaml := generateWorkflowYAML(cron, retention)

	if !strings.Contains(yaml, cron) {
		t.Errorf("generateWorkflowYAML() should include cron schedule %s", cron)
	}

	if !strings.Contains(yaml, "60") {
		t.Error("generateWorkflowYAML() should include custom retention value 60")
	}
}

func TestDescribeCron(t *testing.T) {
	tests := []struct {
		cron     string
		expected string
	}{
		{"0 2 * * 0", "Weekly at 2am on Sunday"},
		{"0 2 * * *", "Daily at 2am"},
		{"0 */6 * * *", "Every 6 hours"},
		{"0 0 1 * *", "Monthly on the 1st at midnight"},
		{"0 2 * * 1", "Weekly at 2am on Monday"},
		{"* * * * *", "Custom schedule"},
		{"0 3 * * 5", "Custom schedule"},
	}

	for _, tt := range tests {
		t.Run(tt.cron, func(t *testing.T) {
			result := describeCron(tt.cron)
			if result != tt.expected {
				t.Errorf("describeCron(%q) = %q, want %q", tt.cron, result, tt.expected)
			}
		})
	}
}

func TestGenerateWorkflowYAML_ValidYAML(t *testing.T) {
	yaml := generateWorkflowYAML("0 2 * * 0", 30)

	// Basic YAML structure validation
	lines := strings.Split(yaml, "\n")
	if len(lines) < 10 {
		t.Error("generateWorkflowYAML() should produce multi-line YAML")
	}

	// Should not have tabs (YAML uses spaces)
	if strings.Contains(yaml, "\t") {
		t.Error("generateWorkflowYAML() should not contain tabs")
	}

	// Should start with name
	if !strings.HasPrefix(strings.TrimSpace(yaml), "name:") {
		t.Error("generateWorkflowYAML() should start with 'name:'")
	}
}

func TestGenerateWorkflowYAML_GitOperations(t *testing.T) {
	yaml := generateWorkflowYAML("0 2 * * 0", 30)

	// Verify git operations are included
	gitOperations := []string{
		"git config",
		"git fetch",
		"git for-each-ref",
		"git update-ref",
		"git push",
	}

	for _, op := range gitOperations {
		if !strings.Contains(yaml, op) {
			t.Errorf("generateWorkflowYAML() should include git operation: %s", op)
		}
	}
}

func TestGenerateWorkflowYAML_BackupReferences(t *testing.T) {
	yaml := generateWorkflowYAML("0 2 * * 0", 30)

	// Verify backup ref handling
	if !strings.Contains(yaml, "refs/backups") {
		t.Error("generateWorkflowYAML() should reference refs/backups")
	}

	if !strings.Contains(yaml, "RETENTION_DAYS") {
		t.Error("generateWorkflowYAML() should use RETENTION_DAYS variable")
	}
}

func TestGenerateWorkflowYAML_Summary(t *testing.T) {
	yaml := generateWorkflowYAML("0 2 * * 0", 30)

	// Verify GitHub Actions summary is included
	if !strings.Contains(yaml, "GITHUB_STEP_SUMMARY") {
		t.Error("generateWorkflowYAML() should include step summary")
	}

	if !strings.Contains(yaml, "Ghost Backup Cleanup Complete") {
		t.Error("generateWorkflowYAML() should include completion message")
	}
}
