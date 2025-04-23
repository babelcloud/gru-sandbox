#!/bin/bash

# Script to start a detached container for a given image suffix.

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
REGISTRY=${REGISTRY:-babelcloud} # Default if not set via env
CONTAINER_PREFIX=${CONTAINER_PREFIX:-gbox-test} # Default if not set via env
CONTAINER_NAME="${CONTAINER_PREFIX}-${IMAGE_SUFFIX}"
IMAGE_NAME="${REGISTRY}/gbox-${IMAGE_SUFFIX}:latest"

# Check if container already exists
if docker container inspect "$CONTAINER_NAME" > /dev/null 2>&1; then
  echo "Container ${CYAN}${CONTAINER_NAME}${RESET} already exists (may be stopped or running)."
  echo "Use './scripts/stop.sh ${IMAGE_SUFFIX}' first if you want to restart it."
  exit 0
fi

# Add GPU flag specifically for 'viewer' image
GPU_FLAG=""
if [ "${IMAGE_SUFFIX}" = "viewer" ]; then
  if docker info --format '{{.Swarm.LocalNodeState}}' 2>/dev/null | grep -q inactive && \
     docker info --format '{{json .Runtimes}}' 2>/dev/null | grep -q nvidia; then
     GPU_FLAG="--gpus all"
     echo "Detected NVIDIA runtime, adding ${GPU_FLAG} for viewer container."
  else
     echo "${YELLOW}Warning: NVIDIA runtime not detected or Swarm mode active. Cannot add --gpus all flag.${RESET}" >&2
     echo "${YELLOW}         NVENC/CUDA features in the viewer may not work.${RESET}" >&2
  fi
fi

echo "Starting container ${CYAN}${CONTAINER_NAME}${RESET} from image ${CYAN}${IMAGE_NAME}${RESET}..."
# Run detached, publish all exposed ports (-P), add GPU flag if set
docker run -d --name "${CONTAINER_NAME}" -P ${GPU_FLAG} "${IMAGE_NAME}"

echo "Container ${CYAN}${CONTAINER_NAME}${RESET} started."
echo "Use './scripts/stop.sh ${IMAGE_SUFFIX}' to stop and remove." 