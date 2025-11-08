# Objectives View Feature Specification

## Overview

The Objectives View feature allows users to organize and track larger goals by linking multiple todos together. An objective consists of one parent note and zero or more child todos, all connected through a unique objective ID. This provides a hierarchical view where you can see all related work in one place, track progress, and manage tasks associated with a specific goal.

## Core Concepts

### What is an Objective?

An **objective** is a group of linked notes consisting of:
- **One parent note**: The main objective with overview, context, and high-level planning
- **Multiple child todos**: Individual tasks that contribute to completing the objective

### Key Principles

1. **Parent notes** can have their own content, checklists, and any other content a regular todo has
2. **Child todos** function as normal todos but are linked to their parent objective
3. **Linking is managed** through unique objective IDs in the note metadata
4. **Children are independent**: Deleting a parent objective unlinks children but doesn't delete them
5. **Single hierarchy**: Parent objectives cannot be children of other objectives
6. **Inheritance**: Children automatically inherit tags from parent when linked (only tags they don't already have)

## Metadata Structure

### Parent Objective Frontmatter
```yaml
---
title: Launch New Feature
date-created: 2025-11-08
tags: [project, frontend]
priority: P1
objective-role: parent
objective-id: 7a8f9b2c
---
```

### Child Todo Frontmatter
```yaml
---
title: Implement API endpoints
date-created: 2025-11-08
tags: [project, frontend, backend]
priority: P1
date-due: 2025-11-15
done: false
objective-id: 7a8f9b2c
---
```

### ID Format
- **objective-id**: 8-character hash (e.g., `7a8f9b2c`)
- Generated when parent objective is created
- Shared across parent and all children

## Commands Reference

### Global Commands

| Command | Description |
|---------|-------------|
| `ob` | Open objectives list view |
| `cpo` | Convert selected todo to parent objective |

### Objectives List View Commands

| Command | Description |
|---------|-------------|
| `j/k` | Navigate up/down through objectives |
| `o` | Open selected objective |
| `n` | Create new objective |
| `l` | Link existing todo to selected objective |
| `dd` | Delete selected objective (with confirmation) |
| `q` | Quit objectives view |

### Single Objective View Commands

| Command | Description |
|---------|-------------|
| `j/k` | Navigate through parent and child todos |
| `o` | Open selected item in editor |
| `n` | Create new child todo (auto-linked) |
| `l` | Link existing todo to this objective |
| `e` | Edit parent objective in editor |
| `u` | Unlink selected child todo |
| `s` | Change sort order |
| `f` | Change filter (show all/incomplete/complete) |
| `q` | Back to objectives list |

### Child Todo Operations (from Objective View)

When a child todo is selected in the objective view, all standard todo commands work:

| Command | Description |
|---------|-------------|
| `p1`, `p2`, `p3` | Set priority |
| `d` | Set specific due date |
| `t` | Set due date to tomorrow |
| `m`, `tu`, `w`, `th`, `f`, `sa`, `su` | Set due date to next occurrence of day |

## User Workflows

### 1. Creating a New Objective

**Command**: `ob` → `n`

```
Create new objective
Title: Launch New Feature_

[User enters title and presses Enter]

Created objective: "Launch New Feature" (7a8f9b2c)

Press any key to return to objectives list...
```

The system:
1. Creates a new note with `objective-role: parent` and generates unique `objective-id`
2. Returns to objectives list showing the new objective

---

### 2. Converting Existing Todo to Parent Objective

**From any todo view** (e.g., `gt`, `gto`):

```
================================
TODOS
================================
> [P1] Build new dashboard feature (due: 2025-11-30)
  [P2] Fix login bug (due: 2025-11-09)
  [P3] Update documentation

[User navigates to desired todo and types 'cpo']

Convert "Build new dashboard feature" to parent objective? (y/n): y_

Converted to parent objective.
Objective ID: 7a8f9b2c

Press any key to continue...
```

The system:
1. Adds `objective-role: parent` and generates `objective-id`
2. Preserves all existing content, frontmatter, and checklists
3. Returns to previous view

---

### 3. Viewing Objectives List

**Command**: `ob`

```
================================
OBJECTIVES
================================
> Launch New Feature (3/5 complete)
  Refactor Authentication (1/8 complete)
  Q1 Planning (0/3 complete)

j/k=navigate, o=open, n=create, l=link, dd=delete, q=quit
```

The system:
- Lists objectives sorted by most recently created first
- Shows completion status: (completed/total child todos)
- Allows navigation with j/k

---

### 4. Viewing Single Objective

**Command**: `ob` → select objective → `o`

```
================================
OBJECTIVE: Launch New Feature [7a8f9b2c]
================================

## Overview
Build and ship the new dashboard feature

## Key Requirements
- [ ] Performance under 2s load time
- [ ] Mobile responsive
- [x] Design approved

## Implementation Notes
Focus on the checkout flow first, then user profile pages.

─────────────────────────────────
LINKED TODOS (2 incomplete, 3 complete)
─────────────────────────────────

INCOMPLETE (sort: due date → priority):
> [P1] Implement API endpoints (due: 2025-11-15)
  [P2] Write integration tests (due: 2025-11-20)

COMPLETE (sort: due date → priority):
  [P3] Design mockups
  [P1] Security review
  [P2] Database schema

j/k=navigate, o=open, n=new child, l=link existing,
e=edit parent, u=unlink, s=sort, f=filter, q=back
```

The system:
- Shows complete parent note content at top
- Lists linked child todos below, separated into incomplete/complete sections
- Default sort: due date → priority
- Default filter: show all

---

### 5. Creating New Child Todo

**From objective view**: `n`

```
[Runs standard 'ct' create todo flow]

Create Todo Title: Implement error handling_

[User completes creation flow]

Created and linked: [P2] Implement error handling

Press any key to return to objective view...
```

The system:
1. Runs the same flow as `ct` command
2. Automatically adds `objective-id` to the new todo
3. Inherits tags from parent (only tags child doesn't already have)
4. Returns to objective view

---

### 6. Linking Existing Todo to Objective

**From objectives list**: `l`

```
================================
OBJECTIVES
================================
  Launch New Feature (3/5 complete)
> Refactor Authentication (1/8 complete)
  Q1 Planning (0/3 complete)

Link existing todo to: Refactor Authentication
Enter search query: api,endpoint_

[User enters search terms separated by commas]

Searching for todos containing "api" AND "endpoint"...

> Fix API endpoint rate limiting (P1, due: 2025-11-12)
  Update API endpoint docs (P2, no due date) [OBJ: Launch Feature]
  API endpoint error handling (P1, due: 2025-11-10)

j/k=navigate, Enter=link, Esc=cancel

Note: "Update API endpoint docs" already linked to "Launch Feature"
```

After selecting a todo:

```
Linked "Fix API endpoint rate limiting" to objective "Refactor Authentication"

Press any key to return...
```

**From objective view**: Same flow triggered with `l`

The system:
1. Searches through all non-done notes using comma-separated grep chains
2. Excludes todos already linked to this objective
3. Shows todos already linked to other objectives with `[OBJ: name]` indicator
4. Adds `objective-id` to selected todo
5. Inherits tags from parent

---

### 7. Unlinking Child Todo

**From objective view**: Select child → `u`

```
================================
OBJECTIVE: Launch New Feature
================================

[...parent content...]

INCOMPLETE:
> [P1] Implement API endpoints (due: 2025-11-15)
  [P2] Write integration tests (due: 2025-11-20)

[User navigates to "Implement API endpoints" and presses 'u']

Unlink "Implement API endpoints" from this objective? (y/n): y_

Unlinked successfully.

Press any key to continue...
```

The system:
1. Shows confirmation prompt
2. Removes `objective-id` from the child todo
3. Todo remains in view until screen refreshes
4. Todo continues to exist as independent note

---

### 8. Deleting Parent Objective

**From objectives list**: Select objective → `dd`

```
================================
OBJECTIVES
================================
  Launch New Feature (3/5 complete)
> Refactor Authentication (1/8 complete)
  Q1 Planning (0/3 complete)

[User presses 'dd']

Delete objective "Refactor Authentication"?
(1 linked todo will be unlinked but not deleted)
(y/n): y_

Deleted successfully.

Press any key to continue...
```

The system:
1. Shows warning about number of linked children
2. Requires confirmation
3. Deletes parent note file
4. Removes `objective-id` from all child todos (unlinks them)
5. Child todos continue to exist as independent notes

---

### 9. Sorting and Filtering

**Changing Sort Order**: `s`

```
Select sort order:
1. Due date → Priority (default)
2. Priority → Due date

Selection: _
```

**Changing Filter**: `f`

```
Filter todos:
1. Show all (default)
2. Incomplete only
3. Complete only

Selection: _
```

The system:
- Applies selected sort/filter to current view
- Resets to defaults when returning to the objective later

---

### 10. Editing Parent from Objective View

**From objective view**: `e`

Opens the parent objective file in the system editor (same behavior as `o` on regular notes).

After saving and closing editor, returns to objective view with updated content.

## Sorting & Filtering

### Default Display Rules

**Sort Order**: Due date → Priority
- Todos with due dates come before todos without due dates
- Within same due date (or no due date), sorted by priority (P1 > P2 > P3)
- Both incomplete and complete sections use same sort rules

**Filter**: Show all
- Displays both incomplete and complete sections
- Complete section always shown below incomplete section

### Filter Options

1. **Show all**: Default, shows both sections
2. **Incomplete only**: Shows only incomplete section
3. **Complete only**: Shows only complete section

### State Persistence

Sort and filter settings reset to defaults each time you enter an objective view.

## Edge Cases & Behaviors

### Converting Child to Parent

If you try to convert a todo that is already a child (`cpo` on a todo with `objective-id`):

```
================================
WARNING
================================
"Implement API endpoints" is currently linked to:
  Parent Objective: "Launch New Feature" (7a8f9b2c)

Converting to parent objective will:
  • Unlink from "Launch New Feature"
  • Create new objective ID
  • Become independent parent objective

Continue? (y/n): _
```

System behavior:
1. Shows warning with details
2. If confirmed, removes existing `objective-id`
3. Adds `objective-role: parent` and generates new `objective-id`
4. Original parent objective no longer references this todo

### Linking Parent as Child

If search results include a parent objective when linking:

```
Searching for todos containing "dashboard"...

> [P1] Dashboard performance improvements (P1, due: 2025-11-15)
  Dashboard Feature Objective (P1) [PARENT]

[User tries to select parent objective]

Error: Cannot link parent objectives as children.

Press any key to continue...
```

System behavior:
1. Parent objectives appear in search results with `[PARENT]` indicator
2. Attempting to link shows error message
3. Returns to search results

### Empty Objectives

When viewing an objective with no linked children:

```
================================
OBJECTIVE: Q1 Planning [9f3a7e1b]
================================

## Goals
Plan out Q1 priorities and set team objectives

─────────────────────────────────
LINKED TODOS (0 incomplete, 0 complete)
─────────────────────────────────

No linked todos yet.

j/k=navigate, o=open parent, n=new child, l=link existing,
e=edit parent, q=back
```

### Integration with Existing Views

When viewing todos in existing views (`gt`, `gto`, `gtnd`, etc.):
- Todos linked to objectives display normally
- No indication of objective membership in these views
- Objective information only visible in objective view (`ob`)

### Tags Inheritance

When linking a child to a parent:
- Child inherits all tags from parent that it doesn't already have
- Existing child tags are preserved
- Example:
  - Parent tags: `[project, frontend]`
  - Child tags before linking: `[backend, api]`
  - Child tags after linking: `[backend, api, project, frontend]`

## Navigation Summary

### Objectives List View
```
j/k      Navigate through objectives
o        Open selected objective
n        Create new objective
l        Link existing todo to selected objective
dd       Delete selected objective
q        Quit to main menu
```

### Single Objective View
```
j/k      Navigate through parent and children
o        Open selected item in editor
n        Create new child todo
l        Link existing todo
e        Edit parent objective
u        Unlink selected child
s        Change sort order
f        Change filter
q        Back to objectives list

When child selected:
  p1/p2/p3     Set priority
  d            Set specific due date
  t            Tomorrow
  m/tu/w/...   Next occurrence of weekday
```

## Future Considerations

Potential enhancements not included in this initial specification:
- Multi-level hierarchies (parents as children)
- Objective templates
- Progress tracking beyond simple completion percentage
- Objective archiving
- Filtering objectives list by tags or status
- Bulk operations on children
- Objective-specific views in other contexts
