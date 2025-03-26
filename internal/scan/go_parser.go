package scan

import (
	"Ttracker/internal/store"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"
)

// Regular expression to match TODO and FIXME comments
// More relaxed pattern to catch different formats
// var todoRegex = regexp.MustCompile(`(?i)(todo|fixme)`)
var todoRegex = regexp.MustCompile(`.*\s(TODO|FIXME)[\s:\(\[]`)

// GoParser implements the Scanner for Go source Files.
type GoParser struct{}

func (g *GoParser) SupportedExtensions() []string {
	return []string{".go"}
}

func (g *GoParser) ParseFile(filePath string) ([]store.Todo, error) {
	var todos []store.Todo

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file: %v", err)
	}

	fmt.Printf("Checking for TODOs in %s\n", filePath)

	// Iterate over all comment groups
	for _, group := range node.Comments {
		for _, c := range group.List {
			text := c.Text
			if todoRegex.MatchString(text) {
				pos := fset.Position(c.Pos())
				funcName := findEnclosingFunction(node, c.Pos(), fset)

				// Get a preview of the comment for logging
				preview := text
				if len(text) > 40 {
					preview = text[:40] + "..."
				}
				fmt.Printf("Found TODO at line %d: %s\n", pos.Line, strings.TrimSpace(preview))

				todos = append(todos, store.Todo{
					Comment:    text,
					FilePath:   filePath,
					LineNumber: pos.Line,
					Function:   funcName,
				})
			}
		}
	}

	fmt.Printf("Found %d TODOs in file %s\n", len(todos), filePath)
	return todos, nil
}

// findEnclosingFunction traverses the AST to locate the function containing pos
func findEnclosingFunction(node *ast.File, pos token.Pos, fset *token.FileSet) string {
	var funcName string
	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Pos() <= pos && pos <= fn.End() {
				funcName = fn.Name.Name
				return false // found it; breakout
			}
		}
		return true
	})
	return funcName
}
