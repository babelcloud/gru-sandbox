"""
GBox File model module

This module defines the File class, which provides an object-oriented interface
to file operations in the GBox API.
"""

from __future__ import annotations  # For type hinting GBoxClient

import io
import os
from typing import TYPE_CHECKING, Any, Dict, Optional, Union

if TYPE_CHECKING:
    from ..api.file_service import FileService
    from ..client import GBoxClient  # Avoid circular import for type hints


class File:
    """
    Represents a file or directory in the GBox shared volume.

    Provides methods to interact with files and directories in the shared volume.
    Attributes are stored in the `attrs` dictionary and can be refreshed using
    `reload()`.
    """

    def __init__(self, client: "GBoxClient", path: str, attrs: Optional[Dict[str, Any]] = None):
        """
        Initialize a File object.

        Args:
            client: The GBoxClient instance
            path: Path to the file or directory in the shared volume
            attrs: Optional dictionary of file attributes
        """
        self._client = client
        self.path = path
        self.attrs = attrs or {}

    @property
    def name(self) -> str:
        """The name of the file or directory."""
        return self.attrs.get("name", os.path.basename(self.path))

    @property
    def size(self) -> Optional[int]:
        """The size of the file in bytes."""
        return self.attrs.get("size")

    @property
    def mode(self) -> Optional[str]:
        """The file mode/permissions."""
        return self.attrs.get("mode")

    @property
    def mod_time(self) -> Optional[str]:
        """The last modification time of the file."""
        return self.attrs.get("modTime")

    @property
    def type(self) -> Optional[str]:
        """The type of the file (file or directory)."""
        return self.attrs.get("type")

    @property
    def mime(self) -> Optional[str]:
        """The MIME type of the file."""
        return self.attrs.get("mime")

    @property
    def is_directory(self) -> bool:
        """Whether the file is a directory."""
        return self.attrs.get("type") == "directory"

    def reload(self) -> None:
        """
        Refreshes the File's attributes by fetching the latest data from the API.

        Raises:
            APIError: If the API call fails.
            NotFound: If the file does not exist.
        """
        data = self._client.file_service.head(self.path)
        if data:
            self.attrs = data

    def read(self) -> bytes:
        """
        Read the content of the file.

        Returns:
            The raw content of the file as bytes.

        Raises:
            APIError: If the API call fails.
            NotFound: If the file does not exist.
            IsADirectoryError: If the path points to a directory.
        """
        if self.is_directory:
            raise IsADirectoryError(f"Cannot read directory content: {self.path}")

        return self._client.file_service.get(self.path)

    def read_text(self, encoding: str = "utf-8") -> str:
        """
        Read the content of the file as text.

        Args:
            encoding: The encoding to use for decoding the bytes to text.

        Returns:
            The content of the file as a string.

        Raises:
            APIError: If the API call fails.
            NotFound: If the file does not exist.
            IsADirectoryError: If the path points to a directory.
            UnicodeDecodeError: If the file content cannot be decoded with the given encoding.
        """
        return self.read().decode(encoding)

    def __eq__(self, other: object) -> bool:
        if not isinstance(other, File):
            return False
        return self.path == other.path

    def __hash__(self) -> int:
        return hash(self.path)

    def __repr__(self) -> str:
        return f"File(path='{self.path}', type='{self.type}')"
