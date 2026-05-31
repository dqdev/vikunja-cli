"""Task attachment commands (list + download)."""

import json
import click
from vikunja.client import VikunjaClient


@click.group()
def attachments():
    """List and download task attachments."""


@attachments.command("list")
@click.option("--task-id", required=True, type=int, help="Task ID")
def list_attachments(task_id):
    """List all attachments for a task (returns metadata, not file content)."""
    client = VikunjaClient()
    result = client.get(f"/tasks/{task_id}/attachments")
    print(json.dumps(result))


@attachments.command("download")
@click.option("--task-id", required=True, type=int, help="Task ID")
@click.option("--attachment-id", required=True, type=int, help="Attachment ID")
@click.option(
    "--output-dir",
    default=".",
    show_default=True,
    help="Directory to save the downloaded file",
)
def download_attachment(task_id, attachment_id, output_dir):
    """Download an attachment to a local directory.

    The filename is taken from the server's Content-Disposition header.
    The response includes the absolute path the file was saved to.
    """
    client = VikunjaClient()
    result = client.download(f"/tasks/{task_id}/attachments/{attachment_id}", output_dir)
    print(json.dumps(result))
