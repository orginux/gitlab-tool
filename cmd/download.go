package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

var download = &cobra.Command{
	Use:     "download",
	Short:   "Download a single file or all archive from artifacts",
	Example: "gitlab-tool download --token BebRx.. --project-id 111 --job-name build --file-name file_name.txt --refspec testing",
	Aliases: []string{"d", "dl"},
	Run: func(cmd *cobra.Command, args []string) {
		runDownloadFile(cmd)
	},
}

func init() {
	rootCmd.AddCommand(download)

	dlFlags := download.Flags()

	var jobName string
	dlFlags.StringVarP(&jobName, "job-name", "j", "", "job name")
	download.MarkFlagRequired("job-name")

	var fileName string
	dlFlags.StringVarP(&fileName, "file-name", "f", "", "download the only file which this name")

	var output string
	dlFlags.StringVarP(&output, "dest", "d", "./", "destination directory")

	var createDir bool
	dlFlags.BoolVarP(&createDir, "create-dirs", "c", false, "create necessary local directory hierarchy")

	var keepSourceArchive bool
	dlFlags.BoolVarP(&keepSourceArchive, "keep-src", "", false, "save archive after extract")
	// TODO specifc pipeline status
	// var status string
	// dlFlags.StringVarP(&status, "pipeline-status", "s", "success", "status of pipelines, one of: manual, failed, canceled;")

	var pipelineID int
	dlFlags.IntVarP(&pipelineID, "pipeline-id", "", 0, "Download artifacts from specific pipeline ID")

	var acceptablePipelinesStatus string
	dlFlags.StringVarP(&acceptablePipelinesStatus, "acceptable-status", "a", "", "acceptable pipeline status for download artefacts")

	var extract bool
	dlFlags.BoolVarP(&extract, "extract", "x", false, "extract files from an archive")
}

// get pipeline ID
func getPipelineID(gl *gitlab.Client, projectID int, branch, acceptableStatus string) (int, error) {
	opt := &gitlab.ListProjectPipelinesOptions{
		Status: gitlab.BuildState(gitlab.Success),
		Ref:    gitlab.String(branch),
	}

	pipelines, _, err := gl.Pipelines.ListProjectPipelines(projectID, opt)
	if err != nil {
		return -1, err
	}
	// latest success pipeline
	pipeline := pipelines[0]

	if acceptableStatus != "" {
		switch acceptableStatus {
		case "manual":
			opt.Status = gitlab.BuildState(gitlab.Manual)
		case "failed":
			opt.Status = gitlab.BuildState(gitlab.Failed)
		case "canceled":
			opt.Status = gitlab.BuildState(gitlab.Canceled)
		}

		acceptablePipelines, _, err := gl.Pipelines.ListProjectPipelines(projectID, opt)
		if err != nil {
			return -1, err
		}

		// latest acceptable pipeline
		acceptablePipeline := acceptablePipelines[0]

		if acceptablePipeline.UpdatedAt.After(*pipeline.UpdatedAt) {
			pipeline = acceptablePipelines[0]
		}
	}
	return pipeline.ID, nil
}

// get latest job ID
func getJobID(gl *gitlab.Client, projectID, pipelineID int, jobName string) (int, error) {
	fmt.Println("Find job in pipeline: ", pipelineID)
	listJobs, _, err := gl.Jobs.ListPipelineJobs(projectID, pipelineID, nil)
	if err != nil {
		return -1, err
	}
	for _, job := range listJobs {
		if job.Name == jobName {
			fmt.Println(job)
			jobID := job.ID
			return jobID, nil
		}
	}
	return -1, fmt.Errorf("The job '%s' is not found in pipeline %d", jobName, pipelineID)
}

// download all artifacts in zip archive
func downloadArtifacts(gl *gitlab.Client, projectID int, refspec, jobName, filePath string) error {
	opt := &gitlab.DownloadArtifactsFileOptions{
		Job: &jobName,
	}

	content, _, err := gl.Jobs.DownloadArtifactsFile(projectID, refspec, opt)
	if err != nil {
		return err
	}

	archive, _ := ioutil.ReadAll(content)
	err = ioutil.WriteFile(filePath, archive, 0644)
	if err != nil {
		return err
	}

	return nil
}

