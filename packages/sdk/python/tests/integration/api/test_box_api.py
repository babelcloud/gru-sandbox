# tests/api/test_box_api.py
import unittest
from unittest.mock import Mock

from gbox.api.box_api import BoxApi
from gbox.api.client import Client as ApiClient  # For type hinting/mocking spec
from gbox.config import GBoxConfig  # For type hinting/mocking spec


class TestBoxApi(unittest.TestCase):

    def setUp(self):
        """Set up test fixtures."""
        self.mock_api_client = Mock(spec=ApiClient)
        self.mock_config = Mock(spec=GBoxConfig)
        self.mock_config.logger = Mock()  # BoxApi uses the logger from config
        self.box_api = BoxApi(client=self.mock_api_client, config=self.mock_config)
        self.box_id = "box-test-123"

    def test_init(self):
        """Test BoxApi initialization."""
        self.assertIs(self.box_api.client, self.mock_api_client)
        self.assertIs(self.box_api.logger, self.mock_config.logger)

    def test_list_no_filters(self):
        """Test list boxes without filters."""
        expected_response = {"boxes": []}
        self.mock_api_client.get.return_value = expected_response

        response = self.box_api.list()

        self.mock_api_client.get.assert_called_once_with("/api/v1/boxes", params={})
        self.assertEqual(response, expected_response)

    def test_list_with_filters(self):
        """Test list boxes with various filters."""
        filters = {
            "label": "env=prod",
            "ancestor": ["ubuntu:latest", "alpine:latest"],
            "id": "box-specific",
        }
        expected_params = {
            "filter": [
                "label=env=prod",
                "ancestor=ubuntu:latest",
                "ancestor=alpine:latest",
                "id=box-specific",
            ]
        }
        expected_response = {"boxes": [{"id": "box-specific"}]}
        self.mock_api_client.get.return_value = expected_response

        response = self.box_api.list(filters=filters)

        # Use assert_called_once_with, ensuring params order doesn't matter for the list
        self.mock_api_client.get.assert_called_once()
        call_args, call_kwargs = self.mock_api_client.get.call_args
        self.assertEqual(call_args[0], "/api/v1/boxes")
        # Sort the lists before comparison to handle potential order differences
        self.assertListEqual(
            sorted(call_kwargs["params"]["filter"]), sorted(expected_params["filter"])
        )
        self.assertEqual(response, expected_response)

    def test_get(self):
        """Test get box details."""
        expected_response = {"id": self.box_id, "status": "running"}
        self.mock_api_client.get.return_value = expected_response

        response = self.box_api.get(self.box_id)

        self.mock_api_client.get.assert_called_once_with(f"/api/v1/boxes/{self.box_id}")
        self.assertEqual(response, expected_response)

    def test_create_minimal(self):
        """Test create box with minimal arguments."""
        image = "alpine"
        expected_data = {"image": image}
        expected_response = {"box": {"id": "new-box"}, "message": "Created"}
        self.mock_api_client.post.return_value = expected_response

        response = self.box_api.create(image=image)

        self.mock_api_client.post.assert_called_once_with("/api/v1/boxes", data=expected_data)
        self.assertEqual(response, expected_response)

    def test_create_all_args(self):
        """Test create box with all arguments."""
        args = {
            "image": "ubuntu:latest",
            "image_pull_secret": "mysecret",
            "env": {"VAR1": "value1"},
            "cmd": "/bin/bash",
            "args": ["-c", "echo hello"],
            "working_dir": "/app",
            "labels": {"owner": "test"},
            "volumes": [
                {
                    "source": "/host/path",
                    "target": "/container/path",
                    "readOnly": True,
                    "propagation": "rprivate",
                }
            ],
        }
        expected_data = {
            "image": "ubuntu:latest",
            "imagePullSecret": "mysecret",
            "env": {"VAR1": "value1"},
            "cmd": "/bin/bash",
            "args": ["-c", "echo hello"],
            "workingDir": "/app",
            "labels": {"owner": "test"},
            "volumes": [
                {
                    "source": "/host/path",
                    "target": "/container/path",
                    "readOnly": True,
                    "propagation": "rprivate",
                }
            ],
        }
        expected_response = {"box": {"id": "full-box"}, "message": "Created"}
        self.mock_api_client.post.return_value = expected_response

        response = self.box_api.create(**args)

        self.mock_api_client.post.assert_called_once_with("/api/v1/boxes", data=expected_data)
        self.assertEqual(response, expected_response)

    def test_delete_no_force(self):
        """Test delete box without force."""
        expected_response = {"message": "Deleted"}
        self.mock_api_client.delete.return_value = expected_response

        response = self.box_api.delete(self.box_id, force=False)

        self.mock_api_client.delete.assert_called_once_with(f"/api/v1/boxes/{self.box_id}", data={})
        self.assertEqual(response, expected_response)

    def test_delete_force(self):
        """Test delete box with force."""
        expected_response = {"message": "Force Deleted"}
        self.mock_api_client.delete.return_value = expected_response

        response = self.box_api.delete(self.box_id, force=True)

        self.mock_api_client.delete.assert_called_once_with(
            f"/api/v1/boxes/{self.box_id}", data={"force": True}
        )
        self.assertEqual(response, expected_response)

    def test_delete_all_no_force(self):
        """Test delete all boxes without force."""
        expected_response = {"count": 0, "message": "Deleted 0 boxes"}
        self.mock_api_client.delete.return_value = expected_response

        response = self.box_api.delete_all(force=False)

        self.mock_api_client.delete.assert_called_once_with("/api/v1/boxes", data={})
        self.assertEqual(response, expected_response)

    def test_delete_all_force(self):
        """Test delete all boxes with force."""
        expected_response = {"count": 2, "message": "Deleted 2 boxes"}
        self.mock_api_client.delete.return_value = expected_response

        response = self.box_api.delete_all(force=True)

        self.mock_api_client.delete.assert_called_once_with("/api/v1/boxes", data={"force": True})
        self.assertEqual(response, expected_response)

    def test_start(self):
        """Test start box."""
        expected_response = {"success": True}
        self.mock_api_client.post.return_value = expected_response

        response = self.box_api.start(self.box_id)

        self.mock_api_client.post.assert_called_once_with(f"/api/v1/boxes/{self.box_id}/start")
        self.assertEqual(response, expected_response)

    def test_stop(self):
        """Test stop box."""
        expected_response = {"success": True}
        self.mock_api_client.post.return_value = expected_response

        response = self.box_api.stop(self.box_id)

        self.mock_api_client.post.assert_called_once_with(f"/api/v1/boxes/{self.box_id}/stop")
        self.assertEqual(response, expected_response)

    def test_run(self):
        """Test run command in box."""
        command = ["echo", "hello world"]
        expected_data = {"cmd": command[:1], "args": command[1:]}  # API splits cmd/args
        expected_response = {"exitCode": 0, "stdout": "hello world\n"}
        self.mock_api_client.post.return_value = expected_response

        response = self.box_api.run(self.box_id, command=command)

        self.mock_api_client.post.assert_called_once_with(
            f"/api/v1/boxes/{self.box_id}/run", data=expected_data
        )
        self.assertEqual(response, expected_response)

    def test_reclaim_specific_box_no_force(self):
        """Test reclaim for a specific box without force."""
        expected_data = {"force": False}  # boxId is in the path
        expected_response = {"message": "Reclaimed"}
        self.mock_api_client.post.return_value = expected_response

        response = self.box_api.reclaim(box_id=self.box_id, force=False)

        self.mock_api_client.post.assert_called_once_with(
            f"/api/v1/boxes/{self.box_id}/reclaim", data=expected_data
        )  # Path includes box_id
        self.assertEqual(response, expected_response)

    def test_reclaim_all_boxes_force(self):
        """Test reclaim for all boxes with force."""
        expected_data = {"force": True}  # No boxId
        expected_response = {"message": "All reclaimed"}
        self.mock_api_client.post.return_value = expected_response

        response = self.box_api.reclaim(box_id=None, force=True)

        self.mock_api_client.post.assert_called_once_with(
            "/api/v1/boxes/reclaim", data=expected_data
        )
        self.assertEqual(response, expected_response)

    def test_get_archive(self):
        """Test get archive from box."""
        path = "/data/archive.tar"
        expected_params = {"path": path}
        expected_response_bytes = b"\x01\x02\x03tar data"
        self.mock_api_client.get.return_value = (
            expected_response_bytes  # Assuming client.get returns bytes for this
        )

        response = self.box_api.get_archive(self.box_id, path=path)

        self.mock_api_client.get.assert_called_once_with(
            f"/api/v1/boxes/{self.box_id}/archive",
            params=expected_params,
            headers={"Accept": "application/x-tar"},  # Add expected headers
            raw_response=True,  # Add expected raw_response flag
        )
        self.assertEqual(response, expected_response_bytes)

    def test_extract_archive(self):
        """Test extract archive into box."""
        path = "/extract/here"
        archive_data = b"some tar data"
        expected_params = {"path": path}
        expected_headers = {"Content-Type": "application/x-tar"}  # Add expected headers
        expected_response = {"message": "Extracted"}
        self.mock_api_client.put.return_value = expected_response  # Use PUT

        response = self.box_api.extract_archive(self.box_id, path=path, archive_data=archive_data)

        self.mock_api_client.put.assert_called_once_with(  # Check PUT call
            f"/api/v1/boxes/{self.box_id}/archive",  # Use correct endpoint for PUT archive
            params=expected_params,
            data=archive_data,
            headers=expected_headers,  # Check headers
        )
        self.assertEqual(response, expected_response)

    def test_head_archive(self):
        """Test head archive in box."""
        path = "/check/this/file"
        expected_params = {"path": path}
        expected_headers = {
            "X-Gbox-Path-Stat": '{"size": 1024}',
            "Content-Type": "application/x-tar",
        }
        self.mock_api_client.head.return_value = expected_headers

        response = self.box_api.head_archive(self.box_id, path=path)

        self.mock_api_client.head.assert_called_once_with(
            f"/api/v1/boxes/{self.box_id}/archive", params=expected_params
        )
        self.assertEqual(response, expected_headers)


if __name__ == "__main__":
    unittest.main()
