import uuid

import pytest

from gbox import Box, GBoxClient
from gbox.exceptions import APIError

# --- Test Constants ---
TEST_IMAGE = "alpine:latest"  # Should match conftest.py


@pytest.mark.e2e
def test_list_all_boxes(gbox: GBoxClient):
    """Test listing all boxes, creating some first."""
    print("\n--- Testing List All Boxes ---")
    initial_boxes = gbox.boxes.list()
    initial_count = len(initial_boxes)
    print(f"Initial box count: {initial_count}")

    # Create a couple of boxes
    box1: Box = None
    box2: Box = None
    label_key = "e2e_list_test"
    label_val_1 = "box1"
    label_val_2 = "box2"
    try:
        print("Creating test boxes...")
        box1 = gbox.boxes.create(image=TEST_IMAGE, labels={label_key: label_val_1})
        box2 = gbox.boxes.create(image=TEST_IMAGE, labels={label_key: label_val_2})
        print(f"Created boxes: {box1.short_id}, {box2.short_id}")

        # List again and check count and presence
        print("Listing boxes again...")
        all_boxes = gbox.boxes.list()
        print(f"New box count: {len(all_boxes)}")
        assert len(all_boxes) >= initial_count + 2

        found_box1 = any(b.id == box1.id for b in all_boxes)
        found_box2 = any(b.id == box2.id for b in all_boxes)
        assert found_box1, f"Box {box1.short_id} not found in list"
        assert found_box2, f"Box {box2.short_id} not found in list"

    finally:
        # Cleanup created boxes
        print("Cleaning up list test boxes...")
        if box1:
            try:
                box1.delete(force=True)
                print(f"Deleted {box1.short_id}")
            except APIError as e:
                print(f"Error deleting {box1.short_id}: {e}")
        if box2:
            try:
                box2.delete(force=True)
                print(f"Deleted {box2.short_id}")
            except APIError as e:
                print(f"Error deleting {box2.short_id}: {e}")


@pytest.mark.e2e
def test_list_boxes_with_label_filter(gbox: GBoxClient):
    """Test filtering boxes by a specific label."""
    print("\n--- Testing List Boxes with Label Filter ---")
    # Use unique labels for this test to avoid conflicts
    label_key = "e2e_filter_test_label"
    target_value = f"target_{uuid.uuid4().hex[:6]}"  # Unique value
    other_value = f"other_{uuid.uuid4().hex[:6]}"  # Unique value
    target_label = {label_key: target_value}
    other_label = {label_key: other_value}
    filter_str = f"{label_key}={target_value}"

    target_box: Box = None
    other_box: Box = None
    try:
        print(f"Creating boxes with labels: {target_label}, {other_label}")
        target_box = gbox.boxes.create(image=TEST_IMAGE, labels=target_label)
        other_box = gbox.boxes.create(image=TEST_IMAGE, labels=other_label)
        print(f"Created target box: {target_box.short_id}, other box: {other_box.short_id}")

        # Filter for the target label
        print(f"Filtering boxes with label: '{filter_str}'")
        # Pass the filter string directly to the 'label' key in the filters dict
        filtered_boxes = gbox.boxes.list(filters={"label": filter_str})
        print(f"Found {len(filtered_boxes)} filtered boxes.")

        # Verify only the target box is returned
        assert len(filtered_boxes) == 1, "Expected exactly one box with the target label"
        assert (
            filtered_boxes[0].id == target_box.id
        ), "The filtered box ID does not match the target box ID"
        assert (
            label_key in filtered_boxes[0].labels
        ), f"Label key '{label_key}' not found in filtered box labels"
        assert (
            filtered_boxes[0].labels[label_key] == target_value
        ), "Filtered box label value does not match target value"

    finally:
        # Cleanup
        print("Cleaning up filter test boxes...")
        if target_box:
            try:
                target_box.delete(force=True)
                print(f"Deleted target box {target_box.short_id}")
            except APIError as e:
                print(f"Error deleting target box {target_box.short_id}: {e}")
        if other_box:
            try:
                other_box.delete(force=True)
                print(f"Deleted other box {other_box.short_id}")
            except APIError as e:
                print(f"Error deleting other box {other_box.short_id}: {e}")


# Add more filter tests here if needed (e.g., filtering by name, id, status if supported)
