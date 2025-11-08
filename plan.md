# Objectives View Feature - Implementation Plan

## Executive Summary

The Objectives View feature will add hierarchical task organization to the CLI notes system, allowing users to group related todos under parent objectives. This requires implementing new metadata fields, data queries, presentation views, and command handlers while maintaining consistency with existing architectural patterns.

**Key Architecture Insight**: The codebase follows a clean 3-layer architecture:
- **Data Layer** (`scripts/data/`): File I/O and persistence
- **Business Logic Layer** (`scripts/`): Domain operations
- **Presentation Layer** (`scripts/presentation/`): UI rendering and command parsing
- **Main Loop** (`main.go`): Command orchestration and keyboard handling

**Implementation Complexity**: Medium-High
**Estimated Components**: 8 new files, 5 modified files
**Risk Level**: Medium (requires careful metadata handling and view state management)
**Timeline**: 6 weeks

---

## Progress Tracking

### Overall Progress
- [ ] Phase 1: Core Data Model Extensions (Week 1)
- [ ] Phase 2: Data Repository Layer (Week 1-2)
- [ ] Phase 3: Business Logic Layer (Week 2)
- [ ] Phase 4: View State Management (Week 3)
- [ ] Phase 5: Presentation Layer (Week 3)
- [ ] Phase 6: Main Loop Integration (Week 4)
- [ ] Phase 7: Advanced Features (Week 5)
- [ ] Phase 8: Testing & Polish (Week 6)

---

## Phase 1: Core Data Model Extensions

**Goal**: Extend the File struct and metadata handling to support objectives

### Phase 1.1: File Struct Extension
- [ ] Add `ObjectiveRole string` field to File struct in `scripts/file.go`
- [ ] Add `ObjectiveID string` field to File struct in `scripts/file.go`
- [ ] Verify struct compiles without errors

**File**: `/home/billy/repos/cli-notes/scripts/file.go`

```go
type File struct {
    Name          string
    Title         string
    Tags          []string
    CreatedAt     time.Time
    DueAt         time.Time
    Done          bool
    Content       string
    Priority      Priority
    ObjectiveRole string   // NEW: "parent" or "" (empty for non-objectives)
    ObjectiveID   string   // NEW: 8-char hash linking parent and children
}
```

### Phase 1.2: Metadata Writing Extension
- [ ] Modify `WriteFile` function in `scripts/data/file_repository.go`
- [ ] Add `objective-role` to metadata array (conditional write if non-empty)
- [ ] Add `objective-id` to metadata array (conditional write if non-empty)
- [ ] Test writing note with objective metadata

**File**: `/home/billy/repos/cli-notes/scripts/data/file_repository.go`

### Phase 1.3: Metadata Reading Extension
- [ ] Modify `getFileIfQueryMatches` in `scripts/data/file_repository.go`
- [ ] Add case for `"objective-role"` in metadata parsing
- [ ] Add case for `"objective-id"` in metadata parsing
- [ ] Update `readLatestFileContent` in `scripts/update.go` to preserve objective fields
- [ ] Test reading note with objective metadata

**Files**:
- `/home/billy/repos/cli-notes/scripts/data/file_repository.go`
- `/home/billy/repos/cli-notes/scripts/update.go`

### Phase 1.4: Objective ID Generator
- [ ] Create new file `scripts/objective_id.go`
- [ ] Implement `GenerateObjectiveID()` function (8-char hex from crypto/rand)
- [ ] Implement `ValidateObjectiveID(id string)` function
- [ ] Write unit tests in `scripts/objective_id_test.go`
- [ ] Test ID generation uniqueness (generate 1000 IDs, verify no collisions)
- [ ] Test ID validation with valid and invalid inputs

**New File**: `/home/billy/repos/cli-notes/scripts/objective_id.go`

**Dependencies**: None
**Risk**: Low - Simple struct and metadata additions

---

## Phase 2: Data Repository Layer - Objective Queries

**Goal**: Create functions to query objectives and their children

