# Ttracker
 I wanted to make a project in Go to learn the language and I have to say, its a pretty awesome language! at first i wanted to make a simple todo app but found myself wanting to write and see more of what go has to offer so ive added a few things i didn't originally planned. at the moment the tracker only scans for todo comments in .go and .py files, though you can make a parser for any language and add it to Ttracker.

 That said, expect things to break. I've done minimal testing and only tested this tool on linux (rhel 9). im not sure if I'll keep working on this project but if I do the next thing would be adding links to the todos in the list command to open the file they live in in whatever editor I want.

## What is Ttracker?

Ttracker is a tool that helps keep track of TODOs across multiple projects. It:
- Scans your project files for TODO comments
- Watches for file changes in real-time
- Supports multiple projects simultaneously
- Allows custom parsers for different file types
- Provides an ignore system (similar to .gitignore) to exclude files/directories

## Installation

1. Clone the repository:
```bash
git clone https://github.com/AdrielMendezRios/Ttracker.git
cd Ttracker
```

2. Build the project:
```bash
go build -o tt
```

3. Install globally:
```bash
./tt install
```

## Basic Usage

```bash
# Start tracking a new project
tt track /path/to/your/project

# List all tracked projects
tt projects

# Switch to a different project
tt switch project-name

# List TODOs in the current project
tt list

# Start the daemon to watch for file changes
tt daemon

# list files/directories to ignore
tt ignore project-name
```

## Adding Custom Parsers

Ttracker supports custom parsers for different file types. To create your own parser:

1. Create a new parser script (see `parsers/python_todo_parser.py` for an example)
2. Add your parser using the plugins command:
```bash
tt plugins --add \
  --id "your-parser-id" \
  --lang "your-language" \
  --cmd "./parsers/your_parser.py" \
  --ext ".ext1,.ext2"
```

### Parser Requirements

Your parser should:
1. Accept file content via stdin
2. Output JSON in the following format:
```json
{
    "todos": [
        {
            "line": 42,
            "content": "TODO: Your todo message"
        }
    ]
}
```

### Managing Plugins

```bash
# List all available plugins
tt plugins --list


# Remove a plugin
tt plugins --remove --id "plugin-id"

# Set a plugin as default for its language
tt plugins --default --id "plugin-id"
```

## Contributing

Feel free to contribute by:
- Adding new parsers for different languages
- Improving existing functionality
- Reporting bugs
- Suggesting new features
