"""Task comment commands."""

import json
import click
from vikunja.client import VikunjaClient


@click.group()
def comments():
    """Manage task comments."""


@comments.command("list")
@click.option("--task-id", required=True, type=int, help="Task ID")
def list_comments(task_id):
    """List all comments on a task."""
    client = VikunjaClient()
    result = client.get(f"/tasks/{task_id}/comments")
    print(json.dumps(result))


@comments.command("create")
@click.option("--task-id", required=True, type=int, help="Task ID")
@click.option("--text", required=True, help="Comment text (Markdown supported)")
def create_comment(task_id, text):
    """Add a comment to a task."""
    client = VikunjaClient()
    result = client.put(f"/tasks/{task_id}/comments", data={"comment": text})
    print(json.dumps(result))


@comments.command("delete")
@click.option("--task-id", required=True, type=int, help="Task ID")
@click.option("--comment-id", required=True, type=int, help="Comment ID to delete")
def delete_comment(task_id, comment_id):
    """Delete a comment from a task."""
    client = VikunjaClient()
    result = client.delete(f"/tasks/{task_id}/comments/{comment_id}")
    print(json.dumps(result))
