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

echo "--- Connection Info for ${CONTAINER_NAME} ---"

# Get port mappings (suppress errors if port not found)
NOVNC_PORT_MAPPING=$(docker port "${CONTAINER_NAME}" 6080/tcp 2>/dev/null || true)
STREAM_PORT_MAPPING=$(docker port "${CONTAINER_NAME}" 8090/tcp 2>/dev/null || true)

# Check if BOTH mappings failed (unlikely if container check passed, but good practice)
if [ -z "${NOVNC_PORT_MAPPING}" ] && [ -z "${STREAM_PORT_MAPPING}" ]; then
  echo "Error: Could not get port mapping for either noVNC (6080) or stream (8090) in ${CONTAINER_NAME}." >&2
  echo "       Ensure the container exposed ports and was run with -P or -p." >&2
  exit 1
fi

# Get VNC password
VNC_PASS=$(docker exec "${CONTAINER_NAME}" cat /root/.vnc/plaintext_passwd 2>/dev/null || true)

# Print basic info
echo "  Container Name:  ${CONTAINER_NAME}"

# --- Wait for container to be healthy ---
echo "--- Waiting for Container ---"
WAIT_TIMEOUT=120 # Seconds
WAIT_INTERVAL=5  # Seconds
ELAPSED=0
while [ $ELAPSED -lt $WAIT_TIMEOUT ]; do
  HEALTH_STATUS=$(docker inspect --format='{{.State.Health.Status}}' "${CONTAINER_NAME}" 2>/dev/null || echo "no_healthcheck")

  if [ "$HEALTH_STATUS" == "healthy" ]; then
    echo "Container is healthy."
    break
  elif [ "$HEALTH_STATUS" == "unhealthy" ]; then
     echo "Error: Container is unhealthy. Check container logs." >&2
     docker logs "${CONTAINER_NAME}" >&2 # Show logs for debugging
     exit 1
  elif [ "$HEALTH_STATUS" == "no_healthcheck" ]; then
     echo "${YELLOW}Warning: Container does not have a health check. Proceeding immediately.${RESET}"
     break # Proceed if no healthcheck is configured
  else # starting or unknown
      echo "  Status: $HEALTH_STATUS (waiting... ${ELAPSED}s / ${WAIT_TIMEOUT}s)"
      sleep $WAIT_INTERVAL
      ELAPSED=$((ELAPSED + WAIT_INTERVAL))
  fi

  if [ $ELAPSED -ge $WAIT_TIMEOUT ]; then
      echo "Error: Timeout waiting for container ${CONTAINER_NAME} to become healthy." >&2
      exit 1
  fi
done
# --- Container is ready (or proceeded without health check) ---


# --- VNC Access ---
echo "--- VNC Access ---"
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
  if [ -n "${VNC_PASS}" ]; then
    echo "  VNC Password:    ${YELLOW}${VNC_PASS}${RESET}"
  fi

  # Launch noVNC browser on host FIRST
  echo "Launching default web browser for noVNC on host..."
  if [[ "$(uname)" == "Darwin" ]]; then
    open "${NOVNC_URL}"
  else
    echo "(Cannot automatically open browser on this OS. Please open the URL manually)"
  fi
else
  echo "  noVNC URL:       ${YELLOW}Not available (port 6080 not mapped?)${RESET}"
fi


# --- Stream Access ---
echo "--- Stream Access ---"
if [ -n "${STREAM_PORT_MAPPING}" ]; then
  STREAM_HOST_PORT=$(echo ${STREAM_PORT_MAPPING} | sed 's/.*://')
  STREAM_URL="tcp://localhost:${STREAM_HOST_PORT}"
  echo "  Stream URL:      ${CYAN}${STREAM_URL}${RESET}"
  echo "  Stream Player:   ${CYAN}ffplay ${STREAM_URL}${RESET}"

  # Attempt to launch ffplay automatically
  echo "Attempting to launch ffplay to play the stream..."
  if command -v ffplay &> /dev/null; then
    # Launch ffplay directly with the TCP URL, quiet loglevel
    # Run in foreground - user can close it when done
    ffplay -loglevel error ${STREAM_URL} &
  else
    echo "${YELLOW}Warning: 'ffplay' command not found. Cannot automatically start player.${RESET}" >&2
    echo "${YELLOW}         Please install ffmpeg (which includes ffplay), then run manually:${RESET}" >&2
    # Correct manual command example (show command in cyan)
    echo "         ${CYAN}ffplay ${STREAM_URL}${RESET}" >&2
  fi

else
  echo "  Stream URL:      ${YELLOW}Not available (port 8090 not mapped?)${RESET}"
fi

# Launch Chromium last, only if VNC is available
if [ -n "${NOVNC_URL}" ]; then
  # Check if Chromium is already running inside the container
  echo "Checking for existing Chromium process inside ${CONTAINER_NAME}..."
  # Use pgrep -f to match the command line, redirect stderr to hide "no process found"
  if ! docker exec "${CONTAINER_NAME}" pgrep -f chromium > /dev/null 2>&1; then
    echo "Launching Chromium inside container ${CONTAINER_NAME}..."
    docker exec -d "${CONTAINER_NAME}" env DISPLAY=:1 chromium --no-sandbox --disable-gpu > /dev/null 2>&1 || \
      echo "${YELLOW}Warning: Failed to launch Chromium inside container. Is chromium installed and DISPLAY=:1 available? Is pgrep installed?${RESET}" >&2
  else
    echo "Chromium appears to be already running inside the container."
  fi
fi