package scan

import (
	plugin "Ttracker/internal/plugins"
	"Ttracker/internal/store"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Manager handles all file parsing operations for finding TODOs
type Manager struct {
	Parsers []Scanner
}

// ExternalParser implements Scanner for external plugin-based parsers
type ExternalParser struct {
	Command             string
	ExtensionsSupported []string
}

// SupportedExtensions returns file extensions supported by this parser
func (ep *ExternalParser) SupportedExtensions() []string {
	return ep.ExtensionsSupported
}

// ParseFile executes the external parser and returns extracted TODOs
func (ep *ExternalParser) ParseFile(filePath string) ([]store.Todo, error) {
	cmd := exec.Command(ep.Command, filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error executing external parser: %v", err)
	}

	var todos []store.Todo
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split to get line number and comment
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		lineNumber := parseLineNumber(parts[0])
		commentText := strings.TrimSpace(parts[1])

		// Extract function information if available
		// Format: "Comment text [in function functionName]"
		functionName := ""
		functionMatch := regexp.MustCompile(`(.*)\s+\[in function\s+([^\]]+)\]`).FindStringSubmatch(commentText)

		if len(functionMatch) == 3 {
			// Update comment text and function name if pattern matches
			commentText = strings.TrimSpace(functionMatch[1])
			functionName = functionMatch[2]
		}

		todo := store.Todo{
			Comment:    commentText,
			FilePath:   filePath,
			LineNumber: lineNumber,
			Function:   functionName,
		}

		todos = append(todos, todo)
	}

	return todos, nil
}

// parseLineNumber converts a string to an integer safely
func parseLineNumber(line string) int {
	var num int
	fmt.Sscanf(line, "%d", &num)
	return num
}

// NewManager creates a manager with all available parsers
func NewManager(configPath string) (*Manager, error) {
	// Initialize with built-in parsers
	manager := &Manager{
		Parsers: []Scanner{&GoParser{}},
	}

	// Load plugin configurations
	pluginMgr, err := plugin.NewPluginManager()
	if err != nil {
		return manager, fmt.Errorf("warning: failed to create plugin manager: %v", err)
	}

	if err := pluginMgr.LoadPlugins(); err != nil {
		return manager, fmt.Errorf("warning: failed to load plugins: %v", err)
	}

	// Create and add external parsers
	for _, cfg := range pluginMgr.Plugins {
		parser := &ExternalParser{
			Command:             cfg.Command,
			ExtensionsSupported: cfg.Extensions,
		}
		manager.Parsers = append(manager.Parsers, parser)
	}

	return manager, nil
}

// GetParser selects an appropriate parser based on file extension
func (m *Manager) GetParser(path string) (Scanner, error) {
	ext := strings.ToLower(filepath.Ext(path))
	for _, parser := range m.Parsers {
		for _, supported := range parser.SupportedExtensions() {
			if supported == ext {
				return parser, nil
			}
		}
	}
	return nil, fmt.Errorf("no parsers available for extension: %s", ext)
}
