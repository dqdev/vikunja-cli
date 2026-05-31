"""Config management for ~/.vikunja.json."""

import json
from pathlib import Path

CONFIG_PATH = Path.home() / ".vikunja.json"


def load_config() -> dict:
    """Load and validate config from disk. Raises RuntimeError if missing or invalid."""
    if not CONFIG_PATH.exists():
        raise RuntimeError(
            f"Config not found at {CONFIG_PATH}. "
            "Run: python vikunja.py config set --url URL --token TOKEN"
        )
    with CONFIG_PATH.open() as f:
        data = json.load(f)
    for key in ("url", "token"):
        if not data.get(key):
            raise RuntimeError(
                f"Config missing required field '{key}'. "
                "Run: python vikunja.py config set --url URL --token TOKEN"
            )
    data["url"] = data["url"].rstrip("/")
    return data


def save_config(url: str, token: str) -> None:
    """Write config to disk."""
    config = {"url": url.rstrip("/"), "token": token}
    with CONFIG_PATH.open("w") as f:
        json.dump(config, f, indent=2)
    CONFIG_PATH.chmod(0o600)
