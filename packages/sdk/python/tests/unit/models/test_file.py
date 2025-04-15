from unittest.mock import MagicMock

import pytest

from gbox import File, GBoxClient, NotFound
from gbox.exceptions import APIError
from gbox.models.files import FileStat


class TestFileModel:
    """Unit tests for the File model class."""

    def test_init_and_properties(self):
        """Test initialization and properties of File objects."""
        # Create mock client
        mock_client = MagicMock(spec=GBoxClient)
        file_path = "/path/to/file.txt"

        # Create a File with attributes using FileStat
        file_attrs_dict = {
            "name": "test_file.txt",
            "path": file_path,  # Ensure path matches the model
            "size": 100,
            "mode": "-rw-r--r--",
            "modTime": "2023-01-01T12:00:00Z",
            "type": "file",
            "mime": "text/plain",
        }
        file_stat = FileStat(**file_attrs_dict)
        file = File(client=mock_client, path=file_path, attrs=file_stat)

        # Test file path (the one used for fetching)
        assert file.path == file_path

        # Test properties derived from FileStat
        assert file.name == "test_file.txt"
        assert file.size == 100
        assert file.mode == "-rw-r--r--"
        assert file.mod_time == "2023-01-01T12:00:00Z"  # Access via mod_time now
        assert file.type == "file"
        assert file.mime == "text/plain"
        assert file.is_directory is False

        # Test directory using FileStat
        dir_path = "/path/to/dir"
        dir_attrs_dict = {
            "name": "dir",
            "path": dir_path,
            "size": 4096,
            "mode": "drwxr-xr-x",
            "modTime": "2023-01-02T10:00:00Z",
            "type": "directory",
            "mime": "inode/directory",
        }
        dir_stat = FileStat(**dir_attrs_dict)
        dir_file = File(client=mock_client, path=dir_path, attrs=dir_stat)
        assert dir_file.is_directory is True
        assert dir_file.type == "directory"

    def test_reload_method(self):
        """Test reload method fetches updated attributes."""
        # Create mock client with file_api.head method
        mock_client = MagicMock(spec=GBoxClient)
        mock_client.file_api = MagicMock()
        mock_head = mock_client.file_api.head
        file_path = "/path/to/file.txt"

        # Setup initial and updated attribute dictionaries for FileStat
        initial_attrs_dict = {
            "name": "file.txt",
            "path": file_path,
            "size": 100,
            "mode": "-rw-----",
            "modTime": "2023-01-01T00:00:00Z",
            "type": "file",
            "mime": "text/plain",
        }
        updated_attrs_dict = {
            "name": "file_updated.txt",
            "path": file_path,
            "size": 200,
            "mode": "-rwxr-xr-x",
            "modTime": "2023-01-01T12:00:00Z",
            "type": "file",
            "mime": "text/updated",
        }

        # Mock head API call to return the updated attributes dictionary
        mock_head.return_value = updated_attrs_dict

        # Create file with initial attributes via FileStat
        initial_stat = FileStat(**initial_attrs_dict)
        file = File(client=mock_client, path=file_path, attrs=initial_stat)
        assert file.size == 100
        assert file.name == "file.txt"

        # Reload and check that attributes are updated
        file.reload()
        mock_head.assert_called_once_with(file_path)
        # Verify the internal attrs is now the updated FileStat object
        assert isinstance(file.attrs, FileStat)
        assert file.size == 200
        assert file.name == "file_updated.txt"
        assert file.type == "file"
        assert file.mime == "text/updated"

        # Test reload when head raises NotFound (simulating file deleted)
        mock_head.reset_mock()
        mock_head.side_effect = NotFound("File not found")
        with pytest.raises(NotFound):
            file.reload()  # Reload should propagate the NotFound exception
        mock_head.assert_called_once_with(file_path)
        # Attributes should remain as they were before the failed reload
        assert file.size == 200

        # Test reload when head raises other APIError
        mock_head.reset_mock()
        mock_head.side_effect = APIError("Server error", status_code=500)
        with pytest.raises(APIError):
            file.reload()
        mock_head.assert_called_once_with(file_path)

    def test_read_methods(self):
        """Test read and read_text methods."""
        # Create mock client with file_api.get method
        mock_client = MagicMock(spec=GBoxClient)
        mock_client.file_api = MagicMock()
        mock_get = mock_client.file_api.get
        file_path = "/path/to/file.txt"

        # Setup return value for get method
        file_content = b"Hello, world!"
        mock_get.return_value = file_content

        # Create file as non-directory using FileStat
        file_attrs_dict = {
            "name": "file.txt",
            "path": file_path,
            "size": len(file_content),
            "mode": "-rw-r--r--",
            "modTime": "t1",
            "type": "file",
            "mime": "text/plain",
        }
        file_stat = FileStat(**file_attrs_dict)
        file = File(client=mock_client, path=file_path, attrs=file_stat)

        # Test read method
        content = file.read()
        mock_get.assert_called_once_with(file_path)
        assert content == file_content

        # Test read_text method
        mock_get.reset_mock()
        mock_get.return_value = b"Hello, world!"
        text = file.read_text()
        mock_get.assert_called_once_with(file_path)
        assert text == "Hello, world!"

        # Test read on directory
        dir_path = "/path/to/dir"
        dir_attrs_dict = {
            "name": "dir",
            "path": dir_path,
            "size": 0,
            "mode": "drwx------",
            "modTime": "t2",
            "type": "directory",
            "mime": "inode/directory",
        }
        dir_stat = FileStat(**dir_attrs_dict)
        dir_file = File(client=mock_client, path=dir_path, attrs=dir_stat)
        with pytest.raises(IsADirectoryError):
            dir_file.read()

        # Test read_text with custom encoding
        mock_get.reset_mock()
        encoded_content = "Hello, World with UTF-16!".encode("utf-16")
        mock_get.return_value = encoded_content
        # Update file size for accuracy in mock FileStat
        file_attrs_dict["size"] = len(encoded_content)
        file = File(client=mock_client, path=file_path, attrs=FileStat(**file_attrs_dict))

        with pytest.raises(UnicodeDecodeError):  # Default utf-8 should fail
            file.read_text()

        mock_get.reset_mock()
        mock_get.return_value = encoded_content
        text = file.read_text(encoding="utf-16")
        assert text == "Hello, World with UTF-16!"
        mock_get.assert_called_once_with(file_path)

    def test_equality_and_hash(self):
        """Test equality and hash methods."""
        # Create mock clients
        mock_client1 = MagicMock(spec=GBoxClient)
        mock_client2 = MagicMock(spec=GBoxClient)

        # Create minimal FileStat for testing equality (only path matters)
        common_path = "/path/to/file.txt"
        attrs1 = FileStat(
            name="f", path=common_path, size=1, mode="m", modTime="t", type="file", mime="m"
        )
        attrs2 = FileStat(
            name="f2", path=common_path, size=2, mode="m2", modTime="t2", type="file", mime="m2"
        )
        attrs3 = FileStat(
            name="f3", path="/other.txt", size=1, mode="m", modTime="t", type="file", mime="m"
        )

        # Create files with same path but different clients/attributes
        file1 = File(client=mock_client1, path=common_path, attrs=attrs1)
        file2 = File(client=mock_client2, path=common_path, attrs=attrs2)
        file3 = File(client=mock_client1, path="/other.txt", attrs=attrs3)

        # Test equality (based on path)
        assert file1 == file2  # Same path, different attributes (equal)
        assert file1 != file3  # Different path (not equal)
        assert file1 != "not a file"  # Different type (not equal)

        # Test hash (based on path)
        file_set = {file1, file2, file3}
        assert len(file_set) == 2  # file1 and file2 should hash to same value

    def test_repr(self):
        """Test string representation."""
        # Create mock client
        mock_client = MagicMock(spec=GBoxClient)
        file_path = "/path/to/file.txt"

        # Test representation with type='file'
        file_attrs_dict = {
            "name": "f",
            "path": file_path,
            "size": 1,
            "mode": "m",
            "modTime": "t",
            "type": "file",
            "mime": "m",
        }
        file_stat = FileStat(**file_attrs_dict)
        file = File(client=mock_client, path=file_path, attrs=file_stat)
        assert repr(file) == f"File(path='{file_path}', type='file')"

        # Test representation with type='directory'
        dir_path = "/path/to/dir"
        dir_attrs_dict = {
            "name": "d",
            "path": dir_path,
            "size": 1,
            "mode": "m",
            "modTime": "t",
            "type": "directory",
            "mime": "m",
        }
        dir_stat = FileStat(**dir_attrs_dict)
        dir_file = File(client=mock_client, path=dir_path, attrs=dir_stat)
        assert repr(dir_file) == f"File(path='{dir_path}', type='directory')"
