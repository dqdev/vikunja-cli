"""HTTP client wrapping the Vikunja REST API."""

import os
import re
import requests
from vikunja.config import load_config


class VikunjaError(Exception):
    """Raised when the Vikunja API returns an error."""

    def __init__(self, message: str, code: int = 0, http_status: int = 0):
        super().__init__(message)
        self.code = code
        self.http_status = http_status


class VikunjaClient:
    def __init__(self):
        config = load_config()
        self.base_url = config["url"] + "/api/v1"
        self.session = requests.Session()
        self.session.headers.update({
            "Authorization": f"Bearer {config['token']}",
            "Content-Type": "application/json",
        })

    # ------------------------------------------------------------------
    # Internal helpers
    # ------------------------------------------------------------------

    def _raise_for_error(self, resp: requests.Response) -> None:
        try:
            err = resp.json()
        except Exception:
            err = {}
        message = err.get("message") or resp.reason or "Unknown error"
        raise VikunjaError(message, code=err.get("code", 0), http_status=resp.status_code)

    def _handle(self, resp: requests.Response, paginated: bool = False) -> dict:
        if not resp.ok:
            self._raise_for_error(resp)
        if resp.status_code == 204 or not resp.content:
            result: dict = {"data": None}
        else:
            result = {"data": resp.json()}
        if paginated:
            result["_pagination"] = {
                "total_pages": int(resp.headers.get("x-pagination-total-pages", 1)),
                "result_count": int(resp.headers.get("x-pagination-result-count", 0)),
            }
        return result

    # ------------------------------------------------------------------
    # Public API methods
    # ------------------------------------------------------------------

    def get(self, path: str, params: dict | None = None, paginated: bool = False) -> dict:
        resp = self.session.get(f"{self.base_url}{path}", params=params)
        return self._handle(resp, paginated=paginated)

    def put(self, path: str, data: dict | None = None) -> dict:
        resp = self.session.put(f"{self.base_url}{path}", json=data or {})
        return self._handle(resp)

    def post(self, path: str, data: dict | None = None) -> dict:
        resp = self.session.post(f"{self.base_url}{path}", json=data or {})
        return self._handle(resp)

    def delete(self, path: str) -> dict:
        resp = self.session.delete(f"{self.base_url}{path}")
        return self._handle(resp)

    def download(self, path: str, output_dir: str) -> dict:
        """Stream a binary response to disk. Returns {"data": {"saved_to": "...", "filename": "..."}}."""
        resp = self.session.get(f"{self.base_url}{path}", stream=True)
        if not resp.ok:
            self._raise_for_error(resp)

        filename = _extract_filename(resp.headers.get("content-disposition", ""))
        if not filename:
            filename = path.split("/")[-1]

        os.makedirs(output_dir, exist_ok=True)
        output_path = os.path.join(output_dir, filename)

        with open(output_path, "wb") as f:
            for chunk in resp.iter_content(chunk_size=8192):
                f.write(chunk)

        return {"data": {"saved_to": os.path.abspath(output_path), "filename": filename}}


def _extract_filename(content_disposition: str) -> str:
    """Parse filename from a Content-Disposition header value."""
    if not content_disposition:
        return ""
    match = re.search(r'filename\*?=(?:UTF-8\'\'|["\']?)([^;"\'\r\n]+)', content_disposition, re.IGNORECASE)
    if match:
        return match.group(1).strip().strip('"\'')
    return ""
