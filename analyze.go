package main

import (
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
