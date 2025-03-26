package scan

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"Ttracker/internal/ignore"
	"Ttracker/internal/store"
)

// RunScan scans the provided project path and updates the store
func RunScan(projectPath, projectName, pluginConfigPath, storeFile string) error {
	// If no project name is provided, fall back to the directory name
	if projectName == "" {
		projectName = filepath.Base(projectPath)
	}

	fmt.Printf("Starting scan for project '%s' at path: %s\n", projectName, projectPath)

	// Ensure the store file's directory exists
	storeDir := filepath.Dir(storeFile)
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for store file: %v", err)
	}

	// Load existing store or create new one
	st := store.NewStore()
	storeExists := false

	if _, err := os.Stat(storeFile); !os.IsNotExist(err) {
		if existingStore, err := store.LoadStore(storeFile); err == nil {
			fmt.Printf("Loaded existing TODOs from %s\n", storeFile)
			st = existingStore
			storeExists = true
		} else {
			fmt.Printf("Warning: Failed to load existing store: %v\n", err)
		}
	}

	// Create a new collection to hold current TODOs
	currentTodos := make([]store.Todo, 0)

	// Create manager
	mgr, err := NewManager(pluginConfigPath)
	if err != nil {
		log.Printf("Warning: %v\n", err)
		// Continue with available parsers
	}

	// Initialize ignore manager with default patterns
	ignoreMgr := ignore.NewIgnoreManager()
	for _, pattern := range ignore.GetDefaultPatterns() {
		ignoreMgr.AddPattern(pattern.Pattern, pattern.IsDir)
	}

	// Try to load project-specific ignore file
	ignoreFile := filepath.Join(projectPath, ".ttignore")
	if err := ignoreMgr.LoadFromFile(ignoreFile); err != nil {
		// Not an error if file doesn't exist
		if !os.IsNotExist(err) {
			log.Printf("Warning: Failed to load ignore file: %v\n", err)
		}
	}

	fmt.Printf("Walking directory: %s\n", projectPath)

	todoCount := 0
	fileCount := 0

	err = filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing %s: %v\n", path, err)
			return nil // continue walking
		}

		// Check if path should be ignored
		if ignoreMgr.ShouldIgnore(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		// Skip files that are not source code
		ext := filepath.Ext(path)
		if ext == "" {
			return nil
		}

		fileCount++

		parser, err := mgr.GetParser(path)
		if err != nil {
			// Not an error, just means we don't have a parser for this file type
			return nil
		}

		fmt.Printf("Parsing file: %s\n", path)
		todos, err := parser.ParseFile(path)
		if err != nil {
			fmt.Printf("Error parsing %s: %v\n", path, err)
			return nil
		}

		if len(todos) > 0 {
			fmt.Printf("Found %d TODOs in %s\n", len(todos), path)
			todoCount += len(todos)
		}

		// Add the found TODOs to our current collection
		currentTodos = append(currentTodos, todos...)

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking directory: %v", err)
	}

	fmt.Printf("Scan complete. Found %d TODOs in %d files for project '%s'\n",
		todoCount, fileCount, projectName)

	// If the store exists and the project has TODOs, we need to update it
	if storeExists {
		// Update existing project or add new one
		st.UpdateProject(projectName, currentTodos)
	} else {
		// Create new project entry
		st.AddProject(projectName, currentTodos)
	}

	// Save the updated store
	if err := st.Save(storeFile); err != nil {
		return fmt.Errorf("failed to save store: %v", err)
	}

	return nil
}
