package config

import (
	"os"
	"path/filepath"
	
	"github.com/joho/godotenv"
	"LogS/shared"
)

const (
	appName              = "jiralogs"
	defaultDataDirectory = ".jiralogs"
)

type Config struct {
	workingDir string
	dataDir    string
	
	// JIRA Configuration
	JiraBaseURL string
	JiraEmail   string
	JiraToken   string
}

// New creates a new configuration instance
func New() *Config {
	// Load .env file
	shared.LogErrorf("CONFIG_INIT", "Loading .env file")
	if err := godotenv.Load(); err != nil {
		shared.LogError("CONFIG_INIT", err)
		shared.LogErrorf("CONFIG_INIT", "Failed to load .env file: %s", err.Error())
	} else {
		shared.LogErrorf("CONFIG_INIT", ".env file loaded successfully")
	}
	
	homeDir, _ := os.UserHomeDir()
	workingDir, _ := os.Getwd()
	
	cfg := &Config{
		workingDir:  workingDir,
		dataDir:     filepath.Join(homeDir, defaultDataDirectory),
		JiraBaseURL: os.Getenv("JIRA_BASE_URL"),
		JiraEmail:   os.Getenv("JIRA_EMAIL"),
		JiraToken:   os.Getenv("JIRA_API_TOKEN"),
	}
	
	shared.LogErrorf("CONFIG_INIT", "Configuration created - JiraBaseURL: %s, JiraEmail: %s, HasToken: %v", 
		cfg.JiraBaseURL, cfg.JiraEmail, cfg.JiraToken != "")
	
	return cfg
}

// WorkingDir returns the current working directory
func (c *Config) WorkingDir() string {
	return c.workingDir
}

// DataDir returns the data directory path
func (c *Config) DataDir() string {
	return c.dataDir
}

// IsConfigured returns true if the basic JIRA configuration is present
func (c *Config) IsConfigured() bool {
	return c.JiraBaseURL != "" && c.JiraEmail != "" && c.JiraToken != ""
}