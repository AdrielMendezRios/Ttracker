package ignore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IgnorePattern represents a single ignore pattern
type IgnorePattern struct {
	Pattern string
	IsDir   bool
}

// IgnoreManager handles ignore patterns for a project
type IgnoreManager struct {
	patterns []IgnorePattern
}

// NewIgnoreManager creates a new ignore manager
func NewIgnoreManager() *IgnoreManager {
	return &IgnoreManager{
		patterns: make([]IgnorePattern, 0),
	}
}

// LoadFromFile loads ignore patterns from a file
func (im *IgnoreManager) LoadFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read ignore file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		// Handle directory patterns (ending with /)
		isDir := strings.HasSuffix(line, "/")
		if isDir {
			line = strings.TrimSuffix(line, "/")
		}

		im.patterns = append(im.patterns, IgnorePattern{
			Pattern: line,
			IsDir:   isDir,
		})
	}

	return nil
}

// AddPattern adds a single ignore pattern
func (im *IgnoreManager) AddPattern(pattern string, isDir bool) {
	im.patterns = append(im.patterns, IgnorePattern{
		Pattern: pattern,
		IsDir:   isDir,
	})
}

// ShouldIgnore checks if a path should be ignored
func (im *IgnoreManager) ShouldIgnore(path string) bool {
	// Convert path to use forward slashes for consistency
	path = filepath.ToSlash(path)

	for _, pattern := range im.patterns {
		// Handle directory patterns
		if pattern.IsDir {
			if strings.HasSuffix(path, pattern.Pattern) {
				return true
			}
			continue
		}

		// Handle file patterns
		if matched, _ := filepath.Match(pattern.Pattern, filepath.Base(path)); matched {
			return true
		}
	}

	return false
}

// GetDefaultPatterns returns common patterns that should be ignored by default
func GetDefaultPatterns() []IgnorePattern {
	return []IgnorePattern{
		{Pattern: ".git", IsDir: true},
		{Pattern: "node_modules", IsDir: true},
		{Pattern: "vendor", IsDir: true},
		{Pattern: "dist", IsDir: true},
		{Pattern: "build", IsDir: true},
		{Pattern: ".DS_Store", IsDir: false},
		{Pattern: "*.log", IsDir: false},
		{Pattern: "*.tmp", IsDir: false},
	}
}
