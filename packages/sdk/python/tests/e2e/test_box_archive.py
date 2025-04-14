import json
import os
import tarfile
import tempfile
import time

import pytest

from gbox import APIError, Box

# --- Test Constants ---
TEST_IMAGE = "alpine:latest"  # Should match conftest.py

# === Archive Tests ===


# Helper function to ensure box is running
def ensure_box_running(box: Box):
    box.reload()
    if box.status != "running":
        print(f"Box {box.short_id} not running, starting it...")
        box.start()
        # Wait longer for start if necessary, adjust sleep time
        time.sleep(3)  # TODO: Use a proper wait mechanism instead of fixed sleep
        box.reload()
        if box.status != "running":
            pytest.fail(
                f"Box {box.short_id} failed to start for archive tests. Status: {box.status}"
            )
        print(f"Box {box.short_id} is running.")


# Helper function to create a simple text file in the box
def create_test_file_in_box(box: Box, file_path: str, content: str):
    # Ensure parent directory exists
    parent_dir = os.path.dirname(file_path)
    if parent_dir and parent_dir != "/":
        # Use mkdir -p to create parent dirs if they don't exist
        mkdir_cmd = ["mkdir", "-p", parent_dir]
        print(f"Running command in box {box.short_id}: {' '.join(mkdir_cmd)}")
        exit_code, _, stderr = box.run(command=mkdir_cmd)
        # BoxService.run might return -1 for exit code in certain scenarios (e.g., non-blocking exec),
        # especially if the connection closes before the command fully finishes and reports status.
        # Treat 0 (success) and -1 (potentially successful but unknown exit code) as acceptable for setup commands.
        if exit_code not in (0, -1):
            pytest.fail(
                f"Failed to create directory {parent_dir} in box. Exit code: {exit_code}, stderr: {stderr}"
            )

    # Use sh -c to handle potential quotes in content and redirection
    # Escape single quotes in content for the shell command
    escaped_content = content.replace("'", "'''")
    write_cmd = ["sh", "-c", f"echo '{escaped_content}' > {file_path}"]
    print(f"Running command in box {box.short_id}: {' '.join(write_cmd)}")
    exit_code, _, stderr = box.run(command=write_cmd)
    if exit_code not in (0, -1):  # Allow -1 due to FIXME (uncertain exit code propagation)
        pytest.fail(
            f"Failed to create file {file_path} in box. Exit code: {exit_code}, stderr: {stderr}"
        )
    print(f"Successfully created {file_path} in box.")


