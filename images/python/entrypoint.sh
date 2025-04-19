#!/bin/bash
set -e

# Start playwright server in the background
echo "Starting Playwright server in background..."
npx playwright@1.51.1 run-server --port 3000 --host 0.0.0.0 &
playwright_pid=$!

echo "Waiting for Playwright server on port 3000 (PID: $playwright_pid)..."
max_wait=30 # Maximum wait time in seconds
waited=0

# Loop until port 3000 is listening or timeout
# Use curl to check. It exits with 0 on success (2xx/3xx), non-zero otherwise.
# --fail makes it exit non-zero for server errors (4xx/5xx).
# --silent prevents output to stdout/stderr.
# --head only fetches headers, faster.
while ! curl --fail --silent --head http://localhost:3000 > /dev/null; do
  # Check if the playwright process is still alive
  if ! kill -0 $playwright_pid 2>/dev/null; then
      echo "Error: Playwright server process (PID: $playwright_pid) exited prematurely." >&2
      # Consider adding code here to show logs if playwright writes any
      exit 1 # Exit container if the server died
  fi

  # Check for timeout
  if [ $waited -ge $max_wait ]; then
    echo "Error: Timeout waiting for Playwright server on port 3000." >&2
    # Optional: Attempt to kill the background process before exiting
    # kill $playwright_pid 2>/dev/null || true 
    exit 1
  fi
  
  sleep 1
  waited=$((waited + 1))
done

# Execute the command passed into the entrypoint (e.g., the CMD or runtime command)
# This becomes the main foreground process managed by tini
exec "$@" 