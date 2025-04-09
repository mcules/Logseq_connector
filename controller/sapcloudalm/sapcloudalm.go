package sapcloudalm

import (
	"Logseq_connector/controller/fileFunctions"
	"Logseq_connector/controller/logseq"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Config represents the primary configuration structure for interacting with external APIs and services.
type Config struct {
	Name         string
	Graph        string
	ClientId     string
	ClientSecret string
	UserId       string
	Url          string
	TokenUrl     string
	Token        string
}

var config Config

// TokenResponse represents the structure of an OAuth2 token response.
// AccessToken is the issued token used for authentication.
// TokenType indicates the type of token issued, typically "Bearer".
// ExpiresIn specifies the lifespan of the token in seconds.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// getAccessToken requests and retrieves an OAuth2 access token using client credentials and token URL.
// It returns the access token as a string or an error if the request fails or the response cannot be decoded.
func getAccessToken(clientID string, clientSecret string, tokenURL string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed: %s\n%s", resp.Status, string(bodyBytes))
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

// getProjects retrieves a list of projects from the external API and returns it as a slice of map objects or an error.
func getProjects() ([]map[string]interface{}, error) {
	req, err := http.NewRequest("GET", config.Url+"api/calm-projects/v1/projects", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+config.Token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get projects: %s\n%s", resp.Status, string(body))
	}

	// JSON wird direkt als Array zurückgegeben, daher verwenden wir eine einfache Slice-Deklaration
	var projects []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return projects, nil
}

// getTasksForProject retrieves tasks for a specific project by its ID using an API call and returns them as a slice of maps.
// It returns an error if the request fails, the response status code is not OK, or the JSON decoding is unsuccessful.
func getTasksForProject(projectID string) ([]map[string]interface{}, error) {
	// Erstellen Sie die API-URL mit der projectId
	apiURL := fmt.Sprintf("%sapi/calm-tasks/v1/tasks?projectId=%s", config.Url, projectID)

	// HTTP-Request erstellen
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+config.Token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	// Fehler abfangen, wenn der Statuscode nicht 200 ist
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get tasks for project %s: %s\n%s", projectID, resp.Status, string(body))
	}

	// JSON-Daten in ein Array von Aufgaben dekodieren
	var tasks []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks response for project %s: %w", projectID, err)
	}

	return tasks, nil
}

// createTaskEntry generates a formatted task entry string and a unique identifier based on the given task data.
// It processes task information including status, priority, tags, due date, and project metadata.
func createTaskEntry(task map[string]interface{}) (string, string) {
	var result string

	result += "- " + getTaskType(task["status"].(string))
	result += " " + getTaskPriority(task["priorityId"].(float64))
	uniqueStr := " [[" + task["displayId"].(string) + "]]"
	result += uniqueStr
	result += " " + task["title"].(string)

	result += "\n  Project:: [[" + task["projectName"].(string) + "@" + config.Name + "]]"

	// get Tags
	if tags, ok := task["tags"].([]interface{}); ok {
		var tagList []string
		for _, tag := range tags {
			if tagStr, valid := tag.(string); valid {
				tagList = append(tagList, "[["+tagStr+"]]")
			}
		}
		if len(tagList) > 0 {
			result += "\n  tags:: " + strings.Join(tagList, ", ")
		}
	} else if tagString, ok := task["tags"].(string); ok {
		result += "\n  tags:: " + tagString
	}

	// get DueDate
	if task["dueDate"] != nil && task["dueDate"].(string) != "" {
		dueDate := logseq.GetScheduledDateFormat(task["dueDate"].(string))
		if dueDate != "" {
			result += "\n  " + dueDate
		}
	}

	return result, uniqueStr
}

// isUserInvolved checks whether the user (from config.UserId) is involved in a task by assigneeId or involvedParties.
func isUserInvolved(task map[string]interface{}) bool {
	assigneeID, ok := task["assigneeId"].(string)
	if ok && assigneeID == config.UserId {
		return true
	}

	involvedParties, ok := task["involvedParties"].(string)
	if !ok {
		return false
	}

	entries := strings.Split(involvedParties, ",")
	for _, entry := range entries {
		if strings.TrimSpace(entry) == config.UserId {
			return true
		}
	}

	return false
}

// getTaskType maps a task status string to a task type string such as TODO, DOING, WAIT, DONE, CLOSED, or UNKNOWN.
func getTaskType(status string) string {
	statusToType := map[string]string{
		"CIPTKOPEN":    "TODO",
		"CIPUSOPEN":    "TODO",
		"CIPREQUOPEN":  "TODO",
		"CIPDFCTOPEN":  "TODO",
		"CIPTKINP":     "DOING",
		"CIPUSINP":     "DOING",
		"CIPREQUINP":   "DOING",
		"CIPDFCTINP":   "DOING",
		"CIPTKBLK":     "WAIT",
		"CIPUSBLK":     "WAIT",
		"CIPREQUBLK":   "WAIT",
		"CIPDFCTBLK":   "WAIT",
		"CIPTKNO":      "CLOSED",
		"CIPUSNO":      "CLOSED",
		"CIPREQUNO":    "CLOSED",
		"CIPTKCLOSE":   "DONE",
		"CIPUSCLOSE":   "DONE",
		"CIPREQUCLOSE": "DONE",
		"CIPDFCTDONE":  "DONE",
	}

	if taskType, ok := statusToType[status]; ok {
		return taskType
	}
	return "UNKNOWN"
}

// getTaskPriority returns the priority type as a string for a given priority ID.
// The mapping is predefined in the function using a map of float64 to string.
func getTaskPriority(priorityId float64) string {
	priorityToType := map[float64]string{
		10: "[#A]",
		20: "[#B]",
		30: "[#C]",
		40: "[#D]",
	}

	return priorityToType[priorityId]
}

// Process retrieves tasks from multiple projects, applies user-specific filters, and returns a list of relevant tasks.
func Process(extConf Config, path string) []map[string]interface{} {
	config = extConf

	var err error
	config.Token, err = getAccessToken(config.ClientId, config.ClientSecret, config.TokenUrl)
	if err != nil {
		panic(err)
	}

	projects, err := getProjects()
	if err != nil {
		panic(err)
	}

	var allUserTasks []map[string]interface{}
	var fileContent string

	for _, project := range projects {
		projectID, ok := project["id"].(string)
		if !ok {
			fmt.Printf("Invalid project ID format: %v\n", project["id"])
			continue
		}
		projectName, _ := project["name"].(string)

		tasks, err := getTasksForProject(projectID)
		if err != nil {
			fmt.Printf("Failed to get tasks for project %s: %v\n", projectID, err)
			continue
		}

		for _, task := range tasks {
			if isUserInvolved(task) {
				task["projectName"] = projectName
				allUserTasks = append(allUserTasks, task)
				taskLine, uniqueStr := createTaskEntry(task)
				fileContent = logseq.AddOrReplaceEntry(uniqueStr, taskLine, fileContent)
			}
		}
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

	return allUserTasks
}
