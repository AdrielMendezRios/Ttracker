/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
// listCommand()
// 1. Load active project from storage
// 2. If no active project, print message and exit
// 3. Scan all files in the active project directory
// 4. Read each file line-by-line
// 5. If line contains 'TODO:', store it with file path and line number
// 6. Print the collected TODOs in a readable format
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"Ttracker/internal/config"
	"Ttracker/internal/scan"
	"Ttracker/internal/store"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	allProjects bool
	treeView    bool
	forceScan   bool
)

var (
	bold   = color.New(color.Bold).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list [project-name]",
	Short: "List TODOs for the active or specified project",
	Long: `List displays all TODOs for the active project or a specified project.
	
If no project is specified, the active project is used. You can also display
TODOs for all projects using the --all flag.

Example:
  tt list                   # List TODOs for active project
  tt list "My Project"      # List TODOs for a specific project
  tt list --all             # List TODOs for all projects
  tt list --rescan          # Force a scan before listing
`,
	Run: listRun,
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Add flags
	listCmd.Flags().BoolVarP(&allProjects, "all", "a", false, "List TODOs for all projects")
	listCmd.Flags().BoolVarP(&treeView, "tree", "t", true, "Display TODOs in a tree view (default)")
	listCmd.Flags().BoolVarP(&forceScan, "rescan", "r", false, "Force a scan before listing TODOs")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func listRun(cmd *cobra.Command, args []string) {
	// Define paths
	storeFilePath := filepath.Join("data", "todos.json")
	pluginConfigPath := filepath.Join("parsers", "plugin.json")

	// Ensure directories and files exist
	if err := os.MkdirAll("data", 0755); err != nil {
		fmt.Printf("Error creating data directory: %v\n", err)
		return
	}

	if err := os.MkdirAll("parsers", 0755); err != nil {
		fmt.Printf("Error creating parsers directory: %v\n", err)
		return
	}

	// Create an empty plugin.json if it doesn't exist
	if _, err := os.Stat(pluginConfigPath); os.IsNotExist(err) {
		if err := os.WriteFile(pluginConfigPath, []byte("[]"), 0644); err != nil {
			fmt.Printf("Error creating empty plugin config: %v\n", err)
			return
		}
	}

	// Force scan if requested
	if forceScan {
		if len(args) > 0 {
			// Scan specific project
			projectName := args[0]
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Printf("Error loading config: %v\n", err)
				return
			}

			path, exists := cfg.Projects[projectName]
			if !exists {
				fmt.Printf("Project '%s' not found\n", projectName)
				return
			}

			fmt.Printf("Scanning project '%s'...\n", projectName)
			if err := scan.RunScan(path, projectName, pluginConfigPath, storeFilePath); err != nil {
				fmt.Printf("Error scanning project: %v\n", err)
				return
			}
		} else if allProjects {
			// Scan all projects
			fmt.Println("Scanning all projects...")
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Printf("Error loading config: %v\n", err)
				return
			}

			for name, path := range cfg.Projects {
				fmt.Printf("Scanning project '%s'...\n", name)
				if err := scan.RunScan(path, name, pluginConfigPath, storeFilePath); err != nil {
					fmt.Printf("Error scanning project '%s': %v\n", name, err)
					// Continue with other projects
				}
			}
		} else {
			// Scan active project
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Printf("Error loading config: %v\n", err)
				return
			}

			if cfg.Active == "" {
				fmt.Println("No active project set")
				return
			}

			// Find project name
			projectName := ""
			for name, path := range cfg.Projects {
				if path == cfg.Active {
					projectName = name
					break
				}
			}

			if projectName == "" {
				fmt.Println("Active project not found in config")
				return
			}

			fmt.Printf("Scanning active project '%s'...\n", projectName)
			if err := scan.RunScan(cfg.Active, projectName, pluginConfigPath, storeFilePath); err != nil {
				fmt.Printf("Error scanning project: %v\n", err)
				return
			}
		}
	}

	// Load the store
	if _, err := os.Stat(storeFilePath); os.IsNotExist(err) {
		fmt.Println("No TODOs found. Use 'tt track' to track a project first.")
		return
	}

	st, err := store.LoadStore(storeFilePath)
	if err != nil {
		fmt.Printf("Error loading TODO store: %v\n", err)
		return
	}

	if len(st.Projects) == 0 {
		fmt.Println("No TODOs found in any tracked projects.")
		return
	}

	// Determine which projects to display
	var projectsToShow []string
	if allProjects {
		// Show all projects
		for name := range st.Projects {
			projectsToShow = append(projectsToShow, name)
		}
		sort.Strings(projectsToShow)
	} else if len(args) > 0 {
		// Show the specified project
		projectName := args[0]
		if _, ok := st.Projects[projectName]; !ok {
			fmt.Printf("Project '%s' not found or has no TODOs.\n", projectName)
			fmt.Println("Available projects:")
			for name := range st.Projects {
				fmt.Printf("  - %s\n", name)
			}
			return
		}
		projectsToShow = append(projectsToShow, projectName)
	} else {
		// Show the active project
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		if cfg.Active == "" {
			fmt.Println("No active project set. Use 'tt active <project-name>' to set an active project.")
			fmt.Println("Or specify a project name: 'tt list <project-name>'")
			fmt.Println("Or use --all to show all projects: 'tt list --all'")
			return
		}

		// Find the project name that matches the active path
		activeProjectName := ""
		for name, _ := range cfg.Projects {
			if name == cfg.Active {
				activeProjectName = name
				break
			}
		}

		if activeProjectName == "" {
			fmt.Println("No active project found in your configuration.")
			fmt.Println("The active path doesn't match any tracked project.")
			fmt.Println("Try setting the active project again with: tt active <project-name>")
			fmt.Println("Available projects:")
			for name := range cfg.Projects {
				fmt.Printf("  - %s\n", name)
			}
			return
		}

		// Make sure the project has TODOs
		if _, ok := st.Projects[activeProjectName]; !ok {
			fmt.Printf("Active project '%s' doesn't have any TODOs yet.\n", activeProjectName)
			fmt.Println("Try adding a TODO comment to one of your source files.")

			// If there are other projects with TODOs, suggest them
			if len(st.Projects) > 0 {
				fmt.Println("\nOther projects with TODOs:")
				for name := range st.Projects {
					fmt.Printf("  - %s\n", name)
				}
				fmt.Println("\nYou can view them with: tt list \"<project-name>\"")
			}
			return
		}

		projectsToShow = append(projectsToShow, activeProjectName)
	}

	// Display TODOs
	if treeView {
		displayTreeView(st, projectsToShow)
	} else {
		displayListView(st, projectsToShow)
	}
}

