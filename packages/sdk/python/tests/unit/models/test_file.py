from unittest.mock import MagicMock

import pytest

from gbox import File, GBoxClient


class TestFileModel:
    """Unit tests for the File model class."""

    def test_init_and_properties(self):
        """Test initialization and properties of File objects."""
        # Create mock client
        mock_client = MagicMock(spec=GBoxClient)

        # Create a File with attributes
        file_attrs = {
            "name": "test_file.txt",
            "size": 100,
            "mode": "-rw-r--r--",
            "modTime": "2023-01-01T12:00:00Z",
            "type": "file",
            "mime": "text/plain",
        }
        file = File(client=mock_client, path="/path/to/file.txt", attrs=file_attrs)

        # Test file path
        assert file.path == "/path/to/file.txt"

        # Test properties
        assert file.name == "test_file.txt"
        assert file.size == 100
        assert file.mode == "-rw-r--r--"
        assert file.mod_time == "2023-01-01T12:00:00Z"
        assert file.type == "file"
        assert file.mime == "text/plain"
        assert file.is_directory is False

        # Test with minimal attributes
        file_min = File(client=mock_client, path="/path/to/minimal.txt")
        assert file_min.path == "/path/to/minimal.txt"
        assert file_min.name == "minimal.txt"  # Should be extracted from path
        assert file_min.size is None
        assert file_min.is_directory is False  # Default when type is not set

        # Test directory
        dir_attrs = {"type": "directory"}
        dir_file = File(client=mock_client, path="/path/to/dir", attrs=dir_attrs)
        assert dir_file.is_directory is True

    def test_reload_method(self):
        """Test reload method fetches updated attributes."""
        # Create mock client with file_service.head method
        mock_client = MagicMock(spec=GBoxClient)
        # Explicitly create file_service as a mock
        mock_client.file_service = MagicMock()
        mock_head = mock_client.file_service.head

        # Setup return value for head method
        initial_attrs = {"name": "file.txt", "size": 100}
        updated_attrs = {"name": "file.txt", "size": 200, "type": "file"}
        mock_head.return_value = updated_attrs

        # Create file with initial attributes
        file = File(client=mock_client, path="/path/to/file.txt", attrs=initial_attrs)
        assert file.size == 100

        # Reload and check that attributes are updated
        file.reload()
        mock_head.assert_called_once_with("/path/to/file.txt")
        assert file.size == 200
        assert file.type == "file"

        # Test reload when head returns None
        mock_head.return_value = None
        file.reload()
        assert file.attrs == updated_attrs  # Should not change if head returns None

    def test_read_methods(self):
        """Test read and read_text methods."""
        # Create mock client with file_service.get method
        mock_client = MagicMock(spec=GBoxClient)
        # Explicitly create file_service as a mock
        mock_client.file_service = MagicMock()
        mock_get = mock_client.file_service.get

        # Setup return value for get method
        file_content = b"Hello, world!"
        mock_get.return_value = file_content

        # Create file as non-directory
        file_attrs = {"type": "file"}
        file = File(client=mock_client, path="/path/to/file.txt", attrs=file_attrs)

        # Test read method
        content = file.read()
        mock_get.assert_called_once_with("/path/to/file.txt")
        assert content == file_content

        # Test read_text method
        mock_get.reset_mock()
        mock_get.return_value = b"Hello, world!"
        text = file.read_text()
        mock_get.assert_called_once_with("/path/to/file.txt")
        assert text == "Hello, world!"

        # Test read on directory
        dir_file = File(client=mock_client, path="/path/to/dir", attrs={"type": "directory"})
        with pytest.raises(IsADirectoryError):
            dir_file.read()

        # Test read_text with custom encoding
        mock_get.reset_mock()
        mock_get.return_value = "Hello, World with UTF-16!".encode("utf-16")
        with pytest.raises(UnicodeDecodeError):  # Default utf-8 should fail
            file.read_text()

        mock_get.reset_mock()
        mock_get.return_value = "Hello, World with UTF-16!".encode("utf-16")
        text = file.read_text(encoding="utf-16")
        assert text == "Hello, World with UTF-16!"

    def test_equality_and_hash(self):
        """Test equality and hash methods."""
        # Create mock clients
        mock_client1 = MagicMock(spec=GBoxClient)
        mock_client2 = MagicMock(spec=GBoxClient)

        # Create files with same path but different clients/attributes
        file1 = File(client=mock_client1, path="/path/to/file.txt", attrs={"size": 100})
        file2 = File(client=mock_client2, path="/path/to/file.txt", attrs={"size": 200})
        file3 = File(client=mock_client1, path="/path/to/other.txt", attrs={"size": 100})

        # Test equality
        assert file1 == file2  # Same path, different attributes (equal)
        assert file1 != file3  # Different path (not equal)
        assert file1 != "not a file"  # Different type (not equal)

        # Test hash
        file_set = {file1, file2, file3}
        assert len(file_set) == 2  # file1 and file2 should hash to same value

    def test_repr(self):
        """Test string representation."""
        # Create mock client
        mock_client = MagicMock(spec=GBoxClient)

        # Test representation with type
        file_attrs = {"type": "file"}
        file = File(client=mock_client, path="/path/to/file.txt", attrs=file_attrs)
        assert repr(file) == "File(path='/path/to/file.txt', type='file')"

        # Test representation without type
        file_no_type = File(client=mock_client, path="/path/to/file.txt")
        assert repr(file_no_type) == "File(path='/path/to/file.txt', type='None')"
