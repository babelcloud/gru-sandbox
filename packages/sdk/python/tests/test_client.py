# tests/test_client.py
import unittest
from unittest.mock import Mock, patch

from gbox.api.box_service import BoxService as ApiBoxService
from gbox.api.client import Client as ApiClient
from gbox.client import GBoxClient

# Modules to patch
from gbox.config import GBoxConfig
from gbox.exceptions import APIError
from gbox.managers.boxes import BoxManager


class TestGBoxClient(unittest.TestCase):

    # Use patch to replace classes with Mocks during tests
    @patch("gbox.client.BoxManager")
    @patch("gbox.client.ApiBoxService")
    @patch("gbox.client.ApiClient")
    @patch("gbox.client.GBoxConfig")
    def test_init(self, MockGBoxConfig, MockApiClient, MockApiBoxService, MockBoxManager):
        """Test GBoxClient initialization."""
        # Arrange
        mock_config_instance = Mock(spec=GBoxConfig)
        mock_config_instance.logger = Mock()  # Mock the logger attribute
        MockGBoxConfig.return_value = mock_config_instance

        mock_api_client_instance = Mock(spec=ApiClient)
        MockApiClient.return_value = mock_api_client_instance

        mock_api_box_service_instance = Mock(spec=ApiBoxService)
        MockApiBoxService.return_value = mock_api_box_service_instance

        mock_box_manager_instance = Mock(spec=BoxManager)
        MockBoxManager.return_value = mock_box_manager_instance

        base_url = "http://test.gbox.local:8080"
        timeout = 90

        # Act
        client = GBoxClient(base_url=base_url, timeout=timeout)

        # Assert
        # Verify GBoxConfig was called (or used default)
        MockGBoxConfig.assert_called_once()
        # Verify ApiClient was instantiated correctly
        MockApiClient.assert_called_once_with(
            base_url=base_url, timeout=timeout, logger=mock_config_instance.logger
        )
        self.assertEqual(client.api_client, mock_api_client_instance)

        # Verify ApiBoxService was instantiated correctly
        MockApiBoxService.assert_called_once_with(
            client=mock_api_client_instance, config=mock_config_instance
        )
        self.assertEqual(client.box_service, mock_api_box_service_instance)

        # Verify BoxManager was instantiated correctly
        MockBoxManager.assert_called_once_with(client=client)
        self.assertEqual(client.boxes, mock_box_manager_instance)

    @patch("gbox.client.BoxManager")
    @patch("gbox.client.ApiBoxService")
    @patch("gbox.client.ApiClient")
    @patch("gbox.client.GBoxConfig")
    def test_init_with_config(
        self, MockGBoxConfig, MockApiClient, MockApiBoxService, MockBoxManager
    ):
        """Test GBoxClient initialization with a provided config."""
        # Arrange
        provided_config = Mock(spec=GBoxConfig)
        provided_config.logger = Mock()

        mock_api_client_instance = Mock(spec=ApiClient)
        MockApiClient.return_value = mock_api_client_instance

        # Act: Pass the pre-made config
        client = GBoxClient(base_url="http://another.url", config=provided_config)

        # Assert
        # Verify GBoxConfig was NOT called (because one was provided)
        MockGBoxConfig.assert_not_called()
        self.assertEqual(client.config, provided_config)
        # Verify ApiClient used the logger from the provided config
        MockApiClient.assert_called_once_with(
            base_url="http://another.url",
            timeout=60,  # Default timeout when not specified
            logger=provided_config.logger,
        )
        # Verify ApiBoxService used the provided config
        MockApiBoxService.assert_called_once_with(
            client=mock_api_client_instance, config=provided_config
        )
        # Verify BoxManager was instantiated correctly
        MockBoxManager.assert_called_once_with(client=client)

    # Patch dependencies for the version method tests
    @patch("gbox.client.BoxManager")
    @patch("gbox.client.ApiBoxService")
    @patch("gbox.client.ApiClient")
    @patch("gbox.client.GBoxConfig")
    def test_version_success(
        self, MockGBoxConfig, MockApiClient, MockApiBoxService, MockBoxManager
    ):
        """Test the version() method successfully retrieving version info."""
        # Arrange
        mock_api_client_instance = Mock(spec=ApiClient)
        MockApiClient.return_value = mock_api_client_instance
        expected_version_info = {"Version": "1.2.3", "ApiVersion": "1.0"}
        mock_api_client_instance.get.return_value = expected_version_info

        client = GBoxClient(base_url="http://test.version")

        # Act
        version_info = client.version()

        # Assert
        mock_api_client_instance.get.assert_called_once_with("/api/v1/version")
        self.assertEqual(version_info, expected_version_info)

    @patch("gbox.client.BoxManager")
    @patch("gbox.client.ApiBoxService")
    @patch("gbox.client.ApiClient")
    @patch("gbox.client.GBoxConfig")
    def test_version_api_error(
        self, MockGBoxConfig, MockApiClient, MockApiBoxService, MockBoxManager
    ):
        """Test the version() method when the API client raises APIError."""
        # Arrange
        mock_api_client_instance = Mock(spec=ApiClient)
        MockApiClient.return_value = mock_api_client_instance
        original_error = APIError("Server error", status_code=500)
        mock_api_client_instance.get.side_effect = original_error

        client = GBoxClient(base_url="http://test.error")

        # Act & Assert
        with self.assertRaises(APIError) as cm:
            client.version()

        # Check that the *original* APIError was re-raised
        self.assertIs(cm.exception, original_error)
        mock_api_client_instance.get.assert_called_once_with("/api/v1/version")

    @patch("gbox.client.BoxManager")
    @patch("gbox.client.ApiBoxService")
    @patch("gbox.client.ApiClient")
    @patch("gbox.client.GBoxConfig")
    def test_version_other_error(
        self, MockGBoxConfig, MockApiClient, MockApiBoxService, MockBoxManager
    ):
        """Test the version() method when the API client raises a non-APIError."""
        # Arrange
        mock_api_client_instance = Mock(spec=ApiClient)
        MockApiClient.return_value = mock_api_client_instance
        original_exception = ValueError("Unexpected response format")
        mock_api_client_instance.get.side_effect = original_exception

        client = GBoxClient(base_url="http://test.other.error")

        # Act & Assert
        with self.assertRaises(APIError) as cm:
            client.version()

        # Check that a new APIError was raised, wrapping the original exception
        self.assertIsNot(cm.exception, original_exception)
        self.assertIn("Failed to get server version", str(cm.exception))
        self.assertIs(cm.exception.__cause__, original_exception)
        mock_api_client_instance.get.assert_called_once_with("/api/v1/version")


if __name__ == "__main__":
    unittest.main()