// Get the appropriate function keyword based on file extension
func getFunctionKeyword(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	// Map of extensions to function keywords
	keywordMap := map[string]string{
		".go":    "func",
		".py":    "def",
		".js":    "function",
		".ts":    "function",
		".java":  "method",
		".c":     "function",
		".cpp":   "function",
		".cs":    "method",
		".php":   "function",
		".rb":    "def",
		".rs":    "fn",
		".swift": "func",
		".kt":    "fun",
		".scala": "def",
		".r":     "function",
		".sh":    "function",
		".pl":    "sub",
		".lua":   "function",
	}

	// Return the appropriate keyword or a default
	if keyword, ok := keywordMap[ext]; ok {
		return keyword
	}
	return "function" // Default keyword
}

// displayTreeView shows TODOs in a file system tree-like structure
func displayTreeView(st *store.Store, projectNames []string) {

	// Load config to get project paths
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Warning: Could not load config: %v\n", err)
	}

	for _, projectName := range projectNames {
		todos := st.Projects[projectName]
		if len(todos) == 0 {
			continue
		}

		// Get project root path
		projectRoot := ""
		if cfg.Projects != nil {
			projectRoot = cfg.Projects[projectName]
		}

		// Print project name and path
		if projectRoot != "" {
			fmt.Printf("%s ( %s )\n", bold(projectName), cyan(projectRoot))
		} else {
			fmt.Printf("%s\n", bold(projectName))
		}

		// Group TODOs by directory and file
		todosByPath := make(map[string]map[string][]store.Todo)

		for _, todo := range todos {
			dir := filepath.Dir(todo.FilePath)
			file := filepath.Base(todo.FilePath)

			// Make relative paths if project root is known
			if projectRoot != "" {
				if relDir, err := filepath.Rel(projectRoot, dir); err == nil {
					dir = "/" + relDir
				}
			}

			if _, ok := todosByPath[dir]; !ok {
				todosByPath[dir] = make(map[string][]store.Todo)
			}

			todosByPath[dir][file] = append(todosByPath[dir][file], todo)
		}

		// Sort directories
		var dirs []string
		for dir := range todosByPath {
			dirs = append(dirs, dir)
		}
		sort.Strings(dirs)

		for i, dir := range dirs {
			// Print directory with appropriate tree characters
			dirPrefix := "├── "
			if i == len(dirs)-1 {
				dirPrefix = "└── "
			}
			fmt.Printf("%s%s\n", dirPrefix, cyan(dir))

			// Sort files
			var files []string
			for file := range todosByPath[dir] {
				files = append(files, file)
			}
			sort.Strings(files)

			for j, file := range files {
				// Print file with appropriate tree characters
				filePrefix := "│   ├── "
				if i == len(dirs)-1 {
					filePrefix = "    ├── "
				}
				if j == len(files)-1 {
					if i == len(dirs)-1 {
						filePrefix = "    └── "
					} else {
						filePrefix = "│   └── "
					}
				}
				fmt.Printf("%s%s\n", filePrefix, yellow(file))

				// Print TODOs in this file
				todos := todosByPath[dir][file]
				sort.Slice(todos, func(a, b int) bool {
					return todos[a].LineNumber < todos[b].LineNumber
				})

				for k, todo := range todos {
					// First line prefix (for line number and function)
					firstLinePrefix := "│   │   ├── "
					if i == len(dirs)-1 {
						firstLinePrefix = "    │   ├── "
					}
					if j == len(files)-1 {
						if i == len(dirs)-1 {
							firstLinePrefix = "        ├── "
						} else {
							firstLinePrefix = "│       ├── "
						}
					}

					// Second line prefix (for the actual comment)
					secondLinePrefix := "│   │   │   └── "
					if i == len(dirs)-1 {
						secondLinePrefix = "    │   │   └── "
					}
					if j == len(files)-1 {
						if i == len(dirs)-1 {
							secondLinePrefix = "        │   └── "
						} else {
							secondLinePrefix = "│       │   └── "
						}
					}

					// Last item handling
					if k == len(todos)-1 {
						firstLinePrefix = strings.Replace(firstLinePrefix, "├── ", "└── ", 1)
					}

					// Clean the comment
					comment := strings.TrimSpace(todo.Comment)
					// Replace newlines with spaces in multi-line comments
					comment = strings.ReplaceAll(comment, "\n", " ")
					// Normalize whitespace
					comment = strings.Join(strings.Fields(comment), " ")

					if len(comment) > 81 {
						comment = comment[:80] + "..."
					}

					// Get appropriate function keyword
					functionInfo := ""
					if todo.Function != "" {
						keyword := getFunctionKeyword(file)
						functionInfo = fmt.Sprintf(" @ %s %s", keyword, green(todo.Function))
					}

					// Print the first line with metadata
					fmt.Printf("%sLine %d%s:\n",
						firstLinePrefix,
						todo.LineNumber,
						functionInfo)

					// Print the second line with the comment
					fmt.Printf("%s%s\n",
						secondLinePrefix,
						cyan(comment))
				}
			}
		}

		fmt.Println()
	}
}

