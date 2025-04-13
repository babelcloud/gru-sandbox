# gbox/exceptions.py

"""Base exception for gbox."""


class GBoxError(Exception):
    """Base exception for gboxsdk."""

    pass


class APIError(GBoxError):
    """Indicates an error returned by the GBox API."""

    def __init__(self, message, status_code=None, explanation=None):
        super().__init__(message)
        self.status_code = status_code
        self.explanation = explanation  # Potentially more details from API response

    def __str__(self):
        msg = super().__str__()
        if self.status_code:
            msg = f"{self.status_code} Server Error: {msg}"
        if self.explanation:
            msg = f"{msg} ({self.explanation})"
        return msg


class NotFound(APIError):
    """Resource could not be found (404)."""

    pass


class ConflictError(APIError):
    """Request conflicts with the current state (409)."""

    pass


# ... other specific errors based on HTTP status or API error codes
