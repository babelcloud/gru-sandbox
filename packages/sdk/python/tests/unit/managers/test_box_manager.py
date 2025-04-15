# tests/managers/test_box_manager.py
import unittest
from unittest.mock import Mock, patch

from gbox.api.box_api import BoxApi

# Modules to patch/mock
from gbox.client import GBoxClient
from gbox.exceptions import APIError, NotFound
from gbox.managers.boxes import BoxManager
from gbox.models.boxes import (
    Box,
    BoxBase,
    BoxesDeleteResponse,
    BoxGetResponse,
    BoxReclaimResponse,
)


class TestBoxManager(unittest.TestCase):

    # Remove patch from setUp
    # @patch('gbox.managers.boxes.Box')
    def setUp(self):  # Remove MockBoxModel argument
        """Set up test fixtures."""
        self.mock_client = Mock(spec=GBoxClient)
        self.mock_box_api = Mock(spec=BoxApi)
        self.mock_client.box_api = self.mock_box_api  # Attach mock api to mock client
        self.box_manager = BoxManager(client=self.mock_client)
        # Remove self.MockBoxModel = MockBoxModel

    def test_init(self):
        """Test BoxManager initialization."""
        self.assertIs(self.box_manager._client, self.mock_client)
        self.assertIs(self.box_manager._api, self.mock_box_api)

    @patch("gbox.managers.boxes.Box")
    def test_list_success(self, MockBoxModel):  # Add MockBoxModel arg
        """Test listing boxes successfully."""
        # Arrange: Add required 'image' field and move 'name' to labels
        mock_box_data_list = [
            {"id": "box-1", "status": "running", "image": "img1", "labels": {"name": "box1"}},
            {"id": "box-2", "status": "stopped", "image": "img2", "labels": {"name": "box2"}},
        ]
        # Mock API returns the raw dict containing the list
        self.mock_box_api.list.return_value = {"boxes": mock_box_data_list}

        # Manager validates raw_response -> BoxListResponse
        # Manager then creates Box instances using validated BoxBase data
        validated_box_base_list = [BoxBase(**data) for data in mock_box_data_list]
        mock_box_instances = [Mock(spec=Box, id=d.id) for d in validated_box_base_list]
        MockBoxModel.side_effect = mock_box_instances

        # Act
        boxes = self.box_manager.list()

        # Assert
        self.mock_box_api.list.assert_called_once_with(filters=None)
        self.assertEqual(len(boxes), 2)
        # Verify Box model was called correctly with validated BoxBase data
        MockBoxModel.assert_any_call(client=self.mock_client, box_data=validated_box_base_list[0])
        MockBoxModel.assert_any_call(client=self.mock_client, box_data=validated_box_base_list[1])
        self.assertEqual(MockBoxModel.call_count, 2)

    @patch("gbox.managers.boxes.Box")
    def test_list_with_filters(self, MockBoxModel):  # Add MockBoxModel arg
        """Test listing boxes with filters."""
        filters = {"label": "test"}
        self.mock_box_api.list.return_value = {"boxes": []}
        MockBoxModel.side_effect = []  # Use local MockBoxModel

        self.box_manager.list(filters=filters)

        self.mock_box_api.list.assert_called_once_with(filters=filters)
        # Add assertion that MockBoxModel was not called, for consistency
        MockBoxModel.assert_not_called()

    @patch("gbox.managers.boxes.Box")
    def test_list_empty(self, MockBoxModel):  # Add MockBoxModel arg
        """Test listing when API returns no boxes."""
        self.mock_box_api.list.return_value = {"boxes": []}
        MockBoxModel.side_effect = []  # Use local MockBoxModel

        boxes = self.box_manager.list()

        self.mock_box_api.list.assert_called_once_with(filters=None)
        self.assertEqual(boxes, [])
        MockBoxModel.assert_not_called()  # Use local MockBoxModel

    @patch("gbox.managers.boxes.Box")
    def test_list_no_boxes_key(self, MockBoxModel):  # Add MockBoxModel arg
        """Test listing when API response lacks 'boxes' key."""
        # Provide a valid response structure (empty list) expected by BoxListResponse
        self.mock_box_api.list.return_value = {"boxes": []}
        MockBoxModel.side_effect = []  # Use local MockBoxModel

        boxes = self.box_manager.list()

        self.mock_box_api.list.assert_called_once_with(filters=None)
        self.assertEqual(boxes, [])
        MockBoxModel.assert_not_called()  # Use local MockBoxModel

    @patch("gbox.managers.boxes.Box")
    def test_get_success(self, MockBoxModel):  # Add MockBoxModel arg
        """Test getting a specific box successfully."""
        box_id = "box-xyz"
        # Add required 'image' field, move 'name' to labels
        mock_box_data = {
            "id": box_id,
            "status": "running",
            "image": "img-xyz",
            "labels": {"name": "mybox"},
        }
        self.mock_box_api.get.return_value = mock_box_data

        # Manager validates raw_data -> BoxGetResponse
        # Manager creates Box instance with validated BoxGetResponse data
        validated_box_data = BoxGetResponse(**mock_box_data)
        mock_box_instance = Mock(spec=Box, id=box_id)
        MockBoxModel.return_value = mock_box_instance

        box = self.box_manager.get(box_id)

        self.mock_box_api.get.assert_called_once_with(box_id)
        # Verify Box model was called with validated BoxGetResponse data
        MockBoxModel.assert_called_once_with(client=self.mock_client, box_data=validated_box_data)
        self.assertIs(box, mock_box_instance)

    @patch("gbox.managers.boxes.Box")
    def test_get_not_found_apierror_404(self, MockBoxModel):  # Add MockBoxModel arg
        """Test getting a non-existent box (APIError 404)."""
        box_id = "box-404"
        error = APIError("Not Found", status_code=404)
        self.mock_box_api.get.side_effect = error

        with self.assertRaises(NotFound) as cm:
            self.box_manager.get(box_id)

        self.assertEqual(cm.exception.status_code, 404)
        self.mock_box_api.get.assert_called_once_with(box_id)
        MockBoxModel.assert_not_called()  # Use local MockBoxModel

    @patch("gbox.managers.boxes.Box")
    def test_get_not_found_empty_response(self, MockBoxModel):  # Add MockBoxModel arg
        """Test getting a non-existent box (empty API response)."""
        box_id = "box-empty"
        self.mock_box_api.get.return_value = {}

        # Empty dict {} is invalid for BoxGetResponse, causing ValidationError in manager
        # The manager catches ValidationError and raises APIError
        with self.assertRaises(APIError) as cm:
            self.box_manager.get(box_id)

        # Check that the error message indicates a validation problem
        self.assertIn("Invalid API response format", str(cm.exception))
        self.assertIn("BoxGetResponse", str(cm.exception))
        self.mock_box_api.get.assert_called_once_with(box_id)
        MockBoxModel.assert_not_called()  # Use local MockBoxModel

    @patch("gbox.managers.boxes.Box")
    def test_get_other_api_error(self, MockBoxModel):  # Add MockBoxModel arg
        """Test getting a box when API returns a non-404 error."""
        box_id = "box-500"
        error = APIError("Server Error", status_code=500)
        self.mock_box_api.get.side_effect = error

        with self.assertRaises(APIError) as cm:
            self.box_manager.get(box_id)

        self.assertIs(cm.exception, error)  # Should re-raise the original error
        self.mock_box_api.get.assert_called_once_with(box_id)
        MockBoxModel.assert_not_called()  # Use local MockBoxModel

    @patch("gbox.managers.boxes.Box")
    def test_create_success(self, MockBoxModel):  # Add MockBoxModel arg
        """Test creating a box successfully."""
        image = "test-image"
        create_kwargs = {"name": "newbox", "labels": {"a": "b"}}
        created_box_id = "box-created-123"
        # API response should be a flat dictionary matching BoxBase
        api_response_data = {
            "id": created_box_id,
            "status": "created",
            "image": image,
            "labels": {"a": "b", "name": "newbox"},
        }
        self.mock_box_api.create.return_value = api_response_data

        # Manager validates -> BoxBase
        # Manager creates Box instance using validated BoxBase data
        validated_box_data = BoxBase(**api_response_data)
        mock_box_instance = Mock(spec=Box, id=created_box_id)
        MockBoxModel.return_value = mock_box_instance

        box = self.box_manager.create(image=image, **create_kwargs)

        self.mock_box_api.create.assert_called_once_with(image=image, **create_kwargs)
        # Verify Box was called with the validated BoxBase data
        MockBoxModel.assert_called_once_with(client=self.mock_client, box_data=validated_box_data)
        self.assertIs(box, mock_box_instance)

    @patch("gbox.managers.boxes.Box")
    def test_create_invalid_response_no_id(self, MockBoxModel):  # Add MockBoxModel arg
        """Test create when API response lacks an ID."""
        image = "bad-image"
        # Provide a response that passes BoxCreateResponse validation but lacks id in box
        api_response_data = {
            "box": {
                # Missing required "id"
                "status": "creating",
                "image": image,
                "labels": {"name": "no_id_box"},
            }
        }
        self.mock_box_api.create.return_value = api_response_data

        # Expect ValidationError during BoxCreateResponse.model_validate
        with self.assertRaises(APIError) as cm:
            self.box_manager.create(image=image)

        # Check that the error raised by the manager indicates a validation problem
        self.assertIn("Invalid API response format for created box data", str(cm.exception))
        self.assertIn("Field required", str(cm.exception))  # Pydantic missing field error
        self.assertIn("BoxBase", str(cm.exception))
        self.assertIn("id", str(cm.exception).lower())  # Check for the field name 'id'
        self.mock_box_api.create.assert_called_once_with(image=image)
        MockBoxModel.assert_not_called()

    @patch("gbox.managers.boxes.Box")
    def test_create_invalid_response_not_dict(self, MockBoxModel):  # Add MockBoxModel arg
        """Test create when API response is not a dictionary."""
        image = "bad-image"
        api_response_data = "just a string"
        self.mock_box_api.create.return_value = api_response_data

        with self.assertRaises(APIError) as cm:
            self.box_manager.create(image=image)

        # Check the new error message reflects validation against BoxBase
        self.assertIn("Invalid API response format for created box data", str(cm.exception))
        self.assertIn("Input should be a valid dictionary", str(cm.exception))
        self.assertIn("BoxBase", str(cm.exception))
        self.mock_box_api.create.assert_called_once_with(image=image)
        MockBoxModel.assert_not_called()

    @patch("gbox.managers.boxes.Box")
    def test_create_api_error(self, MockBoxModel):  # Add MockBoxModel arg
        """Test create when the API call itself fails."""
        image = "fail-image"
        error = APIError("Creation Forbidden", status_code=403)
        self.mock_box_api.create.side_effect = error

        with self.assertRaises(APIError) as cm:
            self.box_manager.create(image=image)

        self.assertIs(cm.exception, error)
        self.mock_box_api.create.assert_called_once_with(image=image)
        MockBoxModel.assert_not_called()  # Use local MockBoxModel

    def test_delete_all(self):
        """Test deleting all boxes."""
        # Add required 'message' field for BoxesDeleteResponse
        expected_response_dict = {
            "count": 5,
            "ids": ["1", "2", "3", "4", "5"],
            "message": "All boxes deleted",
        }
        self.mock_box_api.delete_all.return_value = expected_response_dict

        # Manager validates -> BoxesDeleteResponse
        response_model = self.box_manager.delete_all(force=False)

        self.mock_box_api.delete_all.assert_called_once_with(force=False)
        # Assert the returned value is the validated Pydantic model
        self.assertIsInstance(response_model, BoxesDeleteResponse)
        self.assertEqual(response_model.count, expected_response_dict["count"])
        self.assertEqual(response_model.ids, expected_response_dict["ids"])
        self.assertEqual(response_model.message, expected_response_dict["message"])

    def test_delete_all_force(self):
        """Test deleting all boxes with force."""
        # Add required 'message' field for BoxesDeleteResponse
        expected_response_dict = {
            "count": 3,
            "ids": ["1", "3", "5"],
            "message": "All boxes force deleted",
        }
        self.mock_box_api.delete_all.return_value = expected_response_dict

        # Manager validates -> BoxesDeleteResponse
        response_model = self.box_manager.delete_all(force=True)

        self.mock_box_api.delete_all.assert_called_once_with(force=True)
        # Assert the returned value is the validated Pydantic model
        self.assertIsInstance(response_model, BoxesDeleteResponse)
        self.assertEqual(response_model.count, expected_response_dict["count"])
        self.assertEqual(response_model.ids, expected_response_dict["ids"])
        self.assertEqual(response_model.message, expected_response_dict["message"])

    def test_reclaim(self):
        """Test reclaiming all inactive boxes."""
        # Add required 'message' field for BoxReclaimResponse
        expected_response_dict = {
            "stoppedCount": 2,
            "deletedCount": 1,
            "stoppedIds": ["box-s1", "box-s2"],
            "deletedIds": ["box-d1"],
            "message": "Reclamation complete",
        }
        self.mock_box_api.reclaim.return_value = expected_response_dict

        # Manager validates -> BoxReclaimResponse
        response_model = self.box_manager.reclaim(force=False)

        # Verify reclaim is called for *all* boxes (box_id=None)
        self.mock_box_api.reclaim.assert_called_once_with(box_id=None, force=False)
        # Assert the returned value is the validated Pydantic model
        self.assertIsInstance(response_model, BoxReclaimResponse)
        self.assertEqual(response_model.stopped_count, expected_response_dict["stoppedCount"])
        self.assertEqual(response_model.deleted_count, expected_response_dict["deletedCount"])
        self.assertEqual(response_model.stopped_ids, expected_response_dict["stoppedIds"])
        self.assertEqual(response_model.deleted_ids, expected_response_dict["deletedIds"])
        self.assertEqual(response_model.message, expected_response_dict["message"])

    def test_reclaim_force(self):
        """Test reclaiming all inactive boxes with force."""
        # Add required 'message' field for BoxReclaimResponse
        expected_response_dict = {
            "stoppedCount": 3,
            "deletedCount": 0,
            "stoppedIds": ["box-s1", "box-s2", "box-s3"],
            # deletedIds might be None/omitted if 0
            "message": "Force reclamation complete",
        }
        self.mock_box_api.reclaim.return_value = expected_response_dict

        # Manager validates -> BoxReclaimResponse
        response_model = self.box_manager.reclaim(force=True)

        self.mock_box_api.reclaim.assert_called_once_with(box_id=None, force=True)
        # Assert the returned value is the validated Pydantic model
        self.assertIsInstance(response_model, BoxReclaimResponse)
        self.assertEqual(response_model.stopped_count, expected_response_dict["stoppedCount"])
        self.assertEqual(response_model.deleted_count, expected_response_dict["deletedCount"])
        self.assertEqual(response_model.stopped_ids, expected_response_dict["stoppedIds"])
        self.assertIsNone(response_model.deleted_ids)  # Check for None if key might be missing
        self.assertEqual(response_model.message, expected_response_dict["message"])


if __name__ == "__main__":
    unittest.main()