def test_put_get_head_archive_local_path(test_box: Box):
    """
    Tests putting a local file, heading it, and getting it back to a local path.
    Focuses on the single file upload/download functionality via local paths.
    """
    print(f"Testing archive operations (local path) on box: {test_box.short_id}")
    ensure_box_running(test_box)

    # --- Test Data ---
    upload_dir_in_box = "/tmp/archive_path_tests"
    test_file_content = f"Hello from single file test! Timestamp: {time.time()}"
    # test_filename = "single_test_file.txt" # No longer strictly needed for box path construction
    # full_path_in_box = f"{upload_dir_in_box}/{test_filename}" # Path will be determined dynamically

    # Ensure the base directory exists for uploads
    create_test_file_in_box(
        test_box, f"{upload_dir_in_box}/.keep", ""
    )  # Create dummy file to ensure dir exists

    # --- 1. Test put_archive with Local File Path ---
    print(f"--- 1. Testing put_archive with local file path to {upload_dir_in_box} ---")
    local_file_path = None  # Define outside try/finally for cleanup
    correct_full_path_in_box = None  # Will be set after file creation and upload
    try:
        # Create a temporary local file
        # Use a suggestive prefix/suffix for clarity if needed, but basename matters
        with tempfile.NamedTemporaryFile(
            mode="w", delete=False, suffix="_test_upload.txt", prefix="local_"
        ) as tmp_file:
            tmp_file.write(test_file_content)
            tmp_file.flush()
            os.fsync(tmp_file.fileno())
            local_file_path = tmp_file.name
            # Determine the correct path inside the box based on the local file's basename
            local_filename_basename = os.path.basename(local_file_path)
            correct_full_path_in_box = f"{upload_dir_in_box}/{local_filename_basename}"
        print(f"Created temporary local file: {local_file_path}")
        print(f"Expected path inside box: {correct_full_path_in_box}")

        print(f"Uploading local file {local_file_path} to place inside {upload_dir_in_box}")
        # Pass the local file path directly to 'data'
        test_box.put_archive(path=upload_dir_in_box, data=local_file_path)
        print("put_archive (local path) call successful.")

        # Verify using box.run with the correct path
        print(f"Verifying content of {correct_full_path_in_box} inside the box...")
        exit_code, stdout, stderr = test_box.run(command=["cat", correct_full_path_in_box])
        if exit_code not in (0, -1):
            pytest.fail(
                f"Failed to cat uploaded file {correct_full_path_in_box}. Exit code: {exit_code}, Stdout: {stdout}, Stderr: {stderr}"
            )
        # Strip potential trailing newline from 'cat' output
        assert (
            stdout.strip() == test_file_content
        ), f"Content mismatch. Expected: '{test_file_content}', Got: '{stdout.strip()}'"
        print("File content verified successfully inside the box after put_archive (local path).")

    except FileNotFoundError as e:
        pytest.fail(f"put_archive (local path) failed with FileNotFoundError: {e}")
    except APIError as e:
        pytest.fail(f"put_archive (local path) failed with APIError: {e}")
    finally:
        # Clean up the temporary local file
        if local_file_path and os.path.exists(local_file_path):
            os.remove(local_file_path)
            print(f"Removed temporary local file: {local_file_path}")

    # --- 2. Test head_archive (Using the uploaded file) ---
    # Ensure correct_full_path_in_box was set
    assert correct_full_path_in_box, "Test setup error: correct_full_path_in_box was not set."
    print(f"--- 2. Testing head_archive for {correct_full_path_in_box} ---")
    try:
        headers = test_box.head_archive(path=correct_full_path_in_box)
        print(f"head_archive successful, headers: {headers}")
        assert headers is not None
        stat_header_key = "X-Gbox-Path-Stat"
        assert stat_header_key in headers, f"'{stat_header_key}' header missing."
        stat_json_str = headers[stat_header_key]
        assert stat_json_str, f"'{stat_header_key}' header is empty."
        stat_info = json.loads(stat_json_str)
        # Use the dynamically determined basename for verification
        assert (
            "name" in stat_info and os.path.basename(correct_full_path_in_box) in stat_info["name"]
        )
        assert "size" in stat_info and stat_info["size"] == len(test_file_content.encode("utf-8"))
        assert "mode" in stat_info  # Check mode exists
        # Add more stat checks if needed (e.g., mode type)
    except APIError as e:
        # Check if it's a 'NotFound' scenario disguised as a 500 error temporarily
        if e.status_code == 500 and "no such file or directory" in str(e).lower():
            pytest.fail(
                f"head_archive failed: File {correct_full_path_in_box} reported as not found by API (500 error). Stderr: {e}"
            )
        else:
            pytest.fail(f"head_archive failed with APIError: {e}")

    # Test heading a non-existent file (path remains the same)
    non_existent_path = f"{upload_dir_in_box}/non_existent_file.txt"
    print(f"--- Testing head_archive for non-existent path {non_existent_path} ---")
    with pytest.raises(
        APIError
    ) as excinfo:  # Expect APIError (likely 500) until NotFound is properly returned
        test_box.head_archive(path=non_existent_path)
    # TODO: Change this check to `isinstance(excinfo.value, NotFound)` when API returns 404
    assert (
        excinfo.value.status_code == 500
    ), "Expected APIError 500 for non-existent file (temporary check)"
    print(
        "APIError(500) raised for non-existent file as expected (temporary workaround for lack of 404)."
    )

    # --- 3. Test get_archive Saving Directly to Local Path ---
    # Ensure correct_full_path_in_box was set
    assert correct_full_path_in_box, "Test setup error: correct_full_path_in_box was not set."
    print(f"--- 3. Testing get_archive (direct save) for {correct_full_path_in_box} ---")
    try:
        # Create a temporary directory for download
        with tempfile.TemporaryDirectory() as tmp_dir:
            # Use a fixed local name for the downloaded file for simplicity
            local_download_path = os.path.join(tmp_dir, "downloaded_via_get_archive.txt")
            print(
                f"Attempting to download {correct_full_path_in_box} directly to {local_download_path}"
            )

            # Call get_archive with local_path argument, using the correct remote path
            returned_stream, stat_info = test_box.get_archive(
                path=correct_full_path_in_box, local_path=local_download_path
            )
            print(f"get_archive (direct save) successful, stat_info: {stat_info}")

            # Verify stream is None and file exists locally with correct content
            assert (
                returned_stream is None
            ), "get_archive should return None for stream when local_path is provided"
            assert os.path.exists(
                local_download_path
            ), f"Local file {local_download_path} was not created."

            with open(local_download_path, "r") as f:
                downloaded_content = f.read()
            assert (
                downloaded_content == test_file_content
            ), f"Downloaded file content mismatch. Expected: '{test_file_content}', Got: '{downloaded_content}'"
            print("get_archive (direct save) successful and content verified locally.")

    except FileNotFoundError as e:
        pytest.fail(
            f"get_archive (direct save) failed, file not found in archive or path issue?: {e}"
        )
    except IsADirectoryError as e:
        pytest.fail(f"get_archive (direct save) failed, remote path is a directory?: {e}")
    except tarfile.TarError as e:
        pytest.fail(f"get_archive (direct save) failed with TarError (archive issue?): {e}")
    except APIError as e:
        pytest.fail(f"get_archive (direct save) failed with APIError: {e}")
    except Exception as e:
        pytest.fail(f"An unexpected error occurred during get_archive (direct save): {e}")

    # --- 4. Test Error Cases --- #
    print("--- 4. Testing Error Cases ---")

    # a) put_archive with local directory path
    print("Testing put_archive with local directory...")
    with tempfile.TemporaryDirectory() as tmp_dir_local:
        with pytest.raises(IsADirectoryError):
            test_box.put_archive(path=upload_dir_in_box, data=tmp_dir_local)
        print("Successfully caught IsADirectoryError for put_archive with local dir.")

    # b) get_archive with local_path for a remote directory
    remote_dir_to_get = upload_dir_in_box  # This directory exists and contains at least .keep
    print(f"Testing get_archive with local_path for remote directory {remote_dir_to_get}...")
    with tempfile.TemporaryDirectory() as tmp_dir_local:
        local_dl_path_for_dir = os.path.join(tmp_dir_local, "should_fail_dir_download.txt")
        # This should fail because the remote path is a directory, leading to multiple files
        # in the tar stream, which get_archive with local_path cannot handle.
        with pytest.raises(tarfile.TarError) as excinfo:
            test_box.get_archive(path=remote_dir_to_get, local_path=local_dl_path_for_dir)
        print(
            f"Successfully caught tarfile.TarError for get_archive with local_path on remote dir: {excinfo.value}"
        )
        # Also check that the dummy file wasn't created
        assert not os.path.exists(
            local_dl_path_for_dir
        ), "Target file for directory download should not have been created."

    # c) get_archive with local_path for a non-existent remote file
    # Ensure correct_full_path_in_box was set for context, though we test a different path here
    assert correct_full_path_in_box, "Test setup error: correct_full_path_in_box was not set."
    non_existent_remote_path = f"{upload_dir_in_box}/does_not_exist.txt"
    print(
        f"Testing get_archive with local_path for non-existent remote file {non_existent_remote_path}..."
    )
    with tempfile.TemporaryDirectory() as tmp_dir_local:
        local_dl_path_non_existent = os.path.join(tmp_dir_local, "should_not_exist.txt")
        with pytest.raises(APIError) as excinfo:  # Expect APIError(500) / NotFound
            test_box.get_archive(
                path=non_existent_remote_path, local_path=local_dl_path_non_existent
            )
        # TODO: Change this check to `isinstance(excinfo.value, NotFound)` when API returns 404
        assert (
            excinfo.value.status_code == 500
        ), "Expected APIError 500 for get_archive on non-existent remote file (temporary check)"
        print(
            "Successfully caught APIError 500 for get_archive with local_path on non-existent remote file."
        )
        assert not os.path.exists(
            local_dl_path_non_existent
        ), "Target file for non-existent download should not have been created."

    # --- Cleanup --- #
    # Ensure correct_full_path_in_box was set for context, though we remove the whole dir
    assert correct_full_path_in_box, "Test setup error: correct_full_path_in_box was not set."
    # Optional: Clean up the created file/directory inside the box
    try:
        print(f"Cleaning up test directory {upload_dir_in_box} inside the box...")
        # Use rm -rf to remove the directory and its contents
        exit_code, stdout, stderr = test_box.run(command=["rm", "-rf", upload_dir_in_box])
        if exit_code != 0:
            # Log stderr if removal failed
            print(
                f"Warning: Failed to remove test directory {upload_dir_in_box}. Exit Code: {exit_code}, Stderr: {stderr}"
            )
        else:
            print("Test directory removed successfully.")
    except Exception as e:
        print(f"Warning: Exception during cleanup inside the box: {e}")


