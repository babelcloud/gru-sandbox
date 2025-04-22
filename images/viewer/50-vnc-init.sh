#!/bin/bash
set -e

echo "Initializing VNC..."

# --- Resolution Setup ---
DEFAULT_VNC_RESOLUTION="1280x1024x24"
# Pattern: WxHxD, where D must be 8, 16, 24, or 32
RESOLUTION_PATTERN='^[0-9]+x[0-9]+x(8|16|24|32)$'

if [ -z "$VNC_RESOLUTION" ]; then
    echo "VNC_RESOLUTION not set, using default: $DEFAULT_VNC_RESOLUTION"
    VNC_RESOLUTION="$DEFAULT_VNC_RESOLUTION"
else
    echo "Using VNC_RESOLUTION from environment: $VNC_RESOLUTION"
    # Validate format WxHxD with specific depth
    if [[ ! "$VNC_RESOLUTION" =~ $RESOLUTION_PATTERN ]]; then
        echo >&2 -e "\033[1;33mWarning: Invalid VNC_RESOLUTION format '\$VNC_RESOLUTION'. Should be WxHxD with depth 8, 16, 24, or 32 (e.g., 1280x1024x24). Falling back to default.\033[0m"
        VNC_RESOLUTION="$DEFAULT_VNC_RESOLUTION"
    fi
fi

# Extract WxH for x11vnc geometry flag
VNC_GEOMETRY="${VNC_RESOLUTION%x*}" # Removes the last 'x' and the depth

# Export variables for supervisord programs
export VNC_RESOLUTION
export VNC_GEOMETRY

echo "Using effective VNC resolution (Xvfb): $VNC_RESOLUTION"
echo "Using effective VNC geometry (x11vnc): $VNC_GEOMETRY"
# --- End Resolution Setup ---


# --- Password Setup ---
# ANSI Color Codes
COLOR_BOLD_YELLOW='\033[1;33m'
COLOR_RESET='\033[0m'

PLAINTEXT_PASSWD_PATH=/root/.vnc/plaintext_passwd # Path for readable password
VNC_PASSWORD_TO_USE=""

if [ -n "${VNC_PASSWORD}" ]; then
    echo "Using VNC password from VNC_PASSWORD environment variable."
    VNC_PASSWORD_TO_USE="${VNC_PASSWORD}"
else
    # Generate a random password
    RANDOM_VNC_PASSWORD=$(pwgen -s 8 1)
    VNC_PASSWORD_TO_USE="${RANDOM_VNC_PASSWORD}"
    echo "VNC_PASSWORD environment variable not set."
    # Log the password in bold yellow using variables
    echo -e "Generating random VNC password: ${COLOR_BOLD_YELLOW}${VNC_PASSWORD_TO_USE}${COLOR_RESET}"
fi

# Write the plaintext password to the file
echo "${VNC_PASSWORD_TO_USE}" > "${PLAINTEXT_PASSWD_PATH}"

# Set permissions for the plaintext file only
chmod 600 "${PLAINTEXT_PASSWD_PATH}"

echo "VNC password set in plaintext file."
# --- End Password Setup --- 