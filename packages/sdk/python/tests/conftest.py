import os
from typing import Generator  # Import Generator

import pytest

from gbox import APIError, Box, GBoxClient

# API server URL for testing - use environment variable or default
# Make sure your API server is running at this address before starting tests!
GBOX_API_URL = os.environ.get("GBOX_TEST_API_URL", "http://localhost:28081")
TEST_IMAGE = "alpine:latest"  # Image used for testing


@pytest.fixture(scope="session")
def gbox() -> GBoxClient:
    """Fixture to initialize the GBoxClient for the test session."""
    print(f"\nInitializing GBox client for API: {GBOX_API_URL}")
    try:
        gbox_instance = GBoxClient(base_url=GBOX_API_URL, timeout=60)
        # Optional: Ping the server or get version to ensure connectivity early
        gbox_instance.version()
        print("Client initialized and connected successfully.")
        return gbox_instance
    except APIError as e:
        pytest.fail(f"Failed to connect to GBox API at {GBOX_API_URL}: {e}")
    except Exception as e:
        pytest.fail(f"An unexpected error occurred during client initialization: {e}")


@pytest.fixture(scope="function")  # Use function scope for isolation
def test_box(gbox: GBoxClient) -> Generator[Box, None, None]:
    """Fixture to create a Box for a test and ensure cleanup."""
    created_box: Box = None
    print(f"\n--- [Fixture Setup] Creating test box (image: {TEST_IMAGE}) ---")
    try:
        created_box = gbox.boxes.create(
            image=TEST_IMAGE, labels={"creator": "pytest_fixture", "purpose": "e2e_testing"}
        )
        print(f"[Fixture Setup] Box created: {created_box.short_id}")
        yield created_box  # Provide the box to the test function
    except APIError as e:
        pytest.fail(f"Failed to create test box in fixture: {e}")
    except Exception as e:
        pytest.fail(f"Unexpected error during test box creation: {e}")
    finally:
        # --- Cleanup ---
        if created_box:
            print(f"\n--- [Fixture Teardown] Cleaning up test box: {created_box.short_id} ---")
            try:
                # Attempt to stop first if it's running (optional, force delete handles it)
                try:
                    created_box.reload()
                    if created_box.status == "running":
                        print(f"[Fixture Teardown] Stopping box {created_box.short_id}...")
                        created_box.stop()
                except APIError as stop_err:
                    print(
                        f"[Fixture Teardown] Info: Error stopping box (may already be stopped/gone): {stop_err}"
                    )

                print(f"[Fixture Teardown] Deleting box {created_box.short_id}...")
                created_box.delete(force=True)
                print(f"[Fixture Teardown] Box {created_box.short_id} deleted.")
            except APIError as e:
                # Log error but don't fail the test run during cleanup
                print(f"[Fixture Teardown] Error: Failed to delete box {created_box.short_id}: {e}")
            except Exception as e:
                print(f"[Fixture Teardown] Error: Unexpected error during box cleanup: {e}")
        else:
            print("\n--- [Fixture Teardown] No test box was created, skipping cleanup. ---")
