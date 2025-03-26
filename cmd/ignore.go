package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"Ttracker/internal/config"

	"github.com/spf13/cobra"
)

// ignoreCmd represents the ignore command
var ignoreCmd = &cobra.Command{
	Use:   "ignore [project-name]",
	Short: "Manage ignore patterns for a project",
	Long: `Manage ignore patterns for a project. These patterns determine which files and directories
should be ignored when scanning for TODOs.

The ignore patterns are stored in a .ttignore file in the project root directory.
Patterns can be specified in a format similar to .gitignore:

  # Ignore specific files
  *.log
  *.tmp
  .DS_Store

  # Ignore directories
  node_modules/
  vendor/
  dist/

  # Ignore specific paths
  build/out/
  test/data/

If no project name is provided or if "." is used, the command will use the default project.

Example:
  tt ignore                    # Show patterns for default project
  tt ignore .                  # Show patterns for default project
  tt ignore my-project         # Show patterns for specific project
  tt ignore --add "*.log"      # Add pattern to default project
  tt ignore . --add "*.log"    # Add pattern to default project
  tt ignore my-project --add "*.log"  # Add pattern to specific project`,
	Run: ignoreRun,
}

func init() {
	rootCmd.AddCommand(ignoreCmd)

	ignoreCmd.Flags().StringP("add", "a", "", "Add a pattern to ignore")
	ignoreCmd.Flags().StringP("remove", "r", "", "Remove a pattern")
	ignoreCmd.Flags().BoolP("list", "l", false, "List all ignore patterns")
	ignoreCmd.Flags().BoolP("clear", "c", false, "Clear all ignore patterns")
}

func ignoreRun(cmd *cobra.Command, args []string) {
	// Load config to get project path
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// Determine which project to use
	var projectName string
	var projectPath string

	if len(args) == 0 || args[0] == "." {
		// Use default project
		if cfg.Active == "" {
			fmt.Println("Error: No default project set. Please specify a project name or set a default project.")
			return
		}

		// Find project name for the active path
		for name, path := range cfg.Projects {
			if path == cfg.Active {
				projectName = name
				projectPath = path
				break
			}
		}

		if projectName == "" {
			fmt.Println("Error: Default project not found in config")
			return
		}
	} else {
		// Use specified project
		projectName = args[0]
		path, exists := cfg.Projects[projectName]
		if !exists {
			fmt.Printf("Error: Project '%s' not found\n", projectName)
			return
		}
		projectPath = path
	}

	ignoreFile := filepath.Join(projectPath, ".ttignore")

	// Ensure the project directory exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		fmt.Printf("Error: Project directory '%s' does not exist\n", projectPath)
		return
	}

	// Read existing patterns
	patterns := make([]string, 0)
	if content, err := os.ReadFile(ignoreFile); err == nil {
		patterns = strings.Split(string(content), "\n")
		// Remove empty lines and comments
		patterns = filterPatterns(patterns)
	}

	// Handle commands
	isAdd := cmd.Flags().Lookup("add").Changed
	isRemove := cmd.Flags().Lookup("remove").Changed
	isList := cmd.Flags().Lookup("list").Changed
	isClear := cmd.Flags().Lookup("clear").Changed

	if isAdd {
		pattern, _ := cmd.Flags().GetString("add")
		if pattern == "" {
			fmt.Println("Error: Pattern is required for --add")
			return
		}
		patterns = append(patterns, pattern)
		fmt.Printf("Added pattern: %s\n", pattern)
	}

	if isRemove {
		pattern, _ := cmd.Flags().GetString("remove")
		if pattern == "" {
			fmt.Println("Error: Pattern is required for --remove")
			return
		}
		patterns = removePattern(patterns, pattern)
		fmt.Printf("Removed pattern: %s\n", pattern)
	}

	if isClear {
		patterns = make([]string, 0)
		fmt.Println("Cleared all ignore patterns")
	}

	if isList || (!isAdd && !isRemove && !isClear) {
		if len(patterns) == 0 {
			fmt.Println("No ignore patterns defined")
			return
		}
		fmt.Printf("Current ignore patterns for project '%s':\n", projectName)
		for _, pattern := range patterns {
			fmt.Printf("  %s\n", pattern)
		}
		return
	}

	// Save patterns back to file
	content := strings.Join(patterns, "\n")
	if err := os.WriteFile(ignoreFile, []byte(content), 0644); err != nil {
		fmt.Printf("Error saving ignore patterns: %v\n", err)
		return
	}
}

func filterPatterns(patterns []string) []string {
	filtered := make([]string, 0)
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern != "" && !strings.HasPrefix(pattern, "#") {
			filtered = append(filtered, pattern)
		}
	}
	return filtered
}

func removePattern(patterns []string, pattern string) []string {
	filtered := make([]string, 0)
	for _, p := range patterns {
		if p != pattern {
			filtered = append(filtered, p)
		}
	}
	return filtered
}
