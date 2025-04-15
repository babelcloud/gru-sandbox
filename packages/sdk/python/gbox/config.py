"""
GBox SDK Configuration Module
"""

from typing import Any, Optional, Protocol


class Logger(Protocol):
    """Logger interface protocol"""

    def debug(self, msg: str, *args: Any, **kwargs: Any) -> None:
        """Log debug level message"""
        ...

    def info(self, msg: str, *args: Any, **kwargs: Any) -> None:
        """Log info level message"""
        ...

    def warning(self, msg: str, *args: Any, **kwargs: Any) -> None:
        """Log warning level message"""
        ...

    def error(self, msg: str, *args: Any, **kwargs: Any) -> None:
        """Log error level message"""
        ...

    def critical(self, msg: str, *args: Any, **kwargs: Any) -> None:
        """Log critical error level message"""
        ...


class GBoxConfig:
    """
    GBox SDK configuration class, used to store configuration information for SDK client
    """

    def __init__(
        self,
        api_url: str = "http://localhost:28080",
        logger: Optional[Logger] = None,
        debug_mode: bool = False,
    ):
        """
        Initialize configuration object

        Args:
            api_url: URL address of the API server
            logger: Optional logger object
            debug_mode: Whether to enable debug mode
        """
        self.api_url = (api_url or "http://localhost:28080").rstrip("/")
        self.logger = logger
        self.debug_mode = debug_mode
