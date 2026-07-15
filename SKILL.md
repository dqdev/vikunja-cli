---
name: vikunja
description: Manage Vikunja tasks, projects, labels, comments, relations, and attachments through the bundled Go CLI and the Vikunja REST API. Use when the user asks to view, search, create, update, complete, delete, label, comment on, relate, triage, decompose, or download attachments for Vikunja tasks or projects.
---

# Vikunja

Use the bundled Go CLI to operate on Vikunja. Run commands from this skill's
base directory unless the `vikunja` binary is already on `PATH`.

```bash
./vikunja <group> <command> [options]
```

All command output is JSON:

- Single item: `{"data": {...}}`
- List: `{"data": [...], "_pagination": {"total_pages": N, "result_count": N}}`
- Delete or no content: `{"data": null}`
- Error: `{"error": "...", "code": N, "http_status": N}`

## Prerequisites

1. A compiled `vikunja` binary. From the skill directory, run `go build -o vikunja .`.
2. A Vikunja API token from Vikunja Settings > API Tokens.
3. One-time config:

```bash
./vikunja config set --url https://vikunja.example.com --token YOUR_API_TOKEN
```

Config is stored as `.vikunja.json` in this skill's installed directory, next
to the `vikunja` binary, with file mode `0600` when possible. Use
`./vikunja config show` to confirm the URL; the token is masked.

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
- Apply the Task standards below whenever creating, updating, decomposing, or
  interpreting tasks for AI or human execution.

## Task standards

- Every created task must have exactly one complexity label:
  `complexity:simple`, `complexity:medium`, or `complexity:hard`.
- Use `complexity:simple` for work a smaller/cheaper model such as GPT mini or
  Claude Haiku can complete; use `complexity:medium` for work appropriate for a
  frontier general model such as GPT-5.5 or Claude Sonnet; use
  `complexity:hard` for work that needs the strongest available models, such as
  Anthropic Fable or GPT-5.6.
- Add the `Human-Only` label when the deliverable requires a human decision,
  credential, approval, physical action, account ownership, legal judgment, or
  other capability an AI agent cannot safely complete.
- Create one task per deliverable target. The task must be granular enough to be
  implemented on its own git branch, reviewed, and merged without blocking or
  changing the scope of another task.
- If a task cannot be started or finished until another task is done, mark it
  blocked: add a `blocked` label if available, create a `blocked` relation from
  the waiting task to each blocking task, and mention the blocking task IDs in
  the description or a comment.
- Before creating or applying required labels, search existing labels by exact
  title. Create missing labels only when the user has not forbidden label
  creation.

## config — Connection settings

```bash
# Save connection details
./vikunja config set --url https://vikunja.example.com --token YOUR_API_TOKEN

# Show current config (token masked)
./vikunja config show
```

## projects — Project management

```bash
# List projects (paginated)
./vikunja projects list [--page N] [--per-page N] [--search TEXT] [--archived]

# Get a single project
./vikunja projects get PROJECT_ID

# Create a project
./vikunja projects create --title "My Project" [--description TEXT] [--parent-id N]

# Update a project
./vikunja projects update PROJECT_ID [--title TEXT] [--description TEXT]

# Delete a project (and all its tasks)
./vikunja projects delete PROJECT_ID
```

## tasks — Task management

```bash
# List all tasks (optionally filtered)
./vikunja tasks list \
  [--project-id N] \
  [--page N] [--per-page N] \
  [--search TEXT] \
  [--filter 'due_date < now && !done'] \
  [--sort-by due_date] [--order-by asc|desc]

# Get a single task
./vikunja tasks get TASK_ID

# Create a task in a project
./vikunja tasks create \
  --project-id N \
  --title "Task title" \
  [--description TEXT] \
  [--due-date 2024-12-31T23:59:59Z] \
  [--priority 0-5] \
  [--label-ids 1,2,3]

# Update a task (provide only fields to change)
./vikunja tasks update TASK_ID \
  [--title TEXT] \
  [--description TEXT] \
  [--due-date 2024-12-31T23:59:59Z] \
  [--priority 0-5] \
  [--done true|false] \
  [--percent-done 0.0-1.0]

# Delete a task
./vikunja tasks delete TASK_ID
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
./vikunja labels list [--search TEXT] [--page N] [--per-page N]

# Get a label
./vikunja labels get LABEL_ID

# Create a label
./vikunja labels create --title "Bug" [--color "#FF5733"] [--description TEXT]

# Update a label
./vikunja labels update LABEL_ID [--title TEXT] [--color "#RRGGBB"] [--description TEXT]

# Delete a label
./vikunja labels delete LABEL_ID

# Apply a label to a task
./vikunja labels apply --task-id N --label-id N

# Remove a label from a task
./vikunja labels remove --task-id N --label-id N
```

## comments — Task comments

```bash
# List comments on a task
./vikunja comments list --task-id N

# Add a comment
./vikunja comments create --task-id N --text "Comment text (Markdown supported)"

# Delete a comment
./vikunja comments delete --task-id N --comment-id N
```

## attachments — Task attachments

```bash
# List attachment metadata for a task
./vikunja attachments list --task-id N

# Download an attachment to disk
./vikunja attachments download --task-id N --attachment-id N [--output-dir ./downloads]
# Returns: {"data": {"saved_to": "/abs/path/to/file.pdf", "filename": "file.pdf"}}
```

## relations — Task relations

Supported relation kinds: `subtask`, `parenttask`, `related`, `duplicateof`,
`duplicates`, `blocking`, `blocked`, `precedes`, `follows`, `copiedfrom`, `copiedto`.

```bash
# List all relations for a task
./vikunja relations list TASK_ID

# Create a relation
./vikunja relations create --task-id N --other-id N --kind subtask

# Delete a relation
./vikunja relations delete --task-id N --other-id N --kind subtask
```

## Common workflows

**Find overdue tasks:**
```bash
./vikunja tasks list --filter "due_date < now && !done" --sort-by due_date
```

**Mark a task done:**
```bash
./vikunja tasks update 42 --done true
```

**Create a task with labels:**
```bash
./vikunja tasks create --project-id 5 --title "Fix login bug" --label-ids 1,3 --priority 4
```

**Download all attachments for a task:**
```bash
# First list to get IDs
./vikunja attachments list --task-id 42
# Then download each one
./vikunja attachments download --task-id 42 --attachment-id 7 --output-dir ./downloads
```
