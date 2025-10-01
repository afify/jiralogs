package jira

import (
	"fmt"
	"time"

	"LogS/internal/leaves"
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
	client       *JiraClient
	user         *shared.User
	leaveManager *leaves.Manager
}

func NewWorklogService(client *JiraClient) (*WorklogService, error) {
	user, err := client.GetCurrentUser()
	if err != nil {
		shared.LogError("NewWorklogService", err)
		return nil, fmt.Errorf("initializing worklog service: %w", err)
	}

	return &WorklogService{
		client:       client,
		user:         user,
		leaveManager: leaves.NewManager(),
	}, nil
}

func (w *WorklogService) IsWeekend(t time.Time) bool {
	day := t.Weekday()
	return day == time.Friday || day == time.Saturday
}

func (w *WorklogService) IsLeaveDay(t time.Time) bool {
	return w.leaveManager.IsLeaveDay(t)
}

func (w *WorklogService) ShouldSkipDay(t time.Time) bool {
	return w.IsWeekend(t) || w.IsLeaveDay(t)
}

func (w *WorklogService) LogWork(ticketKey string, hours float64, description string, date time.Time) error {
	return w.client.AddWorklog(ticketKey, date.Format("2006-01-02"), hours, description)
}

func (w *WorklogService) GetMyTickets() ([]shared.Issue, error) {
	jql := fmt.Sprintf("(assignee = currentUser()) OR (worklogAuthor = '%s') ORDER BY updated DESC", w.user.AccountID)
	return w.client.SearchIssues(jql)
}

func (w *WorklogService) GetMyRecentTickets() ([]shared.Issue, error) {
	jql := fmt.Sprintf("worklogAuthor = '%s' AND worklogDate >= -30d ORDER BY updated DESC", w.user.AccountID)
	return w.client.SearchIssues(jql)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
