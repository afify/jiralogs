package config

import (
	"os"
	"path/filepath"

	"LogS/shared"
	"github.com/joho/godotenv"
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
	envFileExists := true
	if err := godotenv.Load(); err != nil {
		shared.LogError("CONFIG_INIT", err)
		shared.LogErrorf("CONFIG_INIT", "Failed to load .env file: %s", err.Error())
		envFileExists = false
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

	// Show helpful message if configuration is missing
	if !cfg.IsConfigured() {
		if !envFileExists {
			printEnvSetupHelp()
		} else {
			printMissingConfigHelp(cfg)
		}
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

// printEnvSetupHelp prints instructions for setting up the .env file
func printEnvSetupHelp() {
	println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	println("📋 JIRA Configuration Required")
	println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	println()
	println("No .env file found. Please create a .env file with the following:")
	println()
	println("1. Create a JIRA API token:")
	println("   🔗 https://id.atlassian.com/manage-profile/security/api-tokens")
	println()
	println("2. Create a .env file in the current directory with:")
	println()
	println("   JIRA_BASE_URL=https://your-company.atlassian.net")
	println("   JIRA_EMAIL=your.email@company.com")
	println("   JIRA_API_TOKEN=your_api_token_here")
	println()
	println("Example .env file:")
	println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	println("JIRA_BASE_URL=https://acme.atlassian.net")
	println("JIRA_EMAIL=john.doe@acme.com")
	println("JIRA_API_TOKEN=ATATT3xFfGF0abcdef123456...")
	println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	println()
}

// printMissingConfigHelp prints which specific configuration values are missing
func printMissingConfigHelp(cfg *Config) {
	println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	println("⚠️  Incomplete JIRA Configuration")
	println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	println()
	println("The following environment variables are missing or empty:")
	println()

	if cfg.JiraBaseURL == "" {
		println("  ❌ JIRA_BASE_URL - Your JIRA instance URL")
		println("     Example: https://your-company.atlassian.net")
		println()
	}

	if cfg.JiraEmail == "" {
		println("  ❌ JIRA_EMAIL - Your JIRA account email")
		println("     Example: your.email@company.com")
		println()
	}

	if cfg.JiraToken == "" {
		println("  ❌ JIRA_API_TOKEN - Your JIRA API token")
		println("     Create one at: https://id.atlassian.com/manage-profile/security/api-tokens")
		println()
	}

	println("Please update your .env file with the missing values.")
	println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	println()
}
