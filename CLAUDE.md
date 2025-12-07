# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run Commands

```bash
# Run the application
go run .

# Build the binary
go build -o cli-notes .

# Run all tests
go test ./...

# Run a single test file
go test ./e2e -run TestName -v

# Run e2e tests only
go test ./e2e/... -v

# Run unit tests in scripts package
go test ./scripts/... -v
```

## Architecture Overview

This is an interactive CLI note manager written in Go. Notes are stored as Markdown files with YAML frontmatter in the `notes/` directory.

### Package Structure

- **`main.go`** - Entry point, keyboard event loop, command routing via `handleCommand()`
- **`test_mode.go`** - Alternative stdin-based input for e2e testing (activated via `CLI_NOTES_TEST_MODE=true`)
- **`input/`** - Input abstraction (`InputReader` interface) for keyboard vs stdin reading
- **`scripts/`** - Core business logic
  - `file.go` - `File` struct (the central domain model with Title, Tags, DueAt, Priority, ObjectiveID, etc.)
  - `create.go` - Note creation functions
  - `get.go` - Query/fetch operations
  - `update.go` - Note modification operations
  - `objectives.go` - Objective linking/unlinking logic
  - `sort.go` - Sorting algorithms for todos
- **`scripts/data/`** - Data layer (file I/O, repositories)
  - `file_repository.go` - Read/write notes to disk, YAML frontmatter parsing
  - `objective_repository.go` - Objective-specific queries
  - `week_planner.go`, `week_planner_state.go` - Weekly planner data/state
  - `objectives_state.go` - Objectives view state management
  - `talk_to_state.go` - Talk-to feature state
- **`scripts/presentation/`** - UI rendering and input handling
  - `command.go` - Main command parser (`CommandHandler`, `WIPCommand`, `CompletedCommand`)
  - `week_planner_ui.go`, `week_planner_input.go` - Weekly planner TUI
  - `objectives_ui.go`, `objectives_input.go` - Objectives view TUI
  - `talk_to_ui.go`, `talk_to_input.go` - Talk-to feature TUI
- **`e2e/`** - End-to-end tests using `TestHarness` framework

### Key Patterns

1. **Command Flow**: Keyboard input → `CommandHandler` → `WIPCommand` (in-progress) or `CompletedCommand` → `handleCommand()` executes action

2. **Repository Pattern**: Business logic in `scripts/` calls repository functions from `scripts/data/` via function parameters (dependency injection), e.g., `data.WriteFile`, `data.QueryFilesByDone`

3. **State Management**: Interactive views (week planner, objectives, talk-to) use dedicated state structs in `scripts/data/` with corresponding UI/input handlers in `scripts/presentation/`

4. **Note Format**: Markdown with YAML frontmatter containing: `title`, `date-created`, `tags`, `priority`, `date-due`, `done`, `objective-role`, `objective-id`

### Testing

- **E2E tests** use `TestHarness` from `e2e/framework.go` which builds the binary, creates temp directories, and runs the CLI with piped stdin
- Set `CLI_NOTES_TEST_MODE=true` to use stdin-based input instead of keyboard library
- Set `TEST_FIXED_DATE=2025-11-28` for deterministic date-based tests
