package jira

import (
	"encoding/json"
	"fmt"

	"LogS/shared"
)

func (j *JiraClient) SearchIssues(jql string) ([]shared.Issue, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"jql":        jql,
		"fields":     []string{"key", "summary"},
		"maxResults": shared.MaxResults,
	})
	if err != nil {
		shared.LogError("JiraClient.SearchIssues", err)
		return nil, fmt.Errorf("creating search payload: %w", err)
	}

	data, err := j.request("POST", "/rest/api/3/search/jql", payload)
	if err != nil {
		shared.LogError("JiraClient.SearchIssues", err)
		return nil, fmt.Errorf("searching issues: %w", err)
	}

	var result IssueSearchResult
	if err := json.Unmarshal(data, &result); err != nil {
		shared.LogError("JiraClient.SearchIssues", err)
		return nil, fmt.Errorf("parsing search response: %w", err)
	}

	if result.Total > len(result.Issues) {
		shared.LogErrorf("SearchIssues", "WARNING: JIRA returned %d/%d issues (missing %d)", len(result.Issues), result.Total, result.Total-len(result.Issues))
	}

	return result.Issues, nil
}

func (j *JiraClient) CreateIssue(projectKey string, summary string, description string, issueType string) (string, error) {
	payload := map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]string{
				"key": projectKey,
			},
			"summary":     summary,
			"description": description,
			"issuetype": map[string]string{
				"name": issueType,
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		shared.LogError("JiraClient.CreateIssue", err)
		return "", fmt.Errorf("creating issue payload: %w", err)
	}

	data, err := j.request("POST", "/rest/api/2/issue", jsonPayload)
	if err != nil {
		shared.LogError("JiraClient.CreateIssue", err)
		return "", fmt.Errorf("creating issue: %w", err)
	}

	var result struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		shared.LogError("JiraClient.CreateIssue", err)
		return "", fmt.Errorf("parsing create issue response: %w", err)
	}

	return result.Key, nil
}

func (j *JiraClient) GetProjects() ([]shared.ProjectData, error) {
	data, err := j.request("GET", "/rest/api/2/project", nil)
	if err != nil {
		shared.LogError("JiraClient.GetProjects", err)
		return nil, fmt.Errorf("getting projects: %w", err)
	}

	var projects []shared.ProjectData
	if err := json.Unmarshal(data, &projects); err != nil {
		shared.LogError("JiraClient.GetProjects", err)
		return nil, fmt.Errorf("parsing projects response: %w", err)
	}

	return projects, nil
}
