"""
GBox Box API Service Module

This module provides direct mappings to the Box-related API endpoints in the GBox API server.
All methods return the raw server response data without any transformation.
"""

from typing import Any, Dict, List, Optional, Union

from ..config import GBoxConfig
from .client import Client


class BoxService:
    """
    Box service, provides low-level operations for Box resources.

    Each method directly corresponds to an API endpoint in the server and returns
    the raw server response data.
    """

    def __init__(self, client: Client, config: GBoxConfig):
        """
        Initialize Box service

        Args:
            client: HTTP client
            config: GBox configuration
        """
        self.client = client
        self.logger = config.logger

    def list(self, filters: Optional[Dict[str, Union[str, List[str]]]] = None) -> Dict[str, Any]:
        """
        Get all Boxes with optional filtering.
        Maps to GET /api/v1/boxes endpoint.

        Args:
            filters: Optional filter conditions with the following options:
                - id: Filter by box ID(s)
                - label: Filter by label(s), can use format 'key=value' or just 'key'
                - ancestor: Filter by ancestor image

        Returns:
            Raw API response with the following structure:
            {
                "boxes": [
                    {
                        "id": "box-id",
                        "status": "status",
                        "image": "image-name",
                        "labels": {"key": "value", ...}
                    },
                    ...
                ]
            }
        """
        params = {}
        if filters:
            filter_params = []
            for k, v in filters.items():
                if isinstance(v, list):
                    # Handle list of values for the same filter key
                    for item in v:
                        filter_params.append(f"{k}={item}")
                else:
                    # Handle single value
                    filter_params.append(f"{k}={v}")

            if filter_params:
                params["filter"] = filter_params

        response = self.client.get("/api/v1/boxes", params=params)
        return response

    def get(self, box_id: str) -> Dict[str, Any]:
        """
        Get detailed information of a specific Box.
        Maps to GET /api/v1/boxes/{id} endpoint.

        Args:
            box_id: ID of the Box

        Returns:
            Raw API response with Box details including:
            {
                "id": "box-id",
                "status": "status",
                "image": "image-name",
                "labels": {"key": "value", ...}
            }
        """
        response = self.client.get(f"/api/v1/boxes/{box_id}")
        return response

    def create(
        self,
        image: str,
        image_pull_secret: Optional[str] = None,
        env: Optional[Dict[str, str]] = None,
        cmd: Optional[str] = None,
        args: Optional[List[str]] = None,
        working_dir: Optional[str] = None,
        labels: Optional[Dict[str, str]] = None,
        volumes: Optional[List[Dict[str, Any]]] = None,
    ) -> Dict[str, Any]:
        """
        Create a new Box.
        Maps to POST /api/v1/boxes endpoint.

        Args:
            image: Container image to use
            image_pull_secret: Secret for pulling images (For docker: base64 encoded auth string, for k8s: secret name)
            env: Environment variables
            cmd: Command
            args: Command arguments
            working_dir: Working directory
            labels: Additional labels
            volumes: List of volumes to mount, each containing:
                  - source: Host path
                  - target: Container path
                  - readOnly: Whether read-only
                  - propagation: Mount propagation mode (private, rprivate, shared, rshared, slave, rslave)

        Returns:
            Raw API response with the following structure:
            {
                "box": {
                    "id": "box-id",
                    "status": "status",
                    "image": "image-name",
                    "labels": {"key": "value", ...}
                },
                "message": "Box created successfully"
            }
        """
        data = {"image": image}

        # Add optional parameters
        if image_pull_secret:
            data["imagePullSecret"] = image_pull_secret
        if env:
            data["env"] = env
        if cmd:
            data["cmd"] = cmd
        if args:
            data["args"] = args
        if working_dir:
            data["workingDir"] = working_dir
        if labels:
            data["labels"] = labels
        if volumes:
            data["volumes"] = volumes

        response = self.client.post("/api/v1/boxes", data=data)
        return response

    def delete(self, box_id: str, force: bool = False) -> Dict[str, Any]:
        """
        Delete a Box.
        Maps to DELETE /api/v1/boxes/{id} endpoint.

        Args:
            box_id: ID of the Box
            force: Whether to force deletion

        Returns:
            Raw API response with the following structure:
            {
                "message": "Box deleted successfully"
            }
        """
        data = {"force": force} if force else {}
        response = self.client.delete(f"/api/v1/boxes/{box_id}", data=data)
        return response

    def delete_all(self, force: bool = False) -> Dict[str, Any]:
        """
        Delete all Boxes.
        Maps to DELETE /api/v1/boxes endpoint.

        Args:
            force: Whether to force deletion

        Returns:
            Raw API response with the following structure:
            {
                "count": 3,
                "message": "3 boxes deleted",
                "ids": ["box-1", "box-2", "box-3"]
            }
        """
        data = {"force": force} if force else {}
        response = self.client.delete("/api/v1/boxes", data=data)
        return response

    def start(self, box_id: str) -> Dict[str, Any]:
        """
        Start a Box.
        Maps to POST /api/v1/boxes/{id}/start endpoint.

        Args:
            box_id: ID of the Box

        Returns:
            Raw API response with the following structure:
            {
                "success": true,
                "message": "Box started successfully"
            }
        """
        response = self.client.post(f"/api/v1/boxes/{box_id}/start")
        return response

    def stop(self, box_id: str) -> Dict[str, Any]:
        """
        Stop a Box.
        Maps to POST /api/v1/boxes/{id}/stop endpoint.

        Args:
            box_id: ID of the Box

        Returns:
            Raw API response with the following structure:
            {
                "success": true,
                "message": "Box stopped successfully"
            }
        """
        response = self.client.post(f"/api/v1/boxes/{box_id}/stop")
        return response

    def run(self, box_id: str, command: List[str]) -> Dict[str, Any]:
        """
        Run a command in the Box (non-interactive).
        Maps to POST /api/v1/boxes/{id}/run endpoint.

        Args:
            box_id: ID of the Box
            command: Command to run (first item is the command, rest are arguments)

        Returns:
            Raw API response with the following structure:
            {
                "box": {
                    "id": "box-id",
                    "status": "status",
                    "image": "image-name",
                    "labels": {"key": "value", ...}
                },
                "exitCode": 0,
                "stdout": "command output",
                "stderr": "error output"
            }
        """
        data = {"cmd": command[:1], "args": command[1:] if len(command) > 1 else []}
        response = self.client.post(f"/api/v1/boxes/{box_id}/run", data=data)
        return response

    # FIXME: This is not implemented yet
    def exec() -> Any:
        return

    def reclaim(self, box_id: Optional[str] = None, force: bool = False) -> Dict[str, Any]:
        """
        Reclaim Box resources.
        Maps to POST /api/v1/boxes/reclaim or POST /api/v1/boxes/{id}/reclaim endpoints.

        Args:
            box_id: ID of the Box to reclaim, if None, reclaim all inactive Boxes
            force: Whether to force reclamation

        Returns:
            Raw API response with the following structure:
            {
                "message": "Resources reclaimed",
                "stoppedIds": ["box-1", "box-2"],
                "deletedIds": ["box-3"],
                "stoppedCount": 2,
                "deletedCount": 1
            }
        """
        data = {"force": force}

        if box_id:
            path = f"/api/v1/boxes/{box_id}/reclaim"
            response = self.client.post(path, data=data)
        else:
            path = "/api/v1/boxes/reclaim"
            response = self.client.post(path, data=data)

        return response

    def get_archive(self, box_id: str, path: str) -> bytes:
        """
        Get files from box as tar archive.
        Maps to GET /api/v1/boxes/{id}/archive endpoint.

        Args:
            box_id: ID of the Box
            path: Path to get files from within the Box

        Returns:
            Raw binary tar archive data
        """
        params = {"path": path}
        response = self.client.get(
            f"/api/v1/boxes/{box_id}/archive",
            params=params,
            headers={"Accept": "application/x-tar"},
            raw_response=True,
        )
        return response

    def extract_archive(self, box_id: str, path: str, archive_data: bytes) -> Dict[str, Any]:
        """
        Extract tar archive to box.
        Maps to PUT /api/v1/boxes/{id}/archive endpoint.

        Args:
            box_id: ID of the Box
            path: Path to extract files to within the Box
            archive_data: Binary tar archive data

        Returns:
            Raw API response indicating success or failure
        """
        params = {"path": path}
        response = self.client.put(
            f"/api/v1/boxes/{box_id}/archive",
            params=params,
            data=archive_data,
            headers={"Content-Type": "application/x-tar"},
        )
        return response

    def head_archive(self, box_id: str, path: str) -> Dict[str, Any]:
        """
        Get metadata about files in box.
        Maps to HEAD /api/v1/boxes/{id}/archive endpoint.

        Args:
            box_id: ID of the Box
            path: Path to get metadata from within the Box

        Returns:
            Raw API response with file metadata
        """
        params = {"path": path}
        response = self.client.head(f"/api/v1/boxes/{box_id}/archive", params=params)
        return response