def test_copy_method(test_box: Box):
    """
    Tests the unified Box.copy() method for uploads and downloads.
    """
    print(f"Testing copy method on box: {test_box.short_id}")
    ensure_box_running(test_box)

    upload_dir_in_box = "/tmp/copy_method_tests"
    local_file_content = f"Hello from copy test! Timestamp: {time.time()}"

    # Ensure the base directory exists for uploads/downloads inside the box
    create_test_file_in_box(
        test_box, f"{upload_dir_in_box}/.keep", ""
    )  # Create dummy file to ensure dir exists

    local_tmp_file_path = None
    local_tmp_dir_path = None

    try:
        # --- 1. Test Upload (Local -> Box) ---
        print("--- 1. Testing copy (upload: local -> box) ---")
        with tempfile.NamedTemporaryFile(mode="w", delete=False, suffix="_copy_up.txt") as tmp_file:
            tmp_file.write(local_file_content)
            tmp_file.flush()
            os.fsync(tmp_file.fileno())
            local_tmp_file_path = tmp_file.name
            local_filename_basename = os.path.basename(local_tmp_file_path)

        print(f"Created temporary local file: {local_tmp_file_path}")
        target_box_path = f"box:{upload_dir_in_box}"
        expected_file_in_box = f"{upload_dir_in_box}/{local_filename_basename}"
        print(
            f"Uploading {local_tmp_file_path} to {target_box_path} (expected in box: {expected_file_in_box})"
        )

        test_box.copy(source=local_tmp_file_path, target=target_box_path)
        print("copy (upload) call successful.")

        # Verify using box.run
        print(f"Verifying content of {expected_file_in_box} inside the box...")
        exit_code, stdout, stderr = test_box.run(command=["cat", expected_file_in_box])
        if exit_code not in (0, -1):
            pytest.fail(
                f"Failed to cat uploaded file {expected_file_in_box}. Exit code: {exit_code}, Stdout: {stdout}, Stderr: {stderr}"
            )
        assert (
            stdout.strip() == local_file_content
        ), f"Content mismatch. Expected: '{local_file_content}', Got: '{stdout.strip()}'"
        print("File content verified successfully inside the box after copy (upload).")

        # --- 2. Test Download (Box -> Local) ---
        print("--- 2. Testing copy (download: box -> local) ---")
        local_tmp_dir = tempfile.TemporaryDirectory()
        local_tmp_dir_path = local_tmp_dir.name  # Store path for potential cleanup info
        download_target_local_path = os.path.join(local_tmp_dir_path, "downloaded_via_copy.txt")
        source_box_path = f"box:{expected_file_in_box}"

        print(f"Downloading {source_box_path} to {download_target_local_path}")
        test_box.copy(source=source_box_path, target=download_target_local_path)
        print("copy (download) call successful.")

        # Verify local file content
        assert os.path.exists(
            download_target_local_path
        ), f"Local download file {download_target_local_path} not found."
        with open(download_target_local_path, "r") as f:
            downloaded_content = f.read()
        assert (
            downloaded_content == local_file_content
        ), f"Downloaded content mismatch. Expected: '{local_file_content}', Got: '{downloaded_content}'"
        print("Downloaded file content verified locally after copy (download).")

        # Clean up temp dir explicitly here for clarity
        local_tmp_dir.cleanup()
        local_tmp_dir = None
        local_tmp_dir_path = None

        # --- 3. Test Error Cases ---
        print("--- 3. Testing copy error cases ---")

        # a) Local to Local
        with pytest.raises(ValueError, match="Cannot copy between two local paths"):
            print("Testing copy (local -> local)")
            test_box.copy(source=local_tmp_file_path, target="another_local_file.txt")

        # b) Box to Box
        with pytest.raises(ValueError, match="Cannot copy directly between two Box paths"):
            print("Testing copy (box -> box)")
            test_box.copy(source=f"box:{expected_file_in_box}", target="box:/another/path")

        # c) Upload non-existent local file
        with pytest.raises(FileNotFoundError):
            print("Testing copy (upload non-existent local file)")
            test_box.copy(source="/no/such/local/file.xyz", target=target_box_path)

        # d) Download non-existent box file
        with pytest.raises(APIError) as excinfo:
            print("Testing copy (download non-existent box file)")
            test_box.copy(source="box:/no/such/remote/file.abc", target="local_download_fail.txt")
        # TODO: Change this check to `isinstance(excinfo.value, NotFound)` when API returns 404
        assert (
            excinfo.value.status_code == 500
        ), "Expected APIError 500 for non-existent file (temporary check)"

        # e) Upload local directory (should fail)
        with tempfile.TemporaryDirectory() as tmp_dir_for_upload:
            with pytest.raises(IsADirectoryError):
                print("Testing copy (upload local directory)")
                test_box.copy(source=tmp_dir_for_upload, target=target_box_path)

        # f) Download box directory to local file (should fail)
        with tempfile.TemporaryDirectory() as tmp_dir_for_download:
            local_dl_path_for_dir = os.path.join(tmp_dir_for_download, "should_fail_dir.txt")
            # Downloading a directory ('upload_dir_in_box') to a file path ('local_dl_path_for_dir')
            # This fails inside get_archive's tar processing when local_path is set.
            with pytest.raises(tarfile.TarError):
                print("Testing copy (download box directory to local file)")
                test_box.copy(source=f"box:{upload_dir_in_box}", target=local_dl_path_for_dir)

        print("All copy error cases tested successfully.")

    except Exception as e:
        pytest.fail(f"An unexpected error occurred during copy tests: {e}")
    finally:
        # Clean up the temporary local file
        if local_tmp_file_path and os.path.exists(local_tmp_file_path):
            os.remove(local_tmp_file_path)
            print(f"Cleaned up temporary local file: {local_tmp_file_path}")
        # Clean up temporary local directory if it wasn't cleaned up by its context manager (e.g., due to error)
        if local_tmp_dir_path and os.path.exists(local_tmp_dir_path):
            # Be cautious cleaning directories manually if not using TemporaryDirectory context
            # For this test setup, relying on the context manager or TemporaryDirectory object cleanup is safer.
            print(
                f"Note: Temporary directory {local_tmp_dir_path} might need manual cleanup if test failed mid-download."
            )

        # Clean up inside the box
        try:
            print(f"Cleaning up test directory {upload_dir_in_box} inside the box...")
            exit_code, _, stderr = test_box.run(command=["rm", "-rf", upload_dir_in_box])
            if exit_code != 0:
                print(
                    f"Warning: Failed to remove test directory {upload_dir_in_box} in box. Exit Code: {exit_code}, Stderr: {stderr}"
                )
            else:
                print("Test directory in box removed successfully.")
        except Exception as e:
            print(f"Warning: Exception during cleanup inside the box: {e}")


# Add tests for put_archive edge cases (e.g., empty file, permissions, invalid paths) if needed.
# Add tests for get_archive on directories (without local_path).

# Test cases will be added below
