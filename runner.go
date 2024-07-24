package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
)

func run(scanType string) error {
	projects, err := getProjects()

	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("Total number of projects: ", projects.Count)
	fmt.Println("Scan Type: ", scanType)
	fmt.Println("--------------------------------------------------")
	pipelineStats := PipelineStats{}

	// iterate over all the projects
	devsSet := make(map[string]int)
	repoSet := make(map[string]int)
	for i, project := range projects.Value {
		fmt.Printf("Project %d:\n", i+1)
		fmt.Printf("  Name: %s\n", project.Name)
		if scanType == "pipelines" {
			getPipelines(project.Name, &pipelineStats)
		} else if scanType == "builds" {
			err := getBuilds(project.Name, &pipelineStats)
			if err != nil {
				fmt.Println(err)
				return err
			}
		} else if scanType == "devs" {
			err := getActiveDevs(project.Name, devsSet, repoSet)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
		fmt.Println("--------------------------------------------------")
	}
	if scanType == "devs" {
		fmt.Println("	Total number of devs: ", len(devsSet))
	}

	if len(devsSet) != 0 {
		writeToFile(devsSet, "dev")
	}

	if len(repoSet) != 0 {
		writeToFile(repoSet, "email")
	}

	printStats(pipelineStats)

	return nil
}

func writeToFile(set map[string]int, fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for k, v := range set {
		_, err := writer.WriteString(k + " - " + strconv.Itoa(v) + "\n")
		if err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}
	}
	if err := writer.Flush(); err != nil {
		log.Fatalf("Failed to flush buffer: %v", err)
	}

	fmt.Println("Email addresses written to emails.txt")
}
