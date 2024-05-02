package main

import "fmt"

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
	for i, project := range projects.Value {
		fmt.Printf("Project %d:\n", i+1)
		fmt.Printf("  Name: %s\n", project.Name)
		if scanType != "build" {
			getPipelines(project.Name, &pipelineStats)
		} else {
			err := getBuilds(project.Name, &pipelineStats)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
		fmt.Println("--------------------------------------------------")
	}

	printStats(pipelineStats)

	return nil
}
