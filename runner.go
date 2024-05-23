package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
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
	devsSet := make(map[string]string)
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
			err := getActiveDevs(project.Name, devsSet)
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

	writeToaFile(devsSet)
	printStats(pipelineStats)

	return nil
}

func writeToaFile(devsSet map[string]string) {
	// Create a file to write the email addresses
	file, err := os.Create("emails.txt")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	// Create a buffered writer for efficient writing
	writer := bufio.NewWriter(file)

	// Write each email address to the file
	for email, url := range devsSet {
		_, err := writer.WriteString(email + " - " + url + "\n")
		if err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}
	}

	// Flush the buffer to ensure all data is written
	if err := writer.Flush(); err != nil {
		log.Fatalf("Failed to flush buffer: %v", err)
	}

	fmt.Println("Email addresses written to emails.txt")
}
