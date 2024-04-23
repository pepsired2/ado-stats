package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func checkYAML(yamlString string, pipelineStats *PipelineStats) string {
	lines := strings.Split(yamlString, "\n")
	foundSnykTask, foundSnykCli := false, false
	var debugInfo string

	for _, line := range lines {
		if strings.Contains(line, "SnykSecurityScan") && !strings.Contains(line, "#") {
			debugInfo += "\n- Snyk SCA scan task is present"

			foundSnykTask = true

			pipelineStats.SnykScansTasks += 1
		}

		if strings.Contains(line, "snyk test") && !strings.Contains(line, "#") {
			debugInfo += "\n- Snyk SCA scan command is present"
			foundSnykTask = true

			pipelineStats.SnykScansCli += 1
		}

		if strings.Contains(line, "FortifyScanCentralSAST") && !strings.Contains(line, "#") {
			debugInfo += "\n- Fortify scan task is present"

			pipelineStats.FortifyScan += 1
		}

		if strings.Contains(line, "SonarQubePublish") && !strings.Contains(line, "#") {
			debugInfo += "\n- SonarQube publish task is present"

			pipelineStats.SonarQubePublish += 1
		}

		if strings.Contains(line, "SonarQubePrepare") && !strings.Contains(line, "#") {
			debugInfo += "\n- SonarQube prepare task is present"
			pipelineStats.SonarQubePrepare += 1
		}
		if strings.Contains(line, "SonarQubeAnalyze") && !strings.Contains(line, "#") {
			debugInfo += "\n- SonarQube analyze task is present"
			pipelineStats.SonarQubeAnalyze += 1
		}

		if foundSnykTask {
			if strings.Contains(line, "failOnIssues: true") && !strings.Contains(line, "#") {
				debugInfo += "\n" + "- Snyk scan task is configured OK"

			}
			if strings.HasPrefix(line, "-") {
				debugInfo += "\n" + "- Snyk scan task is NOT OK -- please set failOnIssues: true"
				foundSnykTask = false
			}
		}
		if foundSnykCli {
			if strings.Contains(line, "continueOnError: false") && !strings.Contains(line, "#") {
				debugInfo += "\n" + "- Snyk scan cli is configured OK"

			}
			if strings.HasPrefix(line, "-") {
				debugInfo += "\n" + "- Snyk scan cli is NOT OK -- please set continueOnError: false"
				foundSnykTask = false
			}
		}
	}
	if foundSnykTask {
		debugInfo += "\n" + "- Snyk scan task is NOT OK -- please set failOnIssues: true"
	}
	if foundSnykCli {
		debugInfo += "\n" + "- Snyk scan cli is NOT OK -- please set continueOnError: false"
	}
	return debugInfo
}

func checkTimeline(buildId int, pipelineStats *PipelineStats, projectName string) string {
	var username string = os.Getenv("USER")
	var passwd string = os.Getenv("PASS")

	// Declared an empty interface
	var tasks map[string]interface{}
	client := &http.Client{}
	var debugInfo string

	url := "https://dev.azure.com/PepsiCoIT/" + projectName + "/_apis/build/builds/" + strconv.Itoa(buildId) + "/timeline/" + strconv.Itoa(buildId) + "?api-version=7.2-preview.2"
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return ""
	}

	req.SetBasicAuth(username, passwd)
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}

	bodyText, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return ""
	}
	json.Unmarshal(bodyText, &tasks)

	if tasks["records"] == nil {
		return ""
	}
	allTasks := tasks["records"].([]interface{})

	for _, task := range allTasks {
		if t, ok := task.(map[string]interface{}); ok {
			// fmt.Println(t["name"])
			if adoTask, ok := t["task"].(map[string]interface{}); ok {
				if adoTask["id"] == "55a97a52-a238-46af-9de9-e4245ab45e72" {
					debugInfo += "\n- Fortify scan task is present"
					pipelineStats.FortifyScan += 1
				}

				if adoTask["id"] == "15b84ca1-b62f-4a2a-a403-89b77a063157" {
					debugInfo += "\n- SonarQube prepare task is present"
					pipelineStats.SonarQubePrepare += 1
				}

				if adoTask["id"] == "291ed61f-1ee4-45d3-b1b0-bf822d9095ef" {
					debugInfo += "\n- SonarQube prepare task is present"
					pipelineStats.SonarQubePublish += 1
				}

				if adoTask["id"] == "826d5fe9-3983-4643-b918-487964d7cc87" {
					debugInfo += "\n- Snyk scan task is present"
					pipelineStats.SnykScansTasks += 1
				}
			}
		}
	}
	return debugInfo
}
