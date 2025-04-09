package main

import (
	"Logseq_connector/controller/calendar"
	"Logseq_connector/controller/gitlab"
	"Logseq_connector/controller/paperless"
	"Logseq_connector/controller/sapcloudalm"
	"encoding/json"
	"github.com/shomali11/util/xconditions"
	"log"
	"os"
)

type Config struct {
	Graph       map[string]string
	Calendar    []calendar.Config
	Gitlab      []gitlab.Config
	Paperless   []paperless.Config
	SapCloudAlm []sapcloudalm.Config
}

var config *Config

func main() {
	var path string

	if len(os.Args) > 1 {
		path = os.Args[1] + xconditions.IfThenElse(string(os.Args[1][len(os.Args[1])-1:]) == "/", "", "/").(string)
	}

	getConfig(path + "config.json")

	// region Calendar
	for _, instance := range config.Calendar {
		log.Println("get Calendar:", instance.Name)
		calendar.GetCalendar(instance, path+config.Graph[instance.Graph])
	}
	// endregion

	// region gitlab
	for _, instance := range config.Gitlab {
		log.Println("get Gitlab:", instance.Name)
		gitlab.Process(instance, path+config.Graph[instance.Graph]+"/pages/gitlab___")
	}
	// endregion

	// region Paperless
	for _, instance := range config.Paperless {
		log.Println("get Paperless:", instance.Name)
		paperless.Process(instance, path+config.Graph[instance.Graph]+"/pages/documents___paperless___")
	}
	// endregion

	// region SapCloudAlm
	for _, instance := range config.SapCloudAlm {
		log.Println("get SAP Cloud ALM:", instance.Name)
		sapcloudalm.Process(instance, path+config.Graph[instance.Graph]+"/pages/sapcloudalm___")
	}
	// endregion
}

func getConfig(filename string) {
	f, err := os.ReadFile(filename)
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal(f, &config)
	if err != nil {
		panic(err)
	} //nolint:errcheck
}
