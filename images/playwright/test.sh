#!/bin/bash

# ANSI color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get target container name from argument
TARGET_CONTAINER=$1
if [ -z "$TARGET_CONTAINER" ]; then
  echo -e "${RED}Error: Target container name not provided as argument.${NC}"
  exit 1
fi

# Check if target container is running
if [ -z "$(docker ps -q -f name=^${TARGET_CONTAINER}$)" ]; then
    echo -e "${RED}Error: Target container '${TARGET_CONTAINER}' is not running. Start it first (e.g., make start-playwright).${NC}"
    exit 1
fi
echo -e "Target Playwright server container: ${BLUE}${TARGET_CONTAINER}${NC}"

# Create output directory relative to script location
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
OUTPUT_DIR="$SCRIPT_DIR/test/output"
mkdir -p "$OUTPUT_DIR"

# Build test client image
echo -e "Building test client image..."
docker build -t playwright-test:latest -f test/Dockerfile test/

# Get mapped port from the existing target container
echo -e "Getting mapped port from ${TARGET_CONTAINER}..."
PORT=""
RETRY_COUNT=0
MAX_RETRIES=5
while [ -z "$PORT" ] && [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
  sleep 1
  PORT=$(docker port "$TARGET_CONTAINER" 3000/tcp | cut -d: -f2)
  RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ -z "$PORT" ]; then
  echo -e "${RED}Error: Could not get port mapping for 3000/tcp from container '${TARGET_CONTAINER}' after $MAX_RETRIES attempts.${NC}"
  docker logs "$TARGET_CONTAINER"
  exit 1
fi
echo -e "Playwright server is running on port: ${GREEN}${PORT}${NC}"

# Wait for server to be ready (Healthcheck should ensure this, but keep a small delay)
echo -e "Waiting for server to be ready (briefly)..."
sleep 2

# Run test client
echo -e "Running test client against ${TARGET_CONTAINER} on port ${PORT}..."
docker run --rm \
    --add-host=host.docker.internal:host-gateway \
    -e PLAYWRIGHT_SERVER_HOST=host.docker.internal \
    -e PLAYWRIGHT_SERVER_PORT=$PORT \
    -v "$OUTPUT_DIR:/app/output" \
    playwright-test:latest
TEST_EXIT_CODE=$?

# Check if screenshot was saved
if [ -f "$OUTPUT_DIR/screenshot.png" ]; then
    echo -e "Screenshot saved to: ${GREEN}$OUTPUT_DIR/screenshot.png${NC}"
else
    echo -e "Warning: ${YELLOW}Screenshot was not saved.${NC}"
fi

echo -e "Test finished with exit code: ${TEST_EXIT_CODE}"
exit $TEST_EXIT_CODE 