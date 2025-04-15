# tests/models/test_boxes.py
import io
import unittest
from unittest.mock import MagicMock, Mock, patch

import pytest

from gbox.client import GBoxClient
from gbox.exceptions import APIError, NotFound  # Import exceptions
from gbox.models.boxes import Box, BoxBase

# from gbox.api.box_api import BoxApi


class TestBox(unittest.TestCase):

    def setUp(self):
        """Set up for test methods."""
        self.mock_client = Mock(spec=GBoxClient)
        self.mock_client.box_api = Mock()

        self.box_id = "box-12345678-abcd"
        self.box_attrs = {
            "id": self.box_id,
            "status": "stopped",
            "labels": {"env": "testing", "name": "test-box"},
            "image": "ubuntu:latest",
        }
        box_data = BoxBase(**self.box_attrs)
        self.box = Box(self.mock_client, box_data)

        simple_box_data = BoxBase(id="box123", status="running", image="simple:image")
        self.box_simple_id = Box(self.mock_client, simple_box_data)

    def test_init(self):
        """Test Box initialization."""
        self.assertEqual(self.box.id, self.box_id)
        self.assertEqual(self.box.attrs.id, self.box_attrs["id"])
        self.assertEqual(self.box.attrs.status, self.box_attrs["status"])
        self.assertEqual(self.box.attrs.labels, self.box_attrs["labels"])
        self.assertEqual(self.box.attrs.image, self.box_attrs["image"])
        self.assertEqual(self.box._client, self.mock_client)

    def test_short_id(self):
        """Test the short_id property."""
        self.assertEqual(self.box.short_id, "box-12345678")
        self.assertEqual(self.box_simple_id.short_id, "box123")

    def test_short_id_no_hyphen(self):
        """Test short_id property when ID has no hyphen."""
        self.assertEqual(self.box_simple_id.short_id, "box123")

    def test_name(self):
        """Test the name property."""
        self.assertEqual(self.box.name, "test-box")
        box_no_name_data = BoxBase(id="box-noname", status="stopped", image="img")
        box_no_name = Box(self.mock_client, box_no_name_data)
        self.assertIsNone(box_no_name.name)

    def test_status(self):
        """Test the status property."""
        self.assertEqual(self.box.status, "stopped")
        box_unknown_status_data = BoxBase(id="box-nostatus", status="unknown", image="img")
        box_unknown_status = Box(self.mock_client, box_unknown_status_data)
        self.assertEqual(box_unknown_status.status, "unknown")

    def test_labels(self):
        """Test the labels property."""
        self.assertEqual(self.box.labels, {"env": "testing", "name": "test-box"})
        box_no_labels_data = BoxBase(id="box-nolabels", status="stopped", image="img")
        box_no_labels = Box(self.mock_client, box_no_labels_data)
        self.assertEqual(box_no_labels.labels, {})

    def test_reload(self):
        """Test the reload method."""
        new_attrs_dict = {
            "id": self.box_id,
            "status": "running",
            "image": "reloaded:image",
            "labels": {"name": "reloaded-box"},
        }
        self.mock_client.box_api.get.return_value = new_attrs_dict

        self.box.reload()

        self.mock_client.box_api.get.assert_called_once_with(self.box_id)
        self.assertIsInstance(self.box.attrs, BoxBase)
        self.assertEqual(self.box.attrs.status, "running")
        self.assertEqual(self.box.attrs.image, "reloaded:image")
        self.assertEqual(self.box.name, "reloaded-box")

    def test_reload_api_error(self):
        """Test reload method when API call fails."""
        error = APIError("Failed to get", status_code=500)
        self.mock_client.box_api.get.side_effect = error
        with self.assertRaises(APIError) as cm:
            self.box.reload()
        self.assertIs(cm.exception, error)
        self.mock_client.box_api.get.assert_called_once_with(self.box_id)

    def test_start(self):
        """Test the start method."""
        start_response = {"success": True, "message": "Box started"}
        reload_attrs = {
            "id": self.box_id,
            "status": "running",
            "image": "ubuntu:latest",
            "labels": self.box_attrs["labels"],
        }
        self.mock_client.box_api.start.return_value = start_response
        self.mock_client.box_api.get.return_value = reload_attrs

        self.box.start()
        self.mock_client.box_api.start.assert_called_once_with(self.box_id)
        self.mock_client.box_api.get.assert_called_once_with(self.box_id)
        self.assertEqual(self.box.status, "running")

    def test_start_api_error(self):
        """Test start method when API call fails."""
        error = APIError("Failed to start", status_code=503)
        self.mock_client.box_api.start.side_effect = error
        with self.assertRaises(APIError) as cm:
            self.box.start()
        self.assertIs(cm.exception, error)
        self.mock_client.box_api.start.assert_called_once_with(self.box_id)
        self.mock_client.box_api.get.assert_not_called()

    def test_stop(self):
        """Test the stop method."""
        stop_response = {"success": True, "message": "Box stopped"}
        reload_attrs = {
            "id": self.box_id,
            "status": "stopped",
            "image": "ubuntu:latest",
            "labels": self.box_attrs["labels"],
        }
        self.mock_client.box_api.stop.return_value = stop_response
        self.mock_client.box_api.get.return_value = reload_attrs

        self.box.stop()
        self.mock_client.box_api.stop.assert_called_once_with(self.box_id)
        self.mock_client.box_api.get.assert_called_once_with(self.box_id)
        self.assertEqual(self.box.status, "stopped")

    def test_stop_api_error(self):
        """Test stop method when API call fails."""
        error = APIError("Failed to stop", status_code=503)
        self.mock_client.box_api.stop.side_effect = error
        with self.assertRaises(APIError) as cm:
            self.box.stop()
        self.assertIs(cm.exception, error)
        self.mock_client.box_api.stop.assert_called_once_with(self.box_id)

    def test_delete(self):
        """Test the delete method."""
        delete_response = {"message": "Box deleted"}
        self.mock_client.box_api.delete.return_value = delete_response

        self.box.delete()
        self.mock_client.box_api.delete.assert_called_once_with(self.box_id, force=False)

    def test_delete_api_error(self):
        """Test delete method when API call fails."""
        error = APIError("Failed to delete", status_code=500)
        self.mock_client.box_api.delete.side_effect = error
        with self.assertRaises(APIError) as cm:
            self.box.delete()
        self.assertIs(cm.exception, error)
        self.mock_client.box_api.delete.assert_called_once_with(self.box_id, force=False)

    def test_delete_force(self):
        """Test the delete method with force=True."""
        # Mock the API response to be a valid dictionary for BoxDeleteResponse
        delete_response_dict = {"message": "Box deleted forcibly"}
        self.mock_client.box_api.delete.return_value = delete_response_dict

        self.box.delete(force=True)

        self.mock_client.box_api.delete.assert_called_once_with(self.box_id, force=True)
        # Add check: no exception should be raised if validation passes

    @patch("logging.getLogger")
    def test_run(self, mock_get_logger):
        """Test the run method."""
        mock_logger = Mock()
        mock_get_logger.return_value = mock_logger

        command = ["echo", "hello"]
        run_response_box_attrs = self.box_attrs.copy()
        run_response_box_attrs["status"] = "exited"
        expected_response = {
            "box": run_response_box_attrs,
            "exitCode": 0,
            "stdout": "hello\n",
            "stderr": "",
        }
        self.mock_client.box_api.run.return_value = expected_response

        exit_code, stdout, stderr = self.box.run(command)

        self.mock_client.box_api.run.assert_called_once_with(self.box_id, command=command)
        self.assertEqual(exit_code, 0)
        self.assertEqual(stdout, "hello\n")
        self.assertEqual(stderr, "")
        mock_logger.debug.assert_called()

    @patch("logging.getLogger")
    def test_run_api_error(self, mock_get_logger):
        """Test run method when API call fails."""
        command = ["echo", "fail"]
        error = APIError("Failed to run command", status_code=500)
        self.mock_client.box_api.run.side_effect = error
        with self.assertRaises(APIError) as cm:
            self.box.run(command)
        self.assertIs(cm.exception, error)
        self.mock_client.box_api.run.assert_called_once_with(self.box_id, command=command)

    def test_reclaim(self):
        """Test the reclaim method."""
        expected_response_dict = {
            "message": "reclaimed",
            "stoppedIds": [],
            "deletedIds": [],
            "stoppedCount": 0,
            "deletedCount": 0,
        }
        self.mock_client.box_api.reclaim.return_value = expected_response_dict

        response_model = self.box.reclaim()

        self.mock_client.box_api.reclaim.assert_called_once_with(box_id=self.box_id, force=False)
        self.assertEqual(response_model.message, expected_response_dict["message"])
        self.assertEqual(response_model.stopped_ids, expected_response_dict["stoppedIds"])
        self.assertEqual(response_model.deleted_ids, expected_response_dict["deletedIds"])
        self.assertEqual(response_model.stopped_count, expected_response_dict["stoppedCount"])
        self.assertEqual(response_model.deleted_count, expected_response_dict["deletedCount"])

    def test_reclaim_api_error(self):
        """Test reclaim method when API call fails."""
        error = APIError("Failed to reclaim", status_code=500)
        self.mock_client.box_api.reclaim.side_effect = error
        with self.assertRaises(APIError) as cm:
            self.box.reclaim()
        self.assertIs(cm.exception, error)
        self.mock_client.box_api.reclaim.assert_called_once_with(box_id=self.box_id, force=False)

    def test_reclaim_force(self):
        """Test the reclaim method with force=True."""
        expected_response_dict = {
            "message": "force reclaimed",
            "stoppedIds": ["box-stopped"],
            "deletedIds": ["box-deleted"],
            "stoppedCount": 1,
            "deletedCount": 1,
        }
        self.mock_client.box_api.reclaim.return_value = expected_response_dict

        response_model = self.box.reclaim(force=True)

        self.mock_client.box_api.reclaim.assert_called_once_with(box_id=self.box_id, force=True)
        self.assertEqual(response_model.message, expected_response_dict["message"])
        self.assertEqual(response_model.stopped_ids, expected_response_dict["stoppedIds"])
        self.assertEqual(response_model.deleted_ids, expected_response_dict["deletedIds"])
        self.assertEqual(response_model.stopped_count, expected_response_dict["stoppedCount"])
        self.assertEqual(response_model.deleted_count, expected_response_dict["deletedCount"])

    def test_head_archive(self):
        """Test the head_archive method."""
        path = "/data/file.txt"
        expected_headers = {"Content-Length": "1024", "X-GBox-Mode": "0644"}
        self.mock_client.box_api.head_archive.return_value = expected_headers

        headers = self.box.head_archive(path)

        self.mock_client.box_api.head_archive.assert_called_once_with(self.box_id, path=path)
        self.assertEqual(headers, expected_headers)

    def test_head_archive_api_error(self):
        """Test head_archive method when API call fails with a generic error."""
        path = "/data/file.txt"
        error = APIError("HEAD failed", status_code=500)
        self.mock_client.box_api.head_archive.side_effect = error
        with self.assertRaises(APIError) as cm:
            self.box.head_archive(path)
        self.assertIs(cm.exception, error)
        self.mock_client.box_api.head_archive.assert_called_once_with(self.box_id, path=path)

    def test_head_archive_not_found(self):
        """Test head_archive method when the path is not found."""
        path = "/data/not_found.txt"
        error = NotFound("Path not found")
        self.mock_client.box_api.head_archive.side_effect = error
        with self.assertRaises(NotFound) as cm:
            self.box.head_archive(path)
        self.assertIs(cm.exception, error)
        self.mock_client.box_api.head_archive.assert_called_once_with(self.box_id, path=path)

    def test_get_archive(self):
        """Test the get_archive method."""
        path = "/data"
        mock_stats = {"Content-Type": "application/x-tar", "X-GBox-Size": "2048"}
        mock_tar_data = b"tar data bytes"

        self.mock_client.box_api.head_archive.return_value = mock_stats
        # Mock get_archive to return bytes directly, not a tuple
        self.mock_client.box_api.get_archive.return_value = mock_tar_data

        content_stream, stats = self.box.get_archive(path)

        self.mock_client.box_api.head_archive.assert_called_once_with(self.box_id, path=path)
        self.mock_client.box_api.get_archive.assert_called_once_with(self.box_id, path=path)

        self.assertIsInstance(content_stream, io.BytesIO)
        self.assertEqual(content_stream.read(), mock_tar_data)
        self.assertEqual(stats, mock_stats)

    def test_get_archive_head_fails(self):
        """Test get_archive when the initial head_archive call fails."""
        path = "/data"
        error = APIError("HEAD failed first", status_code=500)
        self.mock_client.box_api.head_archive.side_effect = error

        with self.assertRaises(APIError) as cm:
            self.box.get_archive(path)

        self.assertIs(cm.exception, error)
        self.mock_client.box_api.head_archive.assert_called_once_with(self.box_id, path=path)
        self.mock_client.box_api.get_archive.assert_not_called()

    def test_get_archive_get_fails(self):
        """Test get_archive when the get_archive API call fails."""
        path = "/data"
        mock_stats = {"Content-Type": "application/x-tar"}
        error = APIError("GET archive failed", status_code=500)

        self.mock_client.box_api.head_archive.return_value = mock_stats
        self.mock_client.box_api.get_archive.side_effect = error

        with self.assertRaises(APIError) as cm:
            self.box.get_archive(path)

        self.assertIs(cm.exception, error)
        self.mock_client.box_api.head_archive.assert_called_once_with(self.box_id, path=path)
        self.mock_client.box_api.get_archive.assert_called_once_with(self.box_id, path=path)

    def test_put_archive_bytes(self):
        """Test put_archive with bytes data."""
        path = "/uploads"
        tar_data = b"some tar data"

        self.box.put_archive(path, tar_data)

        self.mock_client.box_api.extract_archive.assert_called_once_with(
            self.box_id, path=path, archive_data=tar_data
        )

    def test_put_archive_file(self):
        """Test put_archive with a file-like object."""
        path = "/uploads"
        tar_data = b"tar data from file"

        mock_file = MagicMock(spec=io.BufferedReader)
        mock_file.read.return_value = tar_data

        self.box.put_archive(path, mock_file)

        mock_file.read.assert_called_once()
        self.mock_client.box_api.extract_archive.assert_called_once_with(
            self.box_id, path=path, archive_data=tar_data
        )

    def test_put_archive_invalid_type(self):
        """Test put_archive raises TypeError for invalid data types."""
        with pytest.raises(TypeError):
            self.box.put_archive("/data", data=12345)

        with pytest.raises(FileNotFoundError):
            self.box.put_archive("/data", data="this is a string, not a path")

    def test_put_archive_directory_path(self):
        pass

    def test_put_archive_api_error(self):
        """Test put_archive method when API call fails."""
        path = "/uploads"
        tar_data = b"tar data"
        error = APIError("Extract failed", status_code=500)
        self.mock_client.box_api.extract_archive.side_effect = error

        with self.assertRaises(APIError) as cm:
            self.box.put_archive(path, tar_data)

        self.assertIs(cm.exception, error)
        self.mock_client.box_api.extract_archive.assert_called_once_with(
            self.box_id, path=path, archive_data=tar_data
        )

    def test_eq(self):
        """Test equality comparison."""
        same_id_attrs = self.box_attrs.copy()
        same_id_attrs["status"] = "running"
        same_id_data = BoxBase(**same_id_attrs)
        same_box = Box(self.mock_client, same_id_data)

        diff_id_attrs = self.box_attrs.copy()
        diff_id_attrs["id"] = "box-different-id"
        diff_id_data = BoxBase(**diff_id_attrs)
        diff_box = Box(self.mock_client, diff_id_data)

        self.assertEqual(self.box, same_box)
        self.assertNotEqual(self.box, diff_box)
        self.assertNotEqual(self.box, "not a box")

    def test_hash(self):
        """Test hashing."""
        same_id_attrs = self.box_attrs.copy()
        same_id_attrs["status"] = "running"
        same_id_data = BoxBase(**same_id_attrs)
        same_box = Box(self.mock_client, same_id_data)

        self.assertEqual(hash(self.box), hash(same_box))
        self.assertEqual(hash(self.box), hash(self.box.id))

    def test_repr(self):
        """Test the string representation."""
        # Update expected repr to match the actual format
        expected_repr = f"<Box: {self.box.short_id} ({self.box.status})>"  # Removed 'status='
        self.assertEqual(repr(self.box), expected_repr)


if __name__ == "__main__":
    unittest.main()
