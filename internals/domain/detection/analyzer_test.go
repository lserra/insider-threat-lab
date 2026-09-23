package detection

import (
	"os"
	"strings"
	"testing"
)

func TestAnalyzeFlagsLargeExternalTransfer(t *testing.T) {
	report, err := Analyze(strings.NewReader(`[{"user_id":"user1","file_size":60000000,"timestamp":"2024-10-01T03:00:00Z","destination":"external_drive"}]`), "fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Findings) != 1 || report.Findings[0].Score != 5 {
		t.Fatalf("expected one finding with score 5, got %#v", report.Findings)
	}
	if len(report.Findings[0].Signals) != 2 {
		t.Fatalf("expected two signals, got %#v", report.Findings[0].Signals)
	}
}

func TestAnalyzeDoesNotFlagSmallInternalTransfer(t *testing.T) {
	report, err := Analyze(strings.NewReader(`[{"user_id":"user1","file_size":1000,"timestamp":"2024-10-01T12:00:00Z","destination":"internal_server"}]`), "fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected no findings, got %#v", report.Findings)
	}
}

func TestAnalyzeDirectoryCorrelatesSecuritySignals(t *testing.T) {
	dir := t.TempDir()
	writeJSON(t, dir+"/security_log_modification_logs.json", `[{"user_id":"user9","modifications":"removed entries","timestamp":"2024-10-01T03:00:00Z"}]`)
	writeJSON(t, dir+"/large_file_transfer_logs.json", `[{"user_id":"user9","file_size":60000000,"timestamp":"2024-10-01T03:05:00Z","destination":"external_drive"}]`)

	report, err := AnalyzeDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	if report.Events != 2 || len(report.Findings) != 2 {
		t.Fatalf("expected two events and findings, got %#v", report)
	}
	if len(report.Users) != 1 || report.Users[0].UserID != "user9" || report.Users[0].TotalScore != 11 || report.Users[0].HighestScore != 6 {
		t.Fatalf("expected correlated high-risk user, got %#v", report.Users)
	}
}

func writeJSON(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