### Phase 2.1: Objective-Specific Queries
- [ ] Create new file `scripts/data/objective_repository.go`
- [ ] Implement `QueryAllObjectives()` - returns all parent objectives
- [ ] Implement `QueryChildrenByObjectiveID(objectiveID, includeDone)` - returns children
- [ ] Implement `GetObjectiveByID(objectiveID)` - finds parent by ID
- [ ] Implement `QueryTodosWithoutObjective(query)` - returns unlinked todos
- [ ] Write unit tests in `scripts/data/objective_repository_test.go`
- [ ] Test each query with various scenarios (empty, single, multiple results)
- [ ] Test filtering logic (done/not done, with/without objective-id)

**New File**: `/home/billy/repos/cli-notes/scripts/data/objective_repository.go`

**Dependencies**: Phase 1 (all)
**Risk**: Medium - Query logic needs to handle edge cases

---

## Phase 3: Business Logic Layer - Objective Operations

**Goal**: Implement core objective operations (create, link, unlink, convert, delete)

### Phase 3.1: Objective Creation Operations
- [ ] Create new file `scripts/objectives.go`
- [ ] Implement `CreateParentObjective(title, onFileCreated)` function
- [ ] Implement `ConvertTodoToParentObjective(file, writeFile)` function
- [ ] Implement `CreateChildTodo(title, parentObjective, onFileCreated)` function
- [ ] Implement tag inheritance logic in `CreateChildTodo`
- [ ] Write unit tests in `scripts/objectives_test.go` for creation operations
- [ ] Test CreateParentObjective generates valid ID and metadata
- [ ] Test ConvertTodoToParentObjective preserves existing content
- [ ] Test CreateChildTodo inherits parent tags correctly

**New File**: `/home/billy/repos/cli-notes/scripts/objectives.go`

