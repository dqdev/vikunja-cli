# vikunja-cli

A **GitHub Copilot CLI skill** for managing [Vikunja](https://vikunja.io/) tasks,
projects, labels, comments, relations, and attachments via the Vikunja REST API.

All output is JSON — making it easy for Copilot to parse and act on responses.

---

## Features

- **Projects** — list, get, create, update, delete
- **Tasks** — list (with filter/sort/pagination), get, create, update, mark done, delete
- **Labels** — CRUD + apply/remove on tasks
- **Comments** — list, create, delete
- **Attachments** — list metadata, download to disk
- **Relations** — list, create, delete (subtask, blocking, related, etc.)

## Requirements

- Python 3.10+
- A running Vikunja instance with an API token

## Installation as a Copilot skill

### 1. Clone this repo into your skills directory

**Personal skill** (available in all projects):

```bash
git clone https://github.com/YOUR_USERNAME/vikunja-cli ~/.copilot/skills/vikunja
```

**Project skill** (specific to one repository):

```bash
git clone https://github.com/YOUR_USERNAME/vikunja-cli .github/skills/vikunja
```

### 2. Install dependencies

```bash
pip install -r ~/.copilot/skills/vikunja/requirements.txt
```

### 3. Configure your Vikunja connection

Get an API token from **Vikunja → Settings → API Tokens**, then:

```bash
python ~/.copilot/skills/vikunja/vikunja.py config set \
  --url https://your-vikunja-instance.com \
  --token YOUR_API_TOKEN
```

This writes to `~/.vikunja.json` (permissions: `0600`).

### 4. Reload skills in Copilot CLI

```
/skills reload
```

Or restart the CLI. The skill will appear in `/skills list` as `vikunja`.

---

## Usage

Copilot will automatically use this skill when you ask about Vikunja tasks and projects.
You can also invoke it directly:

```
Show me my overdue Vikunja tasks
```

### Direct CLI usage

```bash
SKILL=~/.copilot/skills/vikunja

# List projects
python $SKILL/vikunja.py projects list

# List overdue tasks
python $SKILL/vikunja.py tasks list --filter "due_date < now && !done" --sort-by due_date

# Create a task
python $SKILL/vikunja.py tasks create --project-id 1 --title "Fix the bug" --priority 4

# Mark a task done
python $SKILL/vikunja.py tasks update 42 --done true

# Download an attachment
python $SKILL/vikunja.py attachments download --task-id 42 --attachment-id 7 --output-dir .
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

Connection settings are stored in `~/.vikunja.json`:

```json
{
  "url": "https://your-vikunja-instance.com",
  "token": "your-api-token"
}
```

Manage with:

```bash
python vikunja.py config set --url URL --token TOKEN
python vikunja.py config show
```

---

## Development

```bash
git clone https://github.com/YOUR_USERNAME/vikunja-cli
cd vikunja-cli
pip install -e ".[dev]"
pytest -v
ruff check .
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
