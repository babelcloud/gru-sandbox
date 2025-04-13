# gbox/managers/boxes.py
from __future__ import annotations

from typing import TYPE_CHECKING, Any, Dict, List, Optional, Union

from ..exceptions import APIError, NotFound  # Import your specific exceptions
from ..models.boxes import Box

if TYPE_CHECKING:
    from ..client import GBoxClient


class BoxManager:
    """
    Manages Box resources. Accessed via `client.boxes`.

    Provides methods for listing, getting, creating, and performing bulk
    actions on Boxes.
    """

    def __init__(self, client: "GBoxClient"):
        self._client = client
        self._service = client.box_service  # Direct access to the API service layer

    def list(self, filters: Optional[Dict[str, Union[str, List[str]]]] = None) -> List[Box]:
        """
        Lists Boxes, optionally filtering them.

        Args:
            filters: A dictionary for filtering (e.g., {'label': 'key=value', 'id': [...]}).

        Returns:
            A list of Box objects matching the criteria.
        """
        raw_response = self._service.list(filters=filters)
        box_list_data = raw_response.get("boxes", [])
        return [
            Box(client=self._client, id=box_data.get("id"), attrs=box_data)
            for box_data in box_list_data
            if box_data.get("id")  # Ensure ID exists
        ]

    def get(self, box_id: str) -> Box:
        """
        Retrieves a specific Box by its ID.

        Args:
            box_id: The ID of the Box.

        Returns:
            A Box object representing the requested Box.
        Raises:
            NotFound: If the Box with the given ID does not exist.
            APIError: For other API-related errors.
        """
        try:
            raw_data = self._service.get(box_id)
            # Check if the API returned meaningful data (e.g., an ID)
            if not raw_data or not raw_data.get("id"):
                raise NotFound(
                    f"Box with ID '{box_id}' not found (empty response)", status_code=404
                )
            return Box(client=self._client, id=box_id, attrs=raw_data)
        except APIError as e:
            # Re-raise specific errors if the service/client layer doesn't already
            if e.status_code == 404:
                raise NotFound(f"Box with ID '{box_id}' not found", status_code=404) from e
            raise  # Re-raise other APIErrors

    def create(self, image: str, **kwargs: Any) -> Box:
        """
        Creates a new Box.

        Args:
            image (str): The image identifier to use.
            **kwargs: Additional keyword arguments passed directly to the
                      `BoxService.create` method (e.g., cmd, args, env, labels,
                      working_dir, volumes, image_pull_secret, name).

        Returns:
            A Box object representing the newly created Box.
        Raises:
            APIError: If the creation fails or the response is invalid.
        """
        # We don't need to explicitly handle 'name' here unless the API layer
        # requires it named differently than how users might pass it.
        # We assume kwargs are passed through correctly by `_service.create`.

        raw_response = self._service.create(image=image, **kwargs)  # Returns the box data directly

        # The raw_response itself should be the box data dictionary
        if not isinstance(raw_response, dict) or not raw_response.get("id"):
            raise APIError(
                "API did not return valid box data after creation", explanation=str(raw_response)
            )

        # Use the ID and attributes from the response to create the Box object
        return Box(client=self._client, id=raw_response["id"], attrs=raw_response)

    def delete_all(self, force: bool = False) -> Dict[str, Any]:
        """
        Deletes all Boxes managed by the service.

        Args:
            force: If True, attempt to force delete (if API supports).

        Returns:
            The raw API response dictionary indicating the result (e.g., count, ids).
        Raises:
            APIError: If the bulk deletion fails.
        """
        return self._service.delete_all(force=force)

    def reclaim(self, force: bool = False) -> Dict[str, Any]:
        """
        Reclaims resources for all inactive Boxes.

        Args:
            force: If True, force reclamation.

        Returns:
            The raw API response dictionary indicating the result (e.g., stopped/deleted counts/ids).
        Raises:
            APIError: If the reclamation fails.
        """
        # Call reclaim without a specific box_id for the bulk operation
        return self._service.reclaim(box_id=None, force=force)
