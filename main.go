package main

import (
	"Logseq_connector/controller/calendar"
	"Logseq_connector/controller/gitlab"
	"Logseq_connector/controller/paperless"
	"encoding/json"
	"github.com/shomali11/util/xconditions"
	"log"
	"os"
)

type Config struct {
	Graph     map[string]string
	Calendar  []calendar.Config
	Gitlab    []gitlab.Config
	Paperless []paperless.Config
}

var config *Config

func main() {
	var path string

	if len(os.Args) > 1 {
		path = os.Args[1] + xconditions.IfThenElse(string(os.Args[1][len(os.Args[1])-1:]) == "/", "", "/").(string)
	}

	getConfig(path + "config.json")

	for _, instance := range config.Calendar {
		calendar.GetCalendar(instance, path+config.Graph[instance.Graph])
	}

	for _, instance := range config.Gitlab {
		gitlab.Process(instance, path+config.Graph[instance.Graph]+"/pages/gitlab___")
	}

	for _, instance := range config.Paperless {
		paperless.Process(instance, path+config.Graph[instance.Graph]+"/pages/documents___paperless___")
	}
}

func getConfig(filename string) {
	f, err := os.ReadFile(filename)
	if err != nil {
		log.Println(err)
	}

	json.Unmarshal(f, &config) //nolint:errcheck
}
