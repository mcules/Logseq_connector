# Summary
This project is designed to seamlessly synchronize data from various systems into the Logseq graph, creating a cohesive and easy-to-navigate data network. At present, the supported systems include:
- ICS Calendar: Enhances the functionality of Logseq by integrating calendar events from the ICS calendar system. This allows users to have a unified view of their event and task schedule in one place.
- GitLab Issues: Syncs active GitLab issue tracking into the Logseq graph. An ideal feature for project management, it assists developers in maintaining an overview of issue statuses and progresses, all within Logseq.
- Paperless-ngx Documents: Implements the document management system of Paperless-ngx into Logseq. By doing so, it eases access to important documents and notes.

With these data integrations, the project takes a significant step towards making Logseq a more comprehensive tool for developers and individuals looking to streamline their digital workflows. The aim is to offer a multi-dimensional data interaction platform inside Logseq, using data from well-established systems in a structured and user-friendly manner.

# Install

## Build
First, you need to [install Golang](https://go.dev/doc/install) in order to be able to compile the project.
```
git clone https://github.com/mcules/Logseq_connector.git
cd Logseq_connector
go build .
```

## Logseq
### config.edn
In order for the icons to be displayed correctly, the following must be entered in the macro section of your config.edn: `"i" "[:span {:class ti} \"&#x$1 \" ]"`

You can find icons with the associated codes here: [tabler-icons.io](https://tabler-icons.io/)

## Configuration
Create a ***config.json*** file. An example of how it could look is provided below. You can use multiple instances for each system, but it's important that they each have distinct names.

### graph
You can use multiple graphs. In the respective settings you have to specify which graph you would like the data to be written with.

| Variable   | Content        |
|------------|----------------|
| Graph_name | Link to folder |

### calendar
Calendar Events are written to the daily journal file in format: `{{i $CALENDAR_ICON$}} *$EVENT_TIME$* [[$CALENDAR_NAME$]]: [[$EVENT_SUMMARY$]]`

| Variable | Content                          | required |
|----------|----------------------------------|----------|
| name     | Name for your calendar in Logseq | yes      |
| graph    | Which graph should used          | yes      |
| ics      | Link to ics file in web          | yes      |
| icon     | Icon to show                     | yes      |

### gitlab
The gitlab Issues are written to File: `gitlab___$GITLAB_PROJECT_NAME$___tickets.md`

| Variable         | Content                                | default | required |
|------------------|----------------------------------------|---------|----------|
| name             | Name for your namespace in Logseq      |         | yes      |
| graph            | Which graph should used                |         | yes      |
| project          | project name                           |         | yes      |
| url              | url to your gitlab api                 |         | yes      |
| authToken        | your gitlab authToken                  |         | yes      |
| username         | your gitlab username                   |         | yes      |
| sort             | asc / desc                             | desc    | optional |
| state            | opened / closed                        |         | yes      |
| scope            | Scopename / all                        | all     | optional |
| assigneeUsername | Only issues which are assigned to user |         | optional |

### paperless
The paperless documents are written in own files per correspondent: `documents___paperless___$PAPERLESS_CONFIG_NAME$___$DOCUMENT_CORRESPONDENT_NAME$.md`

The document line has the following structure: `icon document_type document_link document_name document_tags`

| Variable | Content                            | required |
|----------|------------------------------------|----------|
| name     | Name for your namespace in Logseq  | yes      |
| graph    | Which graph should used            | yes      |
| username | your paperless sync user           | yes      |
| password | your paperless sync password       | yes      |
| url      | url to your paperless installation | yes      |

### Example
```
{
  "graph": {
    "Privat": "Logseq/Privat/",
    "Work": "Logseq/Work/"
  },
  "calendar": [
    {
      "name": "Privat",
      "graph":"Privat",
      "ics": "https://cloud.private.xyz/remote.php/dav/public-calendars/nlcuiesyknuc4e?export",
      "Icon": "ea53"
    },
    {
      "name": "Work",
      "graph":"Work",
      "ics": "https://cloud.work.xyz/remote.php/dav/public-calendars/nlcuiesyknuc4e?export",
      "Icon": "ea53"
    }
  ],
  "gitlab": [
    {
      "name": "git.work.xyz",
      "graph": "Work",
      "project": "Work",
      "url": "https://git.work.xyz/api/v4",
      "authToken": "MySecureAuthToken",
      "username": "MyUsername",
      "sort": "asc",
      "state": "opened"
    }
  ],
  "paperless": [
    {
      "name": "paperless.private.xyz",
      "graph":"Privat",
      "username": "MyUsername",
      "password": "MySecurePassword",
      "url": "https://paperless.private.xyz/"
    }
  ]
}
```

## Graph
Depending on where you want to run your connector, you will need to ensure that your data is synchronized between your devices. Currently, I am using [Syncthing](https://syncthing.net/) for this purpose. However, you can also use [Nextcloud](https://nextcloud.com/) or any other data synchronization tool of your choice.

## Cronjob
I have the connector set up to run every 15 minutes through a cron job on my system. You can adjust the time as per your requirements.
`*/15 * * * * /opt/Logseq_connector/Logseq_connector /opt/Logseq_connector/`