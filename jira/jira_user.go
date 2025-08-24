package jira

import (
	"encoding/json"
	"fmt"

	"LogS/shared"
)

func (j *JiraClient) GetCurrentUser() (*shared.User, error) {
	data, err := j.request("GET", "/rest/api/2/myself", nil)
	if err != nil {
		shared.LogError("JiraClient.GetCurrentUser", err)
		return nil, fmt.Errorf("getting current user: %w", err)
	}

	var user shared.User
	if err := json.Unmarshal(data, &user); err != nil {
		shared.LogError("JiraClient.GetCurrentUser", err)
		return nil, fmt.Errorf("parsing user response: %w", err)
	}

	return &user, nil
}
