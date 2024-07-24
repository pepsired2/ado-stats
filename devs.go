package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func getActiveDevs(projectName string, devsSet, repoSet map[string]int) error {
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
			getDevs(repoSet, devsSet, projectName, r["id"].(string), r["name"].(string))
		}
	}
	fmt.Println("\n\n-----------------------------------------")

	return nil
}

func getDevs(repoSet, devsSet map[string]int, projectName, repoId, repoName string) error {
	username := os.Getenv("USER")
	passwd := os.Getenv("PASS")
	oldDevCnt := len(devsSet)
	fromDate := "06%2F23%2F2024%2012%3A00%3A00%20AM"
	apiVersion := "7.1-preview.1"
	baseURL := fmt.Sprintf("https://dev.azure.com/PepsiCoIT/%s/_apis/git/repositories/%s/commits", projectName, repoId)
	defaultBranch, err := getDefaultBranch(projectName, repoId, username, passwd)
	if defaultBranch == "" {
		return nil
	}
	if err != nil {
		fmt.Println(err)
		return err
	}
	skip := 0
	for {
		url := fmt.Sprintf("%s?searchCriteria.fromDate=%s&searchCriteria.itemVersion.version=%s&searchCriteria.$skip=%d&api-version=%s", baseURL, fromDate, defaultBranch, skip, apiVersion)
		commits, err := fetchCommits(url, username, passwd)
		if err != nil {
			fmt.Println("failed to fetch commits")
			return err
		}

		processCommits(commits, devsSet)

		if len(commits) < 100 {
			break
		}

		skip += 100
	}

	fmt.Printf("Total number of devs: %d\n", len(devsSet))
	devs := len(devsSet) - oldDevCnt
	key := projectName + "/" + repoName
	repoSet[key] = devs

	return nil
}

func fetchCommits(url, username, passwd string) ([]interface{}, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(username, passwd)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyText, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return result["value"].([]interface{}), nil
}

func processCommits(commits []interface{}, devsSet map[string]int) {
	fmt.Printf("Total number of commits: %d\n", len(commits))
	fmt.Println("-----------------------------------------")

	for _, commit := range commits {
		if commitMap, ok := commit.(map[string]interface{}); ok {
			if author, ok := commitMap["author"].(map[string]interface{}); ok {
				if email, ok := author["email"].(string); ok && email != "" {
					fmt.Println(email)
					devsSet[strings.ToLower(email)] += 1
				}
			}
		}
	}
}

func getDefaultBranch(projectName, repoId, username, passwd string) (string, error) {
	client := &http.Client{}
	url := "https://dev.azure.com/PepsiCoIT/" + projectName + "/_apis/git/repositories/" + repoId + "?api-version=7.1-preview.1"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(username, passwd)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(bodyText, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if branchName, ok := result["defaultBranch"].(string); ok {
		branchName = GetLastSegment(result["defaultBranch"].(string))
		return branchName, nil
	}
	return "", nil
}

func GetLastSegment(s string) string {
	if idx := strings.LastIndex(s, "/"); idx != -1 {
		return s[idx+1:]
	}
	return s
}
