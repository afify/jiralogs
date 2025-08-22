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
	HTTPClientError     = 400
	SecondsPerHour      = 3600
	ExitCodeError       = 1
	DayNameLength       = 3
	FilePermissions     = 0666
	DateFormat          = "2006-01-02"
	DateTimeFormat      = "2006-01-02 15:04:05"
	WorkStartTime       = "T09:00:00.000+0300"
	WorklogDateFormat   = "2006-01-02T15:04:05.000-0700"
	ZeroHours           = 0.0
	ZeroTotal           = 0
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
	return day == time.Friday || day == time.Saturday
}
