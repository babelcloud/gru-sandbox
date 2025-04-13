# tests/managers/test_box_manager.py
import unittest
from unittest.mock import MagicMock, Mock, patch

from gbox.api.box_service import BoxService

# Modules to patch/mock
from gbox.client import GBoxClient
from gbox.exceptions import APIError, NotFound
from gbox.managers.boxes import BoxManager
from gbox.models.boxes import Box


class TestBoxManager(unittest.TestCase):

    # Remove patch from setUp
    # @patch('gbox.managers.boxes.Box')
    def setUp(self):  # Remove MockBoxModel argument
        """Set up test fixtures."""
        self.mock_client = Mock(spec=GBoxClient)
        self.mock_box_service = Mock(spec=BoxService)
        self.mock_client.box_service = self.mock_box_service  # Attach mock service to mock client
        self.box_manager = BoxManager(client=self.mock_client)
        # Remove self.MockBoxModel = MockBoxModel

    def test_init(self):
        """Test BoxManager initialization."""
        self.assertIs(self.box_manager._client, self.mock_client)
        self.assertIs(self.box_manager._service, self.mock_box_service)

    @patch("gbox.managers.boxes.Box")
    def test_list_success(self, MockBoxModel):  # Add MockBoxModel arg
        """Test listing boxes successfully."""
        # Arrange
        mock_box_data_list = [
            {"id": "box-1", "name": "box1", "status": "running"},
            {"id": "box-2", "name": "box2", "status": "stopped"},
        ]
        self.mock_box_service.list.return_value = {"boxes": mock_box_data_list}
        # Configure the mock Box model to return mock instances
        mock_box_instances = [Mock(spec=Box, id=d["id"]) for d in mock_box_data_list]
        MockBoxModel.side_effect = mock_box_instances  # Use local MockBoxModel

        # Act
        boxes = self.box_manager.list()

        # Assert
        self.mock_box_service.list.assert_called_once_with(filters=None)
        self.assertEqual(len(boxes), 2)
        # Verify Box model was called correctly for each box
        MockBoxModel.assert_any_call(  # Use local MockBoxModel
            client=self.mock_client, id="box-1", attrs=mock_box_data_list[0]
        )
        MockBoxModel.assert_any_call(  # Use local MockBoxModel
            client=self.mock_client, id="box-2", attrs=mock_box_data_list[1]
        )
        self.assertEqual(MockBoxModel.call_count, 2)  # Use local MockBoxModel

    @patch("gbox.managers.boxes.Box")
    def test_list_with_filters(self, MockBoxModel):  # Add MockBoxModel arg
        """Test listing boxes with filters."""
        filters = {"label": "test"}
        self.mock_box_service.list.return_value = {"boxes": []}
        MockBoxModel.side_effect = []  # Use local MockBoxModel

        self.box_manager.list(filters=filters)

        self.mock_box_service.list.assert_called_once_with(filters=filters)
        # Add assertion that MockBoxModel was not called, for consistency
        MockBoxModel.assert_not_called()

    @patch("gbox.managers.boxes.Box")
    def test_list_empty(self, MockBoxModel):  # Add MockBoxModel arg
        """Test listing when API returns no boxes."""
        self.mock_box_service.list.return_value = {"boxes": []}
        MockBoxModel.side_effect = []  # Use local MockBoxModel

        boxes = self.box_manager.list()

        self.mock_box_service.list.assert_called_once_with(filters=None)
        self.assertEqual(boxes, [])
        MockBoxModel.assert_not_called()  # Use local MockBoxModel

    @patch("gbox.managers.boxes.Box")
    def test_list_no_boxes_key(self, MockBoxModel):  # Add MockBoxModel arg
        """Test listing when API response lacks 'boxes' key."""
        self.mock_box_service.list.return_value = {}
        MockBoxModel.side_effect = []  # Use local MockBoxModel

        boxes = self.box_manager.list()

        self.mock_box_service.list.assert_called_once_with(filters=None)
        self.assertEqual(boxes, [])
        MockBoxModel.assert_not_called()  # Use local MockBoxModel

    @patch("gbox.managers.boxes.Box")
    def test_get_success(self, MockBoxModel):  # Add MockBoxModel arg
        """Test getting a specific box successfully."""
        box_id = "box-xyz"
        mock_box_data = {"id": box_id, "name": "mybox", "status": "running"}
        self.mock_box_service.get.return_value = mock_box_data
        mock_box_instance = Mock(spec=Box, id=box_id)
        MockBoxModel.return_value = mock_box_instance  # Use local MockBoxModel

        box = self.box_manager.get(box_id)

        self.mock_box_service.get.assert_called_once_with(box_id)
        MockBoxModel.assert_called_once_with(  # Use local MockBoxModel
            client=self.mock_client, id=box_id, attrs=mock_box_data
        )

    @patch("gbox.managers.boxes.Box")
    def test_get_not_found_apierror_404(self, MockBoxModel):  # Add MockBoxModel arg
        """Test getting a non-existent box (APIError 404)."""
        box_id = "box-404"
        error = APIError("Not Found", status_code=404)
        self.mock_box_service.get.side_effect = error

        with self.assertRaises(NotFound) as cm:
            self.box_manager.get(box_id)

        self.assertEqual(cm.exception.status_code, 404)
        self.mock_box_service.get.assert_called_once_with(box_id)
        MockBoxModel.assert_not_called()  # Use local MockBoxModel

    @patch("gbox.managers.boxes.Box")
    def test_get_not_found_empty_response(self, MockBoxModel):  # Add MockBoxModel arg
        """Test getting a non-existent box (empty API response)."""
        box_id = "box-empty"
        self.mock_box_service.get.return_value = {}

        with self.assertRaises(NotFound) as cm:
            self.box_manager.get(box_id)

        self.assertEqual(cm.exception.status_code, 404)
        self.mock_box_service.get.assert_called_once_with(box_id)
        MockBoxModel.assert_not_called()  # Use local MockBoxModel

    @patch("gbox.managers.boxes.Box")
    def test_get_other_api_error(self, MockBoxModel):  # Add MockBoxModel arg
        """Test getting a box when API returns a non-404 error."""
        box_id = "box-500"
        error = APIError("Server Error", status_code=500)
        self.mock_box_service.get.side_effect = error

        with self.assertRaises(APIError) as cm:
            self.box_manager.get(box_id)

        self.assertIs(cm.exception, error)  # Should re-raise the original error
        self.mock_box_service.get.assert_called_once_with(box_id)
        MockBoxModel.assert_not_called()  # Use local MockBoxModel

    @patch("gbox.managers.boxes.Box")
    def test_create_success(self, MockBoxModel):  # Add MockBoxModel arg
        """Test creating a box successfully."""
        image = "test-image"
        create_kwargs = {"name": "newbox", "labels": {"a": "b"}}
        created_box_id = "box-created-123"
        # The API response *is* the box data for create
        api_response_data = {
            "id": created_box_id,
            "image": image,
            "name": "newbox",
            "labels": {"a": "b"},
        }
        self.mock_box_service.create.return_value = api_response_data
        mock_box_instance = Mock(spec=Box, id=created_box_id)
        MockBoxModel.return_value = mock_box_instance  # Use local MockBoxModel

        box = self.box_manager.create(image=image, **create_kwargs)

        self.mock_box_service.create.assert_called_once_with(image=image, **create_kwargs)
        MockBoxModel.assert_called_once_with(  # Use local MockBoxModel
            client=self.mock_client, id=created_box_id, attrs=api_response_data
        )

    @patch("gbox.managers.boxes.Box")
    def test_create_invalid_response_no_id(self, MockBoxModel):  # Add MockBoxModel arg
        """Test create when API response lacks an ID."""
        image = "bad-image"
        api_response_data = {"name": "noboxid"}
        self.mock_box_service.create.return_value = api_response_data

        with self.assertRaises(APIError) as cm:
            self.box_manager.create(image=image)

        self.assertIn("API did not return valid box data after creation", str(cm.exception))
        self.mock_box_service.create.assert_called_once_with(image=image)
        MockBoxModel.assert_not_called()  # Use local MockBoxModel

    @patch("gbox.managers.boxes.Box")
    def test_create_invalid_response_not_dict(self, MockBoxModel):  # Add MockBoxModel arg
        """Test create when API response is not a dictionary."""
        image = "bad-image"
        api_response_data = "just a string"
        self.mock_box_service.create.return_value = api_response_data

        with self.assertRaises(APIError) as cm:
            self.box_manager.create(image=image)

        self.assertIn("API did not return valid box data after creation", str(cm.exception))
        self.mock_box_service.create.assert_called_once_with(image=image)
        MockBoxModel.assert_not_called()  # Use local MockBoxModel

    @patch("gbox.managers.boxes.Box")
    def test_create_api_error(self, MockBoxModel):  # Add MockBoxModel arg
        """Test create when the API call itself fails."""
        image = "fail-image"
        error = APIError("Creation Forbidden", status_code=403)
        self.mock_box_service.create.side_effect = error

        with self.assertRaises(APIError) as cm:
            self.box_manager.create(image=image)

        self.assertIs(cm.exception, error)
        self.mock_box_service.create.assert_called_once_with(image=image)
        MockBoxModel.assert_not_called()  # Use local MockBoxModel

    def test_delete_all(self):
        """Test deleting all boxes."""
        expected_response = {"count": 5, "ids": ["1", "2", "3", "4", "5"]}
        self.mock_box_service.delete_all.return_value = expected_response

        response = self.box_manager.delete_all(force=False)

        self.mock_box_service.delete_all.assert_called_once_with(force=False)
        self.assertEqual(response, expected_response)

    def test_delete_all_force(self):
        """Test deleting all boxes with force."""
        expected_response = {"count": 5, "ids": ["1", "2", "3", "4", "5"]}
        self.mock_box_service.delete_all.return_value = expected_response

        response = self.box_manager.delete_all(force=True)

        self.mock_box_service.delete_all.assert_called_once_with(force=True)
        self.assertEqual(response, expected_response)

    def test_reclaim(self):
        """Test reclaiming all inactive boxes."""
        expected_response = {"stoppedCount": 2, "deletedCount": 1}
        self.mock_box_service.reclaim.return_value = expected_response

        response = self.box_manager.reclaim(force=False)

        # Verify reclaim is called for *all* boxes (box_id=None)
        self.mock_box_service.reclaim.assert_called_once_with(box_id=None, force=False)
        self.assertEqual(response, expected_response)

    def test_reclaim_force(self):
        """Test reclaiming all inactive boxes with force."""
        expected_response = {"stoppedCount": 3, "deletedCount": 0}
        self.mock_box_service.reclaim.return_value = expected_response

        response = self.box_manager.reclaim(force=True)

        self.mock_box_service.reclaim.assert_called_once_with(box_id=None, force=True)
        self.assertEqual(response, expected_response)


if __name__ == "__main__":
    unittest.main()
