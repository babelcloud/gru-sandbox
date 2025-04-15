#!/bin/bash

# ANSI color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create output directory
OUTPUT_DIR="$(pwd)/test/output"
mkdir -p "$OUTPUT_DIR"

# Build test client image
echo -e "Building test client image..."
docker build -t playwright-test:latest -f test/Dockerfile test/

# Start Playwright server
echo -e "Starting Playwright server..."
docker run --rm -d -P --name playwright-server babelcloud/gbox-python:latest

# Get mapped port
echo -e "Getting mapped port..."
PORT=$(docker port playwright-server 3000 | cut -d: -f2)
echo -e "Playwright server is running on port: ${GREEN}${PORT}${NC}"

# Wait for server to be ready
echo -e "Waiting for server to be ready..."
sleep 5

# Run test client
echo -e "Running test client..."
docker run --rm \
    --add-host=host.docker.internal:host-gateway \
    -e PLAYWRIGHT_SERVER_HOST=host.docker.internal \
    -e PLAYWRIGHT_SERVER_PORT=$PORT \
    -v "$OUTPUT_DIR:/app/output" \
    playwright-test:latest

# Check if screenshot was saved
if [ -f "$OUTPUT_DIR/screenshot.png" ]; then
    echo -e "Screenshot saved to: ${GREEN}$OUTPUT_DIR/screenshot.png${NC}"
else
    echo -e "Warning: ${RED}Screenshot was not saved${NC}"
fi

# Cleanup
echo -e "Cleaning up..."
docker stop playwright-server || true 