package jira

import "LogS/shared"

type IssueSearchResult struct {
	Issues     []shared.Issue `json:"issues"`
	Total      int            `json:"total"`
	MaxResults int            `json:"maxResults"`
	StartAt    int            `json:"startAt"`
}

type WorklogResult struct {
	Worklogs []shared.Worklog `json:"worklogs"`
}