### Phase 3.2: Link/Unlink Operations
- [ ] Implement `LinkTodoToObjective(todo, parentObjective, writeFile)` function
- [ ] Add validation to prevent linking parent objectives as children
- [ ] Implement tag inheritance in `LinkTodoToObjective`
- [ ] Implement `UnlinkTodoFromObjective(todo, writeFile)` function
- [ ] Implement `DeleteParentObjective(parent, getChildrenFunc, writeFile)` function
- [ ] Write unit tests for link/unlink/delete operations
- [ ] Test LinkTodoToObjective tag inheritance (child doesn't have parent tags)
- [ ] Test LinkTodoToObjective prevents linking parent as child
- [ ] Test UnlinkTodoFromObjective removes only objective-id
- [ ] Test DeleteParentObjective unlinks all children and removes file

**File**: `/home/billy/repos/cli-notes/scripts/objectives.go` (continued)

**Dependencies**: Phase 1.4, Phase 2
**Risk**: Medium - Tag inheritance and file deletion need careful testing

---

## Phase 4: View State Management

**Goal**: Create state management for objectives views

### Phase 4.1: Objectives View State
- [ ] Create new file `scripts/data/objectives_state.go`
- [ ] Define `ObjectivesViewMode` enum (ObjectivesListView, SingleObjectiveView)
- [ ] Define `SortOrder` enum (SortByDueDateThenPriority, SortByPriorityThenDueDate)
- [ ] Define `FilterMode` enum (ShowAll, ShowIncompleteOnly, ShowCompleteOnly)
- [ ] Define `ObjectivesViewState` struct with all necessary fields
- [ ] Implement `NewObjectivesViewState()` - initializes list view
- [ ] Implement `SelectNext()` and `SelectPrevious()` navigation methods
- [ ] Implement `OpenSelectedObjective()` - transition to single view
- [ ] Implement `BackToList()` - return to list view
- [ ] Implement `Refresh()` - reload current view
- [ ] Implement `applySortAndFilter()` - apply sort/filter to children
- [ ] Implement `GetSelectedObjective()` - returns selected objective in list
- [ ] Implement `GetSelectedChild()` - returns selected child in single view
- [ ] Implement `GetCompletionStats(objectiveID)` - returns (complete, total) counts
- [ ] Write unit tests in `scripts/data/objectives_state_test.go`
- [ ] Test navigation methods (boundary conditions)
- [ ] Test view transitions (list→single, single→list)
- [ ] Test sort and filter application
- [ ] Test state refresh after mutations

**New File**: `/home/billy/repos/cli-notes/scripts/data/objectives_state.go`

**Dependencies**: Phase 2, Phase 3
**Risk**: Medium - Complex state transitions need careful testing

---

## Phase 5: Presentation Layer - UI Rendering

**Goal**: Create UI rendering functions for objectives views

### Phase 5.1: Objectives List UI
- [ ] Create new file `scripts/presentation/objectives_ui.go`
- [ ] Implement `RenderObjectivesListView(state)` function
- [ ] Format objectives with completion status (X/Y complete)
- [ ] Add selection indicator (> for selected)
- [ ] Add command help text at bottom
- [ ] Implement `RenderSingleObjectiveView(state)` function
- [ ] Display parent content at top
- [ ] Add separator between parent and children
- [ ] Implement `renderChildTodo()` helper function
- [ ] Separate incomplete and complete sections
- [ ] Apply filter mode to sections
- [ ] Add command help text for single view
- [ ] Test rendering with various states (empty, single, multiple items)
- [ ] Visual testing in terminal (80x24 minimum)

**New File**: `/home/billy/repos/cli-notes/scripts/presentation/objectives_ui.go`

### Phase 5.2: Input Handling
- [ ] Create new file `scripts/presentation/objectives_input.go`
- [ ] Define `ObjectivesAction` enum with all possible actions
- [ ] Define `ObjectivesInput` struct
- [ ] Implement `ParseObjectivesInput(char, key)` function
- [ ] Map j/k to navigation
- [ ] Map o to open/enter
- [ ] Map n to create new
- [ ] Map l to link existing
- [ ] Map d to delete (handle dd in main loop)
- [ ] Map q to quit/back
- [ ] Map e to edit parent
- [ ] Map u to unlink
- [ ] Map s to sort
- [ ] Map f to filter
- [ ] Test input parsing with all commands

**New File**: `/home/billy/repos/cli-notes/scripts/presentation/objectives_input.go`

**Dependencies**: Phase 4
**Risk**: Low - Rendering and input parsing are straightforward

---

## Phase 6: Main Loop Integration

**Goal**: Integrate objectives views into main command loop

### Phase 6.1: Add Global Commands
- [ ] Modify `main.go` to add `ob` command case
- [ ] Implement `ob` → call `runObjectivesView()`
- [ ] Add `cpo` command case to `handleCommand` switch
- [ ] Implement `cpo` → check if file selected
- [ ] Add warning if file is already a child (show parent info)
- [ ] Implement confirmation prompt for child-to-parent conversion
- [ ] Call `ConvertTodoToParentObjective` on confirmation
- [ ] Display success message with new objective ID
- [ ] Test `cpo` on regular todo
- [ ] Test `cpo` on existing child (with warning)
- [ ] Test `cpo` with no file selected

**File**: `/home/billy/repos/cli-notes/main.go`

### Phase 6.2: Objectives View Event Loop
- [ ] Implement `runObjectivesView()` function in `main.go`
- [ ] Initialize ObjectivesViewState
- [ ] Implement main event loop (for infinite loop until quit)
- [ ] Add rendering logic (render appropriate view based on state)
- [ ] Implement 'dd' detection for delete (track last char)
- [ ] Implement navigation handlers (j/k)
- [ ] Implement open handler (o) - different behavior per view
- [ ] Implement create new handler (n) - different per view
- [ ] Implement quit handler (q) - different per view
- [ ] Implement edit parent handler (e) - single view only
- [ ] Implement unlink handler (u) - single view, child selected only
- [ ] Implement sort handler (s) - toggle sort order
- [ ] Implement filter handler (f) - cycle through filters
- [ ] Implement delete objective with confirmation (dd)
- [ ] Add `getLineInput()` helper function (close/reopen keyboard)
- [ ] Add error message display mechanism (lastMessage)
- [ ] Test keyboard state management (proper cleanup on exit)
- [ ] Test all navigation flows
- [ ] Test view transitions
- [ ] Test error handling

**File**: `/home/billy/repos/cli-notes/main.go` (continued)

**Dependencies**: Phase 4, Phase 5
**Risk**: Medium - Event loop complexity, keyboard state management

---

## Phase 7: Advanced Features

**Goal**: Implement search/link functionality and todo operations

### Phase 7.1: Search and Link Interface
- [ ] Create new file `scripts/presentation/objectives_search.go`
- [ ] Implement `SearchAndLinkTodo(parentObjective)` function
- [ ] Prompt for comma-separated search query
- [ ] Parse query string into array
- [ ] Call `QueryTodosWithoutObjective(queries)` to search
- [ ] Display search results with selection
- [ ] Implement j/k navigation in results
- [ ] Show parent objectives with [PARENT] indicator
- [ ] Show already-linked todos with [OBJ: name] indicator
- [ ] Handle Enter to link selected todo
- [ ] Prevent linking parent objectives (show error)
- [ ] Handle Esc to cancel
- [ ] Return selected todo or nil
- [ ] Wire up to `l` command in `runObjectivesView`
- [ ] Test search with various queries
- [ ] Test search with no results
- [ ] Test linking prevention for parents
- [ ] Test keyboard state transitions

**New File**: `/home/billy/repos/cli-notes/scripts/presentation/objectives_search.go`

### Phase 7.2: Todo Operations in Objective View
- [ ] Add priority change handlers (p1/p2/p3) in `runObjectivesView`
- [ ] Check if in single view and child selected
- [ ] Prompt for priority number
- [ ] Call `ChangePriority` on selected child
- [ ] Refresh state after priority change
- [ ] Add due date handler (t) - set to today
- [ ] Add due date handler (d) - set to specific date
- [ ] Add weekday handlers (m/tu/w/th/f/sa/su) - set to next occurrence
- [ ] Test each operation on child todos
- [ ] Test operations don't work when parent selected
- [ ] Test operations don't work in list view
- [ ] Verify state refresh after each operation

**File**: `/home/billy/repos/cli-notes/main.go` (extend `runObjectivesView`)

**Dependencies**: Phase 6
**Risk**: Low - Reusing existing update functions

---

## Phase 8: Testing & Polish

**Goal**: Comprehensive testing and production readiness

### Phase 8.1: Unit Tests
- [ ] Review test coverage for all new files
- [ ] `scripts/objective_id_test.go` - ID generation and validation
- [ ] `scripts/objectives_test.go` - All CRUD operations
- [ ] `scripts/data/objective_repository_test.go` - All query functions
- [ ] `scripts/data/objectives_state_test.go` - State management
- [ ] Ensure >80% code coverage
- [ ] Run `go test ./...` and verify all tests pass
- [ ] Fix any failing tests

### Phase 8.2: Integration Tests
- [ ] Test complete workflow: Create objective → Link child → Verify tag inheritance
- [ ] Test complete workflow: Convert child to parent → Verify unlinking
- [ ] Test complete workflow: Delete parent → Verify children unlinked
- [ ] Test complete workflow: Search and link → Verify filtering of parents
- [ ] Test complete workflow: From `ob` command to linking multiple todos
- [ ] Test with temporary test directory
- [ ] Verify file system changes after each operation
- [ ] Test view state transitions in complete workflows

### Phase 8.3: Edge Case Tests
- [ ] Test empty objectives (no children)
- [ ] Test circular linking prevention (child cannot be parent)
- [ ] Test deleting non-existent objective
- [ ] Test linking parent as child (should error)
- [ ] Test converting child with existing objective-id
- [ ] Test filtering and sorting with empty lists
- [ ] Test navigation bounds checking (wrap-around)
- [ ] Test very long objective titles (display truncation)
- [ ] Test objectives with many children (50+)
- [ ] Test with special characters in titles

### Phase 8.4: Manual Testing Checklist
- [ ] Create new objective via `ob` → `n`
- [ ] Convert existing todo with `cpo`
- [ ] Convert child todo with `cpo` (verify warning)
- [ ] Link child via search (`l` command)
- [ ] Create new child via `n` in objective view
- [ ] Verify tag inheritance (check file content)
- [ ] Change child priority (p1/p2/p3)
- [ ] Change child due date (t, d, weekday shortcuts)
- [ ] Unlink child (`u` command)
- [ ] Delete objective with confirmation (`dd`)
- [ ] Navigate between list and single views
- [ ] Test with empty objectives
- [ ] Test with 0, 1, and many children
- [ ] Test sort options (s command)
- [ ] Test filter options (f command)
- [ ] Verify week planner still works
- [ ] Verify all existing commands still work (`gt`, `gto`, etc.)
- [ ] Test terminal resize handling
- [ ] Test with 80x24 terminal (minimum size)

### Phase 8.5: Performance Testing
- [ ] Test with 100+ objectives
- [ ] Test with 50+ children per objective
- [ ] Measure search performance with large note sets
- [ ] Profile query performance if needed
- [ ] Optimize slow operations if necessary

### Phase 8.6: Documentation & Polish
- [ ] Add code comments to all public functions
- [ ] Update README.md with objectives commands
- [ ] Verify OBJECTIVES_FEATURE_SPEC.md remains accurate
- [ ] Add example usage to documentation
- [ ] Create user guide section for objectives
- [ ] Clean up any debug logging
- [ ] Remove any commented-out code
- [ ] Format all code with `go fmt`
- [ ] Run linter and fix issues

### Phase 8.7: Final Review
- [ ] Code review (self-review or peer review)
- [ ] Check for security issues (command injection, path traversal)
- [ ] Verify no data loss scenarios
- [ ] Test rollback/recovery from errors
- [ ] Final acceptance testing
- [ ] Create list of known issues (if any)
- [ ] Prepare merge/deployment plan

**Dependencies**: All previous phases
**Risk**: Low - Testing and polish

---

## File Structure Summary

### New Files (8 + tests)
```
scripts/
├── objective_id.go                    # ID generation and validation
├── objective_id_test.go               # ID tests
├── objectives.go                      # Business logic operations
├── objectives_test.go                 # Operations tests
└── data/
    ├── objective_repository.go        # Query functions
    ├── objective_repository_test.go   # Query tests
    ├── objectives_state.go            # View state management
    └── objectives_state_test.go       # State tests

scripts/presentation/
├── objectives_ui.go                   # Rendering functions
├── objectives_input.go                # Input parsing
└── objectives_search.go               # Search & link UI
```

### Modified Files (5)
```
scripts/
├── file.go                            # Add ObjectiveRole, ObjectiveID fields
├── update.go                          # Preserve objective fields in readLatestFileContent
└── data/
    └── file_repository.go             # Parse and write objective metadata

main.go                                 # Add ob, cpo commands + runObjectivesView
```

---

## Risk Assessment & Mitigation

### High Risk Areas

**1. Metadata Persistence** (Phase 1.2, 1.3)
- **Risk**: Data loss if metadata writing fails
- **Mitigation**:
  - Test with backup of notes directory
  - Implement conditional writing (only write if non-empty)
  - Add validation before writing

**2. File Deletion** (Phase 3.2)
- **Risk**: Accidental data loss
- **Mitigation**:
  - Always require confirmation
  - Unlink children before deleting parent
  - Consider testing in isolated directory first

**3. Keyboard State Management** (Phase 6.2, 7.1)
- **Risk**: Terminal left in raw mode if crash occurs
- **Mitigation**:
  - Use defer for keyboard cleanup
  - Follow week planner pattern exactly
  - Test error paths thoroughly

### Medium Risk Areas

**4. Tag Inheritance** (Phase 3.1, 3.2)
- **Risk**: Tag conflicts or duplication
- **Mitigation**:
  - Use set-based logic to prevent duplicates
  - Test with various tag combinations
  - Document inheritance rules clearly

**5. State Synchronization** (Phase 4.1)
- **Risk**: View shows stale data after file operations
- **Mitigation**:
  - Always call Refresh() after mutations
  - Reload from disk, don't cache aggressively

**6. Search Performance** (Phase 7.1)
- **Risk**: Slow searches with many notes
- **Mitigation**:
  - Follow existing query patterns (already performant)
  - Profile with large note directories
  - Consider pagination if needed

---

## Technical Specifications

### Objective ID Format
- **Length**: 8 characters
- **Encoding**: Hexadecimal (lowercase)
- **Source**: 4 bytes from crypto/rand
- **Example**: `7a8f9b2c`
- **Validation**: Must be exactly 8 hex characters

### Metadata Schema Extensions
```yaml
---
title: Example Objective
date-created: 2025-11-08
tags: [objective, frontend]
priority: 1
objective-role: parent        # NEW: "parent" or omitted
objective-id: 7a8f9b2c        # NEW: 8-char hash
---
```

### Tag Inheritance Rules
1. When linking child to parent:
   - Child inherits all parent tags it doesn't already have
   - Exclude "objective" tag from inheritance
   - Use set-based deduplication
2. Inherited tags are written to child's frontmatter
3. Unlinking does NOT remove inherited tags (permanent)

### Sort & Filter Defaults
- **Sort**: Due date → Priority (overdue first, then by due date, then priority)
- **Filter**: Show all (both incomplete and complete)
- **Persistence**: Reset to defaults when entering objective view

---

## Success Criteria

### Functional Requirements
- [ ] Can create parent objectives
- [ ] Can convert existing todos to parents
- [ ] Can create child todos linked to parent
- [ ] Can link existing todos to objectives
- [ ] Can unlink children
- [ ] Can delete parents (with confirmation)
- [ ] Tag inheritance works correctly
- [ ] Can navigate objectives list
- [ ] Can view single objective with children
- [ ] Can change child priority and due dates
- [ ] Can sort and filter children
- [ ] Search and link works with comma-separated queries
- [ ] Prevents linking parents as children
- [ ] Handles empty objectives gracefully

### Non-Functional Requirements
- [ ] No data loss in any scenario
- [ ] Keyboard state properly managed
- [ ] Performance acceptable with 100+ notes
- [ ] UI renders correctly in 80x24 terminal minimum
- [ ] Code follows existing patterns
- [ ] Comprehensive test coverage (>80%)
- [ ] No regressions in existing features
- [ ] Clear error messages for all failures

### Documentation Requirements
- [ ] Code comments for all public functions
- [ ] README updated with objectives commands
- [ ] Feature spec remains accurate
- [ ] Example usage documented

---

## Implementation Timeline

| Week | Phase | Focus | Deliverable |
|------|-------|-------|-------------|
| 1 | Phase 1-2 | Foundation & Data Layer | Can create and query objectives at data layer |
| 2 | Phase 3 | Business Logic | All CRUD operations for objectives functional |
| 3 | Phase 4-5 | Views & UI | Can render both views (not yet interactive) |
| 4 | Phase 6 | Integration | Can use objectives view from main app |
| 5 | Phase 7 | Advanced Features | Fully functional objectives feature |
| 6 | Phase 8 | Testing & Polish | Production-ready code |

---

## Next Steps

### Immediate Actions
1. Create feature branch: `git checkout -b feature/objectives-view`
2. Start with Phase 1.1: Add fields to File struct
3. Commit incrementally after each sub-phase
4. Run tests frequently

### First Commits
1. `feat: add objective metadata fields to File struct`
2. `feat: add objective metadata reading/writing`
3. `feat: add objective ID generator`
4. Continue with Phase 2...

---

## Notes

- Follow existing code patterns throughout
- Test incrementally, don't wait until the end
- Keep commits small and focused
- Update this plan.md as you complete tasks
- Document any deviations from the plan
- Add notes about challenges or decisions made

---

**Last Updated**: 2025-11-08
**Status**: Planning Complete - Ready for Implementation
