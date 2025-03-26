/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"Ttracker/internal/config"
	"fmt"

	"github.com/spf13/cobra"
)

// activeCmd represents the active command
var activeCmd = &cobra.Command{
	Use:   "active [project-name]",
	Short: "View or set the active project",
	Long: `View the currently active project or set a new active project.
	
If no project name is specified, this command will display the current active project
and a list of all tracked projects.

Examples:
  tt active               # Show the current active project
  tt active "My Project"  # Set "My Project" as the active project
`,
	Run: activeRun,
}

func init() {
	rootCmd.AddCommand(activeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// activeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// activeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func activeRun(cmd *cobra.Command, args []string) {
	// If a project name is provided, set it as active
	if len(args) > 0 {
		projectName := args[0]
		if err := setActiveProject(projectName); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Active project set to: %s\n", projectName)

		// Load the config to get the path for confirmation
		cfg, _ := config.LoadConfig()
		fmt.Printf("Project path: %s\n", cfg.Active)
		return
	}

	// Otherwise, show the current active project
	showActiveProject()
}

// setActiveProject sets the specified project as active
func setActiveProject(projectName string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config file: %v", err)
	}

	// Check if the project exists
	projectPath, exists := cfg.Projects[projectName]
	if !exists {
		fmt.Println("Available projects:")
		for name := range cfg.Projects {
			fmt.Printf("  - %s\n", name)
		}
		return fmt.Errorf("no project with name '%s' is registered. Use 'tt track' first", projectName)
	}

	// Set the active project using its path
	cfg.Active = projectPath
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("error saving config: %v", err)
	}

	return nil
}

// showActiveProject displays the current active project and all tracked projects
func showActiveProject() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config file: %v\n", err)
		return
	}

	if len(cfg.Projects) == 0 {
		fmt.Println("No projects are currently being tracked.")
		fmt.Println("Use 'tt track <path>' to start tracking a project.")
		return
	}

	// Find active project name
	activeName := "None"
	for name, path := range cfg.Projects {
		if path == cfg.Active {
			activeName = name
			break
		}
	}

	fmt.Printf("Active project: %s\n\n", activeName)
	fmt.Println("All tracked projects:")

	// Print all projects, highlighting the active one
	for name, path := range cfg.Projects {
		if path == cfg.Active {
			fmt.Printf("  * %s (%s)\n", name, path)
		} else {
			fmt.Printf("    %s (%s)\n", name, path)
		}
	}
}
