"""Tests for vikunja.config module."""

import json
import pytest
from pathlib import Path

import vikunja.config as config_module
from vikunja.config import load_config, save_config


@pytest.fixture(autouse=True)
def isolated_config(tmp_path, monkeypatch):
    """Redirect CONFIG_PATH to a temp file for each test."""
    fake_path = tmp_path / ".vikunja.json"
    monkeypatch.setattr(config_module, "CONFIG_PATH", fake_path)
    return fake_path


def test_save_and_load_roundtrip():
    save_config("https://vikunja.example.com", "my-secret-token")
    cfg = load_config()
    assert cfg["url"] == "https://vikunja.example.com"
    assert cfg["token"] == "my-secret-token"


def test_save_strips_trailing_slash():
    save_config("https://vikunja.example.com/", "tok")
    cfg = load_config()
    assert cfg["url"] == "https://vikunja.example.com"


def test_load_missing_file_raises():
    with pytest.raises(RuntimeError, match="Config not found"):
        load_config()


def test_load_missing_url_raises(isolated_config):
    isolated_config.write_text(json.dumps({"token": "tok"}))
    with pytest.raises(RuntimeError, match="missing required field 'url'"):
        load_config()


def test_load_missing_token_raises(isolated_config):
    isolated_config.write_text(json.dumps({"url": "https://x.com"}))
    with pytest.raises(RuntimeError, match="missing required field 'token'"):
        load_config()


def test_config_file_permissions(isolated_config):
    """Config file should be created with restricted permissions (Unix only)."""
    import sys
    if sys.platform == "win32":
        pytest.skip("File permission check not applicable on Windows")
    save_config("https://vikunja.example.com", "tok")
    mode = isolated_config.stat().st_mode & 0o777
    assert mode == 0o600, f"Expected 0o600, got 0o{mode:o}"
