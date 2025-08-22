package main

import "LogS/shared"

type IssueSearchResult struct {
	Issues []shared.Issue `json:"issues"`
}

type WorklogResult struct {
	Worklogs []shared.Worklog `json:"worklogs"`
}

