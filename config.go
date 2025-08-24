package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var (
	APIToken string
	BaseURL  string
	Email    string
)

// Constants have been moved to shared/types.go

func LoadConfig() error {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	APIToken = os.Getenv("JIRA_API_TOKEN")
	BaseURL = os.Getenv("JIRA_BASE_URL")
	Email = os.Getenv("JIRA_EMAIL")

	if APIToken == "" {
		return fmt.Errorf("JIRA_API_TOKEN environment variable is not set")
	}
	if BaseURL == "" {
		return fmt.Errorf("JIRA_BASE_URL environment variable is not set")
	}
	if Email == "" {
		return fmt.Errorf("JIRA_EMAIL environment variable is not set")
	}

	return nil
}

func IsWeekend(t time.Time) bool {
	day := t.Weekday()
	return day == time.Friday || day == time.Saturday
}
