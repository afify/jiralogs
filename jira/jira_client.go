package jira

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"LogS/shared"
)

type JiraClient struct {
	baseURL  string
	email    string
	apiToken string
	client   *http.Client
}

func NewJiraClient(baseURL, email, apiToken string) *JiraClient {
	return &JiraClient{
		baseURL:  baseURL,
		email:    email,
		apiToken: apiToken,
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

	if resp.StatusCode >= shared.HTTPClientError {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(data))
	}

	return data, nil
}

func (j *JiraClient) GetConfiguredEmail() string {
	return j.email
}
