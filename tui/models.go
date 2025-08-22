package tui

import "LogS/shared"

type JiraClientInterface interface {
	GetCurrentUser() (*shared.User, error)
	GetConfiguredEmail() string
	SearchIssues(jql string) ([]shared.Issue, error)
	GetIssueWorklogs(issueKey string) ([]shared.Worklog, error)
	AddWorklog(issueKey string, date string, hours float64, comment string) error
	CreateIssue(projectKey, summary, description, issueType string) (string, error)
	GetProjects() ([]shared.ProjectData, error)
}

type WorklogServiceInterface interface {
	FetchWorklogs(period shared.Period) (map[string]*shared.TicketWorklog, map[string]float64, map[string][]string, error)
	GenerateDailyReports(period shared.Period, dailyHours map[string]float64, dailyTickets map[string][]string) []shared.DailyReport
	GenerateSummary(period shared.Period, reports []shared.DailyReport, ticketLogs map[string]*shared.TicketWorklog) shared.Summary
	LogMissingDays(ticketKey string, dates []string, comment string) error
	LogDistribution(distribution map[string][]string, comment string) error
}

