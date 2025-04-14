# tests/models/test_boxes.py
import io
import unittest
from unittest.mock import MagicMock, Mock, patch

import pytest

# 假设 GBoxClient 和 Box 类在 gbox 包中
# 根据你的项目结构调整导入路径
# TODO: Ensure correct import paths based on actual structure
# This might require adjusting the path or ensuring tests run from the correct root
from gbox.client import GBoxClient  # 需要这个类来模拟
from gbox.exceptions import APIError, NotFound  # Import exceptions
from gbox.models.boxes import Box

# from gbox.api.box_service import BoxService # 如果需要显式模拟 BoxService


class TestBox(unittest.TestCase):

    def setUp(self):
        """Set up for test methods."""
        # 创建 GBoxClient 的模拟实例
        self.mock_client = Mock(spec=GBoxClient)
        # 创建 BoxService 的模拟实例并将其附加到 mock_client
        self.mock_client.box_service = Mock()  # spec=BoxService 如果 BoxService 定义了方法

        # 示例 Box 数据
        self.box_id = "box-12345678-abcd"
        self.box_attrs = {
            "id": self.box_id,
            "name": "test-box",
            "status": "stopped",
            "labels": {"env": "testing"},
            "image": "ubuntu:latest",
        }
        # 创建 Box 实例进行测试
        self.box = Box(self.mock_client, self.box_id, self.box_attrs)

        self.box_simple_id = Box(self.mock_client, "box123", {})

    def test_init(self):
        """Test Box initialization."""
        self.assertEqual(self.box.id, self.box_id)
        self.assertEqual(self.box.attrs, self.box_attrs)
        self.assertEqual(self.box._client, self.mock_client)

    def test_short_id(self):
        """Test the short_id property."""
        self.assertEqual(self.box.short_id, "box-12345678")
        # 测试没有'-'的情况
        self.assertEqual(self.box_simple_id.short_id, "box123")

    def test_short_id_no_hyphen(self):
        """Test short_id property when ID has no hyphen."""
        self.assertEqual(self.box_simple_id.short_id, "box123")

    def test_name(self):
        """Test the name property."""
        self.assertEqual(self.box.name, "test-box")
        box_no_name = Box(self.mock_client, "box-noname", {})
        self.assertIsNone(box_no_name.name)

    def test_status(self):
        """Test the status property."""
        self.assertEqual(self.box.status, "stopped")
        box_no_status = Box(self.mock_client, "box-nostatus", {})
        self.assertIsNone(box_no_status.status)

    def test_labels(self):
        """Test the labels property."""
        self.assertEqual(self.box.labels, {"env": "testing"})
        box_no_labels = Box(self.mock_client, "box-nolabels", {})
        self.assertEqual(box_no_labels.labels, {})

    def test_reload(self):
        """Test the reload method."""
        new_attrs = {"id": self.box_id, "status": "running", "name": "reloaded-box"}
        # 配置 mock_box_service.get 在被调用时返回新的属性
        self.mock_client.box_service.get.return_value = new_attrs

        # 调用 reload
        self.box.reload()

        # 验证 box_service.get 被正确调用
        self.mock_client.box_service.get.assert_called_once_with(self.box_id)
        # 验证 box 的属性已更新
        self.assertEqual(self.box.attrs, new_attrs)
        self.assertEqual(self.box.status, "running")  # 属性也应更新

    def test_reload_api_error(self):
        """Test reload method when API call fails."""
        error = APIError("Failed to get", status_code=500)
        self.mock_client.box_service.get.side_effect = error
        with self.assertRaises(APIError) as cm:
            self.box.reload()
        self.assertIs(cm.exception, error)
        self.mock_client.box_service.get.assert_called_once_with(self.box_id)

    def test_start(self):
        """Test the start method."""
        self.box.start()
        # 验证 box_service.start 被正确调用
        self.mock_client.box_service.start.assert_called_once_with(self.box_id)
        # 注意：我们不在这里测试状态是否真的变为 'running'，
        # 因为这依赖于 reload() 或 API 的实际行为。我们只测试是否调用了正确的服务方法。

    def test_start_api_error(self):
        """Test start method when API call fails."""
        error = APIError("Failed to start", status_code=503)
        self.mock_client.box_service.start.side_effect = error
        with self.assertRaises(APIError) as cm:
            self.box.start()
        self.assertIs(cm.exception, error)
        self.mock_client.box_service.start.assert_called_once_with(self.box_id)

    def test_stop(self):
        """Test the stop method."""
        self.box.stop()
        # 验证 box_service.stop 被正确调用
        self.mock_client.box_service.stop.assert_called_once_with(self.box_id)

    def test_stop_api_error(self):
        """Test stop method when API call fails."""
        error = APIError("Failed to stop", status_code=503)
        self.mock_client.box_service.stop.side_effect = error
        with self.assertRaises(APIError) as cm:
            self.box.stop()
        self.assertIs(cm.exception, error)
        self.mock_client.box_service.stop.assert_called_once_with(self.box_id)

    def test_delete(self):
        """Test the delete method."""
        self.box.delete()
        # 验证 box_service.delete 被正确调用，默认 force=False
        self.mock_client.box_service.delete.assert_called_once_with(self.box_id, force=False)

    def test_delete_api_error(self):
        """Test delete method when API call fails."""
        error = APIError("Failed to delete", status_code=500)
        self.mock_client.box_service.delete.side_effect = error
        with self.assertRaises(APIError) as cm:
            self.box.delete()
        self.assertIs(cm.exception, error)
        self.mock_client.box_service.delete.assert_called_once_with(self.box_id, force=False)

    def test_delete_force(self):
        """Test the delete method with force=True."""
        self.box.delete(force=True)
        # 验证 box_service.delete 被正确调用，force=True
        self.mock_client.box_service.delete.assert_called_once_with(self.box_id, force=True)

    @patch("logging.getLogger")  # 模拟 logging
    def test_run(self, mock_get_logger):
        """Test the run method."""
        # 配置模拟日志记录器
        mock_logger = Mock()
        mock_get_logger.return_value = mock_logger

        command = ["echo", "hello"]
        expected_response = {"exitCode": 0, "stdout": "hello\n", "stderr": ""}
        # 配置 box_service.run 的返回值
        self.mock_client.box_service.run.return_value = expected_response

        exit_code, stdout, stderr = self.box.run(command)

        # 验证 box_service.run 被正确调用
        self.mock_client.box_service.run.assert_called_once_with(self.box_id, command=command)
        # 验证返回值是否正确
        self.assertEqual(exit_code, 0)
        self.assertEqual(stdout, "hello\n")
        self.assertEqual(stderr, "")
        # 验证日志记录是否被调用（可选）
        mock_logger.debug.assert_called()

    @patch("logging.getLogger")
    def test_run_api_error(self, mock_get_logger):
        """Test run method when API call fails."""
        command = ["echo", "fail"]
        error = APIError("Failed to run command", status_code=500)
        self.mock_client.box_service.run.side_effect = error
        with self.assertRaises(APIError) as cm:
            self.box.run(command)
        self.assertIs(cm.exception, error)
        self.mock_client.box_service.run.assert_called_once_with(self.box_id, command=command)

    def test_reclaim(self):
        """Test the reclaim method."""
        expected_response = {"message": "reclaimed", "resources": {}}
        self.mock_client.box_service.reclaim.return_value = expected_response

        response = self.box.reclaim()

        self.mock_client.box_service.reclaim.assert_called_once_with(
            box_id=self.box_id, force=False
        )
        self.assertEqual(response, expected_response)

    def test_reclaim_api_error(self):
        """Test reclaim method when API call fails."""
        error = APIError("Failed to reclaim", status_code=500)
        self.mock_client.box_service.reclaim.side_effect = error
        with self.assertRaises(APIError) as cm:
            self.box.reclaim()
        self.assertIs(cm.exception, error)
        self.mock_client.box_service.reclaim.assert_called_once_with(
            box_id=self.box_id, force=False
        )

    def test_reclaim_force(self):
        """Test the reclaim method with force=True."""
        expected_response = {"message": "force reclaimed", "resources": {}}
        self.mock_client.box_service.reclaim.return_value = expected_response

        response = self.box.reclaim(force=True)

        self.mock_client.box_service.reclaim.assert_called_once_with(box_id=self.box_id, force=True)
        self.assertEqual(response, expected_response)

    def test_head_archive(self):
        """Test the head_archive method."""
        path = "/data/file.txt"
        expected_headers = {"Content-Length": "1024", "X-GBox-Mode": "0644"}
        self.mock_client.box_service.head_archive.return_value = expected_headers

        headers = self.box.head_archive(path)

        self.mock_client.box_service.head_archive.assert_called_once_with(self.box_id, path=path)
        self.assertEqual(headers, expected_headers)

    def test_head_archive_api_error(self):
        """Test head_archive method when API call fails with a generic error."""
        path = "/data/file.txt"
        error = APIError("HEAD failed", status_code=500)
        self.mock_client.box_service.head_archive.side_effect = error
        with self.assertRaises(APIError) as cm:
            self.box.head_archive(path)
        self.assertIs(cm.exception, error)
        self.mock_client.box_service.head_archive.assert_called_once_with(self.box_id, path=path)

    def test_head_archive_not_found(self):
        """Test head_archive method when the path is not found."""
        path = "/data/not_found.txt"
        # Assume NotFound is raised directly, or adjust if it's an APIError(404)
        error = NotFound("Path not found")
        self.mock_client.box_service.head_archive.side_effect = error
        with self.assertRaises(NotFound) as cm:
            self.box.head_archive(path)
        self.assertIs(cm.exception, error)
        self.mock_client.box_service.head_archive.assert_called_once_with(self.box_id, path=path)

    def test_get_archive(self):
        """Test the get_archive method."""
        path = "/data"
        mock_stats = {"Content-Type": "application/x-tar", "X-GBox-Size": "2048"}
        mock_tar_data = b"tar data bytes"

        self.mock_client.box_service.head_archive.return_value = mock_stats
        self.mock_client.box_service.get_archive.return_value = mock_tar_data

        tar_stream, stats = self.box.get_archive(path)

        self.mock_client.box_service.head_archive.assert_called_once_with(self.box_id, path=path)
        self.mock_client.box_service.get_archive.assert_called_once_with(self.box_id, path=path)

        self.assertIsInstance(tar_stream, io.BytesIO)
        self.assertEqual(tar_stream.read(), mock_tar_data)
        self.assertEqual(stats, mock_stats)

    def test_get_archive_head_fails(self):
        """Test get_archive when the initial head_archive call fails."""
        path = "/data"
        error = APIError("HEAD failed first", status_code=500)
        self.mock_client.box_service.head_archive.side_effect = error

        with self.assertRaises(APIError) as cm:
            self.box.get_archive(path)

        self.assertIs(cm.exception, error)
        self.mock_client.box_service.head_archive.assert_called_once_with(self.box_id, path=path)
        # Ensure get_archive itself was not called
        self.mock_client.box_service.get_archive.assert_not_called()

    def test_get_archive_get_fails(self):
        """Test get_archive when the get_archive API call fails."""
        path = "/data"
        mock_stats = {"Content-Type": "application/x-tar"}
        error = APIError("GET archive failed", status_code=500)

        self.mock_client.box_service.head_archive.return_value = mock_stats
        self.mock_client.box_service.get_archive.side_effect = error

        with self.assertRaises(APIError) as cm:
            self.box.get_archive(path)

        self.assertIs(cm.exception, error)
        self.mock_client.box_service.head_archive.assert_called_once_with(self.box_id, path=path)
        self.mock_client.box_service.get_archive.assert_called_once_with(self.box_id, path=path)

    def test_put_archive_bytes(self):
        """Test put_archive with bytes data."""
        path = "/uploads"
        tar_data = b"some tar data"

        self.box.put_archive(path, tar_data)

        # 验证 extract_archive 被正确调用
        self.mock_client.box_service.extract_archive.assert_called_once_with(
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
        self.mock_client.box_service.extract_archive.assert_called_once_with(
            self.box_id, path=path, archive_data=tar_data
        )

    def test_put_archive_invalid_type(self):
        """Test put_archive raises TypeError for invalid data types."""
        with pytest.raises(TypeError):
            self.box.put_archive("/data", data=12345)  # Invalid type

        # Test case for string that is NOT a valid path (should raise FileNotFoundError)
        with pytest.raises(FileNotFoundError):
            self.box.put_archive("/data", data="this is a string, not a path")

    def test_put_archive_directory_path(self):
        # This test case is not provided in the original file or the code block
        # It's assumed to exist based on the test_put_archive_invalid_type method
        # If this test case is needed, it should be implemented here
        pass

    def test_put_archive_api_error(self):
        """Test put_archive method when API call fails."""
        path = "/uploads"
        tar_data = b"tar data"
        error = APIError("Extract failed", status_code=500)
        self.mock_client.box_service.extract_archive.side_effect = error

        with self.assertRaises(APIError) as cm:
            self.box.put_archive(path, tar_data)

        self.assertIs(cm.exception, error)
        self.mock_client.box_service.extract_archive.assert_called_once_with(
            self.box_id, path=path, archive_data=tar_data
        )

    def test_eq(self):
        """Test equality comparison."""
        box1 = Box(self.mock_client, "box-same-id", {})
        box2 = Box(self.mock_client, "box-same-id", {"status": "running"})
        box3 = Box(self.mock_client, "box-different-id", {})
        not_a_box = "box-same-id"

        self.assertEqual(box1, box2)  # 只比较 ID
        self.assertNotEqual(box1, box3)
        self.assertNotEqual(box1, not_a_box)

    def test_hash(self):
        """Test hashing."""
        box1 = Box(self.mock_client, "box-hash-id", {})
        box2 = Box(self.mock_client, "box-hash-id", {"status": "running"})
        self.assertEqual(hash(box1), hash(box2))  # Hash 基于 ID
        self.assertEqual(hash(box1), hash("box-hash-id"))  # 确认 hash 逻辑

    def test_repr(self):
        """Test the representation string."""
        self.assertEqual(repr(self.box), "<Box: box-12345678>")


if __name__ == "__main__":
    unittest.main()
