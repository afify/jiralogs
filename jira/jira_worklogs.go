package jira

import (
	"encoding/json"
	"fmt"

	"LogS/shared"
)

func (j *JiraClient) GetIssueWorklogs(issueKey string) ([]shared.Worklog, error) {
	endpoint := fmt.Sprintf("/rest/api/2/issue/%s/worklog", issueKey)
	data, err := j.request("GET", endpoint, nil)
	if err != nil {
		shared.LogError("JiraClient.GetIssueWorklogs", err)
		return nil, fmt.Errorf("getting worklogs for %s: %w", issueKey, err)
	}

	var result WorklogResult
	if err := json.Unmarshal(data, &result); err != nil {
		shared.LogError("JiraClient.GetIssueWorklogs", err)
		return nil, fmt.Errorf("parsing worklog response: %w", err)
	}

	return result.Worklogs, nil
}

func (j *JiraClient) AddWorklog(issueKey string, date string, hours float64, comment string) error {
	endpoint := fmt.Sprintf("/rest/api/2/issue/%s/worklog", issueKey)
	seconds := int(hours * shared.SecondsPerHour)
	startTime := fmt.Sprintf("%s%s", date, shared.WorkStartTime)
	payload := map[string]interface{}{
		"timeSpentSeconds": seconds,
		"started":          startTime,
	}

	if comment != "" {
		payload["comment"] = comment
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		shared.LogError("JiraClient.AddWorklog", err)
		return fmt.Errorf("creating worklog payload: %w", err)
	}

	_, err = j.request("POST", endpoint, jsonPayload)
	if err != nil {
		shared.LogError("JiraClient.AddWorklog", err)
		return fmt.Errorf("adding worklog to %s: %w", issueKey, err)
	}

	return nil
}
