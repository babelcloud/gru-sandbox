ARG PYTHON_VERSION=3.13

FROM python:${PYTHON_VERSION}-bookworm

# Install dependencies for Chrome and Go
# Install system dependencies in a single layer
RUN apt-get update && apt-get install -y --no-install-recommends \
    \
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
    \
    wget \
    gnupg \
    ca-certificates \
    curl \
    git \
    # Dependencies required for Chrome
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
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js and npm
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install uv package manager
RUN curl -LsSf https://astral.sh/uv/install.sh | sh

# Install Chromium browser which is available for ARM64
RUN apt update && apt install -y --no-install-recommends \
    chromium \
    && apt clean && rm -rf /var/lib/apt/lists/*

# Install Go
ARG GO_VERSION=1.24.0
RUN wget -q https://golang.org/dl/go${GO_VERSION}.linux-arm64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-arm64.tar.gz \
    && rm go${GO_VERSION}.linux-arm64.tar.gz

# Add Go to the PATH
ENV PATH="/usr/local/go/bin:${PATH}"

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

# Create a directory for the Go application
RUN mkdir -p /app
WORKDIR /app

# Set up Go module and install chromedp
RUN go mod init view_by_url \
    && go get github.com/chromedp/chromedp

# Create the headless browser Go program
COPY view_by_url.go.raw ./view_by_url.go
# Or you can use the next line to embed the program directly in the Dockerfile
# RUN echo 'package main...' > main.go

# Build the Go program
RUN go build -o view_by_url view_by_url.go

# Add the Go binary to the PATH
ENV PATH="/app:${PATH}"

# Make the script executable
RUN chmod +x /app/view_by_url

WORKDIR /root

# Install Playwright and its dependencies
RUN npx playwright install-deps \
    && npx playwright install chromium

ENTRYPOINT ["/usr/bin/tini", "--"]

EXPOSE 3000

CMD ["npx", "playwright", "run-server", "--port", "3000"]