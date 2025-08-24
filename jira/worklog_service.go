package jira

import (
	"fmt"
	"time"

	"LogS/shared"
)

// UI-compatible types for the TUI components
type Worklog struct {
	ID             string     `json:"id"`
	Author         User       `json:"author"`
	Comment        string     `json:"comment"`
	Started        *time.Time `json:"started"`
	TimeSpentHours float64    `json:"timeSpentHours"`
	Issue          Issue      `json:"issue"`
}

type Issue struct {
	Key    string `json:"key"`
	Fields Fields `json:"fields"`
}

type Fields struct {
	Summary  string `json:"summary"`
	Status   Status `json:"status"`
	Assignee *User  `json:"assignee"`
}

type Status struct {
	Name string `json:"name"`
}

type User struct {
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
	AccountID    string `json:"accountId"`
}

// UI Messages for TUI components
type WorklogUpdatedMsg struct {
	Worklog Worklog
}

type WorklogService struct {
	client *JiraClient
	user   *shared.User
}

func NewWorklogService(client *JiraClient) (*WorklogService, error) {
	user, err := client.GetCurrentUser()
	if err != nil {
		shared.LogError("NewWorklogService", err)
		return nil, fmt.Errorf("initializing worklog service: %w", err)
	}

	return &WorklogService{
		client: client,
		user:   user,
	}, nil
}

func (w *WorklogService) IsWeekend(t time.Time) bool {
	day := t.Weekday()
	return day == time.Friday || day == time.Saturday
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
