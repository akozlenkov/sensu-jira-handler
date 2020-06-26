package main

import (
	"fmt"
	"gopkg.in/andygrunwald/go-jira.v1"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"os"
)

type HandlerConfig struct {
	sensu.PluginConfig
	url   		string
	user 		string
	password	string
	project 	string
	issueType   string
	summary		string
	description string
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
			Path:      "url",
			Env:       "JIRA_URL",
			Argument:  "url",
			Usage:     "The jira url",
			Value:     &config.url,
		},
		{
			Path:      "user",
			Env:       "JIRA_USER",
			Argument:  "user",
			Usage:     "The jira user",
			Value:     &config.user,
		},
		{
			Path:      "password",
			Env:       "JIRA_PASSWORD",
			Argument:  "password",
			Usage:     "The jira password",
			Value:     &config.user,
		},
		{
			Path:      "project",
			Env:       "JIRA_PROJECT",
			Argument:  "project",
			Usage:     "The jira project ID",
			Value:     &config.project,
		},
		{
			Path:      "issue_type",
			Env:       "JIRA_ISSUE_TYPE",
			Argument:  "issue_type",
			Usage:     "The jira issue type ID",
			Value:     &config.issueType,
		},
		{
			Path:      "summary",
			Argument:  "summary",
			Usage:     "The jira issue summary",
			Value:     &config.summary,
		},
		{
			Path:      "description",
			Argument:  "description",
			Usage:     "The jira issue description",
			Value:     &config.description,
		},
	}
)

func main()  {
	goHandler := sensu.NewGoHandler(&config.PluginConfig, configOptions, checkArgs, sendMessage)
	goHandler.Execute()
}

func checkArgs(_ *corev2.Event) error {
	if url := os.Getenv("JIRA_URL"); url != "" {
		config.url = url
	}

	if user := os.Getenv("JIRA_USER"); user != "" {
		config.user = user
	}

	if password := os.Getenv("JIRA_PASSWORD"); password != "" {
		config.password = password
	}

	if project := os.Getenv("JIRA_PROJECT"); project != "" {
		config.project = project
	}

	if issueType := os.Getenv("JIRA_ISSUE_TYPE"); issueType != "" {
		config.issueType = issueType
	}

	if len(config.url) == 0 {
		return fmt.Errorf("--url or JIRA_URL environment variable is required")
	}

	if len(config.user) == 0 {
		return fmt.Errorf("--user or JIRA_USER environment variable is required")
	}

	if len(config.password) == 0 {
		return fmt.Errorf("--password or JIRA_PASSWORD environment variable is required")
	}

	if len(config.project) == 0 {
		return fmt.Errorf("--project or JIRA_PROJECT environment variable is required")
	}

	if len(config.issueType) == 0 {
		return fmt.Errorf("--issue_type or JIRA_ISSUE_TYPE environment variable is required")
	}

	if len(config.summary) == 0 {
		return fmt.Errorf("--summary is required")
	}

	if len(config.description) == 0 {
		return fmt.Errorf("--description is required")
	}

	return nil
}

func sendMessage(event *corev2.Event) error {
	auth := jira.BasicAuthTransport{
		Username: config.user,
		Password: config.password,
	}

	client, err := jira.NewClient(auth.Client(), config.url)
	if err != nil {
		exitOnErr(err)
	}

	_, _, err = client.Issue.Create(&jira.Issue{
		Fields: &jira.IssueFields{
			Type: jira.IssueType{
				ID: config.issueType,
			},
			Project: jira.Project{
				ID: config.project,
			},
			Summary: config.summary,
			Description: config.description,
		},
	})
	if err != nil {
		exitOnErr(err)
	}
	return nil
}

func exitOnErr(err error) {
	fmt.Println(err.Error())
	os.Exit(2)
}