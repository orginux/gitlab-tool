package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

const (
	// gitlabToken = "osR6f8FqsTdx9xgZSZYf"
	projectID = 17397134
	jobName   = "build_2"
	fileName  = "simple_file.txt"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "gitlab-tools",
	Short:   "CLI for GitLab API",
	Example: "gitlab-tools download --token ***",
	// TODO BashCompletionFunction:
	Version: "v0.3.1",
}

func init() {
	//	cobra.OnInitialize(initConfig)

	var privateToken string
	// TODO get toket from env
	rootCmd.PersistentFlags().StringVarP(&privateToken, "token", "t", "", "GitLab private token")
	rootCmd.MarkPersistentFlagRequired("token")

	rootCmd.AddCommand(download)
}

// func initConfig() {
// }

func loginGitlab() *gitlab.Client {
	privateToken, exists := rootCmd.Flags().GetString("token")
	log.Printf("exists %v", exists)
	gl, err := gitlab.NewClient(privateToken)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	return gl
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Failed to create client: %v", err)
		os.Exit(1)
	}
}
