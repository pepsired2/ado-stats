package main

import "time"

// Project represents the JSON data structure
type Project struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	URL            string    `json:"url"`
	State          string    `json:"state"`
	Revision       int       `json:"revision"`
	Visibility     string    `json:"visibility"`
	LastUpdateTime time.Time `json:"lastUpdateTime"`
}

// ProjectsResponse represents the JSON response containing multiple projects
type ProjectsResponse struct {
	Count int       `json:"count"`
	Value []Project `json:"value"`
}

type PipelinesYaml struct {
	YamlString string `json:"finalYaml"`
}

type Pipeline struct {
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Web struct {
			Href string `json:"href"`
		} `json:"web"`
	} `json:"_links"`
	URL      string `json:"url"`
	ID       int    `json:"id"`
	Revision int    `json:"revision"`
	Name     string `json:"name"`
	Folder   string `json:"folder"`
}

type PipelineResponse struct {
	Count int        `json:"count"`
	Value []Pipeline `json:"value"`
}

type PipelineStats struct {
	TotalPipelines     int
	SnykScansTasks     int
	SnykScansCli       int
	SnykContainerTasks int
	FortifyScan        int
	SonarQubePublish   int
	SonarQubePrepare   int
	SonarQubeAnalyze   int
}
