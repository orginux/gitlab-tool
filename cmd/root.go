package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "gitlab-tool",
	Short:            "Provides a gitlab command-line tool to interact with GitLab",
	TraverseChildren: true,
	// TODO BashCompletionFunction:
	Version: "v0.8.0",
}

func init() {
	var gitlabURL string
	rootCmd.PersistentFlags().StringVarP(&gitlabURL, "gitlab-url", "u", "", "URL for the GitLab server or use CI_SERVER_URL (default 'gitlab.com')")

	var privateToken string
	rootCmd.PersistentFlags().StringVarP(&privateToken, "token", "t", "", "gitLab personal access token or use GITLAB_PRIVATE_TOKEN")

	var verbose bool
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose mode")

	var projectID int
	rootCmd.PersistentFlags().IntVarP(&projectID, "project-id", "p", 0, "project ID")
	// rootCmd.MarkPersistentFlagRequired("project-id")

	var refspec string
	rootCmd.PersistentFlags().StringVarP(&refspec, "refspec", "r", "master", "branch or tag")

}

// Get GitLab token
// If the 'token' flag is not set, try to read from the environment variable
func getToken() (string, error) {
	if flagToken, _ := rootCmd.Flags().GetString("token"); flagToken != "" {
		return flagToken, nil
	}

	if envToken, exists := os.LookupEnv("GITLAB_PRIVATE_TOKEN"); exists && envToken != "" {
		return envToken, nil
	}

	return "", fmt.Errorf("'token' not set\nUse flag '--token' or set Environment variable 'GITLAB_PRIVATE_TOKEN'")
}

// Get GitLab url
// If the 'gitlab-url' flag is not set, try to read from the environment variable "CI_SERVER_URL" (GitLab CI/CD variable),
// if this variable is not set, the default value is used
func getGitlabURL() string {
	if gitlabURL, _ := rootCmd.Flags().GetString("gitlab-url"); gitlabURL != "" {
		return gitlabURL
	}

	if envUrl, exists := os.LookupEnv("CI_SERVER_URL"); exists && envUrl != "" {
		return envUrl
	}

	// default value
	return "https://gitlab.com/"
}

func loginGitlab() *gitlab.Client {
	privateToken, err := getToken()
	if err != nil {
		log.Fatalf("Token cannot be read: %v", err)
	}

	gitlabURL := getGitlabURL()
	gl, err := gitlab.NewClient(privateToken, gitlab.WithBaseURL(gitlabURL))
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
	}
}
