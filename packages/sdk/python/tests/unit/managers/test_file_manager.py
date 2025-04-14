from unittest.mock import MagicMock, patch

import pytest

from gbox import Box, File, GBoxClient
from gbox.exceptions import APIError, NotFound
from gbox.managers.files import FileManager


class TestFileManager:
    """Unit tests for the FileManager class."""

    def test_init(self):
        """Test initialization of FileManager."""
        # Create mock client
        mock_client = MagicMock(spec=GBoxClient)
        mock_client.file_service = MagicMock()

        # Create manager
        manager = FileManager(client=mock_client)

        # Test manager properties
        assert manager._client is mock_client
        assert manager._service is mock_client.file_service

    def test_get_method(self):
        """Test get method returns a File object."""
        # Create mock client and service
        mock_client = MagicMock(spec=GBoxClient)
        mock_client.file_service = MagicMock()
        mock_head = mock_client.file_service.head

        # Setup return value for head method
        file_attrs = {"name": "test.txt", "size": 100, "type": "file"}
        mock_head.return_value = file_attrs

        # Create manager
        manager = FileManager(client=mock_client)

        # Test get with absolute path
        file = manager.get("/path/to/file.txt")
        mock_head.assert_called_once_with("/path/to/file.txt")
        assert isinstance(file, File)
        assert file.path == "/path/to/file.txt"
        assert file.attrs == file_attrs

        # Test get with relative path (should normalize)
        mock_head.reset_mock()
        mock_head.return_value = file_attrs
        file = manager.get("path/to/file.txt")
        mock_head.assert_called_once_with("/path/to/file.txt")
        assert file.path == "/path/to/file.txt"

        # Test get with None response from head
        mock_head.reset_mock()
        mock_head.return_value = None
        with pytest.raises(NotFound):
            manager.get("/not/found.txt")

    def test_exists_method(self):
        """Test exists method checks if a file exists."""
        # Create mock client and service
        mock_client = MagicMock(spec=GBoxClient)
        mock_client.file_service = MagicMock()
        mock_head = mock_client.file_service.head

        # Create manager
        manager = FileManager(client=mock_client)

        # Test exists with file that exists
        mock_head.return_value = {"name": "test.txt"}
        exists = manager.exists("/path/to/file.txt")
        mock_head.assert_called_once_with("/path/to/file.txt")
        assert exists is True

        # Test exists with file that doesn't exist
        mock_head.reset_mock()
        mock_head.return_value = None
        exists = manager.exists("/not/found.txt")
        mock_head.assert_called_once_with("/not/found.txt")
        assert exists is False

        # Test exists with relative path (should normalize)
        mock_head.reset_mock()
        mock_head.return_value = {"name": "test.txt"}
        exists = manager.exists("path/to/file.txt")
        mock_head.assert_called_once_with("/path/to/file.txt")
        assert exists is True

        # Test exists catching NotFound exception
        mock_head.reset_mock()
        mock_head.side_effect = NotFound("File not found")
        exists = manager.exists("/not/found.txt")
        assert exists is False

        # Test exists catching general APIError
        mock_head.reset_mock()
        mock_head.side_effect = APIError("API error")
        exists = manager.exists("/error/path.txt")
        assert exists is False

    def test_share_from_box_method(self):
        """Test share_from_box method with box_id and Box object."""
        # Create mock client, service, and response
        mock_client = MagicMock(spec=GBoxClient)
        mock_client.file_service = MagicMock()
        mock_share = mock_client.file_service.share

        # Mock file list in response
        share_response = {
            "success": True,
            "message": "File shared successfully",
            "fileList": [
                {"name": "test.txt", "path": "/box-123/test.txt", "size": 100, "type": "file"}
            ],
        }
        mock_share.return_value = share_response

        # Setup mock get method to return a File object
        mock_get = MagicMock()
        mock_file = MagicMock(spec=File)
        mock_get.return_value = mock_file

        # Create manager with mocked get method
        manager = FileManager(client=mock_client)
        manager.get = mock_get

        # Test share_from_box with box_id string
        box_id = "box-123"
        box_path = "/var/gbox/share/test.txt"
        file = manager.share_from_box(box_id, box_path)
        mock_share.assert_called_once_with(box_id, box_path)
        mock_get.assert_called_once_with("/box-123/test.txt")
        assert file is mock_file

        # Test share_from_box with Box object
        mock_share.reset_mock()
        mock_get.reset_mock()
        mock_box = MagicMock(spec=Box)
        mock_box.id = "box-456"
        file = manager.share_from_box(mock_box, box_path)
        mock_share.assert_called_once_with("box-456", box_path)
        mock_get.assert_called_once()

        # Test share_from_box with invalid box type
        with pytest.raises(TypeError):
            manager.share_from_box(123, box_path)  # Not a string or Box

        # Test share_from_box with invalid path
        with pytest.raises(ValueError):
            manager.share_from_box(box_id, "/invalid/path.txt")  # Not starting with /var/gbox/

        # Test share_from_box with empty file list
        mock_share.reset_mock()
        mock_share.return_value = {"success": True, "fileList": []}
        with pytest.raises(FileNotFoundError):
            manager.share_from_box(box_id, box_path)

        # Test share_from_box with no file list
        mock_share.reset_mock()
        mock_share.return_value = {"success": True}
        with pytest.raises(FileNotFoundError):
            manager.share_from_box(box_id, box_path)

        # Test share_from_box with no path in file info but with name
        mock_share.reset_mock()
        mock_share.return_value = {"success": True, "fileList": [{"name": "test.txt"}]}
        file = manager.share_from_box(box_id, box_path)
        # Should construct path from box_id and name
        mock_get.assert_called_with("/box-123/test.txt")

    def test_reclaim_method(self):
        """Test reclaim method for files."""
        # Create mock client and service
        mock_client = MagicMock(spec=GBoxClient)
        mock_client.file_service = MagicMock()
        mock_reclaim = mock_client.file_service.reclaim

        # Setup return value for reclaim method
        reclaim_response = {"reclaimed_files": ["/path/to/old.txt"]}
        mock_reclaim.return_value = reclaim_response

        # Create manager
        manager = FileManager(client=mock_client)

        # Test reclaim
        response = manager.reclaim()
        mock_reclaim.assert_called_once()
        assert response == reclaim_response
