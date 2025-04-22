FROM python:3.13-slim

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Python packages
RUN pip install --no-cache-dir playwright

# Create app directory
WORKDIR /app

# Copy test script
COPY test_playwright.py .

ENTRYPOINT ["python", "test_playwright.py"] 