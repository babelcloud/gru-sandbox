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
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Set VNC environment variables
ENV DISPLAY=:1

# Create VNC directory
RUN mkdir -p /root/.vnc && chmod 700 /root/.vnc

# Copy VNC-related supervisord config (merged)
COPY supervisord.conf /etc/supervisor/conf.d/vnc.conf

# Copy VNC initialization script
RUN mkdir -p /entrypoint-init.d
COPY 50-vnc-init.sh /entrypoint-init.d/50-vnc-init.sh
RUN chmod +x /entrypoint-init.d/50-vnc-init.sh

# Expose VNC port (5900), noVNC web port (6080), and MJPEG stream port (8090)
EXPOSE 5900 6080 8090

# Inherit ENTRYPOINT and CMD from base image
# Healthcheck from browser image is also inherited 