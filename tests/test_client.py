"""Tests for vikunja.client module."""

import io
import pytest
from unittest.mock import MagicMock, patch

from vikunja.client import VikunjaClient, VikunjaError, _extract_filename

FAKE_CONFIG = {"url": "https://vikunja.example.com", "token": "test-token"}


def make_client():
    with patch("vikunja.client.load_config", return_value=FAKE_CONFIG):
        return VikunjaClient()


def make_response(status=200, body=None, headers=None, ok=None):
    resp = MagicMock()
    resp.status_code = status
    resp.ok = ok if ok is not None else (status < 400)
    resp.content = b"x" if body is not None else b""
    resp.json.return_value = body
    resp.headers = headers or {}
    resp.reason = "OK" if status < 400 else "Error"
    return resp


# ------------------------------------------------------------------
# Success cases
# ------------------------------------------------------------------

def test_get_returns_data():
    client = make_client()
    resp = make_response(body={"id": 1, "title": "Task"})
    with patch.object(client.session, "get", return_value=resp):
        result = client.get("/tasks/1")
    assert result["data"] == {"id": 1, "title": "Task"}
    assert "_pagination" not in result


def test_get_paginated_extracts_headers():
    client = make_client()
    resp = make_response(
        body=[{"id": 1}],
        headers={"x-pagination-total-pages": "5", "x-pagination-result-count": "100"},
    )
    with patch.object(client.session, "get", return_value=resp):
        result = client.get("/tasks", paginated=True)
    assert result["_pagination"]["total_pages"] == 5
    assert result["_pagination"]["result_count"] == 100


def test_get_paginated_defaults_when_headers_missing():
    client = make_client()
    resp = make_response(body=[])
    with patch.object(client.session, "get", return_value=resp):
        result = client.get("/tasks", paginated=True)
    assert result["_pagination"]["total_pages"] == 1
    assert result["_pagination"]["result_count"] == 0


def test_delete_204_returns_null_data():
    client = make_client()
    resp = make_response(status=204, body=None)
    resp.content = b""
    with patch.object(client.session, "delete", return_value=resp):
        result = client.delete("/tasks/1")
    assert result["data"] is None


# ------------------------------------------------------------------
# Error cases
# ------------------------------------------------------------------

def test_error_response_raises_vikunja_error():
    client = make_client()
    resp = make_response(status=404, ok=False, body={"message": "Task not found", "code": 4002})
    with patch.object(client.session, "get", return_value=resp):
        with pytest.raises(VikunjaError) as exc_info:
            client.get("/tasks/999")
    err = exc_info.value
    assert err.http_status == 404
    assert err.code == 4002
    assert "Task not found" in str(err)


def test_error_response_non_json_body():
    client = make_client()
    resp = make_response(status=500, ok=False)
    resp.json.side_effect = ValueError("no json")
    resp.reason = "Internal Server Error"
    with patch.object(client.session, "get", return_value=resp):
        with pytest.raises(VikunjaError) as exc_info:
            client.get("/tasks")
    assert exc_info.value.http_status == 500


# ------------------------------------------------------------------
# _extract_filename helper
# ------------------------------------------------------------------

@pytest.mark.parametrize("header,expected", [
    ('attachment; filename="report.pdf"', "report.pdf"),
    ("attachment; filename=report.pdf", "report.pdf"),
    ("attachment; filename*=UTF-8''my%20file.txt", "my%20file.txt"),
    ("", ""),
    ("inline", ""),
])
def test_extract_filename(header, expected):
    assert _extract_filename(header) == expected
