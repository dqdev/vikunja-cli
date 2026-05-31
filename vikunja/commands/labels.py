"""Label CRUD and task label commands."""

import json
import click
from vikunja.client import VikunjaClient


@click.group()
def labels():
    """Manage labels and apply them to tasks."""


@labels.command("list")
@click.option("--page", default=1, show_default=True, type=int, help="Page number")
@click.option("--per-page", default=50, show_default=True, type=int, help="Items per page")
@click.option("--search", default=None, help="Search labels by title")
def list_labels(page, per_page, search):
    """List all labels accessible to the current user."""
    client = VikunjaClient()
    params: dict = {"page": page, "per_page": per_page}
    if search:
        params["s"] = search
    result = client.get("/labels", params=params, paginated=True)
    print(json.dumps(result))


@labels.command("get")
@click.argument("label_id", type=int)
def get_label(label_id):
    """Get a single label by ID."""
    client = VikunjaClient()
    result = client.get(f"/labels/{label_id}")
    print(json.dumps(result))


@labels.command("create")
@click.option("--title", required=True, help="Label title")
@click.option("--color", default=None, help="Hex color code, e.g. #FF5733")
@click.option("--description", default=None, help="Optional description")
def create_label(title, color, description):
    """Create a new label."""
    client = VikunjaClient()
    body: dict = {"title": title}
    if color:
        body["hex_color"] = color.lstrip("#")
    if description:
        body["description"] = description
    result = client.put("/labels", data=body)
    print(json.dumps(result))


@labels.command("update")
@click.argument("label_id", type=int)
@click.option("--title", default=None, help="New title")
@click.option("--color", default=None, help="New hex color, e.g. #00BFFF")
@click.option("--description", default=None, help="New description")
def update_label(label_id, title, color, description):
    """Update an existing label (you must be the creator)."""
    client = VikunjaClient()
    body: dict = {}
    if title:
        body["title"] = title
    if color:
        body["hex_color"] = color.lstrip("#")
    if description is not None:
        body["description"] = description
    if not body:
        print(json.dumps({"error": "Provide at least one of --title, --color, or --description"}))
        raise SystemExit(1)
    result = client.put(f"/labels/{label_id}", data=body)
    print(json.dumps(result))


@labels.command("delete")
@click.argument("label_id", type=int)
def delete_label(label_id):
    """Delete a label (you must be the creator)."""
    client = VikunjaClient()
    result = client.delete(f"/labels/{label_id}")
    print(json.dumps(result))


@labels.command("apply")
@click.option("--task-id", required=True, type=int, help="Task ID")
@click.option("--label-id", required=True, type=int, help="Label ID to apply")
def apply_label(task_id, label_id):
    """Apply a label to a task."""
    client = VikunjaClient()
    result = client.put(f"/tasks/{task_id}/labels", data={"label_id": label_id})
    print(json.dumps(result))


@labels.command("remove")
@click.option("--task-id", required=True, type=int, help="Task ID")
@click.option("--label-id", required=True, type=int, help="Label ID to remove")
def remove_label(task_id, label_id):
    """Remove a label from a task."""
    client = VikunjaClient()
    result = client.delete(f"/tasks/{task_id}/labels/{label_id}")
    print(json.dumps(result))
