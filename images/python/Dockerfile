ARG PYTHON_VERSION=3.13
ARG TTYD_VERSION=1.7.7 # Define ttyd version ARG

FROM python:${PYTHON_VERSION}-bookworm

# Install system dependencies in a single layer
RUN apt-get update && apt-get install -y --no-install-recommends \
    # Basic tools
    tini \
    wget \
    gnupg \
    ca-certificates \
    curl \
    git \
    pwgen \
    # Supervisor
    supervisor \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Create directory for supervisord log file
RUN mkdir -p /var/log/supervisor && chmod 755 /var/log/supervisor

# Install uv package manager
RUN curl -LsSf https://astral.sh/uv/install.sh | sh

# Install ttyd using ARG
ARG TTYD_VERSION
RUN curl -L "https://github.com/tsl0922/ttyd/releases/download/${TTYD_VERSION}/ttyd.$(uname -m)" -o /usr/local/bin/ttyd && \
    chmod +x /usr/local/bin/ttyd

WORKDIR /root

# Set base environment variables
ENV PATH="/root/.local/bin:${PATH}"

# Install Python packages
RUN uv pip install --system \
    numpy \
    scipy \
    pandas \
    scikit-learn \
    requests \
    beautifulsoup4 \
    pillow \
    matplotlib \
    reportlab

# Copy base supervisord config (includes ttyd)
COPY supervisord.conf /etc/supervisor/supervisord.conf

# Copy base entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/usr/bin/tini", "--", "/entrypoint.sh"]

# Expose ttyd port
EXPOSE 7681

# Default command
CMD ["sleep", "infinity"]