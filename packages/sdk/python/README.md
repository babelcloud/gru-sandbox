# GBox Python SDK

GBox Python SDK provides a simple interface for interacting with the GBox API server.

## Installation

```bash
pip install gbox
```

## Usage Examples

```python
from gbox import GBoxClient

# Initialize client
# The client uses the default base_url ("http://localhost:28080") if not specified.
# You can override it, e.g.: client = GBoxClient(base_url="YOUR_API_ENDPOINT")
# It might also check environment variables depending on implementation details.
client = GBoxClient()

try:
    # Create a new box
    print("Creating a new box...")
    new_box = client.box.create()
    box_id = new_box.id
    print(f"Box created with ID: {box_id}")

    # Execute a simple command
    # Note: Box creation might take a moment. 
    # In real applications, you might need to check box status before executing.
    print(f"Executing 'echo hello' in box {box_id}...")
    # Allow specifying command as a string or list
    result = client.box.exec(box_id=box_id, command="echo hello") 
    # Or: result = client.box.exec(box_id=box_id, command=["echo", "hello"])
    print(f"Command executed. Exit code: {result.exit_code}")
    print(f"Output:\n{result.output}")

    # List boxes to see the new one
    print("Listing all boxes...")
    boxes = client.box.list()
    print(f"Current boxes: {[box.id for box in boxes]}")

    # Clean up (Optional: Uncomment to delete the box after use)
    # print(f"Deleting box {box_id}...")
    # client.box.delete(box_id=box_id)
    # print("Box deleted.")

except Exception as e:
    print(f"An error occurred: {e}")
```

## Development

### Install Development Dependencies

```bash
pip install -e ".[dev]"
```

### Running Tests

```bash
# Run all tests
pytest

# Run with coverage report
pytest --cov=gbox

# Run specific test file
pytest tests/test_box.py

# Run with verbose output
pytest -v
```

### Test Structure

- `tests/conftest.py`: Contains pytest fixtures used across test files
- `tests/test_config.py`: Tests for the configuration module
- `tests/test_client.py`: Tests for the HTTP client
- `tests/test_box.py`: Tests for the Box service
- `tests/test_file.py`: Tests for the File service
- `tests/test_gbox.py`: Tests for the main GBox class 