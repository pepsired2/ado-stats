package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func getBuilds(projectName string, pipelinestats *PipelineStats) error {
	username := os.Getenv("USER")
	passwd := os.Getenv("PASS")

	// Declare an empty interface
	var builds map[string]interface{}
	client := &http.Client{}
	url := "https://dev.azure.com/PepsiCoIT/" + projectName + "/_apis/build/builds?api-version=7.2-preview.7&maxBuildsPerDefinition=1&queryOrder=queueTimeDescending"
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println(err)
		return err
	}

	req.SetBasicAuth(username, passwd)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
		return err
	}

	bodyText, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	contToken := resp.Header.Get("x-ms-continuationtoken")

	for contToken != "" {
		url := "https://dev.azure.com/PepsiCoIT/" + projectName + "/_apis/build/builds?api-version=7.2-preview.7&maxBuildsPerDefinition=1&queryOrder=queueTimeDescending&continuationToken=" + contToken
		req, err := http.NewRequest("GET", url, nil)

		if err != nil {
			fmt.Println(err)
			return err
		}
		req.SetBasicAuth(username, passwd)
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			fmt.Println(err)
			return err
		}
		contToken = resp.Header.Get("x-ms-continuationtoken")
		bt, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			fmt.Println(err)
			return err
		}
		bodyText = append(bodyText, bt...)
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
					return err
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
	return nil
}
