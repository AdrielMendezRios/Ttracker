#!/usr/bin/env python3
"""
Python TODO Parser for Ttracker

This script parses Python files for TODO and FIXME comments and outputs them
in the format expected by Ttracker's plugin system.

Usage:
    python python_todo_parser.py <file_path>
"""

import sys
import re
import os
import ast
import tokenize
from io import BytesIO

# Regular expression to match TODO and FIXME comments
TODO_PATTERN = re.compile(r'^\s*#\s*(TODO|FIXME)[:]*\s*(.*)', re.IGNORECASE)

class TodoVisitor(ast.NodeVisitor):
    """AST visitor that tracks function and class contexts."""
    
    def __init__(self):
        self.current_function = []
        self.current_class = []
        self.function_ranges = []  # List of (start_line, end_line, function_name)
        
    def visit_FunctionDef(self, node):
        function_name = node.name
        if self.current_class:
            # It's a method
            function_name = f"{self.current_class[-1]}.{function_name}"
        
        self.current_function.append(function_name)
        self.function_ranges.append((
            node.lineno,
            node.end_lineno if hasattr(node, 'end_lineno') else node.lineno,
            function_name
        ))
        
        # Visit children
        self.generic_visit(node)
        
        # Pop the function context
        self.current_function.pop()
        
    def visit_ClassDef(self, node):
        self.current_class.append(node.name)
        
        # Visit children
        self.generic_visit(node)
        
        # Pop the class context
        self.current_class.pop()

def get_function_for_line(function_ranges, line_number):
    """Find which function a line belongs to."""
    for start, end, name in function_ranges:
        if start <= line_number <= end:
            return name
    return ""

def parse_file(file_path):
    """Parse a Python file for TODO and FIXME comments using AST."""
    if not os.path.isfile(file_path):
        print(f"Error: File not found: {file_path}", file=sys.stderr)
        return
    
    # Parse the AST to get function ranges
    with open(file_path, 'r') as f:
        file_content = f.read()
    
    try:
        tree = ast.parse(file_content, filename=file_path)
        visitor = TodoVisitor()
        visitor.visit(tree)
        function_ranges = visitor.function_ranges
    except SyntaxError as e:
        print(f"Error parsing {file_path}: {e}", file=sys.stderr)
        function_ranges = []
    
    # Now tokenize the file to get comments
    with open(file_path, 'rb') as f:
        for tok in tokenize.tokenize(f.readline):
            # Only process comments
            if tok.type == tokenize.COMMENT:
                line_num = tok.start[0]
                comment = tok.string.lstrip('#').strip()
                
                # Check if it's a TODO/FIXME comment
                match = TODO_PATTERN.match(f"#{comment}")
                if match:
                    todo_type = match.group(1)
                    todo_text = match.group(2).strip()
                    
                    # Find which function this comment belongs to
                    function_name = get_function_for_line(function_ranges, line_num)
                    
                    # Output in format: <lineNumber>: <comment> [function]
                    output = f"{line_num}: # {todo_type}: {todo_text}"
                    if function_name:
                        output += f" [in function {function_name}]"
                    
                    print(output)

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} <file_path>", file=sys.stderr)
        sys.exit(1)
    
    parse_file(sys.argv[1]) 