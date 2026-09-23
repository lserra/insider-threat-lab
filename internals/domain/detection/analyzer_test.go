package detection

import (
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
