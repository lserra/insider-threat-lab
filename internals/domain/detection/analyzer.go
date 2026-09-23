package detection

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
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
	Timestamp   string   `json:"timestamp"`
	FileSize    int64    `json:"file_size"`
	Destination string   `json:"destination"`
	Score       int      `json:"score"`
	Signals     []string `json:"signals"`
}

type UserSummary struct {
	UserID       string `json:"user_id"`
	Events       int    `json:"events"`
	Bytes        int64  `json:"bytes"`
	Findings     int    `json:"findings"`
	HighestScore int    `json:"highest_score"`
}

type Report struct {
	Input    string        `json:"input"`
	Events   int           `json:"events"`
	Findings []Finding     `json:"findings"`
	Users    []UserSummary `json:"users"`
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
	sort.Slice(report.Users, func(i, j int) bool {
		if report.Users[i].HighestScore == report.Users[j].HighestScore {
			return report.Users[i].UserID < report.Users[j].UserID
		}
		return report.Users[i].HighestScore > report.Users[j].HighestScore
	})
	return report, nil
}
