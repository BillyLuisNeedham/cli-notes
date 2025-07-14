# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a CLI-based notes management system written in Go that provides an interactive command-line interface for creating, searching, and managing various types of notes (todos, meetings, plans, standups). The application uses keyboard input handling and stores notes as Markdown files with YAML frontmatter in a `notes/` directory.

## Core Architecture

### Main Components

- **main.go**: Entry point with interactive command loop using `github.com/eiannone/keyboard` for real-time key handling
- **scripts/**: Core business logic layer
  - **data/**: Repository layer handling file I/O and data persistence  
  - **presentation/**: UI layer managing command parsing and display
- **notes/**: Directory where all note files are stored as Markdown with YAML frontmatter

### Key Architecture Patterns

- **Command Pattern**: Commands are parsed through `presentation.CommandHandler` and executed via `handleCommand` in main.go
- **Repository Pattern**: Data access abstracted through interfaces in `scripts/data/`
- **Functional Composition**: Heavy use of function injection for dependencies (e.g., `data.QueryFilesByDone`, `data.WriteFile`)

### File Structure

Notes are stored as Markdown files with YAML frontmatter containing:
- `title`, `date-created`, `tags`, `priority` (P1/P2/P3)
- `date-due`, `done` (for todos)
- Content follows standard Markdown with `- [ ]` for tasks

### State Management

- `SearchedFilesStore`: Maintains recently queried files for navigation
- Arrow key navigation cycles through search results
- Selected files can have operations applied (due date changes, etc.)

## Development Commands

### Building and Running
```bash
# Run the application
go run main.go

# Build binary
go build -o main

# Run the built binary
./main
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./scripts/data
go test ./scripts/presentation

# Run single test file
go test -v ./scripts/create_test.go
```

### Dependencies
- Primary dependency: `github.com/eiannone/keyboard` for real-time keyboard input
- Uses Go standard library for file operations, time handling, and string processing

## Key Command Categories

The application supports several command categories executed through the interactive prompt:

- **Todo Management**: `gt`, `gto`, `gtnd`, `gts`, `ct`, `p1`/`p2`/`p3`
- **Note Creation**: `cm` (meetings), `cp` (plans), `cs` (standups)  
- **Search Operations**: `gq`, `gqa`, `gta`, `gat`
- **Due Date Management**: `d`, `t`, `m`/`tu`/`w`/`th`/`f`/`sa`/`su`
- **Navigation**: Arrow keys, `o` (open), `gd` (date range queries)

## Working with the Codebase

### Adding New Commands
1. Add command parsing logic in `scripts/presentation/command.go`
2. Add command execution in `main.go`'s `handleCommand` function
3. Implement business logic in appropriate `scripts/` files
4. Add tests following existing patterns in `*_test.go` files

### File Operations
All file operations go through `data.WriteFile` and query functions in `scripts/data/file_repository.go`. The system expects a `notes/` directory to exist in the working directory.

### Testing Strategy
The codebase uses standard Go testing with comprehensive coverage across:
- Unit tests for individual functions
- Integration tests for command flows
- Repository tests for file operations