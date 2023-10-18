package gitlab

import (
	"Logseq_connector/controller/fileFunctions"
	git "github.com/xanzy/go-gitlab"
	"log"
	"strconv"
	"strings"
)

type Config struct {
	Name      string
	Project   string
	Graph     string
	URL       string
	AuthToken string
	Username  string
	Client    *git.Client
}

var config Config

func Process(extConfig Config, path string) {
	config = extConfig

	config.Client = getClient()

	filename, fileContent := getGitlabIssues()
	fileFunctions.WriteFile(fileContent, fileFunctions.GetFilehandle(path+filename+"tickets.md"))
}

func getClient() *git.Client {
	git, err := git.NewClient(
		config.AuthToken,
		git.WithBaseURL(config.URL),
	)
	if err != nil {
		log.Println(err.Error())
	}

	return git
}

func getGitlabIssues() (filename string, fileContent string) {
	log.Println("get Gitlab Issues: " + config.Name)
	var issues []*git.Issue

	issueOpts := &git.ListIssuesOptions{
		State:            git.String("opened"),
		AssigneeUsername: git.String(config.Username),
		Sort:             git.String("desc"),
		Scope:            git.String("all"),
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
			projectName, _ := getGitlabProjectName(val.ProjectID)
			log.Println(val.State, " ", projectName, " #", val.IID, " ", val.Title, " ", val.Labels)

			var project string
			if len(config.Project) > 0 {
				project = "#" + config.Project + " "
			}

			if val.State == "opened" {
				fileContent = addTicket(projectName+" [#"+strconv.Itoa(val.IID)+"]", "- "+getState(val.State)+" "+getGitlabPriority(val)+project+projectName+" [#"+strconv.Itoa(val.IID)+"]("+val.WebURL+")"+" "+val.Title, fileContent)
			}
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