// download a file from the artifacts
func downloadFile(gl *gitlab.Client, projectID, jobID int, fileName, filePath string) error {
	content, _, err := gl.Jobs.DownloadSingleArtifactsFile(projectID, jobID, fileName)
	if err != nil {
		return err
	}

	data, _ := ioutil.ReadAll(content)
	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// create a parent directory for file
func createDirForFile(file string) error {
	parentDir := filepath.Dir(file)
	err := os.MkdirAll(parentDir, 0775)
	if err != nil {
		return err
	}
	return nil
}

// save file artifacts on host
func saveArtifacts(gl *gitlab.Client, projectID, jobID int, refspec, jobName, fileName, directory string, createDir bool) (string, error) {
	if fileName != "" {
		// save single file
		filePath := path.Join(directory, fileName)
		if createDir {
			err := createDirForFile(filePath)
			if err != nil {
				return "", err
			}
		} else {
			// if file path in artifacts has subdirectories
			// example: --file-name "test/file.txt"
			err := createDirForFile(fileName)
			if err != nil {
				return "", err
			}
		}

		err := downloadFile(gl, projectID, jobID, fileName, filePath)
		if err != nil {
			return "", err
		}
		return filePath, nil
	}

	// save all files
	filePath := path.Join(directory, "artifacts.zip")

	if createDir {
		err := createDirForFile(filePath)
		if err != nil {
			return "", err
		}
	}

	err := downloadArtifacts(gl, projectID, refspec, jobName, filePath)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

func unzip(src, dest string, keepSourceArchive, verbose bool) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) && f.Name != "./" {
			return filenames, fmt.Errorf("illegal file path: '%s'", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if verbose {
			log.Println("Extract: ", fpath)
		}

		if err != nil {
			return filenames, err
		}
	}

	if !keepSourceArchive {
		err := os.Remove(src)
		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

// General func for download file from artifacts
func runDownloadFile(cmd *cobra.Command) {
	gl := loginGitlab()
	projectID, _ := rootCmd.Flags().GetInt("project-id")
	refspec, _ := rootCmd.Flags().GetString("refspec")

	jobName, _ := cmd.Flags().GetString("job-name")
	fileName, _ := cmd.Flags().GetString("file-name")

	directory, _ := cmd.Flags().GetString("destination")
	createDir, _ := cmd.Flags().GetBool("create-dirs")

	// basicStaus, _ := cmd.Flags().GetString("pipeline-status")
	pipelineID, _ := cmd.Flags().GetInt("pipeline-id")
	acceptableStatus, _ := cmd.Flags().GetString("acceptable-status")

	verbose, _ := cmd.Flags().GetBool("verbose")
	extract, _ := cmd.Flags().GetBool("extract")
	keepSourceArchive, _ := cmd.Flags().GetBool("keep-src")

	if pipelineID == 0 {
		var err error
		pipelineID, err = getPipelineID(gl, projectID, refspec, acceptableStatus)
		if err != nil {
			log.Fatal("Error get pipeline: ", err)
		}
	}

	if verbose {
		log.Printf("Pipeline ID: %d\n", pipelineID)
	}

	jobID, err := getJobID(gl, projectID, pipelineID, jobName)
	if err != nil {
		log.Fatal("Error get job: ", err)
	}
	if verbose {
		log.Printf("Job ID: %d", jobID)
	}

	filePath, err := saveArtifacts(gl, projectID, jobID, refspec, jobName, fileName, directory, createDir)
	if err != nil {
		log.Fatal("Error save: ", err)
	}

	if verbose {
		log.Printf("Downloaded: %s", filePath)
	}

	if extract && fileName == "" {
		_, err := unzip(filePath, directory, keepSourceArchive, verbose)
		if err != nil {
			log.Fatal("Unzip error: ", err)
		}
	}
}
