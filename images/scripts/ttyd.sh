#!/bin/bash

# Script to find the ttyd port and open the web terminal.

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
TTYD_CONTAINER_PORT=7681 # Default ttyd port inside container

# Check if container is running
if ! docker container inspect "${CONTAINER_NAME}" > /dev/null 2>&1; then
  echo "${YELLOW}Error: Container ${CONTAINER_NAME} is not found.${RESET}" >&2
  echo "       Use './scripts/start.sh ${IMAGE_SUFFIX}' first." >&2
  exit 1
fi

if [ "$(docker container inspect -f '{{.State.Running}}' "${CONTAINER_NAME}")" = "false" ]; then
 echo "${YELLOW}Error: Container ${CONTAINER_NAME} is not running.${RESET}" >&2
 echo "       Use 'docker start ${CONTAINER_NAME}' or './scripts/start.sh ${IMAGE_SUFFIX}' if it was removed." >&2
 exit 1
fi

echo "Attempting to open ttyd for container ${CYAN}${CONTAINER_NAME}${RESET}..."

# Get port mapping for ttyd
TTYD_PORT_MAPPING=$(docker port "${CONTAINER_NAME}" ${TTYD_CONTAINER_PORT}/tcp 2>/dev/null || true)

if [ -z "${TTYD_PORT_MAPPING}" ]; then
  echo "${YELLOW}Error: Could not get port mapping for ${TTYD_CONTAINER_PORT}/tcp in ${CONTAINER_NAME}.${RESET}" >&2
  echo "       Ensure the container exposed port ${TTYD_CONTAINER_PORT} and was run with -P or -p." >&2
  exit 1
fi

# Extract host port (handle IPv4 and IPv6)
HOST_PORT=$(echo ${TTYD_PORT_MAPPING} | awk -F ':' '{print $NF}')

TTYD_URL="http://localhost:${HOST_PORT}"

echo "  Container Name: ${CYAN}${CONTAINER_NAME}${RESET}"
echo "  ttyd URL:       ${CYAN}${TTYD_URL}${RESET}"

# Open browser (macOS specific)
echo "Launching default web browser..."
if [[ "$(uname)" == "Darwin" ]]; then
  open "${TTYD_URL}"
else
  echo "(Cannot automatically open browser on this OS. Please open the URL manually)"
fi 