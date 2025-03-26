/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"Ttracker/internal/config"
	"Ttracker/internal/scan"
	"Ttracker/internal/watcher"

	"github.com/spf13/cobra"
)

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run Ttracker as a background process",
	Long: `Runs Ttracker as a daemon process that continuously monitors 
tracked projects for changes and updates TODO list automatically.

Example:
  tt daemon  # Start Ttracker in daemon mode
`,
	Run: func(cmd *cobra.Command, args []string) {
		runDaemon()
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}

func runDaemon() {
	fmt.Println("Starting Ttracker daemon...")

	// Create required directories
	if err := os.MkdirAll("data", 0755); err != nil {
		fmt.Printf("Error creating data directory: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll("parsers", 0755); err != nil {
		fmt.Printf("Error creating parsers directory: %v\n", err)
		os.Exit(1)
	}

	// Create an empty plugin.json if it doesn't exist
	pluginConfigPath := filepath.Join("parsers", "plugin.json")
	if _, err := os.Stat(pluginConfigPath); os.IsNotExist(err) {
		if err := os.WriteFile(pluginConfigPath, []byte("[]"), 0644); err != nil {
			fmt.Printf("Error creating empty plugin config: %v\n", err)
			os.Exit(1)
		}
	}

	storeFilePath := filepath.Join("data", "todos.json")

	// Create and start the watcher
	w, err := watcher.NewProjectWatcher(storeFilePath, pluginConfigPath)
	if err != nil {
		fmt.Printf("Error creating file watcher: %v\n", err)
		os.Exit(1)
	}

	// Scan all projects before starting the watcher to ensure we have up-to-date TODOs
	fmt.Println("Scanning all projects for TODOs...")
	if err := scanAllProjects(pluginConfigPath, storeFilePath); err != nil {
		fmt.Printf("Error during initial scan: %v\n", err)
		// Continue anyway - this isn't fatal
	}

	if err := w.Start(); err != nil {
		fmt.Printf("Error starting watcher: %v\n", err)
		os.Exit(1)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Ttracker daemon is running. Press Ctrl+C to stop.")

	// Wait for termination signal
	<-sigChan
	fmt.Println("\nShutting down Ttracker daemon...")

	// Stop the watcher
	if err := w.Stop(); err != nil {
		fmt.Printf("Error stopping watcher: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Ttracker daemon stopped.")
}

// scanAllProjects scans all tracked projects to ensure TODOs are up to date
func scanAllProjects(pluginConfigPath, storeFilePath string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	if len(cfg.Projects) == 0 {
		fmt.Println("No projects to scan.")
		return nil
	}

	for name, path := range cfg.Projects {
		fmt.Printf("Scanning project '%s'...\n", name)
		if err := scan.RunScan(path, name, pluginConfigPath, storeFilePath); err != nil {
			fmt.Printf("Error scanning project '%s': %v\n", name, err)
			// Continue with other projects
		}
	}

	return nil
}
