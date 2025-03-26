/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"Ttracker/internal/config"
	"fmt"

	"github.com/spf13/cobra"
)

var projectPath string

// removeCmd represents the remove command
var untrackCmd = &cobra.Command{
	Use:   "untrack [project-name] or -p [project-path]",
	Short: "untrack project and remove all its tracked data",
	Long: `untrack a project from Ttracker and all its tracked data
	USAGE: tt untrack [project-name]
	USAGE: tt untrack -p [project-path]
	`,
	Run: untrackRun,
}

func init() {
	rootCmd.AddCommand(untrackCmd)

	untrackCmd.Flags().StringVarP(&projectPath, "path", "p", "", "pass the absolute path to the project")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func untrackRun(cmd *cobra.Command, args []string) {
	if len(args) <= 0 && projectPath == "" {
		fmt.Println("must pass a project name or an absolute path to a project")
		fmt.Println("USAGE: tt [project-name]")
		fmt.Println("USAGE: tt -p /abs/path/to/project")
		return
	}

	var project string
	if len(args) > 0 {
		project = args[0]
	}

	if project != "" && projectPath != "" {
		fmt.Println("Either a project name or a project path must be supplied, not both")
		fmt.Println("USAGE: tt [project-name]")
		fmt.Println("USAGE: tt -p /abs/path/to/project")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config file", err)
		return
	}

	if projectPath != "" {
		for projectName, path := range cfg.Projects {
			fmt.Println("project name:", projectName, "path:", path)
			if path == projectPath {
				project = projectName
			}
		}
	}

	isExists := cfg.Projects[project]
	if isExists == "" {
		fmt.Printf("Project name \"%s\" is not a registered name for a project.\n", project)
		return
	}

	if cfg.Active == project {
		cfg.Active = ""
	}
	delete(cfg.Projects, project)
	config.SaveConfig(cfg)
	fmt.Println("Project:", project, "has been untrack, all tracking data has been removed.")

}
