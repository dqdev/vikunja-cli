"""Project CRUD commands."""

import json
import click
from vikunja.client import VikunjaClient


@click.group()
def projects():
    """Manage Vikunja projects."""


@projects.command("list")
@click.option("--page", default=1, show_default=True, type=int, help="Page number")
@click.option("--per-page", default=50, show_default=True, type=int, help="Items per page (max 100)")
@click.option("--search", default=None, help="Search projects by title")
@click.option("--archived", is_flag=True, default=False, help="Include archived projects")
def list_projects(page, per_page, search, archived):
    """List all projects the current user has access to."""
    client = VikunjaClient()
    params = {"page": page, "per_page": per_page}
    if search:
        params["s"] = search
    if archived:
        params["is_archived"] = "true"
    result = client.get("/projects", params=params, paginated=True)
    print(json.dumps(result))


@projects.command("get")
@click.argument("project_id", type=int)
def get_project(project_id):
    """Get a single project by ID."""
    client = VikunjaClient()
    result = client.get(f"/projects/{project_id}")
    print(json.dumps(result))


@projects.command("create")
@click.option("--title", required=True, help="Project title")
@click.option("--description", default=None, help="Project description")
@click.option("--parent-id", default=None, type=int, help="Parent project ID (for sub-projects)")
def create_project(title, description, parent_id):
    """Create a new project."""
    client = VikunjaClient()
    body: dict = {"title": title}
    if description is not None:
        body["description"] = description
    if parent_id is not None:
        body["parent_project_id"] = parent_id
    result = client.put("/projects", data=body)
    print(json.dumps(result))


@projects.command("update")
@click.argument("project_id", type=int)
@click.option("--title", default=None, help="New title")
@click.option("--description", default=None, help="New description")
def update_project(project_id, title, description):
    """Update a project's title or description."""
    client = VikunjaClient()
    body: dict = {}
    if title is not None:
        body["title"] = title
    if description is not None:
        body["description"] = description
    if not body:
        print(json.dumps({"error": "Provide at least one of --title or --description"}))
        raise SystemExit(1)
    result = client.post(f"/projects/{project_id}", data=body)
    print(json.dumps(result))


@projects.command("delete")
@click.argument("project_id", type=int)
def delete_project(project_id):
    """Delete a project and all its tasks."""
    client = VikunjaClient()
    result = client.delete(f"/projects/{project_id}")
    print(json.dumps(result))
