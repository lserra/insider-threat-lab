package detection

import (
	"net/http/httptest"
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
	if len(report.Users) != 1 || report.Users[0].UserID != "user9" || report.Users[0].TotalScore != 11 || report.Users[0].HighestScore != 6 || report.Users[0].Severity != "critical" {
		t.Fatalf("expected correlated high-risk user, got %#v", report.Users)
	}
}

func TestSeverityForScore(t *testing.T) {
	cases := map[int]string{1: "low", 2: "medium", 4: "high", 6: "critical"}
	for score, expected := range cases {
		if got := SeverityForScore(score); got != expected {
			t.Errorf("score %d: expected %s, got %s", score, expected, got)
		}
	}
}

func TestPrometheusIncludesRiskMetrics(t *testing.T) {
	report := Report{
		Events:   2,
		Findings: []Finding{{Severity: "critical"}, {Severity: "high"}},
		Users:    []UserSummary{{UserID: "user\"1", Severity: "critical", TotalScore: 8}},
	}
	metrics := Prometheus(report)
	for _, expected := range []string{
		"insider_events_total 2",
		`insider_findings_by_severity{severity="critical"} 1`,
		`insider_user_risk_score{user_id="user\"1",severity="critical"} 8`,
	} {
		if !strings.Contains(metrics, expected) {
			t.Errorf("metrics missing %q:\n%s", expected, metrics)
		}
	}
}

func TestHandlerExposesMetricsAndReport(t *testing.T) {
	report := Report{Events: 1}
	handler := Handler(report)

	metrics := httptest.NewRecorder()
	handler.ServeHTTP(metrics, httptest.NewRequest("GET", "/metrics", nil))
	if metrics.Code != 200 || !strings.Contains(metrics.Body.String(), "insider_events_total 1") {
		t.Fatalf("unexpected metrics response: %d %s", metrics.Code, metrics.Body.String())
	}

	reportResponse := httptest.NewRecorder()
	handler.ServeHTTP(reportResponse, httptest.NewRequest("GET", "/report", nil))
	if reportResponse.Code != 200 || !strings.Contains(reportResponse.Body.String(), `"events":1`) {
		t.Fatalf("unexpected report response: %d %s", reportResponse.Code, reportResponse.Body.String())
	}
}

func writeJSON(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
