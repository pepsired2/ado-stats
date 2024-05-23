package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func getActiveDevs(projectName string, devsSet map[string]string) error {
	// This function should return the number of active developers in the project
	// and print the number of active developers in the project.
	username := os.Getenv("USER")
	passwd := os.Getenv("PASS")

	// Declare an empty interface
	var repos map[string]interface{}
	client := &http.Client{}
	url := "https://dev.azure.com/PepsiCoIT/" + projectName + "/_apis/git/repositories?api-version=7.2-preview.1"
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
		url := "https://dev.azure.com/PepsiCoIT/" + projectName + "/_apis/git/repositories?api-version=7.2-preview.1&continuationToken=" + contToken
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
	json.Unmarshal(bodyText, &repos)

	allRepos := repos["value"].([]interface{})
	fmt.Println("    Total number of repos: ", len(allRepos))
	for _, repo := range allRepos {
		if r, ok := repo.(map[string]interface{}); ok {
			fmt.Println(r["name"])
			getDevs(projectName, r["id"].(string), devsSet)
		}
	}
	fmt.Println("\n\n-----------------------------------------")

	return nil
}

func getDevs(projectName string, repoId string, devsSet map[string]string) error {
	username := os.Getenv("USER")
	passwd := os.Getenv("PASS")

	// Declare an empty interface
	var commits map[string]interface{}
	client := &http.Client{}
	url := "https://dev.azure.com/PepsiCoIT/" + projectName + "/_apis/git/repositories/" + repoId + "/commits?searchCriteria.fromDate=04/20/2024 12:00:00 AM&api-version=7.2-preview.2"
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

	json.Unmarshal(bodyText, &commits)

	allCommits := commits["value"].([]interface{})
	fmt.Println("	Total number of commits: ", commits["count"])

	// Create a set using a map with struct{} values

	fmt.Println("-----------------------------------------")
	for _, commit := range allCommits {
		if d, ok := commit.(map[string]interface{}); ok {
			auth := d["author"].(map[string]interface{})
			fmt.Println(auth["email"])
			committer := d["committer"].(map[string]interface{})
			fmt.Println(committer["email"])
			if email, ok := auth["email"].(string); ok && email != "" {
				devsSet[strings.ToLower(email)] = d["remoteUrl"].(string)
			}
			if email, ok := committer["email"].(string); ok && email != "" {
				devsSet[strings.ToLower(email)] = d["remoteUrl"].(string)
			}
		}
	}

	skip := 100
	for int(commits["count"].(float64)) == 100 {
		fmt.Println("******Getting more commits******")
		url := "https://dev.azure.com/PepsiCoIT/" + projectName + "/_apis/git/repositories/" + repoId + "/commits?searchCriteria.$skip=" + strconv.Itoa(skip) + "&searchCriteria.fromDate=04/20/2024 12:00:00 AM&api-version=7.2-preview.2"
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

		json.Unmarshal(bodyText, &commits)

		allCommits = commits["value"].([]interface{})
		fmt.Println("	Total number of commits: ", commits["count"])
		skip += int(commits["count"].(float64))

		fmt.Println("-----------------------------------------")
		for _, commit := range allCommits {
			if d, ok := commit.(map[string]interface{}); ok {
				auth := d["author"].(map[string]interface{})
				fmt.Println(auth["email"])
				committer := d["committer"].(map[string]interface{})
				fmt.Println(committer["email"])
				if email, ok := auth["email"].(string); ok && email != "" {
					devsSet[strings.ToLower(email)] = d["remoteUrl"].(string)
				}
				if email, ok := committer["email"].(string); ok && email != "" {
					devsSet[strings.ToLower(email)] = d["remoteUrl"].(string)
				}
			}
		}
	}
	fmt.Println("	Total number of devs: ", len(devsSet))
	return nil
}
