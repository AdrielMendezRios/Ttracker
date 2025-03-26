package store

import (
	"encoding/json"
	"fmt"
	"os"
)

// Todo represents a TODO comment found in source code.
type Todo struct {
	Comment    string `json:"comment"`     // The TODO text (including any extra info)
	FilePath   string `json:"file_path"`   // The file in which it was found
	LineNumber int    `json:"line_number"` // The line number
	Function   string `json:"function"`    // The enclosing function name (if any)
}

// Store holds TODOs for each project.
type Store struct {
	Projects map[string][]Todo `json:"projects"`
}

// NewStore creates an empty Store.
func NewStore() *Store {
	return &Store{Projects: make(map[string][]Todo)}
}

// AddTodo adds a Todo to the given project, avoiding duplicates.
func (s *Store) AddTodo(projectName string, todo Todo) {
	// Check if we already have this project
	if _, exists := s.Projects[projectName]; !exists {
		s.Projects[projectName] = []Todo{}
	}

	// Check for duplicate TODOs (same file and line)
	for _, existing := range s.Projects[projectName] {
		if existing.FilePath == todo.FilePath && existing.LineNumber == todo.LineNumber {
			// Already have this TODO, don't add it again
			return
		}
	}

	// Add the new TODO
	s.Projects[projectName] = append(s.Projects[projectName], todo)
}

// Save writes the store to disk as JSON.
func (s *Store) Save(filePath string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling store: %v", err)
	}
	return os.WriteFile(filePath, data, 0644)
}

// LoadStore loads a Store from the specified JSON file.
func LoadStore(filePath string) (*Store, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read store file: %v\n", err)
	}
	var s Store
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("could not unmarshal store data: %v\n", err)
	}
	return &s, nil
}

// RemoveTodo removes a Todo from the given project.
func (s *Store) RemoveTodo(projectName string, todo Todo) {
	if todos, exists := s.Projects[projectName]; exists {
		updatedTodos := make([]Todo, 0, len(todos))
		for _, t := range todos {
			if t.FilePath != todo.FilePath || t.LineNumber != todo.LineNumber {
				updatedTodos = append(updatedTodos, t)
			}
		}
		s.Projects[projectName] = updatedTodos
	}
}

// FindDeletedTodos compares old and new TODOs and returns ones that have been deleted.
func FindDeletedTodos(oldTodos, newTodos []Todo) []Todo {
	deletedTodos := make([]Todo, 0)

	// Create a map for quick lookup of new TODOs
	newTodoMap := make(map[string]bool)
	for _, todo := range newTodos {
		// Use file path and line number as a unique key
		key := fmt.Sprintf("%s:%d", todo.FilePath, todo.LineNumber)
		newTodoMap[key] = true
	}

	// Check each old TODO to see if it's still in the new list
	for _, oldTodo := range oldTodos {
		key := fmt.Sprintf("%s:%d", oldTodo.FilePath, oldTodo.LineNumber)
		if !newTodoMap[key] {
			deletedTodos = append(deletedTodos, oldTodo)
		}
	}

	return deletedTodos
}

// UpdateTodos updates a project's TODOs, adding new ones and removing deleted ones.
func (s *Store) UpdateTodos(projectName string, currentTodos []Todo) {
	if s.Projects == nil {
		s.Projects = make(map[string][]Todo)
	}

	// If the project doesn't exist yet, just add all TODOs
	if _, exists := s.Projects[projectName]; !exists {
		if len(currentTodos) > 0 {
			s.Projects[projectName] = currentTodos
		}
		return
	}

	// If there are no current TODOs, remove the project
	if len(currentTodos) == 0 {
		delete(s.Projects, projectName)
		return
	}

	// Replace the project's TODOs with the current ones
	s.Projects[projectName] = currentTodos
}

// UpdateProject updates or adds a project's TODOs
func (s *Store) UpdateProject(projectName string, todos []Todo) {
	s.Projects[projectName] = todos
}

// AddProject adds a new project with its TODOs
func (s *Store) AddProject(projectName string, todos []Todo) {
	s.Projects[projectName] = todos
}

// GetProject returns the TODOs for a specific project
func (s *Store) GetProject(projectName string) ([]Todo, bool) {
	todos, exists := s.Projects[projectName]
	return todos, exists
}

// RemoveProject removes a project and its TODOs
func (s *Store) RemoveProject(projectName string) {
	delete(s.Projects, projectName)
}
