package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

const (
	layoutISO = time.RFC3339
)

var commitCmd = &cobra.Command{
	Use:     "commit",
	Short:   "Interaction with commits",
	Example: "",
	// Aliases: []string{""},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use one of the subcommands")
		cmd.Help()
	},
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "Get list commits",
	Example: "",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		commitList(cmd)
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
	commitCmd.AddCommand(listCmd)

	var message string
	commitCmd.PersistentFlags().StringVarP(&message, "message", "m", "", "commit message")

	var allCommits bool
	listCmd.PersistentFlags().BoolVarP(&allCommits, "all", "a", false, "retrieve every commit from the repository")

	var filePath string
	listCmd.PersistentFlags().StringVarP(&filePath, "file", "f", "", "commits for this file")

	var long bool
	listCmd.PersistentFlags().BoolVarP(&long, "long", "l", false, "all info about commits in output")

	var firstParent bool
	listCmd.PersistentFlags().BoolVarP(&firstParent, "first-parent", "", false, "follow only the first parent commit upon seeing a merge commit")

	var withStats bool
	listCmd.PersistentFlags().BoolVarP(&withStats, "with-stats", "", false, "Stats about each commit will be added to the response")

	var since string
	listCmd.PersistentFlags().StringVarP(&since, "since", "", "", "Only commits after or on this date will be returned in ISO 8601 format YYYY-MM-DDTHH:MM:SSZ")

	var until string
	listCmd.PersistentFlags().StringVarP(&until, "until", "", "", "Only commits before or on this date will be returned in ISO 8601 format YYYY-MM-DDTHH:MM:SSZ")
}

// Parse date from string and insert in listOpt struct
func setDateInListOpt(listOpt *gitlab.ListCommitsOptions, since, until string) error {
	if since != "" {
		sinceDate, err := time.Parse(layoutISO, since)
		if err != nil {
			return fmt.Errorf("error convert date 'since': %v\n", err)
		}
		listOpt.Since = gitlab.Time(sinceDate)
	}
	if until != "" {
		untilDate, err := time.Parse(layoutISO, until)
		if err != nil {
			return fmt.Errorf("error convert date 'until': %v\n", err)
		}
		listOpt.Until = gitlab.Time(untilDate)
	}
	return nil
}

func commitList(cmd *cobra.Command) {
	gl := loginGitlab()
	projectID, _ := rootCmd.Flags().GetInt("project-id")
	refspec, _ := rootCmd.Flags().GetString("refspec")

	allCommits, _ := cmd.Flags().GetBool("all")
	filePath, _ := cmd.Flags().GetString("file")
	long, _ := cmd.Flags().GetBool("long")
	firstParent, _ := cmd.Flags().GetBool("first-parent")
	since, _ := cmd.Flags().GetString("since")
	until, _ := cmd.Flags().GetString("until")
	withStats, _ := cmd.Flags().GetBool("with-stats")

	// GitLab API docs: https://docs.gitlab.com/ce/api/commits.html#list-repository-commits
	listOpt := &gitlab.ListCommitsOptions{
		RefName:     gitlab.String(refspec),
		All:         gitlab.Bool(allCommits),
		Path:        gitlab.String(filePath),
		WithStats:   gitlab.Bool(withStats),
		FirstParent: gitlab.Bool(firstParent),
	}

	err := setDateInListOpt(listOpt, since, until)
	if err != nil {
		log.Fatal(err)
	}

	listCommits, _, err := gl.Commits.ListCommits(projectID, listOpt)
	if err != nil {
		log.Fatal(err)
	}

	for _, commit := range listCommits {
		if !long {
			if withStats {
				fmt.Printf("Commit: %v\nAutor: %s <%s>\nDate: %s\nMessage: %v\nWebURL: %v\nStats:\n\tAdditions: %d\n\tDeletions %d\n\tTotal: %d\n\n", commit.ID, commit.AuthorName, commit.AuthorEmail, commit.CommittedDate, commit.Message, commit.WebURL, commit.Stats.Additions, commit.Stats.Deletions, commit.Stats.Total)
			} else {
				fmt.Printf("Commit: %v\nAutor: %s <%s>\nDate: %s\nMessage: %v\nWebURL: %v\n\n", commit.ID, commit.AuthorName, commit.AuthorEmail, commit.CommittedDate, commit.Message, commit.WebURL)
			}
		} else {
			fmt.Println(commit)
		}
	}
}
