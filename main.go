package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"gopkg.in/andygrunwald/go-jira.v1"
	"os"
	"text/template"
)

type HandlerConfig struct {
	sensu.PluginConfig
	jiraUrl             string
	jiraUser            string
	jiraPassword        string
	jiraProject         string
	jiraIssueType       string
	jiraSummaryTmpl     string
	jiraDescriptionTmpl string
}

var (
	config = HandlerConfig{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-jira-handler",
			Short:    "The Sensu Go jira handler",
			Keyspace: "github.com/akozlenkov/sensu-jira-handler",
		},
	}

	configOptions = []*sensu.PluginConfigOption{
		{
			Path:     "jira-url",
			Env:      "JIRA_URL",
			Argument: "jira-url",
			Usage:    "The jira URL",
			Value:    &config.jiraUrl,
		},
		{
			Path:     "jira-user",
			Env:      "JIRA_USER",
			Argument: "jira-user",
			Usage:    "The jira user",
			Value:    &config.jiraUser,
		},
		{
			Path:     "jira-password",
			Env:      "JIRA_PASSWORD",
			Argument: "jira-password",
			Usage:    "The jira password",
			Value:    &config.jiraPassword,
		},
		{
			Path:     "jira-project",
			Env:      "JIRA_PROJECT",
			Argument: "jira-project",
			Usage:    "The jira project key",
			Value:    &config.jiraProject,
		},
		{
			Path:     "jira-issue-type",
			Env:      "JIRA_ISSUE_TYPE",
			Argument: "jira-issue-type",
			Usage:    "The jira issue type",
			Value:    &config.jiraIssueType,
		},
		{
			Path:     "jira-summary",
			Env:      "JIRA_SUMMARY",
			Argument: "jira-summary",
			Usage:    "The template to use to populate the issue summary",
			Default:  "Check {{ .Name }} fired with status {{ .Status }}",
			Value:    &config.jiraSummaryTmpl,
		},
		{
			Path:     "jira-description",
			Env:      "JIRA_DESCRIPTION",
			Argument: "jira-description",
			Usage:    "The template to use to populate the issue description",
			Default:  "{{ .Output }}",
			Value:    &config.jiraDescriptionTmpl,
		},
	}
)

func main() {
	goHandler := sensu.NewGoHandler(&config.PluginConfig, configOptions, checkArgs, sendMessage)
	goHandler.Execute()
}

func checkArgs(_ *corev2.Event) error {
	if jiraUrl := os.Getenv("JIRA_URL"); jiraUrl != "" {
		config.jiraUrl = jiraUrl
	}

	if jiraUser := os.Getenv("JIRA_USER"); jiraUser != "" {
		config.jiraUser = jiraUser
	}

	if jiraPassword := os.Getenv("JIRA_PASSWORD"); jiraPassword != "" {
		config.jiraPassword = jiraPassword
	}

	if jiraProject := os.Getenv("JIRA_PROJECT"); jiraProject != "" {
		config.jiraProject = jiraProject
	}

	if jiraIssueType := os.Getenv("JIRA_ISSUE_TYPE"); jiraIssueType != "" {
		config.jiraIssueType = jiraIssueType
	}

	if jiraSummaryTmpl := os.Getenv("JIRA_SUMMARY"); jiraSummaryTmpl != "" {
		config.jiraSummaryTmpl = jiraSummaryTmpl
	}

	if jiraDescriptionTmpl := os.Getenv("JIRA_DESCRIPTION"); jiraDescriptionTmpl != "" {
		config.jiraDescriptionTmpl = jiraDescriptionTmpl
	}

	if len(config.jiraUrl) == 0 {
		return fmt.Errorf("--jira-url or JIRA_URL environment variable is required")
	}

	if len(config.jiraUser) == 0 {
		return fmt.Errorf("--jira-user or JIRA_USER environment variable is required")
	}

	if len(config.jiraPassword) == 0 {
		return fmt.Errorf("--jira-password or JIRA_PASSWORD environment variable is required")
	}

	if len(config.jiraProject) == 0 {
		return fmt.Errorf("--jira-project or JIRA_PROJECT environment variable is required")
	}

	if len(config.jiraIssueType) == 0 {
		return fmt.Errorf("--jira-issue_type or JIRA_ISSUE_TYPE environment variable is required")
	}
	return nil
}

func sendMessage(event *corev2.Event) error {
	auth := jira.BasicAuthTransport{
		Username: config.jiraUser,
		Password: config.jiraPassword,
	}

	client, err := jira.NewClient(auth.Client(), config.jiraUrl)
	if err != nil {
		exitOnErr(err)
	}

	var project bytes.Buffer
	projectTmpl, err := template.New("project").Parse(config.jiraProject)
	if err != nil {
		exitOnErr(err)
	}
	if err = projectTmpl.Execute(&project, event.Check); err != nil {
		exitOnErr(err)
	}

	var issueType bytes.Buffer
	issueTypeTmpl, err := template.New("issueType").Parse(config.jiraIssueType)
	if err != nil {
		exitOnErr(err)
	}
	if err = issueTypeTmpl.Execute(&issueType, event.Check); err != nil {
		exitOnErr(err)
	}

	var summary bytes.Buffer
	summaryTmpl, err := template.New("summary").Parse(config.jiraSummaryTmpl)
	if err != nil {
		exitOnErr(err)
	}
	if err = summaryTmpl.Execute(&summary, event.Check); err != nil {
		exitOnErr(err)
	}

	var description bytes.Buffer
	descriptionTmpl, err := template.New("description").Parse(config.jiraDescriptionTmpl)
	if err != nil {
		exitOnErr(err)
	}
	if err = descriptionTmpl.Execute(&description, event.Check); err != nil {
		exitOnErr(err)
	}

	_, _, err = client.Issue.Create(&jira.Issue{
		Fields: &jira.IssueFields{
			Type: jira.IssueType{
				Name: issueType.String(),
			},
			Project: jira.Project{
				Key: project.String(),
			},
			Summary:     summary.String(),
			Description: description.String(),
		},
	})
	if err != nil {
		exitOnErr(errors.New("unable create jira ticket"))
	}
	return nil
}

func exitOnErr(err error) {
	fmt.Println(err.Error())
	os.Exit(2)
}
