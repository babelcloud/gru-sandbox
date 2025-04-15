from unittest.mock import MagicMock, Mock

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
        mock_client.file_api = MagicMock()

        # Create manager
        manager = FileManager(client=mock_client)

        # Test manager properties
        assert manager._client is mock_client
        assert manager._service is mock_client.file_api

    def test_get_method(self):
        """Test get method returns a File object."""
        # Create mock client and service
        mock_client = MagicMock(spec=GBoxClient)
        mock_client.file_api = MagicMock()
        mock_head = mock_client.file_api.head

        # Setup return value for head method (needs all FileStat fields)
        file_attrs = {
            "name": "test.txt",
            "path": "/path/to/file.txt",
            "size": 100,
            "mode": "-rw-r--r--",
            "modTime": "t1",
            "type": "file",
            "mime": "text/plain",
        }
        mock_head.return_value = file_attrs

        # Create manager
        manager = FileManager(client=mock_client)

        # Test get with absolute path
        file = manager.get("/path/to/file.txt")
        mock_head.assert_called_once_with("/path/to/file.txt")
        assert isinstance(file, File)
        assert file.path == "/path/to/file.txt"
        # Compare attributes of the FileStat object
        assert file.attrs.name == file_attrs["name"]
        assert file.attrs.path == file_attrs["path"]
        assert file.attrs.size == file_attrs["size"]
        assert file.attrs.mode == file_attrs["mode"]
        assert file.attrs.mod_time == file_attrs["modTime"]
        assert file.attrs.type == file_attrs["type"]
        assert file.attrs.mime == file_attrs["mime"]

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
        mock_client.file_api = MagicMock()
        mock_head = mock_client.file_api.head

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
        mock_client.file_api = MagicMock()
        mock_share = mock_client.file_api.share

        # Mock file list in response (needs all FileStat fields in fileList)
        shared_file_attrs = {
            "name": "test.txt",
            "path": "/box-123/test.txt",
            "size": 100,
            "mode": "-rw-r--r--",
            "modTime": "t1",
            "type": "file",
            "mime": "text/plain",
        }
        share_response = {
            "success": True,
            "message": "File shared successfully",
            "fileList": [shared_file_attrs],
        }
        mock_share.return_value = share_response

        # Mock the head call that happens inside manager.get
        mock_head = mock_client.file_api.head
        # This head call needs to return the full FileStat data for the shared file
        # Use the same attrs dict used in the share_response for consistency
        mock_head.return_value = shared_file_attrs

        # Create manager
        manager = FileManager(client=mock_client)

        # Test share_from_box with box_id string
        box_id = "box-123"
        box_path = "/var/gbox/share/test.txt"
        file = manager.share_from_box(box_id, box_path)

        # Assert API call
        mock_share.assert_called_once_with(box_id, box_path)
        # Assert returned object is a File instance constructed with validated data
        assert isinstance(file, File)
        assert file.name == shared_file_attrs["name"]
        # Path might be constructed differently by the manager, check relevant part
        assert file.path.endswith(shared_file_attrs["name"])
        assert file.size == shared_file_attrs["size"]
        assert file.type == shared_file_attrs["type"]

        # Test share_from_box with Box object
        mock_box = Mock(spec=Box)
        mock_box.id = box_id  # Mock the id attribute
        mock_share.reset_mock()  # Reset mock for the next call
        # Need to mock head again for the second call via manager.get
        mock_head.reset_mock()
        mock_head.return_value = shared_file_attrs

        file2 = manager.share_from_box(mock_box, box_path)

        # Assert API call again
        mock_share.assert_called_once_with(box_id, box_path)
        # Assert returned object is a File instance
        assert isinstance(file2, File)
        assert file.path == file2.path  # Should point to the same shared file path

    def test_reclaim_method(self):
        """Test reclaim method for files."""
        # Create mock client and service
        mock_client = MagicMock(spec=GBoxClient)
        mock_client.file_api = MagicMock()
        mock_reclaim = mock_client.file_api.reclaim

        # Setup return value for reclaim method
        reclaim_response = {"reclaimed_files": ["/path/to/old.txt"]}
        mock_reclaim.return_value = reclaim_response

        # Create manager
        manager = FileManager(client=mock_client)

        # Test reclaim
        response = manager.reclaim()
        mock_reclaim.assert_called_once()
        assert response == reclaim_response
