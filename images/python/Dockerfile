ARG PYTHON_VERSION=3.13

FROM python:${PYTHON_VERSION}-bookworm

# Install dependencies for Chrome and Go
RUN apt update && apt install -y --no-install-recommends \
    tini \
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
    && apt clean && rm -rf /var/lib/apt/lists/* \
    && curl -LsSf https://astral.sh/uv/install.sh | sh

# Install Node.js
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Chromium browser which is available for ARM64
RUN apt update && apt install -y --no-install-recommends \
    chromium \
    && apt clean && rm -rf /var/lib/apt/lists/*

# Install Go
ARG GO_VERSION=1.24.0
ARG TARGETARCH
RUN if [ "$TARGETARCH" = "arm64" ]; then \
        wget -q https://golang.org/dl/go${GO_VERSION}.linux-arm64.tar.gz \
        && tar -C /usr/local -xzf go${GO_VERSION}.linux-arm64.tar.gz \
        && rm go${GO_VERSION}.linux-arm64.tar.gz; \
    else \
        wget -q https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz \
        && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
        && rm go${GO_VERSION}.linux-amd64.tar.gz; \
    fi

# Add Go to the PATH
ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /root

# Set global PATH for uv
ENV PATH="/root/.local/bin:${PATH}"

# Install python packages
RUN uv pip install --system numpy scipy pandas scikit-learn requests beautifulsoup4 pillow matplotlib reportlab

# Create a directory for the Go application
RUN mkdir -p /app
WORKDIR /app

# Create the headless browser Go program
COPY browser_open_url.go ./browser_open_url.go
COPY go.mod ./go.mod
COPY go.sum ./go.sum

RUN go mod tidy
RUN npx playwright install-deps

# Build the Go program
RUN go build -o browser_open_url browser_open_url.go

# Add the Go binary to the PATH
ENV PATH="/app:${PATH}"

# Make the script executable
RUN chmod +x /app/browser_open_url

WORKDIR /root

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["sleep", "infinity"]