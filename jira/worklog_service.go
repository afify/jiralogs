package jira

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"LogS/leaves"
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
	user, err := getCachedOrFetchUser(client)
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

// getCachedOrFetchUser tries to load user from cache, fetches if not found
func getCachedOrFetchUser(client *JiraClient) (*shared.User, error) {
	// Try to load from cache first
	if user := loadCachedUser(); user != nil {
		shared.LogErrorf("USER_CACHE", "Using cached user: %s", user.AccountID)
		return user, nil
	}

	// Cache miss, fetch from API with spinner
	shared.LogErrorf("USER_CACHE", "Cache miss, fetching user from API")
	return fetchUserWithSpinner(client)
}

// fetchUserWithSpinner fetches user data with a spinner
func fetchUserWithSpinner(client *JiraClient) (*shared.User, error) {
	// Import spinner - we'll need to add this import
	// For now, let's fetch without spinner and add it in the main function
	user, err := client.GetCurrentUser()
	if err != nil {
		return nil, err
	}

	// Save to cache
	if err := saveCachedUser(user); err != nil {
		shared.LogError("USER_CACHE", err)
		// Don't fail if we can't cache, just continue
	}

	return user, nil
}

// IsCachedUserAvailable checks if user is cached without loading it
func IsCachedUserAvailable() bool {
	_, err := os.Stat(".id")
	return err == nil
}

// loadCachedUser loads user info from .id file
func loadCachedUser() *shared.User {
	data, err := os.ReadFile(".id")
	if err != nil {
		return nil
	}

	var user shared.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil
	}

	return &user
}

// saveCachedUser saves user info to .id file
func saveCachedUser(user *shared.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return os.WriteFile(".id", data, 0644)
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
