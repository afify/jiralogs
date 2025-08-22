package main

import (
	"fmt"

	"LogS/shared"
)

type WorklogService struct {
	client *JiraClient
	user   *shared.User
}

func NewWorklogService(client *JiraClient) (*WorklogService, error) {
	user, err := client.GetCurrentUser()
	if err != nil {
		shared.LogError("NewWorklogService", err)
		return nil, fmt.Errorf("initializing worklog service: %w", err)
	}

	return &WorklogService{
		client: client,
		user:   user,
	}, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
