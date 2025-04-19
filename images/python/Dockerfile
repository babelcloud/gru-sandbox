ARG PYTHON_VERSION=3.13

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
    # Chrome dependencies
    libnss3 \
    libxss1 \
    libasound2 \
    fonts-liberation \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libatspi2.0-0 \
    libcups2 \
    libdrm2 \
    libgbm1 \
    libgtk-3-0 \
    libxcomposite1 \
    libxdamage1 \
    libxkbcommon0 \
    libxrandr2 \
    xdg-utils \
    # Chromium browser
    chromium \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js and npm
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install uv package manager
RUN curl -LsSf https://astral.sh/uv/install.sh | sh

WORKDIR /root

# Set environment variables
ENV PATH="/root/.local/bin:${PATH}" \
    # Set Chromium as default browser for Playwright
    PLAYWRIGHT_BROWSERS_PATH=/root/.cache/ms-playwright \
    PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1

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

# Install Playwright and its dependencies
RUN npx playwright@1.51.1 install-deps \
    && npx playwright@1.51.1 install chromium

# Copy pm2 config and entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/usr/bin/tini", "--", "/entrypoint.sh"]

EXPOSE 3000

# Add health check to verify Playwright server is running
# Waits 5s initially, checks every 10s, times out after 5s, 3 retries
HEALTHCHECK --interval=10s --timeout=5s --start-period=5s --retries=3 \
    CMD curl --fail http://localhost:3000 || exit 1

CMD ["sleep", "infinity"]