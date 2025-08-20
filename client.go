package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type JiraClient struct {
	baseURL  string
	email    string
	apiToken string
	client   *http.Client
}

func NewJiraClient() *JiraClient {
	return &JiraClient{
		baseURL:  BaseURL,
		email:    Email,
		apiToken: APIToken,
		client:   http.DefaultClient,
	}
}

func (j *JiraClient) request(method, endpoint string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(method, j.baseURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.SetBasicAuth(j.email, j.apiToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(data))
	}

	return data, nil
}

func (j *JiraClient) GetCurrentUser() (*User, error) {
	data, err := j.request("GET", "/rest/api/2/myself", nil)
	if err != nil {
		return nil, fmt.Errorf("getting current user: %w", err)
	}

	var user User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("parsing user response: %w", err)
	}

	return &user, nil
}

func (j *JiraClient) SearchIssues(jql string) ([]Issue, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"jql":        jql,
		"fields":     []string{"key", "summary"},
		"maxResults": MaxResults,
	})
	if err != nil {
		return nil, fmt.Errorf("creating search payload: %w", err)
	}

	data, err := j.request("POST", "/rest/api/3/search/jql", payload)
	if err != nil {
		return nil, fmt.Errorf("searching issues: %w", err)
	}

	var result IssueSearchResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing search response: %w", err)
	}

	return result.Issues, nil
}

func (j *JiraClient) GetIssueWorklogs(issueKey string) ([]Worklog, error) {
	endpoint := fmt.Sprintf("/rest/api/2/issue/%s/worklog", issueKey)
	data, err := j.request("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("getting worklogs for %s: %w", issueKey, err)
	}

	var result WorklogResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing worklog response: %w", err)
	}

	return result.Worklogs, nil
}

func (j *JiraClient) AddWorklog(issueKey string, date string, hours float64, comment string) error {
	endpoint := fmt.Sprintf("/rest/api/2/issue/%s/worklog", issueKey)
	seconds := int(hours * 3600)
	startTime := fmt.Sprintf("%sT09:00:00.000+0300", date)
	payload := map[string]interface{}{
		"timeSpentSeconds": seconds,
		"started":          startTime,
	}

	if comment != "" {
		payload["comment"] = comment
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("creating worklog payload: %w", err)
	}

	_, err = j.request("POST", endpoint, jsonPayload)
	if err != nil {
		return fmt.Errorf("adding worklog to %s: %w", issueKey, err)
	}

	return nil
}
