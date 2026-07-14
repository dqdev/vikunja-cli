---
name: vikunja
description: Manage Vikunja tasks, projects, labels, comments, relations, and attachments through the bundled Python CLI and the Vikunja REST API. Use when the user asks to view, search, create, update, complete, delete, label, comment on, relate, or download attachments for Vikunja tasks or projects.
---

# Vikunja

Use the bundled Python CLI to operate on Vikunja. Run commands from this skill's
base directory unless the agent has installed the package entry point.

```bash
python3 vikunja.py <group> <command> [options]
```

All command output is JSON:

- Single item: `{"data": {...}}`
- List: `{"data": [...], "_pagination": {"total_pages": N, "result_count": N}}`
- Delete or no content: `{"data": null}`
- Error: `{"error": "...", "code": N, "http_status": N}`

## Prerequisites

1. Python 3.10+.
2. Dependencies installed with `pip install -r requirements.txt`.
3. A Vikunja API token from Vikunja Settings > API Tokens.
4. One-time config:

```bash
python3 vikunja.py config set --url https://vikunja.example.com --token YOUR_API_TOKEN
```

Config is stored in `~/.vikunja.json` with file mode `0600` when possible. Use
`python3 vikunja.py config show` to confirm the URL; the token is masked.

## Agent Workflow

- Prefer read-only commands first when IDs, project names, label names, or task
  state are ambiguous.
- Ask for confirmation before deleting projects or tasks unless the user clearly
  requested that exact deletion.
- Parse the JSON response instead of scraping text.
- Treat nonzero exit status plus a JSON `error` payload as the failure details to
  report or recover from.
- Quote filter expressions in the shell, especially when they include `&&`,
  `||`, `<`, `>`, or `!`.

## config — Connection settings

```bash
# Save connection details
python3 vikunja.py config set --url https://vikunja.example.com --token YOUR_API_TOKEN

# Show current config (token masked)
python3 vikunja.py config show
```

## projects — Project management

```bash
# List projects (paginated)
python3 vikunja.py projects list [--page N] [--per-page N] [--search TEXT] [--archived]

# Get a single project
python3 vikunja.py projects get PROJECT_ID

# Create a project
python3 vikunja.py projects create --title "My Project" [--description TEXT] [--parent-id N]

# Update a project
python3 vikunja.py projects update PROJECT_ID [--title TEXT] [--description TEXT]

# Delete a project (and all its tasks)
python3 vikunja.py projects delete PROJECT_ID
```

## tasks — Task management

```bash
# List all tasks (optionally filtered)
python3 vikunja.py tasks list \
  [--project-id N] \
  [--page N] [--per-page N] \
  [--search TEXT] \
  [--filter 'due_date < now && !done'] \
  [--sort-by due_date] [--order-by asc|desc]

# Get a single task
python3 vikunja.py tasks get TASK_ID

# Create a task in a project
python3 vikunja.py tasks create \
  --project-id N \
  --title "Task title" \
  [--description TEXT] \
  [--due-date 2024-12-31T23:59:59Z] \
  [--priority 0-5] \
  [--label-ids 1,2,3]

# Update a task (provide only fields to change)
python3 vikunja.py tasks update TASK_ID \
  [--title TEXT] \
  [--description TEXT] \
  [--due-date 2024-12-31T23:59:59Z] \
  [--priority 0-5] \
  [--done true|false] \
  [--percent-done 0.0-1.0]

# Delete a task
python3 vikunja.py tasks delete TASK_ID
```

### Filter syntax

The `--filter` flag uses Vikunja's filter query language:
- Fields: `title`, `description`, `done`, `due_date`, `priority`, `start_date`, `end_date`, `percent_done`, `created`, `updated`, `project_id`
- Operators: `=`, `!=`, `<`, `<=`, `>`, `>=`, `like`
- Values: bare strings, numbers, `now`, `today`, `true`, `false`
- Combinators: `&&` (and), `||` (or), `!` prefix (not)
- Examples: `!done`, `due_date < now`, `priority >= 3 && !done`, `title like "%bug%"`

Use RFC3339 for dates: `2024-12-31T23:59:59Z` or
`2024-12-31T10:00:00+02:00`.

### Priority levels

| Value | Meaning |
|-------|---------|
| 0 | Unset |
| 1 | Low |
| 2 | Medium |
| 3 | High |
| 4 | Urgent |
| 5 | DO NOW |

## labels — Label management

```bash
# List all labels
python3 vikunja.py labels list [--search TEXT] [--page N] [--per-page N]

# Get a label
python3 vikunja.py labels get LABEL_ID

# Create a label
python3 vikunja.py labels create --title "Bug" [--color "#FF5733"] [--description TEXT]

# Update a label
python3 vikunja.py labels update LABEL_ID [--title TEXT] [--color "#RRGGBB"] [--description TEXT]

# Delete a label
python3 vikunja.py labels delete LABEL_ID

# Apply a label to a task
python3 vikunja.py labels apply --task-id N --label-id N

# Remove a label from a task
python3 vikunja.py labels remove --task-id N --label-id N
```

## comments — Task comments

```bash
# List comments on a task
python3 vikunja.py comments list --task-id N

# Add a comment
python3 vikunja.py comments create --task-id N --text "Comment text (Markdown supported)"

# Delete a comment
python3 vikunja.py comments delete --task-id N --comment-id N
```

## attachments — Task attachments

```bash
# List attachment metadata for a task
python3 vikunja.py attachments list --task-id N

# Download an attachment to disk
python3 vikunja.py attachments download --task-id N --attachment-id N [--output-dir ./downloads]
# Returns: {"data": {"saved_to": "/abs/path/to/file.pdf", "filename": "file.pdf"}}
```

## relations — Task relations

Supported relation kinds: `subtask`, `parenttask`, `related`, `duplicateof`,
`duplicates`, `blocking`, `blocked`, `precedes`, `follows`, `copiedfrom`, `copiedto`.

```bash
# List all relations for a task
python3 vikunja.py relations list TASK_ID

# Create a relation
python3 vikunja.py relations create --task-id N --other-id N --kind subtask

# Delete a relation
python3 vikunja.py relations delete --task-id N --other-id N --kind subtask
```

## Common workflows

**Find overdue tasks:**
```bash
python3 vikunja.py tasks list --filter "due_date < now && !done" --sort-by due_date
```

**Mark a task done:**
```bash
python3 vikunja.py tasks update 42 --done true
```

**Create a task with labels:**
```bash
python3 vikunja.py tasks create --project-id 5 --title "Fix login bug" --label-ids 1,3 --priority 4
```

**Download all attachments for a task:**
```bash
# First list to get IDs
python3 vikunja.py attachments list --task-id 42
# Then download each one
python3 vikunja.py attachments download --task-id 42 --attachment-id 7 --output-dir ./downloads
```
