import os
import time
import sys
from playwright.sync_api import sync_playwright, TimeoutError

# ANSI color codes
RED = '\033[0;31m'
GREEN = '\033[0;32m'
YELLOW = '\033[1;33m'
BLUE = '\033[0;34m'
NC = '\033[0m'  # No Color

def print_info(message):
    print(f"{message}{BLUE}{NC}")

def print_success(message):
    print(f"{message}{GREEN}{NC}")

def print_warning(message):
    print(f"{message}{YELLOW}{NC}")

def print_error(message):
    print(f"{message}{RED}{NC}")

def connect_with_retry(playwright, host, port, max_retries=3, retry_delay=5):
    """Try to connect to the Playwright server with retries"""
    for attempt in range(max_retries):
        try:
            print(f"Attempting to connect to Playwright server (attempt {attempt + 1}/{max_retries})...{BLUE}{NC}")
            browser = playwright.chromium.connect(f"ws://{host}:{port}")
            print(f"Successfully connected to Playwright server")
            return browser
        except Exception as e:
            if attempt < max_retries - 1:
                print(f"Connection failed: {YELLOW}{str(e)}{NC}")
                print(f"{BLUE}Retrying in {retry_delay} seconds...{NC}")
                time.sleep(retry_delay)
            else:
                print(f"{RED}Failed to connect after {max_retries} attempts{NC}")
                raise

def test_playwright():
    # Get server configuration from environment variables
    host = os.getenv('PLAYWRIGHT_SERVER_HOST', 'host.docker.internal')
    port = os.getenv('PLAYWRIGHT_SERVER_PORT')
    if not port:
        print(f"Error: {RED}PLAYWRIGHT_SERVER_PORT environment variable is required{NC}")
        sys.exit(1)
    
    print(f"Connecting to Playwright server at {BLUE}ws://{host}:{port}{NC}")
    
    # Connect to Playwright server
    with sync_playwright() as p:
        try:
            # Connect to the remote browser with retry logic
            browser = connect_with_retry(p, host, port)
            page = browser.new_page()
            
            try:
                # Navigate to the page
                print(f"Navigating to page...")
                page.goto("https://gru.ai", timeout=30000)  # 30 seconds timeout
                
                # Wait for page to load
                print(f"Waiting for page to load...")
                time.sleep(2)  # Give the page time to load completely
                
                # Get page title and content
                title = page.title()
                content = page.content()
                
                print(f"Page title: {GREEN}{title}{NC}")
                print(f"Content length: {GREEN}{len(content)}{NC} characters")
                
                if title == "Unknown":
                    print(f"Warning: {YELLOW}Page title is 'Unknown'{NC}")
                
                if len(content) < 100:
                    print(f"Warning: Page content seems too short{NC}")
                
                # Take a screenshot
                print(f"Taking screenshot...{NC}")
                screenshot_path = "/app/output/screenshot.png"
                page.screenshot(path=screenshot_path, full_page=True)
                print(f"Screenshot saved to(in container): {GREEN}{screenshot_path}{NC}")
                
            except TimeoutError:
                print(f"Error: {RED}Page navigation timed out{NC}")
                raise
            except Exception as e:
                print(f"Error during page operations: {RED}{str(e)}{NC}")
                raise
            finally:
                browser.close()
                
        except Exception as e:
            print(f"Error during test: {RED}{str(e)}{NC}")
            raise

if __name__ == "__main__":
    test_playwright() 