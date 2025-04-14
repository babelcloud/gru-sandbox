"""
GBox File API Service Module

This module provides direct mappings to the File-related API endpoints in the GBox API server.
All methods return the raw server response data or headers.
"""

import json
from typing import Any, Dict, Optional

from ..config import GBoxConfig
from .client import Client


class FileService:
    """
    File service, provides low-level operations for file resources in the shared directory.

    Each method directly corresponds to an API endpoint in the server.
    """

    def __init__(self, client: Client, config: GBoxConfig):
        """
        Initialize File service

        Args:
            client: HTTP client
            config: GBox configuration
        """
        self.client = client
        self.logger = config.logger

    def head(self, path: str) -> Optional[Dict[str, Any]]:
        """
        Get metadata of a file or directory in the shared volume.
        Maps to HEAD /api/v1/files/{path} endpoint.
        The metadata is returned in the 'X-Gbox-File-Stat' header.

        Args:
            path: Path to the file or directory within the shared volume (should start with /).

        Returns:
            A dictionary containing file statistics if found, otherwise None.
            Example:
            {
                "name": "file.txt",
                "size": 1024,
                "mode": "-rw-r--r--",
                "modTime": "2023-10-27T10:00:00Z",
                "type": "file",
                "mime": "text/plain"
            }
        """
        # Ensure path starts with /
        if not path.startswith("/"):
            path = "/" + path

        # The actual path might be prefixed, e.g., /<box_id>/path/to/file
        # The API expects the path relative to the share root.
        response = self.client.head(f"/api/v1/files{path}")

        # Metadata is expected in the 'X-Gbox-File-Stat' header as a JSON string
        stat_header = response.get("x-gbox-file-stat")
        if stat_header:
            try:
                return json.loads(stat_header)
            except json.JSONDecodeError as e:
                self.logger.error(f"Failed to decode file stat header: {e}")
                return None
        return None

    def get(self, path: str) -> bytes:
        """
        Get the content of a file from the shared volume.
        Maps to GET /api/v1/files/{path} endpoint.

        Args:
            path: Path to the file within the shared volume (should start with /).

        Returns:
            Raw binary content of the file.
        """
        # Ensure path starts with /
        if not path.startswith("/"):
            path = "/" + path

        response = self.client.get(
            f"/api/v1/files{path}",
            headers={"Accept": "*/*"},  # Accept any content type
            raw_response=True,
        )
        return response

    def reclaim(self) -> Dict[str, Any]:
        """
        Reclaim unused files in the shared volume (older than retention period).
        Maps to POST /api/v1/files?operation=reclaim endpoint.

        Returns:
            Raw API response with the following structure:
            {
                "reclaimed_files": ["/path/to/old_file.txt", "/path/to/empty_dir"],
                "errors": ["error message if any"]
            }
        """
        params = {"operation": "reclaim"}
        response = self.client.post("/api/v1/files", params=params)
        return response

    def share(self, box_id: str, path: str) -> Dict[str, Any]:
        """
        Share a file or directory from a specific Box's shared directory
        (/var/gbox/share inside the box) to the main shared volume.
        Maps to POST /api/v1/files?operation=share endpoint.

        Args:
            box_id: ID of the source Box.
            path: Path to the file or directory inside the Box's shared directory
                  (e.g., /my_output.txt). This path is relative to the box's
                  internal /var/gbox/share directory.

        Returns:
            Raw API response indicating success and listing shared files:
            {
                "success": true,
                "message": "File shared successfully",
                "fileList": [
                    {
                        "name": "shared_file.txt",
                        "size": 123, ... other stat fields ...
                    }
                ]
            }
        """
        params = {"operation": "share"}
        # Ensure path starts with /
        if not path.startswith("/"):
            path = "/" + path
        data = {"boxId": box_id, "path": path}
        response = self.client.post("/api/v1/files", params=params, data=data)
        return response
