/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"Ttracker/internal/config"
	plugin "Ttracker/internal/plugins"
	"Ttracker/internal/scan"

	"github.com/spf13/cobra"
)

var projectName string

// trackCmd represents the track command
var trackCmd = &cobra.Command{
	Use:   "track [-n <project name>] [path/to/project]",
	Short: "Track the current directory or a passed in path.",
	Long: `Track registers a project directory so that Ttracker can scan it for TODO and FIXME comments.
	USAGE: tt track or tt track [path/to/project]
	`,
	Run: func(cmd *cobra.Command, args []string) {
		directory := args[0]
		trackRun(directory, projectName)
	},
}

func init() {
	trackCmd.Flags().StringVarP(&projectName, "name", "n", "", "Optional name for the project")
	rootCmd.AddCommand(trackCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// trackCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// trackCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func trackRun(dir, name string) {
	if dir == "." {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Error: unable to get current directory:", err)
			return
		}
		dir = cwd
	}

	// validate if the directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("Error: directory does not exist")
		return
	}

	//get absolute path
	absPath, err := filepath.Abs(dir)
	if err != nil {
		fmt.Println("Error resolving absolute path:", err)
	}

	// load existing tracked projects
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Error attempting to load config file:", err)
		return
	}

	if cfg.Projects == nil {
		cfg.Projects = make(map[string]string)
	}

	// if no optional name was given use directory path
	if name == "" {
		name = filepath.Base(dir)
	}

	// check if optional name is already taken
	if _, exists := cfg.Projects[name]; exists {
		fmt.Println("Error: Project name already exists. Choose a different name.")
		return
	}

	// check if already tracked
	for _, path := range cfg.Projects {
		if path == absPath {
			fmt.Println("Project is already being tracked:", absPath)
			return
		}
	}

	// add new project
	cfg.Projects[name] = absPath

	// if no project is active make the new one the active project
	if cfg.Active == "" {
		cfg.Active = absPath
	}

	// save config
	if err := config.SaveConfig(cfg); err != nil {
		fmt.Println("Error registering project:", err)
		return
	}
	fmt.Println("Tracking new project:", absPath)

	// Create the data directory if it doesn't exist
	if err := os.MkdirAll("data", 0755); err != nil {
		fmt.Printf("Error creating data directory: %v\n", err)
		return
	}

	// // Make sure the parsers directory exists
	// if err := os.MkdirAll("parsers", 0755); err != nil {
	// 	fmt.Printf("Error creating parsers directory: %v\n", err)
	// 	return
	// }

	// // Create an empty plugin.json if it doesn't exist
	// pluginConfigPath := filepath.Join("parsers", "plugin.json")
	// if _, err := os.Stat(pluginConfigPath); os.IsNotExist(err) {
	// 	if err := os.WriteFile(pluginConfigPath, []byte("[]"), 0644); err != nil {
	// 		fmt.Printf("Error creating empty plugin config: %v\n", err)
	// 		return
	// 	}
	// }

	storeFilePath := filepath.Join("data", "todos.json")

	// Check if daemon is running and let the user know they don't need to scan manually
	if isDaemonRunning() {
		fmt.Println("Daemon is running: The project will be scanned automatically.")
		fmt.Println("You can view TODOs with: tt list")
		return
	}

	// If daemon is not running, scan the project immediately
	fmt.Println("Scanning project for TODOs...")
	err = scan.RunScan(absPath, name, plugin.PluginConfigsPath, storeFilePath)
	if err != nil {
		fmt.Printf("Error scanning project %s: %v\n", absPath, err)
	}

	fmt.Println("To continuously monitor this project, start the daemon with: tt daemon")
}

// Check if the daemon process is running
func isDaemonRunning() bool {
	// This is a simple implementation - in a production app you might
	// use a PID file or another IPC mechanism for more robust detection

	// For now, we'll just check if there's a running process with "tt daemon" in the command line
	cmd := exec.Command("pgrep", "-f", "tt daemon")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
