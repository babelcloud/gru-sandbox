"""
GBox File Manager Module

This module provides the FileManager class for high-level file operations.
"""

from __future__ import annotations

import os
from typing import TYPE_CHECKING, Any, Dict, List, Optional, Union

from ..exceptions import APIError, NotFound
from ..models.files import File

if TYPE_CHECKING:
    from ..client import GBoxClient
    from ..models.boxes import Box


class FileManager:
    """
    Manages file resources. Accessed via `client.files`.

    Provides methods for interacting with files in the shared volume.
    """

    def __init__(self, client: "GBoxClient"):
        """
        Initialize FileManager.

        Args:
            client: The GBoxClient instance
        """
        self._client = client
        self._service = client.file_service  # Direct access to the API service layer

    def get(self, path: str) -> File:
        """
        Get a File object representing a file or directory at the specified path.

        Args:
            path: Path to the file or directory in the shared volume

        Returns:
            A File object for the specified path

        Raises:
            NotFound: If the file or directory does not exist
            APIError: For other API errors
        """
        # Normalize path
        if not path.startswith("/"):
            path = "/" + path

        # Get file metadata
        attrs = self._service.head(path)
        if not attrs:
            raise NotFound(f"File not found at path: {path}", status_code=404)

        return File(client=self._client, path=path, attrs=attrs)

    def exists(self, path: str) -> bool:
        """
        Check if a file or directory exists.

        Args:
            path: Path to check in the shared volume

        Returns:
            True if the path exists, False otherwise
        """
        # Normalize path
        if not path.startswith("/"):
            path = "/" + path

        try:
            attrs = self._service.head(path)
            return attrs is not None
        except NotFound:
            return False
        except APIError:
            return False

    def share_from_box(self, box: Union["Box", str], box_path: str) -> File:
        """
        Share a file from a Box's shared directory to the main shared volume.

        Args:
            box: Either a Box object or a box_id string
            box_path: Path to the file inside the Box's shared directory.
                     Should be a path starting with /var/gbox/ for valid box paths.

        Returns:
            A File object representing the shared file in the main volume

        Raises:
            APIError: If the API call fails
            TypeError: If the box parameter is not a Box object or string
            ValueError: If the box_path does not start with /var/gbox/
            FileNotFoundError: If the shared file cannot be found after sharing
        """
        # Handle both Box object or box_id string
        if hasattr(box, "id"):
            # It's a Box object
            box_id = box.id
        elif isinstance(box, str):
            # It's a box_id string
            box_id = box
        else:
            raise TypeError(f"Expected Box object or box_id string, got {type(box)}")

        # Verify the path starts with /var/gbox/
        if not box_path.startswith("/var/gbox/"):
            raise ValueError(f"Box path must start with /var/gbox/, got: {box_path}")

        # Share the file
        share_response = self._service.share(box_id, box_path)

        # Extract file information from the response
        # The response structure may include a fileList with information about shared files
        if not share_response or "fileList" not in share_response or not share_response["fileList"]:
            raise FileNotFoundError(f"File sharing succeeded but no file information was returned")

        # For simplicity, assume the first (or only) file in the fileList is the one we want
        file_info = share_response["fileList"][0]

        # Determine the path in the main volume where the file was shared
        # This depends on how the API structures the response, may need adjustment
        shared_file_path = file_info.get("path", None)
        if not shared_file_path:
            # If path isn't directly provided, try to construct from box_id and name
            name = file_info.get("name")
            if not name:
                raise FileNotFoundError("Cannot determine shared file path from response")
            # Construct path based on typical sharing conventions
            shared_file_path = f"/{box_id}/{name}"

        # Create and return a File object for the shared file
        return self.get(shared_file_path)

    def reclaim(self) -> Dict[str, Any]:
        """
        Reclaim unused files in the shared volume.

        Returns:
            The raw API response with information about reclaimed files

        Raises:
            APIError: If the reclamation fails
        """
        return self._service.reclaim()
