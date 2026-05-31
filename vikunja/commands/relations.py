"""Task relation commands."""

import json
import click
from vikunja.client import VikunjaClient

RELATION_KINDS = [
    "subtask", "parenttask", "related",
    "duplicateof", "duplicates",
    "blocking", "blocked",
    "precedes", "follows",
    "copiedfrom", "copiedto",
]


@click.group()
def relations():
    """Manage task relations (subtasks, blocking, related, etc.)."""


@relations.command("list")
@click.argument("task_id", type=int)
def list_relations(task_id):
    """List all relations for a task, grouped by relation kind."""
    client = VikunjaClient()
    result = client.get(f"/tasks/{task_id}/relations")
    print(json.dumps(result))


@relations.command("create")
@click.option("--task-id", required=True, type=int, help="Source task ID")
@click.option("--other-id", required=True, type=int, help="Related task ID")
@click.option(
    "--kind",
    required=True,
    type=click.Choice(RELATION_KINDS),
    help="Relation kind",
)
def create_relation(task_id, other_id, kind):
    """Create a relation between two tasks."""
    client = VikunjaClient()
    result = client.put(
        f"/tasks/{task_id}/relations",
        data={"relation_kind": kind, "other_task_id": other_id},
    )
    print(json.dumps(result))


@relations.command("delete")
@click.option("--task-id", required=True, type=int, help="Source task ID")
@click.option("--other-id", required=True, type=int, help="Related task ID")
@click.option(
    "--kind",
    required=True,
    type=click.Choice(RELATION_KINDS),
    help="Relation kind",
)
def delete_relation(task_id, other_id, kind):
    """Remove a relation between two tasks."""
    client = VikunjaClient()
    result = client.delete(f"/tasks/{task_id}/relations/{kind}/{other_id}")
    print(json.dumps(result))
