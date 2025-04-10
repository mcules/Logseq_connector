package logseq

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type Task struct {
	Id         string
	ConfigName string
	Status     string
	Priority   int
	Project    string
	Title      string
	Url        string
	Tags       []string
	DueDate    string
}

// AddOrReplaceEntry searches for an entry starting with "- ", containing `searchStr`,
// replaces it entirely with `insertStr`, or prepends `insertStr` to the file content if `searchStr` is not found.
// An entry starts with "- " and ends either at the beginning of the next entry or the end of the file.
func AddOrReplaceEntry(searchStr string, insertStr string, fileContent string) string {
	lines := strings.Split(fileContent, "\n")
	var newContent []string
	var found bool

	i := 0
	for i < len(lines) {
		line := lines[i]

		// An entry starts with "- "
		if strings.HasPrefix(line, "- ") {
			start := i // Mark the start of the current entry

			// Find the end of the current entry
			j := i + 1
			for j < len(lines) && !strings.HasPrefix(lines[j], "- ") {
				j++
			}

			// Combine all lines of the current entry
			entryBlock := strings.Join(lines[start:j], "\n")

			// Check if the current entry contains `searchStr`
			if strings.Contains(entryBlock, searchStr) {
				// Replace the current entry with `insertStr`
				newContent = append(newContent, insertStr)
				found = true
			} else {
				// If it doesn't match, retain the current entry
				newContent = append(newContent, lines[start:j]...)
			}

			// Skip the processed lines of the current entry
			i = j
		} else {
			// If the line does not belong to an entry, add it as is
			newContent = append(newContent, line)
			i++
		}
	}

	// If no matching entry was found, prepend `insertStr` to the content
	if !found {
		newContent = append([]string{insertStr}, newContent...)
	}

	return strings.Join(newContent, "\n")
}

// GetScheduledDateFormat formats a given date string into the format "SCHEDULED: <YYYY-MM-DD DDD>".
// Returns an empty string if the input date cannot be parsed.
func GetScheduledDateFormat(date string) string {
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		log.Println("Error parsing date:", err)
		return ""
	}

	return fmt.Sprintf("SCHEDULED: <%s %s>", parsedDate.Format("2006-01-02"), parsedDate.Weekday().String()[:3])
}

// GetPrio returns a string representation of a priority based on the provided integer value.
// Valid inputs range from 1 to 4, with a default return value for unrecognized inputs.
func GetPrio(prio int) string {
	switch prio {
	case 1:
		return "[#A]"
	case 2:
		return "[#B]"
	case 3:
		return "[#C]"
	case 4:
		return "[#D]"
	default:
		return "[#B]"
	}
}

func CreateTask(task Task) (string, string) {
	var result string

	result = "- "
	result += task.Status
	result += " " + GetPrio(task.Priority)
	uniqueStr := "[[" + task.Id + "]]"
	result += " " + uniqueStr

	if task.Url != "" {
		result += " " + "[" + task.Title + "](" + task.Url + ")"
	} else {
		result += " " + task.Title
	}

	if task.Project != "" {
		result += "\n  Project:: [[" + task.Project + "@" + task.ConfigName + "]]"
	}

	if len(task.Tags) > 0 {
		result += "\n  tags:: " + strings.Join(task.Tags, ", ")
	}

	if task.DueDate != "" {
		dueDate := GetScheduledDateFormat(task.DueDate)
		if dueDate != "" {
			result += "\n  " + dueDate
		}
	}

	return result, uniqueStr
}
