package watcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"Ttracker/internal/config"
	"Ttracker/internal/ignore"
	"Ttracker/internal/scan"

	"github.com/fsnotify/fsnotify"
)

// ProjectWatcher watches for file changes in tracked projects
type ProjectWatcher struct {
	watcher      *fsnotify.Watcher
	projects     map[string]string // name -> path
	debounceTime time.Duration
	changes      map[string]time.Time
	mutex        sync.Mutex
	scanningLock sync.Mutex
	storeFile    string
	pluginConfig string
	stopChan     chan struct{}
	ignoreMgr    *ignore.IgnoreManager
}

// NewProjectWatcher creates a new watcher for tracking projects
func NewProjectWatcher(storeFile, pluginConfig string) (*ProjectWatcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %v", err)
	}

	// Initialize ignore manager with default patterns
	ignoreMgr := ignore.NewIgnoreManager()
	for _, pattern := range ignore.GetDefaultPatterns() {
		ignoreMgr.AddPattern(pattern.Pattern, pattern.IsDir)
	}

	return &ProjectWatcher{
		watcher:      fsWatcher,
		projects:     make(map[string]string),
		debounceTime: 2 * time.Second,
		changes:      make(map[string]time.Time),
		storeFile:    storeFile,
		pluginConfig: pluginConfig,
		stopChan:     make(chan struct{}),
		ignoreMgr:    ignoreMgr,
	}, nil
}

// Start begins watching all tracked projects
func (pw *ProjectWatcher) Start() error {
	// Load projects from config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Watch each project
	for name, path := range cfg.Projects {
		pw.projects[name] = path
		if err := pw.watchProject(name, path); err != nil {
			log.Printf("Warning: Could not watch project %s: %v", name, err)
		}
	}

	// Start the watcher goroutine
	go pw.watchLoop()

	fmt.Println("Ttracker file watcher started. Monitoring", len(pw.projects), "projects for changes.")
	return nil
}

// Stop gracefully stops the watcher
func (pw *ProjectWatcher) Stop() error {
	close(pw.stopChan)
	return pw.watcher.Close()
}

// AddProject adds a new project to the watcher
func (pw *ProjectWatcher) AddProject(name, path string) error {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()

	pw.projects[name] = path
	return pw.watchProject(name, path)
}

// watchProject recursively watches a project directory
func (pw *ProjectWatcher) watchProject(name, path string) error {
	// Ensure the directory exists
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access project path: %v", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("project path is not a directory")
	}

	// Try to load project-specific ignore file
	ignoreFile := filepath.Join(path, ".ttignore")
	if err := pw.ignoreMgr.LoadFromFile(ignoreFile); err != nil {
		// Not an error if file doesn't exist
		if !os.IsNotExist(err) {
			log.Printf("Warning: Failed to load ignore file for project %s: %v", name, err)
		}
	}

	// Add watcher to the root directory
	if err := pw.watcher.Add(path); err != nil {
		return fmt.Errorf("failed to watch directory: %v", err)
	}

	// Recursively add watchers to subdirectories
	return filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip inaccessible paths
		}

		// Check if path should be ignored
		if pw.ignoreMgr.ShouldIgnore(subpath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Add watcher to directories
		if info.IsDir() {
			if err := pw.watcher.Add(subpath); err != nil {
				log.Printf("Warning: Failed to watch subdirectory %s: %v", subpath, err)
			}
		}
		return nil
	})
}

// watchLoop processes events from the file watcher
func (pw *ProjectWatcher) watchLoop() {
	for {
		select {
		case event := <-pw.watcher.Events:
			pw.handleFileChange(event.Name)
		case err := <-pw.watcher.Errors:
			log.Printf("Error watching files: %v", err)
		case <-pw.stopChan:
			return
		}
	}
}

func (pw *ProjectWatcher) handleFileChange(path string) {
	// Ignore changes to our own data files
	if filepath.Base(path) == "todos.json" || filepath.Base(path) == "config.json" {
		return
	}

	// Check if path should be ignored
	if pw.ignoreMgr.ShouldIgnore(path) {
		return
	}

	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		// File might have been deleted
		fmt.Printf("Warning: Could not stat file %s: %v\n", path, err)
		return
	}

	// If it's a directory, add a watcher
	if info.IsDir() {
		if err := pw.watcher.Add(path); err != nil {
			log.Printf("Warning: Failed to watch new directory %s: %v\n", path, err)
		}
		return
	}

	// Check file extension - only process source code files
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return // No extension, probably not source code
	}

	// Find which project this file belongs to
	projectName := ""
	projectPath := ""
	for name, projPath := range pw.projects {
		if isInDirectory(path, projPath) {
			projectName = name
			projectPath = projPath
			break
		}
	}

	// If we can't identify the project, just record the change
	// The processChanges method will figure it out later
	if projectName == "" {
		// Record the change time
		pw.mutex.Lock()
		pw.changes[path] = time.Now()
		pw.mutex.Unlock()
		return
	}

	// For known source code files in a known project, process them immediately
	// instead of waiting for the debounce timer
	fmt.Printf("Detected change in source file: %s (project: %s)\n", path, projectName)

	// Only do immediate processing for certain extensions like .go
	if ext == ".go" {
		go pw.scanProject(projectName, projectPath)
	} else {
		// For other file types, use the normal debounce mechanism
		pw.mutex.Lock()
		pw.changes[path] = time.Now()
		pw.mutex.Unlock()
	}
}

// processChanges handles any pending file changes
func (pw *ProjectWatcher) processChanges() {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()

	if len(pw.changes) == 0 {
		return
	}

	now := time.Now()
	projectsToScan := make(map[string]bool)

	// Find changes that have settled (not modified recently)
	for path, lastChange := range pw.changes {
		if now.Sub(lastChange) >= pw.debounceTime {
			// Find which project this file belongs to
			for name, projectPath := range pw.projects {
				if isInDirectory(path, projectPath) {
					projectsToScan[name] = true
					break
				}
			}
			delete(pw.changes, path)
		}
	}

	// Trigger scans for affected projects
	for name := range projectsToScan {
		projectPath := pw.projects[name]
		go pw.scanProject(name, projectPath)
	}
}

// scanProject runs a scan for a specific project
func (pw *ProjectWatcher) scanProject(name, path string) {
	// Use a lock to prevent multiple simultaneous scans of the same project
	pw.scanningLock.Lock()
	defer pw.scanningLock.Unlock()

	fmt.Printf("Change detected in project '%s'. Scanning for TODOs...\n", name)

	err := scan.RunScan(path, name, pw.pluginConfig, pw.storeFile)
	if err != nil {
		log.Printf("Error scanning project %s: %v", name, err)
	} else {
		fmt.Printf("Scan complete for project '%s'\n", name)
	}
}

// isInDirectory checks if a path is inside another directory
func isInDirectory(path, dir string) bool {
	// Get absolute paths
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}

	// Check if absPath starts with absDir
	rel, err := filepath.Rel(absDir, absPath)
	if err != nil {
		return false
	}

	return !filepath.IsAbs(rel) && !strings.HasPrefix(rel, "..")
}
