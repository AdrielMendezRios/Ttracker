/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"text/tabwriter"

	"Ttracker/internal/config"
)

// projectsCmd represents the projects command
var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "prints all project currently registered on Ttracker",
	Long: `prints all projects currently registered on Ttracker`,
	Run: projectsRun,
}

func init() {
	rootCmd.AddCommand(projectsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// projectsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// projectsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func projectsRun(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config file:", err)
		return
	}

	fmt.Println("tracked projects:")
	projects := cfg.Projects
	activeProject := cfg.Active
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for projectName, path := range projects {
		//fmt.Printf("project name: %s, active project: %s\n", projectName, activeProject)
		if projectName == activeProject {
			fmt.Fprintf(w, "active\t%s\t%s\n", projectName, path)
		} else {
			fmt.Fprintf(w, "\t%s\t%s\n", projectName, path)
		}
	}
	w.Flush()
}
