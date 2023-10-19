package paperless

import (
	"Logseq_connector/controller/fileFunctions"
	"encoding/json"
	"github.com/kennygrant/sanitize"
	"io"
	"log"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Name     string
	Graph    string
	Username string
	Password string
	Url      string
	Token    string
}

type ResultHead struct {
	Count    int         `json:"count"`
	Next     string      `json:"next"`
	Previous interface{} `json:"previous"`
	All      []int       `json:"all"`
}

type Tags struct {
	ResultHead
	Tags []Tag `json:"results"`
}

type Tag struct {
	ID                int         `json:"id"`
	Slug              string      `json:"slug"`
	Name              string      `json:"name"`
	Colour            int         `json:"colour"`
	Match             string      `json:"match"`
	MatchingAlgorithm int         `json:"matching_algorithm"`
	IsInsensitive     bool        `json:"is_insensitive"`
	IsInboxTag        bool        `json:"is_inbox_tag"`
	DocumentCount     int         `json:"document_count"`
	Owner             interface{} `json:"owner"`
	UserCanChange     bool        `json:"user_can_change"`
}

type Documents struct {
	ResultHead
	Documents []Document `json:"results"`
}

type Document struct {
	ID                  int           `json:"id"`
	Correspondent       int           `json:"correspondent"`
	DocumentType        int           `json:"document_type"`
	StoragePath         interface{}   `json:"storage_path"`
	Title               string        `json:"title"`
	Content             string        `json:"content"`
	Tags                []int         `json:"tags"`
	Created             time.Time     `json:"created"`
	CreatedDate         string        `json:"created_date"`
	Modified            time.Time     `json:"modified"`
	Added               time.Time     `json:"added"`
	ArchiveSerialNumber interface{}   `json:"archive_serial_number"`
	OriginalFileName    string        `json:"original_file_name"`
	ArchivedFileName    string        `json:"archived_file_name"`
	Owner               interface{}   `json:"owner"`
	UserCanChange       bool          `json:"user_can_change"`
	Notes               []interface{} `json:"notes"`
}

type Correspondents struct {
	ResultHead
	Correspondents []Correspondent `json:"results"`
}

type Correspondent struct {
	ID                 int         `json:"id"`
	Slug               string      `json:"slug"`
	Name               string      `json:"name"`
	Match              string      `json:"match"`
	MatchingAlgorithm  int         `json:"matching_algorithm"`
	IsInsensitive      bool        `json:"is_insensitive"`
	DocumentCount      int         `json:"document_count"`
	LastCorrespondence time.Time   `json:"last_correspondence"`
	Owner              interface{} `json:"owner"`
	UserCanChange      bool        `json:"user_can_change"`
}

type DocumentTypes struct {
	ResultHead
	DocumentTypes []DocumentType `json:"results"`
}

type DocumentType struct {
	ID                int         `json:"id"`
	Slug              string      `json:"slug"`
	Name              string      `json:"name"`
	Match             string      `json:"match"`
	MatchingAlgorithm int         `json:"matching_algorithm"`
	IsInsensitive     bool        `json:"is_insensitive"`
	DocumentCount     int         `json:"document_count"`
	Owner             interface{} `json:"owner"`
	UserCanChange     bool        `json:"user_can_change"`
}

var config Config

func Process(extConf Config, path string) {
	config = extConf

	login(config.Username, config.Password, config.Url)

	result := make(map[string][]string)
	documents := documentsGet("")
	tags := tagsGet("")
	correspondents := correspondentsGet("")
	documentTypes := documentTypesGet("")

	for _, doc := range documents {
		var docTags []string
		for _, tag := range doc.Tags {
			thisTag := getTagName(&tags, tag)
			docTags = append(docTags, "[["+thisTag+"]]")
		}

		correspondent := getCorrespondent(&correspondents, doc.Correspondent)

		line := dateToLogseqDate(doc.CreatedDate) + " " +
			getDocType(&documentTypes, doc.DocumentType) + " " +
			getDocLink(doc.ID, doc.Title) + " " +
			strings.Join(docTags[:], " ")

		result[correspondent] = append(result[correspondent], "- "+line)
	}

	for filename, fileContent := range getFiles(result) {
		fileFunctions.WriteFile(fileContent, fileFunctions.GetFilehandle(path+config.Name+"___"+filename+".md"))
	}
}

