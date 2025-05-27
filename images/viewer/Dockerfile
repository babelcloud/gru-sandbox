ARG REGISTRY=babelcloud
ARG TAG=latest
FROM ${REGISTRY}/gbox-playwright:${TAG}

# Install VNC components, support packages, and FFmpeg
RUN apt-get update && apt-get install -y --no-install-recommends \
    xvfb \
    x11vnc \
    novnc \
    python3-websockify \
    ffmpeg \
    netcat-openbsd \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Set VNC environment variables
ENV DISPLAY=:1

# Create VNC directory
RUN mkdir -p /root/.vnc && chmod 700 /root/.vnc

# Copy VNC-related supervisord config (merged)
COPY supervisor/*.conf /etc/supervisor/conf.d/

# Copy VNC initialization script
RUN mkdir -p /entrypoint-init.d
COPY 50-vnc-init.sh /entrypoint-init.d/50-vnc-init.sh
RUN chmod +x /entrypoint-init.d/50-vnc-init.sh

# Expose VNC port (5900), noVNC web port (6080), and MJPEG stream port (8090)
EXPOSE 5900 6080 8090

# Healthcheck: Verify supervisor-managed ports are listening
HEALTHCHECK --interval=30s --timeout=5s --start-period=60s --retries=3 \
    CMD nc -z localhost 5900 && nc -z localhost 6080 && nc -z localhost 8090 || exit 1

# Inherit ENTRYPOINT and CMD from base image
# Healthcheck from browser image is also inherited 