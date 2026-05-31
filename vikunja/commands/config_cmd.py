"""config set / config show commands."""

import json
import click
from vikunja.config import CONFIG_PATH, load_config, save_config


@click.group()
def config():
    """Manage connection settings (~/.vikunja.json)."""


@config.command("set")
@click.option("--url", required=True, help="Vikunja instance URL (e.g. https://vikunja.example.com)")
@click.option("--token", required=True, help="API token (create one under Settings → API Tokens in Vikunja)")
def config_set(url, token):
    """Write URL and API token to the config file."""
    save_config(url, token)
    print(json.dumps({"data": {"message": f"Config saved to {CONFIG_PATH}", "url": url.rstrip("/")}}))


@config.command("show")
def config_show():
    """Display the current configuration (token is masked)."""
    try:
        cfg = load_config()
    except RuntimeError as e:
        print(json.dumps({"error": str(e)}))
        raise SystemExit(1)
    cfg["token"] = cfg["token"][:6] + "..." + cfg["token"][-4:] if len(cfg["token"]) > 10 else "***"
    print(json.dumps({"data": cfg}))
