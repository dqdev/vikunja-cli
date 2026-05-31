"""Task CRUD commands."""

import json
import click
from vikunja.client import VikunjaClient


@click.group()
def tasks():
    """Manage Vikunja tasks."""


@tasks.command("list")
@click.option("--project-id", default=None, type=int, help="Filter tasks to this project ID")
@click.option("--page", default=1, show_default=True, type=int, help="Page number")
@click.option("--per-page", default=50, show_default=True, type=int, help="Items per page (max 100)")
@click.option("--search", default=None, help="Search tasks by text")
@click.option("--filter", "filter_str", default=None,
              help='Vikunja filter expression, e.g. "due_date < now && !done"')
@click.option("--sort-by", default=None,
              help="Field to sort by: id, title, due_date, priority, created, updated, etc.")
@click.option("--order-by", default=None, type=click.Choice(["asc", "desc"]), help="Sort order")
def list_tasks(project_id, page, per_page, search, filter_str, sort_by, order_by):
    """List tasks. Use --project-id to scope to a project, --filter for advanced filtering."""
    client = VikunjaClient()
    params: dict = {"page": page, "per_page": per_page}
    if search:
        params["s"] = search
    if sort_by:
        params["sort_by"] = sort_by
    if order_by:
        params["order_by"] = order_by

    # Combine project_id and user-supplied filter
    filter_parts = []
    if project_id:
        filter_parts.append(f"project_id = {project_id}")
    if filter_str:
        filter_parts.append(f"({filter_str})")
    if filter_parts:
        params["filter"] = " && ".join(filter_parts)

    result = client.get("/tasks", params=params, paginated=True)
    print(json.dumps(result))


@tasks.command("get")
@click.argument("task_id", type=int)
def get_task(task_id):
    """Get a single task by ID."""
    client = VikunjaClient()
    result = client.get(f"/tasks/{task_id}")
    print(json.dumps(result))


@tasks.command("create")
@click.option("--project-id", required=True, type=int, help="Project to create the task in")
@click.option("--title", required=True, help="Task title")
@click.option("--description", default=None, help="Task description (Markdown supported)")
@click.option("--due-date", default=None,
              help="Due date in RFC3339 format, e.g. 2024-12-31T23:59:59Z")
@click.option("--priority", default=None, type=click.IntRange(0, 5),
              help="Priority 0 (unset) to 5 (highest)")
@click.option("--label-ids", default=None,
              help="Comma-separated label IDs to attach, e.g. 1,4,7")
def create_task(project_id, title, description, due_date, priority, label_ids):
    """Create a new task in a project."""
    client = VikunjaClient()
    body: dict = {"title": title}
    if description is not None:
        body["description"] = description
    if due_date is not None:
        body["due_date"] = due_date
    if priority is not None:
        body["priority"] = priority
    if label_ids:
        ids = [int(x.strip()) for x in label_ids.split(",") if x.strip()]
        body["labels"] = [{"id": i} for i in ids]
    result = client.put(f"/projects/{project_id}/tasks", data=body)
    print(json.dumps(result))


@tasks.command("update")
@click.argument("task_id", type=int)
@click.option("--title", default=None, help="New title")
@click.option("--description", default=None, help="New description")
@click.option("--due-date", default=None, help="New due date (RFC3339) or empty string to clear")
@click.option("--priority", default=None, type=click.IntRange(0, 5), help="Priority 0–5")
@click.option("--done", default=None, type=click.Choice(["true", "false"]),
              help="Mark as done (true) or not done (false)")
@click.option("--percent-done", default=None, type=click.FloatRange(0.0, 1.0),
              help="Completion percentage as a decimal (0.0–1.0)")
def update_task(task_id, title, description, due_date, priority, done, percent_done):
    """Update one or more fields on a task."""
    client = VikunjaClient()
    body: dict = {}
    if title is not None:
        body["title"] = title
    if description is not None:
        body["description"] = description
    if due_date is not None:
        body["due_date"] = due_date if due_date else None
    if priority is not None:
        body["priority"] = priority
    if done is not None:
        body["done"] = done == "true"
    if percent_done is not None:
        body["percent_done"] = percent_done
    if not body:
        print(json.dumps({"error": "Provide at least one field to update"}))
        raise SystemExit(1)
    result = client.post(f"/tasks/{task_id}", data=body)
    print(json.dumps(result))


@tasks.command("delete")
@click.argument("task_id", type=int)
def delete_task(task_id):
    """Delete a task permanently."""
    client = VikunjaClient()
    result = client.delete(f"/tasks/{task_id}")
    print(json.dumps(result))
