package jira

import (
	"Logseq_connector/controller/fileFunctions"
	"Logseq_connector/controller/logseq"
	"context"
	jiraApi "github.com/andygrunwald/go-jira/v2/cloud"
	"log"
	"os"
)

// Config represents configuration details for connecting to external systems such as Jira or similar services.
type Config struct {
	Name     string
	Graph    string
	Url      string
	Username string
	Token    string
}

var config Config

// Process retrieves issues from Jira based on the provided configuration, formats them into tasks, and writes them to a file.
func Process(extConf Config, path string) {
	config = extConf

	tp := jiraApi.BasicAuthTransport{
		Username: config.Username,
		APIToken: config.Token,
	}
	jiraClient, err := jiraApi.NewClient(config.Url, tp.Client())
	if err != nil {
		panic(err)
	}

	fields, _, err := jiraClient.Field.GetList(context.Background())
	if err != nil {
		panic(err)
	}

	jql := "assignee=\"" + config.Username + "\" AND status NOT IN (Done,Canceled,Closed,Completed)"

	options := &jiraApi.SearchOptions{Expand: "renderedFields"}
	issues, _, err := jiraClient.Issue.Search(context.Background(), jql, options)
	if err != nil {
		panic(err)
	}

	var fileContent string

	field := getFieldKey("Target end", fields)

	for _, i := range issues {
		var task logseq.Task
		task.Id = i.Key
		task.ConfigName = config.Name
		task.Status = getTaskType(i.Fields.Status.Name)
		task.Priority = getPrio(i.Fields.Priority.Name)
		task.Project = i.Fields.Project.Name
		task.Url = config.Url + "browse/" + i.Key
		task.Title = i.Fields.Summary

		if dueDate, ok := i.Fields.Unknowns.Value(field); ok && dueDate != nil {
			task.DueDate = dueDate.(string)
		}

		taskLine, uniqueStr := logseq.CreateTask(task)
		fileContent = logseq.AddOrReplaceEntry(uniqueStr, taskLine, fileContent)
	}

	filename := path + config.Name + ".md"

	fileHandle, handleErr := fileFunctions.GetFilehandle(filename)
	if handleErr != nil {
		log.Println(err)
	}
	defer func(fileHandle *os.File) {
		err := fileHandle.Close()
		if err != nil {
			log.Println(err)
		}
	}(fileHandle)

	fileFunctions.WriteFile(fileContent, fileHandle)
}

// getTaskType maps a given status string to a predefined task type and returns it. Defaults to "UNKNOWN" if no match is found.
func getTaskType(status string) string {
	statusToType := map[string]string{
		"Open":                 "TODO",
		"To Do":                "TODO",
		"Pending":              "TODO",
		"Reopened":             "TODO",
		"Zu erledigen":         "TODO",
		"In Arbeit":            "DOING",
		"In Progress":          "DOING",
		"Escalated":            "WAIT",
		"Waiting for approval": "WAIT",
		"Waiting for customer": "WAIT",
		"Waiting for support":  "WAIT",
		"Warten":               "WAIT",
		"Canceled":             "CLOSED",
		"Closed":               "CLOSED",
		"Done":                 "DONE",
		"Completed":            "DONE",
		"Resolved":             "DONE",
	}

	if taskType, ok := statusToType[status]; ok {
		return taskType
	}
	return "UNKNOWN"
}

// getPrio maps a priority string to a corresponding integer value. Returns 0 for unrecognized priority strings.
func getPrio(priority string) int {
	switch priority {
	case "Highest":
		return 1
	case "High":
		return 1
	case "Medium":
		return 2
	case "Low":
		return 3
	case "Lowest":
		return 4
	}

	return 0
}

// getFieldKey retrieves the key for a specific field name from a list of Jira fields. Returns an empty string if not found.
func getFieldKey(fieldName string, fields []jiraApi.Field) string {
	for _, field := range fields {
		if field.Name == fieldName {
			return field.Key
		}
	}
	return ""
}
