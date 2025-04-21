#!/bin/bash
set -e

# ANSI Color Codes
COLOR_BOLD_YELLOW='\033[1;33m'
COLOR_RESET='\033[0m'

# VNC Password Setup
PLAINTEXT_PASSWD_PATH=/root/.vnc/plaintext_passwd # Path for readable password
VNC_PASSWORD_TO_USE=""

if [ -n "${VNC_PASSWORD}" ]; then
    echo "Using VNC password from VNC_PASSWORD environment variable."
    VNC_PASSWORD_TO_USE="${VNC_PASSWORD}"
    # Write the plaintext password to the file
    echo "${VNC_PASSWORD_TO_USE}" > "${PLAINTEXT_PASSWD_PATH}"
else
    # Generate a random password
    RANDOM_VNC_PASSWORD=$(pwgen -s 8 1)
    VNC_PASSWORD_TO_USE="${RANDOM_VNC_PASSWORD}"
    echo "VNC_PASSWORD environment variable not set."
    # Log the password in bold yellow using variables
    echo -e "Generating random VNC password: ${COLOR_BOLD_YELLOW}${VNC_PASSWORD_TO_USE}${COLOR_RESET}"
    # Write the plaintext password to the file
    echo "${VNC_PASSWORD_TO_USE}" > "${PLAINTEXT_PASSWD_PATH}"
fi

# Use the determined password with vncpasswd - NO LONGER NEEDED
# echo "${VNC_PASSWORD_TO_USE}" | vncpasswd -f > "${VNC_PASSWD_PATH}"

# Set permissions for the plaintext file only
# chmod 600 "${VNC_PASSWD_PATH}" # Removed
chmod 600 "${PLAINTEXT_PASSWD_PATH}"

echo "VNC password set in plaintext file."

# Start supervisord in the background
# Logs will go to stdout/stderr based on supervisord.conf
echo "Starting supervisord in background..."
/usr/bin/supervisord -c /etc/supervisor/conf.d/supervisord.conf &

# Execute the command passed into the entrypoint (CMD from Dockerfile or command override)
echo "Executing CMD: $*"
exec "$@" 