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
	"time"

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
	scanType := "pipeline"
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	if len(os.Args) > 1 {
		scanType = os.Args[1]
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
	fmt.Println("Scan Type: ", scanType)
	fmt.Println("--------------------------------------------------")
	pipelineStats := PipelineStats{}

	// iterate over the projects
	for i, project := range projects.Value {
		fmt.Printf("Project %d:\n", i+1)
		fmt.Printf("  Name: %s\n", project.Name)
		if scanType != "build" {
			getPipelines(project.Name, &pipelineStats)
		} else {
			getBuilds(project.Name, &pipelineStats)
		}
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

func getBuilds(projectName string, pipelinestats *PipelineStats) {
	var username string = os.Getenv("USER")
	var passwd string = os.Getenv("PASS")
	var bodyText []byte
	// Declared an empty interface
	var builds map[string]interface{}
	client := &http.Client{}
	url := "https://dev.azure.com/PepsiCoIT/" + projectName + "/_apis/build/builds?api-version=7.2-preview.7&maxBuildsPerDefinition=1&queryOrder=queueTimeDescending"
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	req.SetBasicAuth(username, passwd)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	bodyText, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	// TODO: handle the continuation token later
	contToken := resp.Header.Get("x-ms-continuationtoken")

	for contToken != "" {
		url := "https://dev.azure.com/PepsiCoIT/" + projectName + "/_apis/build/builds?api-version=7.2-preview.7&maxBuildsPerDefinition=1&queryOrder=queueTimeDescending&continuationToken=" + contToken
		req, err := http.NewRequest("GET", url, nil)

		if err != nil {
			fmt.Println(err)
			return
		}
		req.SetBasicAuth(username, passwd)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}
		contToken = resp.Header.Get("x-ms-continuationtoken")
		bt, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		bodyText = append(bodyText, bt...)
		defer resp.Body.Close()
	}

	json.Unmarshal(bodyText, &builds)

	fmt.Println("  Total number of builds: ", builds["count"])

	allBuilds := builds["value"].([]interface{})

	for _, build := range allBuilds {
		var buildId int

		if build, ok := build.(map[string]interface{}); ok {
			if id, exists := build["id"].(float64); exists {
				// Convert id to int if needed
				buildId = int(id)
				fmt.Println("Build ID:", buildId)
			}
			if definition, exists := build["definition"].(map[string]interface{}); exists {
				// Convert id to int if needed
				defId := int(definition["id"].(float64))
				fmt.Println("Pipeline ID:", defId)
			}

			if queueTime, exists := build["queueTime"].(string); exists {
				// Parse the date string into a time.Time object
				parsedTime, err := time.Parse(time.RFC3339Nano, queueTime)
				if err != nil {
					fmt.Println("Error parsing date:", err)
					return
				}
				difference := time.Since(parsedTime)
				sixMonths := 6 * 30 * 24 * time.Hour

				if difference < sixMonths {
					pipelinestats.TotalPipelines++
					result := checkTimeline(buildId, pipelinestats, projectName)
					if result != "" {
						fmt.Println(result)
					}
				} else {
					fmt.Println("The build is 6 months old or older.")
				}
			}
		}
	}
}
