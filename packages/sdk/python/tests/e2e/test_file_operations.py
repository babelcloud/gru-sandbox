import time

import pytest

from gbox import APIError, Box, GBoxClient, NotFound

# --- Test Constants ---
TEST_IMAGE = "alpine:latest"  # Should match conftest.py

# === File Operations Tests ===


# Helper function to ensure box is running
def ensure_box_running(box: Box):
    """Ensure the test box is running."""
    box.reload()
    if box.status != "running":
        print(f"Box {box.short_id} not running, starting it...")
        box.start()
        # Wait longer for start if necessary
        time.sleep(3)  # TODO: Use a proper wait mechanism instead of fixed sleep
        box.reload()
        if box.status != "running":
            pytest.fail(f"Box {box.short_id} failed to start for file tests. Status: {box.status}")
        print(f"Box {box.short_id} is now running.")


# Helper function to create a test file in the box's shared directory
def create_test_file_in_box_share(box: Box, filename: str, content: str):
    """Create a test file in the box's shared directory."""
    # Use the /var/gbox/share path inside the box
    file_path = f"/var/gbox/share/{filename}"

    # Ensure the share directory exists
    share_cmd = ["mkdir", "-p", "/var/gbox/share"]
    print(f"Running command in box {box.short_id}: {' '.join(share_cmd)}")
    exit_code, _, stderr = box.run(command=share_cmd)
    if exit_code not in (0, -1):
        pytest.fail(
            f"Failed to create share directory in box. Exit code: {exit_code}, stderr: {stderr}"
        )

    # Escape single quotes in content for the shell command
    escaped_content = content.replace("'", "'''")
    write_cmd = ["sh", "-c", f"echo '{escaped_content}' > {file_path}"]
    print(f"Running command in box {box.short_id}: {' '.join(write_cmd)}")
    exit_code, _, stderr = box.run(command=write_cmd)
    if exit_code not in (0, -1):
        pytest.fail(
            f"Failed to create file {file_path} in box. Exit code: {exit_code}, stderr: {stderr}"
        )
    print(f"Successfully created {file_path} in box {box.short_id}.")

    # Verify file exists and has correct content
    verify_cmd = ["cat", file_path]
    exit_code, stdout, stderr = box.run(command=verify_cmd)
    if exit_code not in (0, -1):
        pytest.fail(
            f"Failed to verify file {file_path} in box. Exit code: {exit_code}, stderr: {stderr}"
        )
    if stdout.strip() != content:
        pytest.fail(f"File content mismatch. Expected: '{content}', Got: '{stdout.strip()}'")

    return file_path


def test_file_exists_and_get(gbox: GBoxClient):
    """Test checking if files exist and getting file objects using real API."""
    print("Testing file exists and get functionality with real API")

    # Test with a file that should exist in the shared volume
    test_path = "/mydata/output.txt"  # This path should exist in the server

    # Create a unique test file if needed
    f"/mydata/test-file-{int(time.time())}.txt"

    # Check file existence
    exists = gbox.files.exists(test_path)
    print(f"File {test_path} exists: {exists}")

    if exists:
        # Test get method with existing file
        file = gbox.files.get(test_path)
        print(f"Got file object: {file}")

        # Verify file properties
        assert file.path == test_path, f"File path mismatch: {file.path} != {test_path}"
        assert file.name is not None, "File name should not be None"
        assert file.type is not None, "File type should not be None"
        print(f"File properties: name={file.name}, size={file.size}, type={file.type}")
    else:
        print(f"Standard test file {test_path} not found, skipping")

    # Test with a non-existent file - this should never exist
    random_name = f"nonexistent-{int(time.time())}.txt"
    non_existent_path = f"/non/existent/{random_name}"
    exists = gbox.files.exists(non_existent_path)
    assert exists is False, f"File {non_existent_path} should not exist"

    # Try to get non-existent file
    try:
        gbox.files.get(non_existent_path)
        pytest.fail(f"Expected NotFound error for {non_existent_path}")
    except NotFound:
        print("NotFound raised for non-existent file as expected")


def test_file_read(gbox: GBoxClient):
    """Test reading file content from real API."""
    print("Testing file read functionality with real API")

    # Test with a file that should exist in the shared volume
    test_path = "/mydata/output.txt"  # This path should exist in the server

    if not gbox.files.exists(test_path):
        pytest.skip(f"File {test_path} doesn't exist, skipping read test")

    # Get file object
    file = gbox.files.get(test_path)
    print(f"Got file object: {file}")

    # Test read method (binary)
    content_bytes = file.read()
    assert content_bytes is not None, "File content should not be None"
    assert len(content_bytes) > 0, "File content should not be empty"
    print(f"File binary content length: {len(content_bytes)} bytes")

    # Test read_text method
    try:
        content_text = file.read_text()
        assert content_text is not None, "File text content should not be None"
        assert len(content_text) > 0, "File text content should not be empty"
        print(
            f"File text content: {content_text[:100]}" + ("..." if len(content_text) > 100 else "")
        )
    except UnicodeDecodeError:
        print("File content is not UTF-8 encoded text")


def test_file_share_from_box(gbox: GBoxClient, test_box: Box):
    """Test sharing files from a box to the shared volume using real box and API."""
    print(f"Testing sharing files from box {test_box.short_id} with real API")
    ensure_box_running(test_box)

    # Create a test file in the box's shared directory with unique name
    test_filename = f"test-file-{int(time.time())}.txt"
    test_content = f"Test file content for sharing test. Timestamp: {time.time()}"
    file_path_in_box = create_test_file_in_box_share(test_box, test_filename, test_content)
    print(f"Created test file in box: {file_path_in_box}")

    # Share the file using box_id string
    try:
        # Test sharing with box_id
        shared_file = gbox.files.share_from_box(test_box.id, file_path_in_box)
        print(f"Shared file using box_id: {shared_file}")

        # Verify shared file properties
        assert (
            shared_file.name == test_filename
        ), f"Shared file name mismatch: {shared_file.name} != {test_filename}"

        # Read and verify content
        shared_content = shared_file.read()
        shared_text = shared_content.decode("utf-8").strip()
        assert (
            shared_text == test_content
        ), f"Shared file content mismatch: '{shared_text}' != '{test_content}'"
        print("Shared file content verified successfully")

        # Test sharing with Box object
        shared_file2 = gbox.files.share_from_box(test_box, file_path_in_box)
        print(f"Shared file using Box object: {shared_file2}")

        # Verify second shared file
        assert (
            shared_file2.path == shared_file.path
        ), "Paths of files shared with different methods should match"

    except APIError as e:
        pytest.fail(f"Failed to share file from box: {e}")

    # Test sharing with invalid path
    try:
        gbox.files.share_from_box(test_box.id, "/invalid/path.txt")
        pytest.fail("Expected ValueError for invalid path")
    except ValueError as e:
        assert "must start with /var/gbox/" in str(e)
        print("ValueError raised for invalid path as expected")


def test_file_reclaim(gbox: GBoxClient):
    """Test file reclamation using real API."""
    print("Testing file reclamation with real API")

    # Reclaim files
    try:
        response = gbox.files.reclaim()
        print(f"Reclaim response: {response}")
        assert response is not None, "Reclaim response should not be None"
    except APIError as e:
        pytest.fail(f"Failed to reclaim files: {e}")

    # Check that the response has the expected structure
    assert isinstance(response, dict), "Reclaim response should be a dictionary"
    assert "reclaimed_files" in response, "Reclaim response should have 'reclaimed_files' key"
