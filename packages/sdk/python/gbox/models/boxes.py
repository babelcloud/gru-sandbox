# gbox/models/boxes.py
from __future__ import annotations  # For type hinting GBoxClient

import io
import logging  # <-- Add logging import
import os  # <-- Add os import
import tarfile  # <-- Add tarfile import
from typing import TYPE_CHECKING, Any, Dict, List, Optional, Tuple, Union

# Add Pydantic imports
from pydantic import BaseModel, Field, field_validator

if TYPE_CHECKING:
    from ..client import GBoxClient  # Avoid circular import for type hints


# --- Pydantic Models for API Responses ---


class BoxBase(BaseModel):
    """Base Pydantic model for a Box."""

    id: str
    status: str
    image: str
    labels: Optional[Dict[str, str]] = Field(
        default_factory=dict
    )  # Use alias for ExtraLabels if needed from Go struct

    @field_validator("labels", mode="before")
    @classmethod
    def empty_labels_as_dict(cls, v):
        # Ensure labels is always a dict, even if API returns null/None
        return v or {}


class BoxListResponse(BaseModel):
    """Pydantic model for the /api/v1/boxes GET response."""

    boxes: List[BoxBase]


class BoxCreateResponse(BaseModel):
    """Pydantic model for the /api/v1/boxes POST response."""

    box: BoxBase
    message: Optional[str] = None


class BoxGetResponse(BoxBase):
    """Pydantic model for the /api/v1/boxes/{id} GET response."""

    # Inherits all fields from BoxBase
    pass


class BoxRunResponse(BaseModel):
    """Pydantic model for the /api/v1/boxes/{id}/run POST response."""

    box: BoxBase
    exit_code: Optional[int] = Field(alias="exitCode", default=None)
    stdout: Optional[str] = None
    stderr: Optional[str] = None


class BoxDeleteResponse(BaseModel):
    """Pydantic model for the /api/v1/boxes/{id} DELETE response."""

    message: str


class BoxesDeleteResponse(BaseModel):
    """Pydantic model for the /api/v1/boxes DELETE response."""

    count: int
    message: str
    ids: Optional[List[str]] = None


class BoxStartResponse(BaseModel):
    """Pydantic model for the /api/v1/boxes/{id}/start POST response."""

    success: bool
    message: str


class BoxStopResponse(BaseModel):
    """Pydantic model for the /api/v1/boxes/{id}/stop POST response."""

    success: bool
    message: str


class BoxReclaimResponse(BaseModel):
    """Pydantic model for the /api/v1/boxes/reclaim or /api/v1/boxes/{id}/reclaim POST response."""

    message: str
    stopped_ids: Optional[List[str]] = Field(alias="stoppedIds", default=None)
    deleted_ids: Optional[List[str]] = Field(alias="deletedIds", default=None)
    stopped_count: int = Field(alias="stoppedCount")
    deleted_count: int = Field(alias="deletedCount")


# --- Existing Box Class ---


