package detection

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const LargeTransferThreshold int64 = 50 * 1024 * 1024

type TransferEvent struct {
	UserID      string `json:"user_id"`
	FileSize    int64  `json:"file_size"`
	Timestamp   string `json:"timestamp"`
	Destination string `json:"destination"`
}

type Finding struct {
	UserID      string   `json:"user_id"`
	Source      string   `json:"source,omitempty"`
	Action      string   `json:"action,omitempty"`
	Timestamp   string   `json:"timestamp"`
	FileSize    int64    `json:"file_size"`
	Destination string   `json:"destination"`
	Score       int      `json:"score"`
	Severity    string   `json:"severity"`
	Signals     []string `json:"signals"`
	Evidence    string   `json:"evidence,omitempty"`
}

type UserSummary struct {
	UserID       string `json:"user_id"`
	Events       int    `json:"events"`
	Bytes        int64  `json:"bytes"`
	Findings     int    `json:"findings"`
	TotalScore   int    `json:"total_score"`
	HighestScore int    `json:"highest_score"`
	Severity     string `json:"severity"`
}

type Report struct {
	Input    string        `json:"input"`
	Events   int           `json:"events"`
	Findings []Finding     `json:"findings"`
	Users    []UserSummary `json:"users"`
}

type genericEvent map[string]json.RawMessage

func AnalyzeDirectory(dir string) (Report, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return Report{}, fmt.Errorf("find JSON files: %w", err)
	}
	if len(paths) == 0 {
		return Report{}, fmt.Errorf("no JSON files found in %q", dir)
	}

	report := Report{Input: dir}
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return Report{}, fmt.Errorf("open %s: %w", path, err)
		}
		var events []genericEvent
		err = json.NewDecoder(file).Decode(&events)
		file.Close()
		if err != nil {
			return Report{}, fmt.Errorf("decode %s: %w", path, err)
		}
		for _, event := range events {
			report.Events++
			userID := stringValue(event, "user_id")
			report.Users = appendUserEvent(report.Users, userID)
			if finding, ok := detectGenericEvent(event, filepath.Base(path)); ok {
				report.Findings = append(report.Findings, finding)
			}
		}
	}
	return summarize(report), nil
}

func appendUserEvent(users []UserSummary, userID string) []UserSummary {
	for i := range users {
		if users[i].UserID == userID {
			users[i].Events++
			return users
		}
	}
	return append(users, UserSummary{UserID: userID, Events: 1})
}

func detectGenericEvent(event genericEvent, source string) (Finding, bool) {
	finding := Finding{
		UserID: userID(event), Source: source, Action: stringValue(event, "action"),
		Timestamp: stringValue(event, "timestamp"), FileSize: intValue(event, "file_size"),
		Destination: stringValue(event, "destination"),
	}
	addSignal := func(score int, signal, evidence string) {
		finding.Score += score
		finding.Signals = append(finding.Signals, signal)
		if finding.Evidence == "" {
			finding.Evidence = evidence
		}
	}
	if finding.FileSize >= LargeTransferThreshold {
		addSignal(3, "large_transfer", strconv.FormatInt(finding.FileSize, 10)+" bytes")
	}
	if finding.Destination == "external_drive" || finding.Destination == "cloud_storage" {
		addSignal(2, "external_destination", finding.Destination)
	}
	action := strings.ToLower(finding.Action)
	if action == "delete" || action == "grant" || action == "modify" {
		addSignal(2, "destructive_or_privileged_action", action)
	}
	if strings.Contains(source, "security_log_modification") {
		addSignal(4, "security_log_tampering", stringValue(event, "modifications"))
	}
	if strings.Contains(source, "mass_deletion") {
		addSignal(3, "mass_deletion", strconv.FormatInt(intValue(event, "files_deleted"), 10))
	}
	if strings.Contains(source, "hacking_tools") {
		addSignal(2, "hacking_tool_usage", stringValue(event, "tool_name"))
	}
	if strings.Contains(source, "unauthorized_software") {
		addSignal(2, "unauthorized_software", stringValue(event, "software_name"))
	}
	if parsed, err := time.Parse(time.RFC3339Nano, finding.Timestamp); err == nil && (parsed.Hour() < 6 || parsed.Hour() >= 20) {
		addSignal(1, "outside_business_hours", parsed.Format(time.RFC3339))
	}
	finding.Severity = SeverityForScore(finding.Score)
	return finding, finding.Score > 0
}

func SeverityForScore(score int) string {
	switch {
	case score >= 6:
		return "critical"
	case score >= 4:
		return "high"
	case score >= 2:
		return "medium"
	default:
		return "low"
	}
}

func userID(event genericEvent) string { return stringValue(event, "user_id") }

func stringValue(event genericEvent, key string) string {
	var value string
	_ = json.Unmarshal(event[key], &value)
	return value
}

func intValue(event genericEvent, key string) int64 {
	var value int64
	_ = json.Unmarshal(event[key], &value)
	return value
}

func Analyze(r io.Reader, input string) (Report, error) {
	var events []TransferEvent
	if err := json.NewDecoder(r).Decode(&events); err != nil {
		return Report{}, fmt.Errorf("decode transfer events: %w", err)
	}

	users := map[string]*UserSummary{}
	report := Report{Input: input, Events: len(events)}
	for _, event := range events {
		user := users[event.UserID]
		if user == nil {
			user = &UserSummary{UserID: event.UserID}
			users[event.UserID] = user
		}
		user.Events++
		user.Bytes += event.FileSize

		finding := Finding{
			UserID:      event.UserID,
			Timestamp:   event.Timestamp,
			FileSize:    event.FileSize,
			Destination: event.Destination,
		}
		if event.FileSize >= LargeTransferThreshold {
			finding.Score += 3
			finding.Signals = append(finding.Signals, "large_transfer")
		}
		if event.Destination == "external_drive" || event.Destination == "cloud_storage" {
			finding.Score += 2
			finding.Signals = append(finding.Signals, "external_destination")
		}
		if finding.Score > 0 {
			finding.Severity = SeverityForScore(finding.Score)
			report.Findings = append(report.Findings, finding)
			user.Findings++
			if finding.Score > user.HighestScore {
				user.HighestScore = finding.Score
			}
		}
	}

	for _, user := range users {
		report.Users = append(report.Users, *user)
	}
	return summarize(report), nil
}

func summarize(report Report) Report {
	users := make(map[string]*UserSummary, len(report.Users))
	for i := range report.Users {
		user := report.Users[i]
		users[user.UserID] = &user
	}
	for _, finding := range report.Findings {
		user := users[finding.UserID]
		if user == nil {
			user = &UserSummary{UserID: finding.UserID}
			users[finding.UserID] = user
		}
		user.Findings++
		user.TotalScore += finding.Score
		if finding.FileSize > 0 {
			user.Bytes += finding.FileSize
		}
		if finding.Score > user.HighestScore {
			user.HighestScore = finding.Score
			user.Severity = SeverityForScore(finding.Score)
		}
	}
	report.Users = nil
	for _, user := range users {
		report.Users = append(report.Users, *user)
	}
	sort.Slice(report.Users, func(i, j int) bool {
		if report.Users[i].TotalScore == report.Users[j].TotalScore {
			if report.Users[i].HighestScore == report.Users[j].HighestScore {
				return report.Users[i].UserID < report.Users[j].UserID
			}
			return report.Users[i].HighestScore > report.Users[j].HighestScore
		}
		return report.Users[i].TotalScore > report.Users[j].TotalScore
	})
	for i := range report.Users {
		report.Users[i].Severity = SeverityForScore(report.Users[i].TotalScore)
	}
	return report
}
