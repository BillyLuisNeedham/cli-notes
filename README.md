# Note Manager

This is a simple command-line program for managing notes. It allows you to create and query different notes.

## Prerequisites

- Go programming language (version 1.16 or later)

## Getting Started

1. Clone the repository or download the source code files.

2. Open a terminal or command prompt and navigate to the root directory of the project.

## Running the Program

To run the program using the `go run` command, follow these steps:

1. In the terminal, navigate to the root directory of the project.

2. Run the following command:

   ```
   go run main.go
   ```

   This command tells Go to run the `main.go` file located in the root directory.

3. The program will start running, and you will see the interactive command-line interface prompt:

   ```
   >
   ```

   You can now enter commands to interact with the program.

## Available Commands

### Todo Management

- `gt` - Get all open todos
- `gt <query>` - Search open todos containing the specified query (multiple queries can be separated by commas)
- `gto` - Get all overdue todos
- `gtnd` - Get all todos with no due date
- `gts` - Get todos due soon (within the next week)
- `ct <title>` - Create a new todo with the specified title
- `p1` - Get high priority (P1) todos
- `p2` - Get medium priority (P2) todos
- `p3` - Get low priority (P3) todos

### Meeting Notes

- `cm <title>` - Create a new meeting note with the specified title

### Planning Notes

- `cp <title>` - Create a new planning note with the specified title and seven questions template

### Standup Notes

- `cs` - Create a new standup note with predefined team member sections

### General Note Operations

- `gta <tags>` - Search notes by tags
- `gq <query>` - Search all notes containing the specified query (multiple queries can be separated by commas)
- `gqa <query>` - Search within the previously queried results
- `gat` - Get all uncompleted tasks from previously queried files
- `o <filename>` - Open a specific note in the editor
- `gd <start-date> <end-date>` - Get completed todos between the specified dates (format: YYYY-MM-DD) and create a summary note

### Due Date Management

- `d <days>` - Delay the due date of the selected todo by the specified number of days
- `t` - Set the due date of the selected todo to today
- `m` - Set the due date of the selected todo to next Monday
- `tu` - Set the due date of the selected todo to next Tuesday
- `w` - Set the due date of the selected todo to next Wednesday
- `th` - Set the due date of the selected todo to next Thursday
- `f` - Set the due date of the selected todo to next Friday
- `sa` - Set the due date of the selected todo to next Saturday
- `su` - Set the due date of the selected todo to next Sunday

### Navigation

- `↑` (Up Arrow) - Navigate to previous file in search results and display its tasks
- `↓` (Down Arrow) - Navigate to next file in search results and display its tasks
- `ESC` - Clear the current command line

### Program Control

- `exit`, `quit`, or `q` - Exit the program

## Note Format

Notes are stored as Markdown files with YAML frontmatter containing metadata such as:

- title
- date-created
- tags
- date-due (for todos)
- done status (for todos)

When navigating through files using the arrow keys, any uncompleted tasks (lines containing "- [ ]") will be automatically displayed below the filename. Tasks are shown in the format:
`filename : task content: line_number`
