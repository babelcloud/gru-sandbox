"""
GBox File model module

This module defines the File class, which provides an object-oriented interface
to file operations in the GBox API.
"""

from __future__ import annotations  # For type hinting GBoxClient

from typing import TYPE_CHECKING, List, Literal

# Add Pydantic imports
from pydantic import BaseModel, Field

if TYPE_CHECKING:
    from ..client import GBoxClient  # Avoid circular import for type hints


# --- Pydantic Models for API Responses ---


class FileStat(BaseModel):
    """Pydantic model representing file metadata, based on Go FileStat."""

    name: str
    path: str
    size: int  # Go uses int64, Python int handles large numbers
    mode: str
    mod_time: str = Field(alias="modTime")
    type: Literal[
        "directory", "file", "symlink", "socket", "pipe", "device"
    ]  # Use Literal for FileType
    mime: str

    # Removed validator for path to accept empty strings from API
    # @field_validator('path')
    # @classmethod
    # def validate_path(cls, v: str) -> str:
    #     if not v:
    #         raise ValueError("FileStat path cannot be empty")
    #     # Optional: Enforce leading slash if desired
    #     # if not v.startswith('/'):
    #     #     raise ValueError("FileStat path must start with '/' ")
    #     return v

    # Optional: Add validator if specific format checks are needed for modTime, mode, etc.


class FileShareResponse(BaseModel):
    """Pydantic model for the file sharing API response, based on Go FileShareResponse."""

    success: bool
    message: str
    file_list: List[FileStat] = Field(alias="fileList", default_factory=list)


# --- Updated File Class ---


class File:
    """
    Represents a file or directory in the GBox shared volume.

    Provides methods to interact with files and directories in the shared volume.
    Attributes are stored in the `attrs` Pydantic model and can be refreshed using
    `reload()`.
    """

    # Use the Pydantic model to store attributes
    attrs: FileStat
    path: str  # Keep path separate as it's the identifier

    def __init__(self, client: "GBoxClient", path: str, attrs: FileStat):
        """
        Initialize a File object using a validated Pydantic model.

        Args:
            client: The GBoxClient instance
            path: Path to the file or directory in the shared volume (primary identifier)
            attrs: A validated FileStat Pydantic model containing the file attributes.
                   Note: attrs.path might differ slightly (e.g. leading slash) from the path used for lookup.
        """
        self._client = client
        self.path = path  # The path used to fetch the file
        self.attrs = attrs  # Store the validated Pydantic model

    @property
    def name(self) -> str:
        """The name of the file or directory from metadata."""
        return self.attrs.name

    @property
    def size(self) -> int:
        """The size of the file in bytes."""
        return self.attrs.size

    @property
    def mode(self) -> str:
        """The file mode/permissions."""
        return self.attrs.mode

    @property
    def mod_time(self) -> str:
        """The last modification time of the file."""
        return self.attrs.mod_time

    @property
    def type(self) -> Literal["directory", "file", "symlink", "socket", "pipe", "device"]:
        """The type of the file."""
        return self.attrs.type

    @property
    def mime(self) -> str:
        """The MIME type of the file."""
        return self.attrs.mime

    @property
    def is_directory(self) -> bool:
        """Whether the file is a directory."""
        return self.attrs.type == "directory"

    def reload(self) -> None:
        """
        Refreshes the File's attributes by fetching the latest data from the API
        and validating it.

        Raises:
            APIError: If the API call fails.
            NotFound: If the file does not exist.
            ValidationError: If the API response format is invalid.
        """
        raw_data = self._client.file_api.head(self.path)
        # Validate the raw data using the Pydantic model
        # NotFound should be raised by file_api.head if it returns None/empty
        validated_data = FileStat.model_validate(raw_data)
        # Update the internal attrs with the validated model
        self.attrs = validated_data

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
        # Assuming file_api.get raises NotFound if needed
        return self._client.file_api.get(self.path)

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
        # Compare based on the path used to identify/fetch the file
        return self.path == other.path

    def __hash__(self) -> int:
        return hash(self.path)

    def __repr__(self) -> str:
        # Use the validated type from attrs
        return f"File(path='{self.path}', type='{self.attrs.type}')"


# Add explicit __all__ if needed later
# __all__ = ["File", "FileStat", "FileShareResponse"]