class Box:
    """
    Represents a GBox Box instance.

    Provides methods to interact with a specific Box. Attributes are stored
    in the `attrs` dictionary (validated Pydantic model) and can be refreshed using `reload()`.
    """

    # Use the Pydantic model to store attributes
    attrs: BoxBase

    def __init__(self, client: "GBoxClient", box_data: BoxBase):
        """Initialize Box object using validated Pydantic model."""
        self._client = client
        # self.id is now derived from attrs
        self.attrs = box_data

    @property
    def id(self) -> str:
        """The unique ID of the Box."""
        return self.attrs.id

    @property
    def short_id(self) -> str:
        """A short identifier for the Box (e.g., 'box-xxxxxxxx')."""
        parts = self.id.split("-", 2)  # Split at most 2 times
        if len(parts) >= 2:
            return f"{parts[0]}-{parts[1]}"
        else:
            return self.id  # Return full ID if no hyphen or only one part

    @property
    def name(self) -> Optional[str]:
        """The name of the Box, if set (potentially from labels)."""
        # Adjust if the name is stored differently, e.g., in labels
        return self.attrs.labels.get("name")  # Example: assuming name is a label

    @property
    def status(self) -> str:
        """The current status of the Box."""
        return self.attrs.status

    @property
    def labels(self) -> Dict[str, str]:
        """Labels associated with the Box."""
        # Ensure it always returns a dict, even if None initially
        return self.attrs.labels or {}

    def reload(self) -> None:
        """
        Refreshes the Box's attributes by fetching the latest data from the API
        and validating it.
        """
        raw_data = self._client.box_api.get(self.id)
        # Validate the raw data using the Pydantic model
        validated_data = BoxGetResponse.model_validate(raw_data)
        # Update the internal attrs with the validated model
        self.attrs = validated_data  # Directly assign the validated model

    def start(self) -> None:
        """
        Starts the Box and attempts to refresh its status.
        Raises:
            APIError: If the API call fails.
            ValidationError: If the API response for start or subsequent reload is invalid.
        """
        response_data = self._client.box_api.start(self.id)
        # Optionally validate the start response itself
        BoxStartResponse.model_validate(response_data)
        # Refresh data after action
        self.reload()

    def stop(self) -> None:
        """
        Stops the Box and attempts to refresh its status.
        Raises:
            APIError: If the API call fails.
            ValidationError: If the API response for stop or subsequent reload is invalid.
        """
        response_data = self._client.box_api.stop(self.id)
        # Optionally validate the stop response itself
        BoxStopResponse.model_validate(response_data)
        # Refresh data after action
        self.reload()

    def delete(self, force: bool = False) -> None:
        """
        Deletes the Box.

        Args:
            force: If True, force deletion even if running (if API supports).
        Raises:
            APIError: If the API call fails.
            ValidationError: If the API response for delete is invalid.
        """
        response_data = self._client.box_api.delete(self.id, force=force)
        # Optionally validate the delete response
        BoxDeleteResponse.model_validate(response_data)
        # After deletion, this object is effectively stale. Mark it?
        # self.attrs.status = 'deleted' # Or similar

    def run(self, command: List[str]) -> Tuple[int, Optional[str], Optional[str]]:
        """
        Runs a command in the Box (non-interactive) and validates the response.

        Args:
            command: The command and its arguments as a list.

        Returns:
            A tuple containing (exit_code, stdout_str, stderr_str).
            stdout/stderr might be None if not captured or empty.
        Raises:
            APIError: If the API call fails.
            ValidationError: If the API response is invalid.
        """
        logger = logging.getLogger(__name__)
        raw_response = self._client.box_api.run(self.id, command=command)
        logger.debug(f"Raw API response from BoxApi.run for Box {self.id}: {raw_response}")
        # Validate the response
        validated_response = BoxRunResponse.model_validate(raw_response)

        # Update box attributes if the response contains updated box info
        if validated_response.box:
            self.attrs = validated_response.box  # Update with validated nested BoxBase

        exit_code = validated_response.exit_code if validated_response.exit_code is not None else -1
        stdout = validated_response.stdout
        stderr = validated_response.stderr

        return exit_code, stdout, stderr

    def reclaim(self, force: bool = False) -> BoxReclaimResponse:
        """
        Reclaims resources for this Box and validates the response.

        Args:
            force: If True, force reclamation.

        Returns:
            A validated Pydantic model (`BoxReclaimResponse`) with the result.
        Raises:
            APIError: If the API call fails.
            ValidationError: If the API response is invalid.
        """
        raw_response = self._client.box_api.reclaim(box_id=self.id, force=force)
        validated_response = BoxReclaimResponse.model_validate(raw_response)
        # Potentially reload the box status after reclaim
        # self.reload()
        return validated_response

    def head_archive(self, path: str) -> Dict[str, Any]:
        """
        Gets metadata about a file or directory inside the Box.
        (No specific Pydantic model defined here as it returns headers)

        Args:
            path: The path inside the Box.

        Returns:
            A dictionary containing file metadata (e.g., from headers returned by BoxApi).
        Raises:
            APIError: If the API call fails.
            NotFound: If the path doesn't exist.
        """
        # BoxApi.head_archive returns the raw response/headers.
        # Pydantic validation might be less useful here unless a specific structure is expected.
        return self._client.box_api.head_archive(self.id, path=path)

    def get_archive(
        self, path: str, local_path: Optional[str] = None
    ) -> Tuple[Optional[io.BytesIO], Dict[str, Any]]:
        """
        Retrieves a file or directory from the Box.
        (No specific Pydantic model defined here for the binary data)

        If local_path is None (default), returns the raw tar archive data as a stream.
        If local_path is provided, attempts to download a single file specified by 'path'
        directly to the 'local_path', extracting it from the archive internally.

        Args:
            path: The path to the file or directory inside the Box.
            local_path: Optional. If provided, the local path where the downloaded file
                      will be saved. Parent directories will be created.
                      If set, the method attempts to extract a single file.

        Returns:
            A tuple containing:
            - An io.BytesIO stream with the raw tar data if local_path is None, otherwise None.
            - A dictionary containing metadata about the archive (from head_archive).

        Raises:
            APIError: If the API call fails.
            NotFound: If the remote path doesn't exist.
            tarfile.TarError: If local_path is provided and the archive is invalid,
                              empty, contains multiple items, or the expected file is not found.
            IsADirectoryError: If local_path is provided but the remote path points to a directory.
            FileNotFoundError: If local_path is provided but the remote path does not point to a file.
            Exception: For other potential errors during file I/O or API interaction.
        """
        stats = self.head_archive(path)  # Get metadata first. Raises NotFound/APIError on failure.
        tar_data_bytes = self._client.box_api.get_archive(self.id, path=path)

        if local_path is None:
            # Original behavior: return the raw tar stream
            tar_stream = io.BytesIO(tar_data_bytes)
            return tar_stream, stats
        else:
            # New behavior: extract single file to local_path
            local_dir = os.path.dirname(local_path)
            if local_dir:
                os.makedirs(local_dir, exist_ok=True)

            tar_stream = io.BytesIO(tar_data_bytes)
            try:
                with tarfile.open(fileobj=tar_stream, mode="r:*") as tar:
                    members = tar.getmembers()
                    if not members:
                        raise tarfile.TarError(f"Received empty tar archive for {path}")

                    # Expecting a single file matching the basename of the requested path
                    target_filename = os.path.basename(path)
                    target_member = None

                    # Basic check: if archive has more than one member, it's likely a directory
                    # or something unexpected for a single file download.
                    if len(members) > 1:
                        # Check if it's just a directory entry + file (common case)
                        if not (
                            len(members) == 2
                            and members[0].isdir()
                            and members[1].isfile()
                            and members[1].name == target_filename
                        ):
                            member_names = [m.name for m in members]
                            raise tarfile.TarError(
                                f"Expected a single file archive for '{path}', but found multiple members: {member_names}. Use get_archive without local_path to handle complex archives."
                            )

                    # Find the file member
                    for member in members:
                        if member.isfile() and (
                            member.name == target_filename
                            or member.name.endswith(f"/{target_filename}")
                        ):
                            target_member = member
                            break
                        # Handle case where tar might contain just the file without parent dir entry
                        if member.isfile() and len(members) == 1 and member.name == target_filename:
                            target_member = member
                            break

                    if target_member is None:
                        # Could be that path points to a directory server-side
                        if members[0].isdir():
                            raise IsADirectoryError(
                                f"Remote path '{path}' points to a directory. Use get_archive without local_path and extract manually, or use a future download_directory method."
                            )
                        else:
                            member_names = [m.name for m in members]
                            raise FileNotFoundError(
                                f"File '{target_filename}' not found within the downloaded archive for '{path}'. Archive contains: {member_names}"
                            )

                    # Extract the found file member
                    extracted_file = tar.extractfile(target_member)
                    if extracted_file:  # Ensure file object was returned
                        with open(local_path, "wb") as f_out:
                            while True:
                                chunk = extracted_file.read(io.DEFAULT_BUFFER_SIZE)
                                if not chunk:
                                    break
                                f_out.write(chunk)
                        # Return None for the stream, as requested
                        return None, stats
                    else:
                        raise tarfile.TarError(
                            f"Could not extract file '{target_member.name}' from archive for '{path}'."
                        )
            except tarfile.TarError as e:
                logging.error(f"TarError processing archive for {path}: {e}")
                raise  # Re-raise the specific error
            except Exception as e:
                logging.error(
                    f"Unexpected error extracting file to {local_path} from archive for {path}: {e}"
                )
                raise  # Re-raise other exceptions

        # Fallback in case extraction logic doesn't return explicitly
        # This line should ideally not be reached if logic is correct
        return None, stats  # Or raise an internal error

    def put_archive(self, path: str, data: Union[bytes, io.BufferedReader, str]) -> None:
        """
        Uploads data as a tar archive to the Box.

        Args:
            path: The destination path inside the Box.
            data: The data to upload. Can be:
                  - bytes: Raw tar archive data.
                  - io.BufferedReader: An open file handle (in binary mode) to the tar archive.
                  - str: Path to a local file or directory to be archived and uploaded.
        Raises:
            APIError: If the API call fails.
            FileNotFoundError: If 'data' is a string path and the file/directory doesn't exist.
            TypeError: If 'data' is of an unsupported type.
            tarfile.TarError: If creating the tar archive from a local path fails.
            Exception: For other potential errors during file I/O or API interaction.
        """
        archive_data: bytes

        if isinstance(data, bytes):
            archive_data = data
        elif isinstance(data, io.BufferedReader):
            archive_data = data.read()
        elif isinstance(data, str):
            # Create a tar archive in memory from the local path
            if not os.path.exists(data):
                raise FileNotFoundError(f"Local path '{data}' not found.")
            # Check if the local path is a directory, which is not supported for direct upload
            if os.path.isdir(data):
                raise IsADirectoryError(
                    f"Uploading a directory directly is not supported. Path: '{data}'"
                )

            tar_stream = io.BytesIO()
            mode = "w:gz"  # Use compression
            try:
                with tarfile.open(fileobj=tar_stream, mode=mode) as tar:
                    # Add the file or directory to the archive.
                    # arcname preserves the original basename in the archive.
                    tar.add(data, arcname=os.path.basename(data))
            except Exception as e:
                logging.error(f"Failed to create tar archive from path '{data}': {e}")
                raise tarfile.TarError(f"Failed to create tar archive: {e}") from e

            archive_data = tar_stream.getvalue()
        else:
            raise TypeError(
                f"Unsupported data type for put_archive: {type(data)}. Expected bytes, file handle, or str path."
            )

        # Ensure the path passed to the API doesn't have the 'box:' prefix
        api_path = path
        if api_path.lower().startswith("box:"):
            api_path = api_path[4:]

        # API call expects bytes
        self._client.box_api.extract_archive(self.id, path=api_path, archive_data=archive_data)
        # No specific response validation needed unless extract_archive returns structured data

    def copy(self, source: str, target: str) -> None:
        """
        Copies a file or directory from the Box to a local path, or vice-versa.

        Determines direction based on which path exists locally.
        - If `source` exists locally -> Upload (put_archive)
        - If `target` parent dir exists locally -> Download (get_archive)

        Args:
            source: Source path (local or inside the Box).
            target: Target path (local or inside the Box).

        Raises:
            ValueError: If both or neither paths seem local, or direction is ambiguous.
            FileNotFoundError: If a local source path doesn't exist.
            APIError, NotFound, TarError, etc.: Propagated from get/put_archive.
        """
        source_is_local = os.path.exists(source)
        # Check if target *directory* exists for download case
        target_dir = os.path.dirname(target)
        target_dir_exists_local = (
            os.path.exists(target_dir) if target_dir else True
        )  # Handle target in root

        if source_is_local and not target_dir_exists_local:
            # Upload: source is local file/dir, target is remote path
            # Target path for put_archive is the destination *inside* the box
            self.put_archive(path=target, data=source)
        elif not source_is_local and target_dir_exists_local:
            # Download: source is remote path, target is local file/dir path
            # Use get_archive with local_path set to the target
            self.get_archive(path=source, local_path=target)
        elif source_is_local and target_dir_exists_local:
            raise ValueError(
                f"Ambiguous copy: Both source ('{source}') and target directory ('{target_dir}') exist locally. Cannot determine copy direction."
            )
        else:  # Neither source exists locally nor target directory exists locally
            raise ValueError(
                f"Cannot determine copy direction: Source path '{source}' does not exist locally, and target directory '{target_dir}' does not exist locally."
            )

    def __eq__(self, other):
        if not isinstance(other, Box):
            return False
        return self.id == other.id

    def __hash__(self):
        return hash(self.id)

    def __repr__(self):
        return f"<Box: {self.short_id} ({self.status})>"
