# gbox/models/boxes.py
from __future__ import annotations  # For type hinting GBoxClient

import io
import logging  # <-- Add logging import
import os  # <-- Add os import
import tarfile  # <-- Add tarfile import
from typing import TYPE_CHECKING, Any, Dict, List, Optional, Tuple, Union

if TYPE_CHECKING:
    from ..api.box_service import (  # Ensure BoxService is available for type hint if needed
        BoxService,
    )
    from ..client import GBoxClient  # Avoid circular import for type hints


class Box:
    """
    Represents a GBox Box instance.

    Provides methods to interact with a specific Box. Attributes are stored
    in the `attrs` dictionary and can be refreshed using `reload()`.
    """

    def __init__(self, client: "GBoxClient", id: str, attrs: Optional[Dict[str, Any]] = None):
        self._client = client
        self.id = id
        self.attrs = attrs or {}  # Store attributes like status, image, labels

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
        """The name of the Box, if set."""
        # Adjust if the name attribute key is different
        return self.attrs.get("name")

    @property
    def status(self) -> Optional[str]:
        """The current status of the Box."""
        return self.attrs.get("status")

    @property
    def labels(self) -> Dict[str, str]:
        """Labels associated with the Box."""
        return self.attrs.get("labels", {})

    def reload(self) -> None:
        """
        Refreshes the Box's attributes by fetching the latest data from the API.
        """
        # Assuming GBoxClient exposes box_service correctly
        data = self._client.box_service.get(self.id)
        # Assuming the raw get response *is* the attribute dictionary
        # Handle potential API errors here if needed, or let them propagate
        self.attrs = data

    def start(self) -> None:
        """
        Starts the Box.
        Raises:
            APIError: If the API call fails.
        """
        self._client.box_service.start(self.id)
        # Consider calling reload() or updating status locally
        # self.attrs['status'] = 'running' # Optimistic update
        # self.reload() # More reliable but slower

    def stop(self) -> None:
        """
        Stops the Box.
        Raises:
            APIError: If the API call fails.
        """
        self._client.box_service.stop(self.id)
        # Consider updating status
        # self.attrs['status'] = 'stopped'
        # self.reload()

    def delete(self, force: bool = False) -> None:
        """
        Deletes the Box.

        Args:
            force: If True, force deletion even if running (if API supports).
        Raises:
            APIError: If the API call fails.
        """
        self._client.box_service.delete(self.id, force=force)
        # After deletion, this object is effectively stale.

    def run(self, command: List[str]) -> Tuple[int, Optional[str], Optional[str]]:
        """
        Runs a command in the Box (non-interactive).

        Args:
            command: The command and its arguments as a list.

        Returns:
            A tuple containing (exit_code, stdout_str, stderr_str).
            stdout/stderr might be None if not captured or empty.
        Raises:
            APIError: If the API call fails.
        """
        logger = logging.getLogger(__name__)  # <-- Get logger instance
        response = self._client.box_service.run(self.id, command=command)
        logger.debug(
            f"Raw API response from BoxService.run for Box {self.id}: {response}"
        )  # <-- Log raw response
        exit_code = response.get("exitCode", -1)  # Provide a default?
        stdout = response.get("stdout", "")  # <-- Provide default value ""
        stderr = response.get("stderr", "")  # <-- Provide default value ""
        # Optionally update self.attrs if response['box'] is reliable
        # self.attrs.update(response.get("box", {}))
        return exit_code, stdout, stderr

    # FIXME: This is not implemented yet
    def exec_run() -> Tuple[int, Optional[str], Optional[str]]:
        return

    def reclaim(self, force: bool = False) -> Dict[str, Any]:
        """
        Reclaims resources for this specific Box.

        Args:
            force: Whether to force reclamation.

        Returns:
            The raw API response dictionary containing reclaim details.
        Raises:
            APIError: If the API call fails.
        """
        return self._client.box_service.reclaim(box_id=self.id, force=force)

    def head_archive(self, path: str) -> Dict[str, Any]:
        """
        Gets metadata about a file or directory inside the Box.

        Args:
            path: The path inside the Box.

        Returns:
            A dictionary containing file metadata (e.g., from headers returned by BoxService).
        Raises:
            APIError: If the API call fails.
            NotFound: If the path doesn't exist.
        """
        # BoxService.head_archive returns the raw response/headers.
        return self._client.box_service.head_archive(self.id, path=path)

    def get_archive(
        self, path: str, local_path: Optional[str] = None
    ) -> Tuple[Optional[io.BytesIO], Dict[str, Any]]:
        """
        Retrieves a file or directory from the Box.

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
        tar_data_bytes = self._client.box_service.get_archive(self.id, path=path)

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
                    if extracted_file:
                        with open(local_path, "wb") as f:
                            f.write(extracted_file.read())
                        # Success: return None for the stream part
                        return None, stats
                    else:
                        # Should not happen if isfile() was true
                        raise tarfile.TarError(
                            f"Failed to extract file content for {target_member.name} from archive."
                        )

            except tarfile.TarError as e:
                # Re-raise specific tar errors
                raise tarfile.TarError(
                    f"Failed to process tar archive for '{path}' when saving to '{local_path}': {e}"
                )
            except (IsADirectoryError, FileNotFoundError):
                raise  # Re-raise specific file system errors
            except Exception as e:
                raise Exception(
                    f"An error occurred during file download/extraction for '{path}' to '{local_path}': {e}"
                )

    def put_archive(self, path: str, data: Union[bytes, io.BufferedReader, str]) -> None:
        """
        Uploads data to the Box. The data can be:
        1. Raw tar archive bytes.
        2. A file-like object opened in binary mode containing a tar archive.
        3. A string representing the path to a single local file to be uploaded.

        If raw tar data (bytes/stream) is provided, it's extracted into the 'path'
        directory inside the Box.
        If a local file path (str) is provided, the file is packaged into a tar archive
        internally and then extracted into the 'path' directory inside the Box.
        The name of the file inside the Box will be the basename of the local file path.

        Args:
            path: The directory path inside the Box where the archive will be extracted
                  or the single file will be placed.
            data: The tar archive data (bytes or binary stream) OR the path
                  to a local file (str).
        Raises:
            APIError: If the API call fails.
            TypeError: If data is not bytes, a binary reader, or a string path.
            FileNotFoundError: If data is a string path and the local file does not exist.
            IsADirectoryError: If data is a string path and it points to a directory.
            tarfile.TarError: If creating the internal tar archive fails (when data is a path).
            ValueError: If path is not a valid directory path inside the box (e.g., empty or root).
            Exception: For other potential errors during file I/O or API interaction.
        """
        archive_bytes: bytes

        if isinstance(data, str):
            # Data is a local file path
            local_path = data
            if not os.path.exists(local_path):
                raise FileNotFoundError(f"Local file not found: {local_path}")
            if not os.path.isfile(local_path):
                raise IsADirectoryError(
                    f"Path specified is a directory, not a file: {local_path}. Use a future upload_directory method or provide a tar archive."
                )

            if not path or path == "/":
                # The underlying extract_archive needs a proper directory.
                # Disallow uploading directly to root for clarity, user should specify e.g., /tmp
                raise ValueError(
                    "Target path in Box cannot be empty or root ('/'). Specify a target directory like '/uploads/'."
                )

            remote_filename = os.path.basename(local_path)
            tar_stream = io.BytesIO()
            try:
                with tarfile.open(fileobj=tar_stream, mode="w") as tar:
                    tar.add(local_path, arcname=remote_filename)
                tar_stream.seek(0)
                archive_bytes = tar_stream.getvalue()
            except tarfile.TarError as e:
                raise tarfile.TarError(
                    f"Failed to create internal tar archive for {local_path}: {e}"
                )
            except Exception as e:
                raise Exception(
                    f"An error occurred during internal tar creation for {local_path}: {e}"
                )

        elif isinstance(data, io.BufferedReader):
            archive_bytes = data.read()
        elif isinstance(data, bytes):
            archive_bytes = data
        else:
            raise TypeError(
                "data must be bytes, a file-like object opened in binary mode, or a string path to a local file"
            )

        # Call the underlying API method to extract the archive (either provided or generated)
        self._client.box_service.extract_archive(self.id, path=path, archive_data=archive_bytes)

    def copy(self, source: str, target: str) -> None:
        """
        Copies files or directories between the local filesystem and the Box.

        Uses a URI-like format to specify source and target locations:
        - 'box:/path/in/box' indicates a path inside the Box.
        - '/local/path' or 'relative/local/path' indicates a path on the local filesystem.

        Examples:
        - Upload local file 'myfile.txt' to '/uploads/' inside the Box:
            `box.copy('myfile.txt', 'box:/uploads/')`
        - Download '/data/report.txt' from the Box to the local current directory:
            `box.copy('box:/data/report.txt', 'report.txt')`
        - Download '/data/report.txt' from the Box to local '/tmp/report.txt':
            `box.copy('box:/data/report.txt', '/tmp/report.txt')`

        Args:
            source: The source path (e.g., 'local_file.txt', 'box:/remote/path').
            target: The target path (e.g., 'box:/remote/dir/', 'local_file.txt').

        Raises:
            ValueError: If the source and target combination is invalid (e.g., both local or both remote).
            FileNotFoundError: If a local source file specified for upload does not exist.
            IsADirectoryError: If a local source path for upload points to a directory.
            APIError: If an API call during upload or download fails.
            NotFound: If a remote source path for download does not exist.
            tarfile.TarError: If downloading to a local path encounters tar processing issues.
            Exception: For other potential errors during file I/O or API interaction.
        """
        source_is_box = source.startswith("box:")
        target_is_box = target.startswith("box:")

        if source_is_box and not target_is_box:
            # Download from Box to Local
            box_path = source[len("box:") :]
            local_path = target
            if not box_path:
                raise ValueError("Source path in Box cannot be empty.")
            # Use get_archive with local_path to download directly
            self.get_archive(path=box_path, local_path=local_path)
            # get_archive handles file existence checks (NotFound) and extraction

        elif not source_is_box and target_is_box:
            # Upload from Local to Box
            local_path = source
            box_path = target[len("box:") :]
            if not box_path or box_path == "/":
                raise ValueError(
                    "Target path in Box must be a directory (e.g., 'box:/uploads/'), not root or empty."
                )
            # Use put_archive with local file path
            self.put_archive(path=box_path, data=local_path)
            # put_archive handles local file checks (FileNotFound, IsADirectory)

        elif source_is_box and target_is_box:
            raise ValueError(
                "Cannot copy directly between two Box paths. Use box.run() or download then upload."
            )
        else:  # Not source_is_box and not target_is_box
            raise ValueError("Cannot copy between two local paths using this method.")

    def __eq__(self, other):
        return isinstance(other, Box) and self.id == other.id

    def __hash__(self):
        return hash(self.id)

    def __repr__(self):
        short_id_val = self.short_id
        return f"<Box: {short_id_val}>"
