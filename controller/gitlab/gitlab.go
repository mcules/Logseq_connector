package gitlab

import (
	"Logseq_connector/controller/fileFunctions"
	git "github.com/xanzy/go-gitlab"
	"log"
	"strconv"
	"strings"
)

type Config struct {
	Name             string
	Project          string
	Graph            string
	URL              string
	AuthToken        string
	Username         string
	AssigneeUsername string
	Sort             string
	State            string
	Scope            string
	Client           *git.Client
}

var config Config

func Process(extConfig Config, path string) {
	config = extConfig

	config.Client = getClient()

	filename, fileContent := getGitlabIssues()
	fileFunctions.WriteFile(fileContent, fileFunctions.GetFilehandle(path+filename+"tickets.md"))
}

func getClient() *git.Client {
	gitClient, err := git.NewClient(
		config.AuthToken,
		git.WithBaseURL(config.URL),
	)
	if err != nil {
		log.Println(err.Error())
	}

	return gitClient
}

func getGitlabIssues() (filename string, fileContent string) {
	log.Println("get Gitlab Issues: " + config.Name)
	var issues []*git.Issue
	sort := "desc"
	scope := "all"

	if len(config.Sort) > 0 {
		sort = config.Sort
	}
	if len(config.Scope) > 0 {
		scope = config.Scope
	}

	issueOpts := &git.ListIssuesOptions{
		Sort:  git.String(sort),
		Scope: git.String(scope),
	}

	if len(config.AssigneeUsername) > 0 {
		issueOpts.AssigneeUsername = git.String(config.AssigneeUsername)
	}

	if len(config.State) > 0 {
		issueOpts.State = git.String(config.State)
	}

	for {
		tempIssues, resp, err := config.Client.Issues.ListIssues(issueOpts)
		if err != nil {
			log.Println(err.Error())
		}

		issues = append(issues, tempIssues...)

		if resp.NextPage == 0 {
			break
		}

		issueOpts.Page = resp.NextPage
	}

	if len(issues) > 0 {
		for _, val := range issues {
			var project, labels, milestone, assignee, closed string
			projectName, _ := getGitlabProjectName(val.ProjectID)

			if len(config.Project) > 0 {
				project = "#" + config.Project + " "
			}

			if len(labels) > 0 {
				for _, label := range val.Labels {
					labels = labels + " [[" + label + "]]"
				}
			}

			if val.Milestone != nil && len(val.Milestone.Title) > 0 {
				milestone = " [[Milestone:: " + val.Milestone.Title + "]]"
			}

			if val.Assignee != nil && len(val.Assignee.Username) > 0 {
				assignee = " [[Assignee:: " + val.Assignee.Username + "]]"
			}

			if val.ClosedAt != nil {
				closed = "\n" + "completed:: " + val.ClosedAt.Format("[[01-02-2006]] *15:04*")
			}

			fileContent = addTicket(
				projectName+" [#"+strconv.Itoa(val.IID)+"]",
				"- "+getState(val.State)+" "+getGitlabPriority(val)+project+projectName+" [#"+strconv.Itoa(val.IID)+"]("+val.WebURL+")"+" "+val.Title+labels+milestone+assignee+closed,
				fileContent)
		}
	}

	return getProjectPath(config.Project), fileContent
}

func addTicket(searchStr string, insertStr string, fileContent string) string {
	var added bool
	lines := strings.Split(fileContent, "\n")

	for i, line := range lines {
		if strings.Contains(line, searchStr) {
			lines[i] = insertStr
			added = true
		}
	}

	if !added {
		newLine := []string{insertStr}
		lines = append(newLine, lines...)
	}

	return strings.Join(lines, "\n")
}

func getGitlabProjectName(projectId int) (string, error) {
	project, _, err := config.Client.Projects.GetProject(projectId, &git.GetProjectOptions{})
	if err != nil {
		return "", err
	}

	return project.Name, nil
}

func getProjectPath(project string) string {
	if len(project) > 0 {
		return project + "___"
	}

	return ""
}

func getGitlabPriority(issue *git.Issue) string {
	for _, label := range issue.Labels {
		if strings.HasPrefix(label, "priority::1 ") {
			return "[#A] "
		} else if strings.HasPrefix(label, "priority::2 ") {
			return "[#B] "
		} else if strings.HasPrefix(label, "priority::3 ") {
			return "[#C] "
		} else if strings.HasPrefix(label, "priority::4 ") {
			return "[#D] "
		}
	}

	return ""
}

func getState(state string) string {
	switch state {
	case "opened":
		return "TODO"
	case "closed":
		return "DONE"
	case "TODO":
		return "opened"
	case "DONE":
		return "closed"
	case "DOING":
		return "opened"
	case "NOW":
		return "opened"
	case "LATER":
		return "opened"
	}

	return ""
}
