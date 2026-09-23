package detection

import (
	"fmt"
	"sort"
	"strings"
)

func Prometheus(report Report) string {
	var output strings.Builder
	output.WriteString("# HELP insider_events_total Total behavioral events analyzed.\n")
	output.WriteString("# TYPE insider_events_total gauge\n")
	fmt.Fprintf(&output, "insider_events_total %d\n", report.Events)
	output.WriteString("# HELP insider_findings_total Total behavioral findings.\n")
	output.WriteString("# TYPE insider_findings_total gauge\n")
	fmt.Fprintf(&output, "insider_findings_total %d\n", len(report.Findings))

	severityCounts := map[string]int{}
	for _, finding := range report.Findings {
		severityCounts[finding.Severity]++
	}
	severities := make([]string, 0, len(severityCounts))
	for severity := range severityCounts {
		severities = append(severities, severity)
	}
	sort.Strings(severities)
	output.WriteString("# HELP insider_findings_by_severity Findings grouped by severity.\n")
	output.WriteString("# TYPE insider_findings_by_severity gauge\n")
	for _, severity := range severities {
		fmt.Fprintf(&output, "insider_findings_by_severity{severity=\"%s\"} %d\n", label(severity), severityCounts[severity])
	}

	output.WriteString("# HELP insider_user_risk_score Accumulated behavioral risk score by user.\n")
	output.WriteString("# TYPE insider_user_risk_score gauge\n")
	for _, user := range report.Users {
		fmt.Fprintf(&output, "insider_user_risk_score{user_id=\"%s\",severity=\"%s\"} %d\n", label(user.UserID), label(user.Severity), user.TotalScore)
	}
	return output.String()
}

func label(value string) string {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`).Replace(value)
}
