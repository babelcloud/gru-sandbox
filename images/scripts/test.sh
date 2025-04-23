#!/bin/bash

# Script to run tests for a specific image against its running container.

set -e

# ANSI color codes
CYAN=$(tput setaf 6)
RESET=$(tput sgr0)
YELLOW=$(tput setaf 3)

# Check if image suffix is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <image_suffix>" >&2
  exit 1
fi

IMAGE_SUFFIX="$1"
CONTAINER_PREFIX=${CONTAINER_PREFIX:-gbox-test} # Default if not set via env
CONTAINER_NAME="${CONTAINER_PREFIX}-${IMAGE_SUFFIX}"
TEST_SCRIPT_PATH="${IMAGE_SUFFIX}/test.sh"

# Check if container is running (should be due to Makefile dependency)
if ! docker container inspect "${CONTAINER_NAME}" > /dev/null 2>&1 || \
   [ "$(docker container inspect -f '{{.State.Running}}' "${CONTAINER_NAME}")" = "false" ]; then
 echo "${YELLOW}Error: Container ${CONTAINER_NAME} is not running or not found.${RESET}" >&2
 echo "       Ensure it was started correctly (e.g., 'make start-${IMAGE_SUFFIX}')." >&2
 exit 1
fi

# Check if test script exists
if [ ! -f "${TEST_SCRIPT_PATH}" ]; then
  echo "${YELLOW}No test script found for ${IMAGE_SUFFIX} at ${TEST_SCRIPT_PATH}${RESET}"
  # Optional: Add more checks here if needed, like listing dir contents
  exit 0
fi

echo "${CYAN}Running tests for ${IMAGE_SUFFIX} against container ${CONTAINER_NAME}...${RESET}"

# Change to the image directory and execute the test script
# Pass the container name as an argument to the test script
cd "${IMAGE_SUFFIX}" && chmod +x test.sh && ./test.sh "${CONTAINER_NAME}"

echo "${CYAN}Tests for ${IMAGE_SUFFIX} completed.${RESET}" 