#!/bin/bash

# Script to get viewer info (noVNC, MJPEG stream) for a running container
# and optionally open the noVNC URL.

set -e # Exit on first error

# ANSI color codes
CYAN=$(tput setaf 6)
RESET=$(tput sgr0)
YELLOW=$(tput setaf 3)

# Check if container name is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <container_name>" >&2
  exit 1
fi

CONTAINER_NAME="$1"

# Check if container is running
if ! docker container inspect "${CONTAINER_NAME}" > /dev/null 2>&1; then
  echo "Error: Container ${CONTAINER_NAME} is not running." >&2
  echo "       Use 'make start-viewer' (or similar) first." >&2
  exit 1
fi

echo "Getting connection info for container ${CONTAINER_NAME}..."

# Get port mappings (suppress errors if port not found)
NOVNC_PORT_MAPPING=$(docker port "${CONTAINER_NAME}" 6080/tcp 2>/dev/null || true)
MJPEG_PORT_MAPPING=$(docker port "${CONTAINER_NAME}" 8090/tcp 2>/dev/null || true)

# Check if BOTH mappings failed
if [ -z "${NOVNC_PORT_MAPPING}" ] && [ -z "${MJPEG_PORT_MAPPING}" ]; then
  echo "Error: Could not get port mapping for either noVNC (6080) or MJPEG stream (8090) in ${CONTAINER_NAME}." >&2
  echo "       Ensure the container exposed ports and was run with -P or -p." >&2
  exit 1
fi

# Get VNC password
VNC_PASS=$(docker exec "${CONTAINER_NAME}" cat /root/.vnc/plaintext_passwd 2>/dev/null || true)

# Print info
echo "  Container Name:  ${CONTAINER_NAME}"

# Handle noVNC URL
NOVNC_URL=""
if [ -n "${NOVNC_PORT_MAPPING}" ]; then
  NOVNC_HOST_PORT=$(echo ${NOVNC_PORT_MAPPING} | sed 's/.*://')
  NOVNC_URL="http://localhost:${NOVNC_HOST_PORT}/vnc.html"
  SEP='?'
  if [ -n "${VNC_PASS}" ]; then
    NOVNC_URL="${NOVNC_URL}${SEP}password=${VNC_PASS}"
    SEP='&'
  fi
  NOVNC_URL="${NOVNC_URL}${SEP}autoconnect=true"
  SEP='&'
  NOVNC_URL="${NOVNC_URL}${SEP}reconnect=true"
  echo "  noVNC URL:       ${CYAN}${NOVNC_URL}${RESET}"
else
  echo "  noVNC URL:       ${YELLOW}Not available (port 6080 not mapped?)${RESET}"
fi

# Handle MJPEG URL
if [ -n "${MJPEG_PORT_MAPPING}" ]; then
  MJPEG_HOST_PORT=$(echo ${MJPEG_PORT_MAPPING} | sed 's/.*://')
  MJPEG_URL="http://localhost:${MJPEG_HOST_PORT}"
  echo "  MJPEG Stream URL:${CYAN}${MJPEG_URL}${RESET}"
else
  echo "  MJPEG Stream URL:${YELLOW}Not available (port 8090 not mapped?)${RESET}"
fi

# Only open noVNC if it's available
if [ -n "${NOVNC_URL}" ]; then
  # Launch Chromium inside the container (does not need the full URL)
  echo "Launching Chromium inside container ${CONTAINER_NAME}..."
  docker exec -d "${CONTAINER_NAME}" env DISPLAY=:1 chromium --no-sandbox > /dev/null 2>&1 || \
    echo "${YELLOW}Warning: Failed to launch Chromium inside container. Is chromium installed and DISPLAY=:1 available?${RESET}" >&2

  echo "Launching default web browser for noVNC on host..."
  # Use appropriate command for the OS (macOS example)
  if [[ "$(uname)" == "Darwin" ]]; then
    open "${NOVNC_URL}"
  else
    echo "(Cannot automatically open browser on this OS. Please open the URL manually)"
  fi
fi 