# gbox/client.py
from typing import Any, Dict, Optional

from .api.box_service import BoxService as ApiBoxService
from .api.client import Client as ApiClient  # Low-level HTTP client
from .config import GBoxConfig
from .exceptions import APIError  # Import base APIError
from .managers.boxes import BoxManager

# Optional: Function to initialize from environment or common configs
# def from_env(): ...


class GBoxClient:
    """
    The main entry point for interacting with the GBox API.

    Provides access to resource managers (e.g., `boxes`).
    """

    def __init__(
        self,
        base_url: str = "http://localhost:28080",
        config: Optional[GBoxConfig] = None,
        timeout: int = 60,
    ):
        """
        Initializes the GBoxClient.

        Args:
            base_url: The base URL of the GBox API server. Defaults to "http://localhost:28080".
            config: Optional GBox configuration object. If None, a default one is created.
            timeout: Default timeout for API requests in seconds. Defaults to 60.
        """
        self.config = config or GBoxConfig()  # Use provided or default config
        # Ensure logger from config is passed to the API client
        self.api_client = ApiClient(base_url=base_url, timeout=timeout, logger=self.config.logger)

        # Initialize low-level services
        self.box_service = ApiBoxService(client=self.api_client, config=self.config)
        # Add other services here (e.g., self.file_service) if you have them

        # Initialize high-level managers
        self.boxes = BoxManager(client=self)
        # Add other managers here (e.g., self.files)

    def version(self) -> Dict[str, Any]:
        """
        Gets the GBox server version information.
        (Assuming a GET /version or similar endpoint exists at the root/API root)
        """
        # Assumes the low-level client has a method for this, or add a dedicated method.
        # This example assumes a simple GET request.
        try:
            # 使用正确的API路径
            return self.api_client.get("/api/v1/version")
        except Exception as e:
            # Catch potential exceptions from api_client.get and wrap them
            # Check if it's already an APIError to avoid double wrapping
            if isinstance(e, APIError):
                raise
            raise APIError(f"Failed to get server version: {e}") from e

    # Potentially add other top-level convenience methods if needed
