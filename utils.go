package main

import "fmt"

func printStats(ps PipelineStats) {
	fmt.Println("Total Pipelines: ", ps.TotalPipelines)
	fmt.Println("Total Pipelines with Snyk Security Scan Task: ", ps.SnykScansTasks)
	fmt.Println("Total Pipelines with Snyk Security Scan CLI: ", ps.SnykScansCli)
	fmt.Println("Total Pipelines with Fortify Scan Task: ", ps.FortifyScan)
	fmt.Println("Total Pipelines with SonarQube Publish Task: ", ps.SonarQubePublish)
	fmt.Println("Total Pipelines with SonarQube Prepare Task: ", ps.SonarQubePrepare)
	fmt.Println("Total Pipelines with SonarQube Analyze Task: ", ps.SonarQubeAnalyze)
	fmt.Println("--------------------------------------------------")
}
