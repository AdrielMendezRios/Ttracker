/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"Ttracker/internal/config"
)

// switchCmd represents the switch command
var switchCmd = &cobra.Command{
	Use:   "switch [project]",
	Short: "Switches active project to [project]",
	Long:  `Switch changes the active project to [project] for project specific commands like list`,
	Args:  cobra.ExactArgs(1),
	Run:   switchRun,
}

func init() {
	rootCmd.AddCommand(switchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// switchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// switchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func switchRun(cmd *cobra.Command, args []string) {
	project := args[0]

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config file:", err)
		return
	}
	projects := cfg.Projects
	if exists := cfg.Projects[project]; exists == "" {
		fmt.Println("Error: no project with that name or path is registered. try track command first!")
		return
	}

	cfg.Active = project
	config.SaveConfig(cfg)

	for projectName, _ := range projects {
		if projectName == cfg.Active {
			active := "*" + projectName
			fmt.Println(active)
		} else {
			fmt.Println(projectName)
		}
	}
}
