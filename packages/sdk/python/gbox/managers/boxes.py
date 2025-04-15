# gbox/managers/boxes.py
from __future__ import annotations

from typing import TYPE_CHECKING, Any, Dict, List, Optional, Union

# Import Pydantic models and validation error
from pydantic import ValidationError

from ..exceptions import APIError, NotFound  # Import your specific exceptions
from ..models.boxes import BoxBase  # Needed for type hints if used directly
from ..models.boxes import (  # Add other response models if manager methods return their validated data
    Box,
    BoxesDeleteResponse,
    BoxGetResponse,
    BoxListResponse,
    BoxReclaimResponse,
)

if TYPE_CHECKING:
    from ..client import GBoxClient


class BoxManager:
    """
    Manages Box resources. Accessed via `client.boxes`.

    Provides methods for listing, getting, creating, and performing bulk
    actions on Boxes, validating responses with Pydantic models.
    """

    def __init__(self, client: "GBoxClient"):
        self._client = client
        self._api = client.box_api  # Direct access to the API layer

    def list(self, filters: Optional[Dict[str, Union[str, List[str]]]] = None) -> List[Box]:
        """
        Lists Boxes, optionally filtering them, and validates the response.

        Args:
            filters: A dictionary for filtering (e.g., {'label': 'key=value', 'id': [...]}).

        Returns:
            A list of Box objects matching the criteria.

        Raises:
            APIError: If the API call fails.
            ValidationError: If the API response format is invalid.
        """
        raw_response = self._api.list(filters=filters)
        try:
            # Validate the entire list response structure
            validated_response = BoxListResponse.model_validate(raw_response)
            # Create Box objects from the validated data
            return [
                Box(client=self._client, box_data=box_data) for box_data in validated_response.boxes
            ]
        except ValidationError as e:
            # Handle or re-raise validation errors appropriately
            raise APIError(
                f"Invalid API response format for list boxes: {e}", explanation=str(raw_response)
            ) from e

    def get(self, box_id: str) -> Box:
        """
        Retrieves a specific Box by its ID and validates the response.

        Args:
            box_id: The ID of the Box.

        Returns:
            A Box object representing the requested Box.
        Raises:
            NotFound: If the Box with the given ID does not exist.
            APIError: For other API-related errors.
            ValidationError: If the API response format is invalid.
        """
        try:
            raw_data = self._api.get(box_id)
            # Validate the response using the Pydantic model
            validated_data = BoxGetResponse.model_validate(raw_data)
            # Pass the validated Pydantic model to the Box constructor
            return Box(client=self._client, box_data=validated_data)
        except APIError as e:
            # Re-raise specific errors if the service/client layer doesn't already
            if e.status_code == 404:
                raise NotFound(f"Box with ID '{box_id}' not found", status_code=404) from e
            raise  # Re-raise other APIErrors
        except ValidationError as e:
            raise APIError(
                f"Invalid API response format for get box '{box_id}': {e}",
                explanation=str(raw_data),
            ) from e

    def create(self, image: str, **kwargs: Any) -> Box:
        """
        Creates a new Box and validates the response.
        Assumes the API returns the Box data directly, not nested within a 'box' key.

        Args:
            image (str): The image identifier to use.
            **kwargs: Additional keyword arguments passed directly to the
                      `BoxApi.create` method (e.g., cmd, args, env, labels,
                      working_dir, volumes, image_pull_secret, name).

        Returns:
            A Box object representing the newly created Box.
        Raises:
            APIError: If the creation fails or the response is invalid.
            ValidationError: If the API response format (expected Box data) is invalid.
        """
        raw_response = self._api.create(image=image, **kwargs)  # API returns raw box data dict
        try:
            # Validate the raw response directly as BoxBase data
            # (Assuming API returns the Box data directly, not nested under 'box')
            validated_box_data = BoxBase.model_validate(raw_response)
            # Pass the validated BoxBase model to the Box constructor
            return Box(client=self._client, box_data=validated_box_data)
        except ValidationError as e:
            # If validation fails, it means the raw_response wasn't valid Box data
            raise APIError(
                f"Invalid API response format for created box data: {e}",
                explanation=str(raw_response),
            ) from e
        except Exception as e:  # Catch other potential errors
            raise APIError(
                f"Unexpected error processing create response: {e}", explanation=str(raw_response)
            ) from e

    def delete_all(self, force: bool = False) -> BoxesDeleteResponse:
        """
        Deletes all Boxes managed by the service and validates the response.

        Args:
            force: If True, attempt to force delete (if API supports).

        Returns:
            A validated Pydantic model (`BoxesDeleteResponse`) indicating the result.
        Raises:
            APIError: If the bulk deletion fails.
            ValidationError: If the API response format is invalid.
        """
        raw_response = self._api.delete_all(force=force)
        try:
            validated_response = BoxesDeleteResponse.model_validate(raw_response)
            return validated_response
        except ValidationError as e:
            raise APIError(
                f"Invalid API response format for delete all boxes: {e}",
                explanation=str(raw_response),
            ) from e

    def reclaim(self, force: bool = False) -> BoxReclaimResponse:
        """
        Reclaims resources for all inactive Boxes and validates the response.

        Args:
            force: If True, force reclamation.

        Returns:
            A validated Pydantic model (`BoxReclaimResponse`) indicating the result.
        Raises:
            APIError: If the reclamation fails.
            ValidationError: If the API response format is invalid.
        """
        raw_response = self._api.reclaim(box_id=None, force=force)
        try:
            validated_response = BoxReclaimResponse.model_validate(raw_response)
            return validated_response
        except ValidationError as e:
            raise APIError(
                f"Invalid API response format for reclaim all boxes: {e}",
                explanation=str(raw_response),
            ) from e
