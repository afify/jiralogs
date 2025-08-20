package main

import (
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

const (
	RequiredHoursPerDay = 8.0
	MaxResults          = 500
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	APIToken = os.Getenv("JIRA_API_TOKEN")
	BaseURL = os.Getenv("JIRA_BASE_URL")
	Email = os.Getenv("JIRA_EMAIL")

	if APIToken == "" {
		log.Fatal("JIRA_API_TOKEN environment variable is not set")
	}
	if BaseURL == "" {
		log.Fatal("JIRA_BASE_URL environment variable is not set")
	}
	if Email == "" {
		log.Fatal("JIRA_EMAIL environment variable is not set")
	}
}

func IsWeekend(t time.Time) bool {
	day := t.Weekday()
	// Weekend in Egypt and KSA is Friday and Saturday
	return day == time.Friday || day == time.Saturday
}
