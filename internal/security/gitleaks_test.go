package security

import (
	"strings"
	"testing"
)

func TestScanDiff_NoGitleaks(t *testing.T) {
	diff := `diff --git a/test.txt b/test.txt
index 0000000..1111111 100644
--- a/test.txt
+++ b/test.txt
@@ -1 +1 @@
-old content
+new content
`

	result, err := ScanDiff(diff)

	if err != nil {
		if !strings.Contains(err.Error(), "gitleaks not found") {
			t.Logf("ScanDiff() returned error: %v (expected if gitleaks not installed)", err)
		}

		if result == nil {
			t.Error("ScanDiff() should return non-nil result even on error")
		} else if result.Error == nil {
			t.Error("ScanDiff() result.Error should be set when function returns error")
		}
		return
	}

	if result == nil {
		t.Fatal("ScanDiff() returned nil result")
	}

	if result.HasSecrets {
		t.Error("ScanDiff() found secrets in clean diff")
	}
}

func TestIsGitleaksAvailable(t *testing.T) {
	available := IsGitleaksAvailable()

	t.Logf("IsGitleaksAvailable() = %v", available)
}

func TestScanResult_Structure(t *testing.T) {
	result := &ScanResult{
		HasSecrets: true,
		Output:     "test output",
		Error:      nil,
	}

	if !result.HasSecrets {
		t.Error("HasSecrets should be true")
	}

	if result.Output != "test output" {
		t.Errorf("Output = %s, want 'test output'", result.Output)
	}

	if result.Error != nil {
		t.Errorf("Error should be nil, got %v", result.Error)
	}
}

func TestGitleaksTimeout(t *testing.T) {
	if GitleaksTimeout <= 0 {
		t.Error("GitleaksTimeout should be positive")
	}

	if GitleaksTimeout.Seconds() < 1 {
		t.Errorf("GitleaksTimeout = %v, should be at least 1 second", GitleaksTimeout)
	}
}
