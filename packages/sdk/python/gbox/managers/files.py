"""
GBox File Manager Module

This module provides the FileManager class for high-level file operations.
"""

from __future__ import annotations

import logging  # Add logging
from typing import TYPE_CHECKING, Any, Dict, Union

# Import Pydantic models and validation error
from pydantic import ValidationError

# Only import custom exceptions here
from ..exceptions import APIError, NotFound  # Alias to avoid conflict
from ..models.files import File, FileShareResponse, FileStat  # Import models

if TYPE_CHECKING:
    from ..client import GBoxClient
    from ..models.boxes import Box

logger = logging.getLogger(__name__)  # Setup logger


class FileManager:
    """
    Manages file resources. Accessed via `client.files`.

    Provides methods for interacting with files in the shared volume,
    validating API responses with Pydantic models.
    """

    def __init__(self, client: "GBoxClient"):
        """
        Initialize FileManager.

        Args:
            client: The GBoxClient instance
        """
        self._client = client
        self._service = client.file_api  # Direct access to the API service layer

    def get(self, path: str) -> File:
        """
        Get a File object representing a file or directory at the specified path,
        validating the metadata response.

        Args:
            path: Path to the file or directory in the shared volume

        Returns:
            A File object for the specified path

        Raises:
            NotFound: If the file or directory does not exist.
            APIError: For other API errors or invalid response format.
            ValidationError: If the API response format is invalid.
        """
        # Normalize path
        if not path.startswith("/"):
            path = "/" + path

        try:
            # Get file metadata (raw dictionary)
            raw_attrs = self._service.head(path)
            # Check if head returned None or empty dict before validation
            if not raw_attrs:
                raise NotFound(f"File or directory not found at path: {path}", status_code=404)

            # Validate the raw attributes using the Pydantic model
            validated_attrs = FileStat.model_validate(raw_attrs)
            # Pass validated data to File constructor
            return File(client=self._client, path=path, attrs=validated_attrs)
        except ValidationError as e:
            # Log the error and the raw data that caused it
            logger.error(
                f"Invalid API response format for file metadata '{path}': {e}",
                exc_info=True,
                extra={"raw_response": raw_attrs},  # Include raw_attrs in log
            )
            raise APIError(
                f"Invalid API response format for file metadata '{path}': {e}",
                explanation=str(raw_attrs),
            ) from e
        except NotFound:  # Re-raise NotFound exceptions from head or check
            raise
        except APIError:  # Re-raise other APIErrors from head
            raise

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
            # Use the service layer method, which might raise NotFound
            attrs = self._service.head(path)
            # If head succeeds and returns something, it exists
            return attrs is not None
        except NotFound:
            return False
        except APIError as e:
            # Log other API errors but return False for existence check
            logger.warning(f"API error checking existence for '{path}', assuming not found: {e}")
            return False

    def share_from_box(self, box: Union["Box", str], box_path: str) -> File:
        """
        Share a file from a Box's shared directory to the main shared volume,
        validating the API response.

        Args:
            box: Either a Box object or a box_id string
            box_path: Path to the file inside the Box's shared directory.
                     Should be a path relative to the box root, e.g., /data/file.txt
                     or an absolute path like /var/gbox/box-id/data/file.txt

        Returns:
            A File object representing the shared file in the main volume

        Raises:
            APIError: If the API call fails or the response format is invalid.
            TypeError: If the box parameter is not a Box object or string.
            ValueError: If the box_path format is invalid (optional check).
            FileNotFoundError: If the shared file cannot be found after sharing or response is invalid.
            ValidationError: If the API response format is invalid.
        """
        # Handle both Box object or box_id string
        if hasattr(box, "id"):
            box_id = box.id  # type: ignore[attr-defined]
        elif isinstance(box, str):
            box_id = box
        else:
            raise TypeError(f"Expected Box object or box_id string, got {type(box)}")

        # Optional: Add validation for box_path format if needed
        # Example: if box_path.startswith("/"): raise ValueError("Use relative path inside box")

        try:
            # Share the file via API
            raw_response = self._service.share(box_id, box_path)
            # Validate the response
            validated_response = FileShareResponse.model_validate(raw_response)

            if not validated_response.success:
                # Use message from validated response if available
                raise APIError(
                    f"File sharing failed: {validated_response.message}",
                    explanation=str(raw_response),
                )

            if not validated_response.file_list:
                raise FileNotFoundError(
                    "File sharing succeeded according to API, but no file information was returned in the validated response.",
                    explanation=str(raw_response),
                )

            # Assume the first file in the list is the primary shared file
            # The FileStat model already performed alias mapping for modTime
            shared_file_stat = validated_response.file_list[0]

            # The path in the shared volume is typically in the FileStat
            shared_file_path = shared_file_stat.path

            # --- Restore check for potentially invalid path from response ---
            if not shared_file_path or shared_file_path == "/":
                logger.warning(
                    f"FileStat path from share response for {shared_file_stat.name} is missing or invalid ('{shared_file_path}'). "
                    f"Attempting to construct path as /{box_id}/{shared_file_stat.name}"
                )
                # Construct path based on typical sharing conventions if path from API is missing/invalid
                shared_file_path = f"/{box_id}/{shared_file_stat.name}"
                if not shared_file_path.startswith("/"):
                    shared_file_path = "/" + shared_file_path  # Ensure leading slash

            # Use self.get which handles validation of the final FileStat
            # This ensures the (potentially constructed) path is valid and accessible.
            return self.get(shared_file_path)

        except ValidationError as e:
            logger.error(
                f"Invalid API response format for file share from box '{box_id}' path '{box_path}': {e}"
            )
            # raise APIError(f"Invalid API response format for file share: {e}", explanation=str(raw_response)) from e
            raise e  # Re-raise the validation error
        except APIError as e:  # Catch API errors from self._service.share or self.get
            logger.error(f"API error during file share from box '{box_id}' path '{box_path}': {e}")
            raise
        except Exception as e:
            logger.error(
                f"Unexpected error during file share from box '{box_id}' path '{box_path}': {e}"
            )
            # Re-raise unexpected errors
            raise

    def reclaim(
        self,
    ) -> Dict[str, Any]:  # Keep returning Dict for now unless a Pydantic model is defined
        """
        Reclaim unused files in the shared volume.

        Returns:
            The raw API response dictionary with information about reclaimed files.
            (Consider adding a Pydantic model for this response later if needed).

        Raises:
            APIError: If the reclamation fails
        """
        # TODO: Define a Pydantic model for the reclaim response if its structure is stable
        # and validation is desired.
        return self._service.reclaim()
