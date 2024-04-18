package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type PipelineStats struct {
	TotalPipelines   int
	SnykScansTasks   int
	SnykScansCli     int
	FortifyScan      int
	SonarQubePublish int
	SonarQubePrepare int
	SonarQubeAnalyze int
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println(` _                                        
	(_)                        _                    
	 _ ____ _   _ _____ ____ _| |_ ___   ____ _   _ 
	| |  _ \ | | | ___ |  _ (_   _) _ \ / ___) | | |
	| | | | \ V /| ____| | | || || |_| | |   | |_| |
	|_|_| |_|\_/ |_____)_| |_| \__)___/|_|    \__  |
											 (____/ `)

	projects, err := getProjects()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Total number of projects: ", projects.Count)
	pipelineStats := PipelineStats{}

	// iterate over the projects
	for i, project := range projects.Value {

		fmt.Printf("Project %d:\n", i+1)
		fmt.Printf("  Name: %s\n", project.Name)
		getPipelines(project.Name, &pipelineStats)
		fmt.Println("--------------------------------------------------")

	}

	fmt.Println("Total Pipelines: ", pipelineStats.TotalPipelines)
	fmt.Println("Total Pipelines with Snyk Security Scan Task: ", pipelineStats.SnykScansTasks)
	fmt.Println("Total Pipelines with Snyk Security Scan CLI: ", pipelineStats.SnykScansCli)
	fmt.Println("Total Pipelines with Fortify Scan Task: ", pipelineStats.FortifyScan)
	fmt.Println("Total Pipelines with SonarQube Publish Task: ", pipelineStats.SonarQubePublish)
	fmt.Println("Total Pipelines with SonarQube Prepare Task: ", pipelineStats.SonarQubePrepare)
	fmt.Println("Total Pipelines with SonarQube Analyze Task: ", pipelineStats.SonarQubeAnalyze)
	fmt.Println("--------------------------------------------------")
}

func getProjects() (ProjectsResponse, error) {
	var username string = os.Getenv("USER")
	var passwd string = os.Getenv("PASS")
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://dev.azure.com/PepsiCoIT/_apis/projects?api-version=7.2-preview", nil)

	if err != nil {
		panic(err)
	}

	req.SetBasicAuth(username, passwd)
	resp, err := client.Do(req)
	if err != nil {
		return ProjectsResponse{}, err
	}

	bodyText, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return ProjectsResponse{}, err
	}

	var response ProjectsResponse

	err = json.Unmarshal(bodyText, &response)

	if err != nil {
		return ProjectsResponse{}, err
	}
	return response, nil
}

func getPipelines(projectName string, pipelinestats *PipelineStats) []Pipeline {
	var username string = os.Getenv("USER")
	var passwd string = os.Getenv("PASS")
	var pipelines []Pipeline

	client := &http.Client{}
	url := "https://dev.azure.com/PepsiCoIT/" + projectName + "/_apis/pipelines?api-version=7.1-preview.1"
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil
	}

	req.SetBasicAuth(username, passwd)
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}

	bodyText, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var response PipelineResponse

	err = json.Unmarshal(bodyText, &response)
	if err != nil {
		return nil
	}
	// total number of pipelines
	fmt.Println("  Total number of pipelines: ", response.Count)
	pipelinestats.TotalPipelines += response.Count

	// iterate over the pipelines
	for i, pipeline := range response.Value {

		fmt.Printf("  Pipeline %d:\n", i+1)
		fmt.Printf("    Name: %s\n", pipeline.Name)
		fmt.Printf("    ID: %d\n", pipeline.ID)

		analyzePipeline(pipeline.ID, projectName, pipelinestats)

		pipelines = append(pipelines, pipeline)
	}

	return pipelines
}

func analyzePipeline(pipelineID int, projectName string, pipelineStats *PipelineStats) {
	var username string = os.Getenv("USER")
	var passwd string = os.Getenv("PASS")

	client := &http.Client{}
	body := []byte(`{
		"previewRun": "true"
	}`)
	url := "https://dev.azure.com/PepsiCoIT/" + projectName + "/_apis/pipelines/" + strconv.Itoa(pipelineID) + "/preview?api-version=7.1-preview.1"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))

	req.Header.Add("Content-Type", "application/json")

	if err != nil {
		return
	}

	req.SetBasicAuth(username, passwd)
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	bodyText, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return
	}

	var pipelineYaml PipelinesYaml

	err = json.Unmarshal(bodyText, &pipelineYaml)
	if err != nil {
		return
	}
	result := checkYAML(pipelineYaml.YamlString, pipelineStats)
	if result != "" {
		fmt.Println(result)
	}
}
