# vikunja-cli

A portable **Agent Skill** for managing [Vikunja](https://vikunja.io/) tasks,
projects, labels, comments, relations, and attachments via the Vikunja REST API.
It is packaged around `SKILL.md` so agents that support the Agent Skills
directory format, including Codex, Claude Code, and OpenCode, can load it.

All CLI output is JSON, making responses easy for coding agents to parse and act on.
The CLI is implemented in Go and ships as a single binary with no runtime package
installation step.

---

## Features

- **Projects** — list, get, create, update, delete
- **Tasks** — list (with filter/sort/pagination), get, create, update, mark done, delete
- **Labels** — CRUD + apply/remove on tasks
- **Comments** — list, create, delete
- **Attachments** — list metadata, download to disk
- **Relations** — list, create, delete (subtask, blocking, related, etc.)

## Requirements

- Go 1.22+ to build the binary, or a prebuilt `vikunja` binary
- A running Vikunja instance with an API token

## Installation as an agent skill

### 1. Clone this repo into your agent's skills directory

Use the skills directory configured by your agent. Common layouts are:

```bash
# Codex
git clone https://github.com/YOUR_USERNAME/vikunja-cli ~/.codex/skills/vikunja

# Claude Code
git clone https://github.com/YOUR_USERNAME/vikunja-cli ~/.claude/skills/vikunja

# OpenCode
git clone https://github.com/YOUR_USERNAME/vikunja-cli ~/.config/opencode/skills/vikunja
```

For a project-local skill, clone this repository into the project-specific skills
directory supported by your agent.

### 2. Build the CLI binary

```bash
cd ~/.codex/skills/vikunja
go build -o vikunja .
```

Adjust the path if you installed the skill somewhere else.

### 3. Configure your Vikunja connection

Get an API token from **Vikunja → Settings → API Tokens**, then:

```bash
~/.codex/skills/vikunja/vikunja config set \
  --url https://your-vikunja-instance.com \
  --token YOUR_API_TOKEN
```

This writes to `.vikunja.json` next to the installed `vikunja` binary
(permissions: `0600`).

### 4. Reload skills in your agent

```text
/skills reload
```

Or restart the agent if it does not support live skill reloads. The skill should
appear as `vikunja`.

---

## Usage

Your agent should automatically use this skill when you ask about Vikunja tasks
and projects.
You can also invoke it directly:

```
Show me my overdue Vikunja tasks
```

### Direct CLI usage

```bash
SKILL=~/.codex/skills/vikunja

# List projects
$SKILL/vikunja projects list

# List overdue tasks
$SKILL/vikunja tasks list --filter "due_date < now && !done" --sort-by due_date

# Create a task
$SKILL/vikunja tasks create --project-id 1 --title "Fix the bug" --priority 4

# Mark a task done
$SKILL/vikunja tasks update 42 --done true

# Download an attachment
$SKILL/vikunja attachments download --task-id 42 --attachment-id 7 --output-dir .
```

All commands output JSON. Errors look like:
```json
{"error": "Task not found", "code": 4002, "http_status": 404}
```

---

## Output format

| Operation | Shape |
|-----------|-------|
| Single item | `{"data": {...}}` |
| List | `{"data": [...], "_pagination": {"total_pages": N, "result_count": N}}` |
| Delete/no-content | `{"data": null}` |
| Error | `{"error": "...", "code": N, "http_status": N}` |

---

## Configuration

Connection settings are stored in `.vikunja.json` next to the installed
`vikunja` binary:

```json
{
  "url": "https://your-vikunja-instance.com",
  "token": "your-api-token"
}
```

Manage with:

```bash
./vikunja config set --url URL --token TOKEN
./vikunja config show
```

---

## Development

```bash
git clone https://github.com/YOUR_USERNAME/vikunja-cli
cd vikunja-cli
go build -o vikunja .
./vikunja config show
```

---

## Filter syntax

The `--filter` flag in `tasks list` uses Vikunja's query language:

```
!done
due_date < now
priority >= 3 && !done
title like "%bug%"
project_id = 5 && due_date < now
```

See [Vikunja filter docs](https://vikunja.io/docs/filters) for full reference.

---

## License

MIT — see [LICENSE](LICENSE).
