package cmd

import (
	"bufio"
	"errors"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

// rootCmd represents the base command when called without any subcommands
var download = &cobra.Command{
	Use:   "download",
	Short: "Print the version number of Hugo",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		gl := loginGitlab()
		size, err := downloadFile(gl, projectID, jobName, fileName)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(size)
	},
}

func init() {
}

// Get latest job ID
func getJobID(gl *gitlab.Client, projectID int, jobName string) (int, error) {
	opt := &gitlab.ListProjectPipelinesOptions{
		Status: gitlab.BuildState(gitlab.Success),
		Ref:    gitlab.String("master"),
	}

	pipelines, _, err := gl.Pipelines.ListProjectPipelines(projectID, opt)
	if err != nil {
		return -1, err
	}
	pipeline := pipelines[0]

	listJobs, _, err := gl.Jobs.ListPipelineJobs(projectID, pipeline.ID, nil)
	if err != nil {
		return -1, err
	}
	for _, job := range listJobs {
		if job.Name == jobName {
			jobID := job.ID
			return jobID, nil
		}
	}
	return -1, errors.New("Job not found")
}

// save file in current directory
// TODO save all files or single file
// TODO choice saving path
func savedFile(gl *gitlab.Client, projectID, jobID int, fileName string) (int64, error) {
	content, _, err := gl.Jobs.DownloadSingleArtifactsFile(projectID, jobID, fileName)
	if err != nil {
		return -1, err
	}

	file, err := os.Create(fileName)
	if err != nil {
		return -1, err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	countBytes, err := writer.ReadFrom(content)
	if err != nil {
		os.Remove(fileName)
		return -1, err
	}

	writer.Flush()
	return countBytes, nil
}

func downloadFile(gl *gitlab.Client, projectID int, jobName, fileName string) (int64, error) {
	jobID, err := getJobID(gl, projectID, jobName)
	if err != nil {
		return -1, err
	}

	fileSizeBytes, err := savedFile(gl, projectID, jobID, fileName)
	if err != nil {
		log.Fatal(err)
	}
	return fileSizeBytes, nil
}
