"""
GBox Low-Level API Client Module
"""

import json
import logging  # Import logging
from typing import Any, Dict, Optional, Tuple, Union
from urllib.parse import urljoin

import requests

# Import custom exceptions from the parent directory
from ..exceptions import APIError, ConflictError, NotFound


class Client:
    """
    Low-level HTTP client for communicating with the GBox API server.
    Handles making requests and basic error handling based on status codes.
    """

    def __init__(self, base_url: str, timeout: int = 60, logger: Optional[logging.Logger] = None):
        """
        Initialize client.

        Args:
            base_url: Base URL of the GBox API server.
            timeout: Default request timeout in seconds.
            logger: Optional logger instance.
        """
        if not base_url:
            raise ValueError("base_url cannot be empty")
        self.base_url = base_url
        self.timeout = timeout
        self.logger = logger or logging.getLogger(__name__)  # Use provided logger or default
        self.session = requests.Session()
        self.session.headers.update(
            {
                "Content-Type": "application/json",
                "Accept": "application/json",
            }
        )

    def _log(self, level: str, message: str, *args: Any, **kwargs: Any) -> None:
        """
        Log message using the configured logger.
        """
        if self.logger:
            log_method = getattr(
                self.logger, level, self.logger.debug
            )  # Default to debug if level invalid
            try:
                log_method(message, *args, **kwargs)
            except Exception as e:
                # Fallback logging in case of issues with the primary logger
                print(f"LOGGING FAILED [{level.upper()}]: {message} - Error: {e}")

    def _raise_for_status(self, response: requests.Response) -> None:
        """
        Raises appropriate GBoxError based on response status code.
        """
        try:
            error_data = response.json()
            message = error_data.get("message", response.reason)
            explanation = error_data.get("explanation")  # Check for explanation field
        except json.JSONDecodeError:
            message = response.text or response.reason  # Use text if JSON fails
            explanation = None
        except Exception as e:  # Catch other potential parsing errors
            message = f"Failed to parse error response: {e}"
            explanation = response.text

        status_code = response.status_code

        if status_code == 404:
            self._log(
                "warning", f"Request failed (NotFound): {status_code} {message} {explanation or ''}"
            )
            raise NotFound(message, status_code=status_code, explanation=explanation)
        elif status_code == 409:
            self._log(
                "warning", f"Request failed (Conflict): {status_code} {message} {explanation or ''}"
            )
            raise ConflictError(message, status_code=status_code, explanation=explanation)
        elif 400 <= status_code < 500:
            # General client error
            self._log(
                "warning",
                f"Request failed (Client Error): {status_code} {message} {explanation or ''}",
            )
            raise APIError(message, status_code=status_code, explanation=explanation)
        elif 500 <= status_code < 600:
            # General server error
            self._log(
                "warning",
                f"Request failed (Server Error): {status_code} {message} {explanation or ''}",
            )
            raise APIError(message, status_code=status_code, explanation=explanation)
        # If no exception is raised, the status code is considered OK.

    def request(
        self,
        method: str,
        path: str,
        params: Optional[Dict[str, Any]] = None,
        data: Optional[Any] = None,
        headers: Optional[Dict[str, str]] = None,
        timeout: Optional[Union[float, Tuple[float, float]]] = None,
        raw_response: bool = False,
    ) -> Any:
        """
        Send HTTP request to API server.

        Args:
            method: HTTP method (GET, POST, PUT, etc.)
            path: API path (relative to base_url)
            params: URL query parameters
            data: Request body data
            headers: Request headers
            timeout: Request timeout (overrides default if provided)
            raw_response: If True, return the raw response content instead of parsing JSON

        Returns:
            Parsed JSON response data, raw response content (if raw_response=True),
            or None for 204 status.

        Raises:
            APIError: For 4xx/5xx errors from the server.
            NotFound: For 404 errors.
            ConflictError: For 409 errors.
            GBoxError: For connection errors or other request issues.
        """
        url = urljoin(self.base_url, path.lstrip("/"))  # Ensure path doesn't start with /
        request_headers = self.session.headers.copy()  # Start with session defaults
        if headers:
            request_headers.update(headers)

        # Determine if data is binary based on Content-Type header
        content_type = request_headers.get("Content-Type", "application/json").lower()
        is_binary = "application/json" not in content_type

        request_data = data
        if data is not None and not is_binary:
            try:
                request_data = json.dumps(data)
            except TypeError as e:
                raise TypeError(f"Failed to serialize data to JSON: {e}. Data: {data}") from e

        # Use provided timeout or the client's default
        current_timeout = timeout if timeout is not None else self.timeout

        self._log("debug", f"Request: {method} {url}")
        if params:
            self._log("debug", f"  Query Params: {params}")
        # Avoid logging potentially large binary data
        if request_data and not is_binary:
            # Log potentially sensitive data carefully
            log_data = str(request_data)
            if len(log_data) > 500:  # Truncate long data
                log_data = log_data[:500] + "..."
            self._log("debug", f"  Body: {log_data}")
        elif is_binary and data is not None:
            self._log("debug", "  Body: <binary data>")
        if request_headers != self.session.headers:  # Log only if different from default
            self._log("debug", f"  Headers: {request_headers}")

        try:
            response = self.session.request(
                method=method,
                url=url,
                params=params,
                data=request_data,
                headers=request_headers,
                timeout=current_timeout,
            )

            self._log(
                "info" if response.ok else "warning",
                f"Response: {response.status_code} {response.reason} ({response.url})",
            )

            # Raise custom exceptions for bad status codes
            self._raise_for_status(response)

            # Process successful response
            if raw_response:
                return response.content

            if response.status_code == 204 or not response.content:
                return None  # No content

            # Try to parse as JSON, return raw bytes if it fails (might be tar, etc.)
            try:
                return response.json()
            except json.JSONDecodeError:
                self._log("debug", "Response is not JSON, returning raw content.")
                return response.content  # Return raw bytes if not JSON

        except requests.exceptions.Timeout as e:
            self._log("error", f"Request timed out: {method} {url} after {current_timeout}s")
            raise APIError(f"Request timed out after {current_timeout}s", status_code=408) from e
        except requests.exceptions.ConnectionError as e:
            self._log("error", f"Connection error: {method} {url} - {e}")
            raise APIError(f"Connection error to {self.base_url}", status_code=503) from e
        except requests.RequestException as e:
            # Catch other requests exceptions (e.g., TooManyRedirects)
            self._log("error", f"Request failed: {method} {url} - {e}")
            # Use a generic status code or leave it None if unclear
            raise APIError(
                f"Request failed: {e}", status_code=getattr(e.response, "status_code", None)
            ) from e
        # Fix: Catch specific known SDK errors first to prevent re-wrapping
        except (NotFound, ConflictError, APIError) as e:
            # Let specific SDK errors raised by _raise_for_status pass through
            raise e
        except Exception as e:
            # Catch truly unexpected errors during request processing
            self._log(
                "error", f"Unexpected error during request: {method} {url} - {e}", exc_info=True
            )
            raise APIError(f"An unexpected error occurred: {e}") from e

    def get(self, path: str, **kwargs: Any) -> Any:
        """Send GET request"""
        return self.request("GET", path, **kwargs)

    def post(self, path: str, **kwargs: Any) -> Any:
        """Send POST request"""
        return self.request("POST", path, **kwargs)

    def put(self, path: str, **kwargs: Any) -> Any:
        """Send PUT request"""
        return self.request("PUT", path, **kwargs)

    def delete(self, path: str, **kwargs: Any) -> Any:
        """Send DELETE request"""
        return self.request("DELETE", path, **kwargs)

    def head(self, path: str, **kwargs: Any) -> Dict[str, Any]:
        """
        Send HEAD request. Returns response headers.

        Raises:
            APIError, NotFound, etc. if the HEAD request fails.
        """
        url = urljoin(self.base_url, path.lstrip("/"))
        current_timeout = kwargs.get("timeout", self.timeout)
        headers = kwargs.get("headers")
        params = kwargs.get("params")

        self._log("debug", f"Request: HEAD {url}")
        if params:
            self._log("debug", f"  Query Params: {params}")
        if headers:
            self._log("debug", f"  Headers: {headers}")

        try:
            response = self.session.head(
                url=url,
                params=params,
                headers=headers,
                timeout=current_timeout,
                allow_redirects=True,  # Typically allow redirects for HEAD
            )

            self._log(
                "debug" if response.ok else "warning",
                f"Response: {response.status_code} {response.reason} ({response.url})",
            )

            # Raise exceptions for bad status codes
            self._raise_for_status(response)

            # Return headers as a case-insensitive dict-like object
            return response.headers

        except requests.exceptions.Timeout as e:
            self._log("error", f"HEAD Request timed out: HEAD {url} after {current_timeout}s")
            raise APIError(
                f"HEAD Request timed out after {current_timeout}s", status_code=408
            ) from e
        except requests.exceptions.ConnectionError as e:
            self._log("error", f"Connection error: HEAD {url} - {e}")
            raise APIError(f"Connection error to {self.base_url}", status_code=503) from e
        except requests.RequestException as e:
            self._log("error", f"HEAD Request failed: HEAD {url} - {e}")
            raise APIError(
                f"HEAD Request failed: {e}", status_code=getattr(e.response, "status_code", None)
            ) from e
        # Fix: Catch specific known SDK errors first to prevent re-wrapping
        except (NotFound, ConflictError, APIError) as e:
            # Let specific SDK errors raised by _raise_for_status pass through
            raise e
        except Exception as e:
            # Catch truly unexpected errors during request processing
            # Attempt to get status code only for *unexpected* errors
            status_code = getattr(getattr(e, "response", None), "status_code", None)
            self._log(
                "error", f"Unexpected error during HEAD request: HEAD {url} - {e}", exc_info=True
            )
            # Pass status_code if found
            raise APIError(
                f"An unexpected error occurred during HEAD request: {e}", status_code=status_code
            ) from e
