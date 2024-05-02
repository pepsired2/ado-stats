package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func getProjects() (ProjectsResponse, error) {
	username := os.Getenv("USER")
	passwd := os.Getenv("PASS")
	client := &http.Client{}
	url := "https://dev.azure.com/PepsiCoIT/_apis/projects?api-version=7.2-preview"

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return ProjectsResponse{}, err
	}

	req.SetBasicAuth(username, passwd)

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		fmt.Println("Error getting projects: ", resp.Status)
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
		fmt.Println("Error unmarshalling projects: ", err)
		return ProjectsResponse{}, err
	}
	return response, nil
}
