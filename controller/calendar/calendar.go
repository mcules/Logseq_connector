package calendar

import (
	"Logseq_connector/controller/fileFunctions"
	"fmt"
	"github.com/apognu/gocal"
	"golang.org/x/text/language"
	"golang.org/x/text/search"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	Name  string
	Graph string
	Ics   string
	Icon  string
}

var config Config

func GetCalendar(extConf Config, path string) {
	config = extConf

	log.Println("get Calendar: " + config.Name)
	icsName := config.Name + ".ics"

	filename := path + "journals/"

	err := downloadFile(icsName, config.Ics)
	if err != nil {
		panic(err)
	}

	f, _ := os.Open(icsName)
	defer f.Close()

	start, end := time.Now().Add(-(24 * time.Hour)), time.Now().Add(24*time.Hour)

	c := gocal.NewParser(f)
	c.Start, c.End = &start, &end

	if err := c.Parse(); err != nil {
		log.Println(err.Error())
	}

	for _, e := range c.Events {
		fileHandle := fileFunctions.GetFilehandle(filename + e.Start.Format("2006_01_02.md"))

		fileContent := fileFunctions.GetFileContent(fileHandle)

		_, found := searchInString(fileContent, e.Summary)
		if !found {
			addToFile(fileHandle, config.Icon+" [["+config.Name+"]]: [["+e.Summary+"]]", strings.ReplaceAll(e.Description, "\\n", "\n"))
		}

		fileHandle.Close()
	}
	os.Remove(icsName)
}

func downloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func addToFile(fileHandle *os.File, summary string, description string) {
	var desc string
	if len(description) > 0 {
		desc = "\n  - " + description
	}

	if _, err := fileHandle.WriteString(emptyLine(fileHandle) + summary + desc); err != nil {
		log.Println(err)
	}
}

func searchInString(fileContent string, searchString string) (int, bool) {
	m := search.New(language.English, search.IgnoreCase)
	index, _ := m.IndexString(fileContent, searchString)
	if index == -1 {
		return index, false
	}

	return index, true
}

func emptyLine(fileHandle *os.File) string {
	lastLine := getLastLineWithSeek(fileHandle)

	fileInfo, _ := fileHandle.Stat()
	if fileInfo.Size() == 0 {
		return "- "
	}

	if len(lastLine) == 1 {
		return " "
	}

	if len(lastLine) > 0 {
		return "\n- "
	}

	return "- "
}

func getLastLineWithSeek(fileHandle *os.File) string {
	line := ""
	var cursor int64 = 0
	stat, _ := fileHandle.Stat()

	if stat.Size() > 0 {
		for {
			cursor -= 1
			fileHandle.Seek(cursor, io.SeekEnd)

			char := make([]byte, 1)
			fileHandle.Read(char)

			if cursor != -1 && (char[0] == 10 || char[0] == 13) { // stop if we find a line
				break
			}

			line = fmt.Sprintf("%s%s", string(char), line) // there is more efficient way

			if cursor == -stat.Size() { // stop if we are at the begining
				break
			}
		}
	}

	return line
}
