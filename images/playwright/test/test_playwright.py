import os
import time
import sys
import json
import urllib.parse
from playwright.sync_api import sync_playwright, TimeoutError, Playwright, Error as PlaywrightError

# ANSI color codes
RED = '\033[0;31m'
GREEN = '\033[0;32m'
YELLOW = '\033[1;33m'
BLUE = '\033[0;34m'
NC = '\033[0m'  # No Color

def print_info(message):
    print(f"{BLUE}{message}{NC}")

def print_success(message):
    print(f"{GREEN}{message}{NC}")

def print_warning(message):
    print(f"{YELLOW}{message}{NC}")

def print_error(message):
    print(f"{RED}{message}{NC}")

def test_playwright():
    # Get server configuration from environment variables
    host = os.getenv('PLAYWRIGHT_SERVER_HOST', 'host.docker.internal')
    port = os.getenv('PLAYWRIGHT_SERVER_PORT')

    if not port:
        print_error("Error: PLAYWRIGHT_SERVER_PORT environment variable is required")
        sys.exit(1)

    # Define launch options as a dictionary
    launch_options = {"channel": "chromium"}
    # Convert to compact JSON string
    launch_options_json = json.dumps(launch_options, separators=(',', ':'))
    # URL-encode the JSON string
    launch_options_encoded = urllib.parse.quote(launch_options_json)

    # Construct the endpoint with dynamically encoded launch options
    ws_endpoint = f"ws://{host}:{port}?launch-options={launch_options_encoded}"
    print_info(f"Targeting Playwright browser endpoint at {ws_endpoint}")

    browser = None
    page = None

    with sync_playwright() as p:
        try:
            # Connect to the browser endpoint exposed by the Playwright server
            connected = False
            for attempt in range(3):
                try:
                    print_info(f"Attempting to connect to browser via server (attempt {attempt + 1}/3)...")
                    # Use the correct connect method for the server's browser endpoint
                    browser = p.chromium.connect(ws_endpoint, timeout=20000)
                    print_success("Successfully connected to browser via Playwright server.")
                    connected = True
                    break
                except (PlaywrightError, TimeoutError) as e:
                    print_warning(f"Connection attempt {attempt + 1} failed: {str(e)}")
                    if attempt < 2:
                        print_info("Retrying in 5 seconds...")
                        time.sleep(5)
                    else:
                        print_error("Failed to connect after 3 attempts.")
                        raise

            if not connected or not browser:
                 print_error("Failed to establish browser connection via server.")
                 sys.exit(1)

            # Proceed with browser operations
            page = browser.new_page()
            output_dir = "/app/output"
            os.makedirs(output_dir, exist_ok=True)
            
            try:
                print_info("Navigating to page https://gru.ai ...")
                page.goto("https://gru.ai", timeout=30000)
                print_info("Waiting for page to load slightly...")
                page.wait_for_load_state('domcontentloaded', timeout=10000)
                title = page.title()
                content = page.content()
                print_success(f"Page title: {title}")
                print_success(f"Content length: {len(content)} characters")
                if not title or title == "Unknown":
                    print_warning(f"Page title is missing or 'Unknown': '{title}'")
                if len(content) < 100:
                    print_warning("Page content seems unexpectedly short")
                print_info("Taking screenshot...")
                screenshot_path = os.path.join(output_dir, "screenshot.png")
                page.screenshot(path=screenshot_path, full_page=True)
                print_success(f"Screenshot saved to (in container): {screenshot_path}")

            except TimeoutError as nav_error:
                print_error(f"Page operation timed out: {nav_error}")
                try:
                    error_screenshot_path = os.path.join(output_dir, "error_screenshot.png")
                    page.screenshot(path=error_screenshot_path)
                    print_warning(f"Saved error screenshot to: {error_screenshot_path}")
                except Exception as ss_err:
                    print_warning(f"Could not save error screenshot: {ss_err}")
                raise
            except Exception as page_err:
                print_error(f"Error during page operations: {page_err}")
                raise
            finally:
                if page and not page.is_closed():
                   page.close()
                   print_info("Page closed.")

        except Exception as e:
            print_error(f"Error during test execution: {str(e)}")
            sys.exit(1)
        finally:
            if browser and browser.is_connected():
                browser.close()
                print_info("Browser connection closed.")

if __name__ == "__main__":
    test_playwright() 