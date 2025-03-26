package scan

import "Ttracker/internal/store"



// Scanner is the interface that every language parser must implement
type Scanner interface {
        // ParseFile scans the given file and returns a slice of todo items
        ParseFile(filePath string) ([]store.Todo, error)

        // supportedExtensions return the file extensions that the parser can handle
        SupportedExtensions() []string
}
