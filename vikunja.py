"""Vikunja CLI entry point."""

import json
import sys
import os

# Ensure the skill directory is on sys.path when run as a script
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

import click  # noqa: E402

from vikunja.client import VikunjaError  # noqa: E402
from vikunja.commands.config_cmd import config  # noqa: E402
from vikunja.commands.projects import projects  # noqa: E402
from vikunja.commands.tasks import tasks  # noqa: E402
from vikunja.commands.labels import labels  # noqa: E402
from vikunja.commands.comments import comments  # noqa: E402
from vikunja.commands.attachments import attachments  # noqa: E402
from vikunja.commands.relations import relations  # noqa: E402


@click.group()
def cli():
    """Vikunja REST API command-line interface.

    All output is JSON. Errors are returned as {"error": "...", "http_status": N}.
    """


cli.add_command(config)
cli.add_command(projects)
cli.add_command(tasks)
cli.add_command(labels)
cli.add_command(comments)
cli.add_command(attachments)
cli.add_command(relations)


def main():
    try:
        cli(standalone_mode=False)
    except VikunjaError as e:
        payload: dict = {"error": str(e)}
        if e.code:
            payload["code"] = e.code
        if e.http_status:
            payload["http_status"] = e.http_status
        print(json.dumps(payload))
        sys.exit(1)
    except click.exceptions.UsageError as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(2)
    except click.exceptions.Abort:
        sys.exit(1)
    except RuntimeError as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)
    except SystemExit:
        raise
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)


if __name__ == "__main__":
    main()
