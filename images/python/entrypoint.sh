#!/bin/bash
set -e

# Run initialization scripts
INIT_DIR="/entrypoint-init.d"
if [ -d "$INIT_DIR" ]; then
  echo "Running initialization scripts in $INIT_DIR..."
  # Use process substitution to feed the loop without a subshell for the loop body
  while IFS= read -r f; do
    case "$f" in
      *.sh)
        # Source the script to affect the current environment
        echo "Sourcing $f";
        . "$f" # Now this affects the main script's environment
        ;;
      *)
        echo "Ignoring $f (not a .sh file)"
        ;;
    esac
  done < <(find "$INIT_DIR/" -follow -type f -print | sort -V)
  # Variables sourced above should now be available
  echo "Finished running initialization scripts."
fi

# Optional: Uncomment to verify if variables are set before starting supervisord
# echo "--- Environment Before Supervisord ---"
# env | grep VNC_
# echo "------------------------------------"

# Start supervisord in the background
echo "Starting supervisord in background..."
/usr/bin/supervisord -c /etc/supervisor/supervisord.conf &

# Execute the command passed into the entrypoint (CMD from Dockerfile or command override)
echo "Executing CMD: $*"
exec "$@" 