package main

import (
	"fmt"

	"LogS/jira"

	"github.com/charmbracelet/huh"
)

func createNewTicket(jiraClient *jira.JiraClient) (string, string, error) {
	// Fetch projects
	projects, err := jiraClient.GetProjects()
	if err != nil {
		return "", "", fmt.Errorf("fetching projects: %w", err)
	}

	if len(projects) == 0 {
		return "", "", fmt.Errorf("no projects available")
	}

	// Create project options
	projectOptions := []huh.Option[string]{}
	for _, project := range projects {
		projectOptions = append(projectOptions,
			huh.NewOption(fmt.Sprintf("%s - %s", project.Key, project.Name), project.Key))
	}

	var selectedProject string
	err = huh.NewSelect[string]().
		Title("Select project").
		Options(projectOptions...).
		Filtering(true).
		Height(10).
		Value(&selectedProject).
		WithTheme(huh.ThemeCharm()).
		Run()

	if err != nil {
		return "", "", err
	}

	// Get ticket details
	var summary string
	var description string
	issueType := "Task" // default

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Ticket summary").
				Placeholder("Brief description of the work").
				Value(&summary).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("summary is required")
					}
					return nil
				}),

			huh.NewText().
				Title("Description (optional)").
				Placeholder("Detailed description...").
				Value(&description),

			huh.NewInput().
				Title("Issue type").
				Placeholder("Task").
				Value(&issueType).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("issue type is required")
					}
					return nil
				}),
		),
	).WithTheme(huh.ThemeCharm())

	err = form.Run()
	if err != nil {
		return "", "", err
	}

	// Create the issue
	fmt.Printf("\n🎫 Creating ticket in %s...\n", selectedProject)
	ticketKey, err := jiraClient.CreateIssue(selectedProject, summary, description, issueType)
	if err != nil {
		return "", "", fmt.Errorf("creating ticket: %w", err)
	}

	fmt.Printf("✅ Created ticket: %s\n", ticketKey)
	return ticketKey, summary, nil
}
