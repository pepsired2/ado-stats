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
	"strings"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	scanType := "pipeline"

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	if len(os.Args) > 1 {
		scanType = os.Args[1]
	}

	art := `                                       
	(_)                        _                    
	 _ ____ _   _ _____ ____ _| |_ ___   ____ _   _ 
	| |  _ \ | | | ___ |  _ (_   _) _ \ / ___) | | |
	| | | | \ V /| ____| | | || || |_| | |   | |_| |
	|_|_| |_|\_/ |_____)_| |_| \__)___/|_|    \__  |
											  (____/ 
												
	`
	// Replace leading tabs with empty strings
	art = strings.ReplaceAll(art, "\t", "    ")
	fmt.Println(art)

	if err := run(scanType); err != nil {
		fmt.Println(err)
	}
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