// displayListView shows TODOs in a simple list
func displayListView(st *store.Store, projectNames []string) {

	// Load config to get project paths
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Warning: Could not load config: %v\n", err)
	}

	for _, projectName := range projectNames {
		todos := st.Projects[projectName]
		if len(todos) == 0 {
			continue
		}

		// Get project root path
		projectRoot := ""
		if cfg.Projects != nil {
			projectRoot = cfg.Projects[projectName]
		}

		// Print project name and path
		if projectRoot != "" {
			fmt.Printf("%s ( %s )\n", bold(projectName), cyan(projectRoot))
		} else {
			fmt.Printf("%s\n", bold(projectName))
		}

		// Sort TODOs by file path and line number
		sort.Slice(todos, func(i, j int) bool {
			if todos[i].FilePath == todos[j].FilePath {
				return todos[i].LineNumber < todos[j].LineNumber
			}
			return todos[i].FilePath < todos[j].FilePath
		})

		// Create a tabwriter for aligned output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		for _, todo := range todos {
			// Clean the comment
			comment := strings.TrimSpace(todo.Comment)
			comment = strings.ReplaceAll(comment, "\n", " ")
			comment = strings.Join(strings.Fields(comment), " ")
			if len(comment) > 81 {
				comment = comment[:80] + "..."
			}

			// Format the function info
			functionInfo := ""
			if todo.Function != "" {
				// Get the file name from the path
				file := filepath.Base(todo.FilePath)
				keyword := getFunctionKeyword(file)
				functionInfo = fmt.Sprintf("@ %s %s", keyword, green(todo.Function))
			}

			// Get relative path if project root is known
			displayPath := todo.FilePath
			if projectRoot != "" {
				if relPath, err := filepath.Rel(projectRoot, todo.FilePath); err == nil {
					displayPath = "/" + relPath
				}
			}

			fmt.Fprintf(w, "%s:%d\t%s\t\n    %s\n",
				yellow(displayPath),
				todo.LineNumber,
				functionInfo,
				cyan(comment))
		}

		w.Flush()
		fmt.Println()
	}
}
