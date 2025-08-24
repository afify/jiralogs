package main

import (
	"LogS/jira"
	"LogS/shared"
)

// jiraClientWrapper implements tui.JiraClientInterface
type jiraClientWrapper struct {
	*jira.JiraClient
}

func (w *jiraClientWrapper) GetCurrentUser() (*shared.User, error) {
	return w.JiraClient.GetCurrentUser()
}

func (w *jiraClientWrapper) SearchIssues(jql string) ([]shared.Issue, error) {
	return w.JiraClient.SearchIssues(jql)
}

func (w *jiraClientWrapper) GetIssueWorklogs(issueKey string) ([]shared.Worklog, error) {
	return w.JiraClient.GetIssueWorklogs(issueKey)
}

func (w *jiraClientWrapper) AddWorklog(issueKey string, date string, hours float64, comment string) error {
	return w.JiraClient.AddWorklog(issueKey, date, hours, comment)
}

func (w *jiraClientWrapper) CreateIssue(projectKey, summary, description, issueType string) (string, error) {
	return w.JiraClient.CreateIssue(projectKey, summary, description, issueType)
}

func (w *jiraClientWrapper) GetProjects() ([]shared.ProjectData, error) {
	return w.JiraClient.GetProjects()
}

// worklogServiceWrapper implements tui.WorklogServiceInterface
type worklogServiceWrapper struct {
	*jira.WorklogService
}

func (w *worklogServiceWrapper) FetchWorklogs(period shared.Period) (map[string]*shared.TicketWorklog, map[string]float64, map[string][]string, error) {
	return w.WorklogService.FetchWorklogs(period)
}

func (w *worklogServiceWrapper) GenerateDailyReports(period shared.Period, dailyHours map[string]float64, dailyTickets map[string][]string) []shared.DailyReport {
	return w.WorklogService.GenerateDailyReports(period, dailyHours, dailyTickets)
}

func (w *worklogServiceWrapper) GenerateSummary(period shared.Period, reports []shared.DailyReport, ticketLogs map[string]*shared.TicketWorklog) shared.Summary {
	return w.WorklogService.GenerateSummary(period, reports, ticketLogs)
}

func (w *worklogServiceWrapper) LogMissingDays(ticketKey string, dates []string, comment string) error {
	return w.WorklogService.LogMissingDays(ticketKey, dates, comment)
}

func (w *worklogServiceWrapper) LogDistribution(distribution map[string][]string, comment string) error {
	return w.WorklogService.LogDistribution(distribution, comment)
}