func login(username string, password string, url string) {
	config.Url = url

	type Token struct {
		Token string
	}

	params := neturl.Values{}
	params.Add("username", username)
	params.Add("password", password)

	resp, err := http.PostForm(config.Url+"api/token/", params)
	if err != nil {
		log.Printf("Request Failed: %s", err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Println(err.Error())
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)

	// Unmarshal result
	tempToken := Token{}
	err = json.Unmarshal(body, &tempToken)
	if err != nil {
		log.Printf("Reading body failed: %s", err)
	}

	config.Token = tempToken.Token
}

func elementsGet(uri string) []byte {
	uri = httpToHttps(uri)
	log.Println("Paperless: " + uri)
	// Create a Bearer string by appending string access token
	var bearer = "Token " + config.Token

	// Create a new request using http
	req, err := http.NewRequest("GET", uri, nil)

	// add authorization header to the req
	req.Header.Add("Authorization", bearer)

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
	}
	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			log.Println(err.Error())
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
	}

	return body
}

func documentsGet(uri string) (documents []Document) {
	if len(uri) == 0 {
		uri = config.Url + "api/documents/?ordering=created&page_size=250&truncate_content=true"
	}

	var temp Documents
	err := json.Unmarshal(elementsGet(uri), &temp)
	if err != nil {
		log.Printf("Reading body failed: %s", err)
	}

	documents = append(documents, temp.Documents...)

	if len(temp.Next) > 0 {
		documents = append(documents, documentsGet(temp.Next)...)
	}

	return documents
}

func tagsGet(uri string) (tags []Tag) {
	if len(uri) == 0 {
		uri = config.Url + "api/tags/?ordering=-added&page_size=250&truncate_content=true"
	}

	var temp Tags
	err := json.Unmarshal(elementsGet(uri), &temp)
	if err != nil {
		log.Printf("Reading body failed: %s", err)
	}

	tags = append(tags, temp.Tags...)

	if len(temp.Next) > 0 {
		tags = append(tags, tagsGet(temp.Next)...)
	}

	return tags
}

func documentTypesGet(uri string) (documentTypes []DocumentType) {
	if len(uri) == 0 {
		uri = config.Url + "api/document_types/?ordering=-added&page_size=250&truncate_content=true"
	}

	var temp DocumentTypes
	err := json.Unmarshal(elementsGet(uri), &temp)
	if err != nil {
		log.Printf("Reading body failed: %s", err)
	}

	documentTypes = append(documentTypes, temp.DocumentTypes...)

	if len(temp.Next) > 0 {
		documentTypes = append(documentTypes, documentTypesGet(temp.Next)...)
	}

	return documentTypes
}

func correspondentsGet(uri string) (correspondents []Correspondent) {
	if len(uri) == 0 {
		uri = config.Url + "api/correspondents/?ordering=-added&page_size=250&truncate_content=true"
	}

	var temp Correspondents
	err := json.Unmarshal(elementsGet(uri), &temp)
	if err != nil {
		log.Printf("Reading body failed: %s", err)
	}

	correspondents = append(correspondents, temp.Correspondents...)

	if len(temp.Next) > 0 {
		correspondents = append(correspondents, correspondentsGet(temp.Next)...)
	}

	return correspondents
}

func httpToHttps(uri string) string {
	if strings.Contains(config.Url, "https://") {
		return strings.ReplaceAll(uri, "http://", "https://")
	}

	return uri
}

func getTagName(tags *[]Tag, id int) string {
	for _, tag := range *tags {
		if tag.ID == id {
			return tag.Name
		}
	}

	return ""
}

func getDocLink(id int, name string) string {
	return "[" + name + "](" + config.Url + "documents/" + strconv.Itoa(id) + "/)"
}

func getDocType(docTypes *[]DocumentType, id int) string {
	for _, tag := range *docTypes {
		if tag.ID == id {
			return "[[" + tag.Name + "]]"
		}
	}

	return ""
}

func getCorrespondent(correspondents *[]Correspondent, id int) string {
	for _, tag := range *correspondents {
		if tag.ID == id {
			return tag.Name
		}
	}

	return ""
}

func dateToLogseqDate(docDate string) string {
	parseTime, err := time.Parse("2006-01-02", docDate)
	if err != nil {
		log.Println(err.Error())
	}

	return parseTime.Format("[[02-01-2006]]")
}

func getFiles(docLines map[string][]string) map[string]string {
	files := make(map[string]string)

	if len(docLines) > 0 {
		for correspondent, lines := range docLines {
			files[sanitize.BaseName(correspondent)] = strings.Join(append([]string{"- Alias:: " + correspondent}, lines...), "\n")
		}
	}

	return files
}
