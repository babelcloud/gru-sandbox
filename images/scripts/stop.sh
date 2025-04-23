#!/bin/bash

# Script to stop and remove a detached container.

set -e

# ANSI color codes
CYAN=$(tput setaf 6)
RESET=$(tput sgr0)

# Check if image suffix is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <image_suffix>" >&2
  exit 1
fi

IMAGE_SUFFIX="$1"
CONTAINER_PREFIX=${CONTAINER_PREFIX:-gbox-test} # Default if not set via env
CONTAINER_NAME="${CONTAINER_PREFIX}-${IMAGE_SUFFIX}"

# Check if container exists
if ! docker container inspect "${CONTAINER_NAME}" > /dev/null 2>&1; then
  echo "Container ${CYAN}${CONTAINER_NAME}${RESET} not found."
  exit 0
fi

echo "Stopping and removing container ${CYAN}${CONTAINER_NAME}${RESET}..."

# Stop the container (ignore errors if already stopped)
docker stop "${CONTAINER_NAME}" > /dev/null 2>&1 || true

# Remove the container
docker rm "${CONTAINER_NAME}" > /dev/null 2>&1

echo "Container ${CYAN}${CONTAINER_NAME}${RESET} stopped and removed." 