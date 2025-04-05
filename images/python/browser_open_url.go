package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/playwright-community/playwright-go"
)

// Output represents the JSON output structure
type Output struct {
	URL             string   `json:"url"`
	ElapsedTime     float64  `json:"elapsed_time_seconds"`
	ContentLength   int      `json:"content_length_bytes"`
	OutputFile      string   `json:"output_file"`
	ScreenshotFiles []string `json:"screenshot_files,omitempty"`
}

func main() {
	// Define command line flags
	takeScreenshot := flag.Bool("screenshot", false, "Take a screenshot of the webpage")
	flag.Parse()

	// Check if URL is provided as command line argument
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: browser_open_url [-screenshot] <url>")
		os.Exit(1)
	}

	url := args[0]
	startTime := time.Now()

	log.Printf("Starting browser for URL: %s", url)

	// Install Playwright browsers if not already installed
	if err := playwright.Install(); err != nil {
		log.Fatalf("Could not install Playwright browsers: %v", err)
	}

	// Start Playwright
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("Could not start Playwright: %v", err)
	}
	defer pw.Stop()

	// Launch browser with additional options
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
		Args: []string{
			"--disable-web-security",
			"--disable-features=IsolateOrigins,site-per-process",
			"--disable-site-isolation-trials",
		},
	})
	if err != nil {
		log.Fatalf("Could not launch browser: %v", err)
	}
	defer browser.Close()

	// Create a new context with viewport size and additional options
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  1280,
			Height: 800,
		},
		IgnoreHttpsErrors: playwright.Bool(true),
		UserAgent:         playwright.String("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"),
	})
	if err != nil {
		log.Fatalf("Could not create context: %v", err)
	}
	defer context.Close()

	// Create a new page
	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("Could not create page: %v", err)
	}

	// Set up page event listeners for better debugging
	page.On("console", func(msg playwright.ConsoleMessage) {
		log.Printf("Browser console: %s", msg.Text())
	})
	page.On("pageerror", func(err error) {
		log.Printf("Browser error: %v", err)
	})

	// Navigate to URL with increased timeout and different wait strategy
	log.Println("Navigating to page...")
	if _, err = page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(60000), // 60 seconds
	}); err != nil {
		log.Printf("Initial navigation failed: %v", err)
		log.Println("Trying alternative navigation strategy...")

		// Try alternative navigation strategy
		if _, err = page.Goto(url, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateLoad,
			Timeout:   playwright.Float(60000), // 60 seconds
		}); err != nil {
			log.Fatalf("Could not navigate to page: %v", err)
		}
	}

	// Wait for network to be idle
	log.Println("Waiting for network to be idle...")
	if err = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(5000), // 5 seconds
	}); err != nil {
		log.Printf("Warning: Network did not become idle: %v", err)
	}

	// Get page content
	content, err := page.Content()
	if err != nil {
		log.Fatalf("Could not get page content: %v", err)
	}

	// Get viewport size
	viewport := page.ViewportSize()
	if viewport == nil {
		log.Fatalf("Could not get viewport size")
	}

	// Calculate page height
	pageHeight, err := page.Evaluate(`document.documentElement.scrollHeight`)
	if err != nil {
		log.Fatalf("Could not get page height: %v", err)
	}

	// Convert page height to int
	var pageHeightInt int
	switch v := pageHeight.(type) {
	case float64:
		pageHeightInt = int(v)
	case int:
		pageHeightInt = v
	default:
		log.Fatalf("Unexpected type for page height: %T", pageHeight)
	}

	// Calculate number of screenshots needed
	numScreenshots := (pageHeightInt + viewport.Height - 1) / viewport.Height

	var screenshots []string
	if *takeScreenshot {
		log.Printf("Taking %d screenshots...", numScreenshots)

		// Create output directory
		outputDir := "/var/gbox"
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			log.Fatalf("Error creating output directory: %v", err)
		}

		timestamp := time.Now().Format("20060102_150405")

		// Take screenshots at each scroll position
		for i := 0; i < numScreenshots; i++ {
			scrollY := i * viewport.Height
			log.Printf("Taking screenshot %d/%d at scroll position %d", i+1, numScreenshots, scrollY)

			// Scroll to position
			if _, err = page.Evaluate(fmt.Sprintf("window.scrollTo(0, %d)", scrollY)); err != nil {
				log.Printf("Error scrolling to position %d: %v", scrollY, err)
				continue
			}

			// Wait for any animations to complete
			if err = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
				State:   playwright.LoadStateNetworkidle,
				Timeout: playwright.Float(2000), // 2 seconds
			}); err != nil {
				log.Printf("Warning: Network did not become idle at position %d: %v", scrollY, err)
			}

			// Take screenshot
			screenshotPath := filepath.Join(outputDir, fmt.Sprintf("screenshot_%s_part%d.png", timestamp, i+1))
			if _, err = page.Screenshot(playwright.PageScreenshotOptions{
				Path: playwright.String(screenshotPath),
			}); err != nil {
				log.Printf("Error taking screenshot at position %d: %v", scrollY, err)
				continue
			}

			screenshots = append(screenshots, screenshotPath)
			log.Printf("Successfully took screenshot %d/%d", i+1, numScreenshots)
		}
	}

	elapsedTime := time.Since(startTime)

	// Create the output directory if it doesn't exist
	outputDir := "/var/gbox"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Generate a filename based on the URL and timestamp
	timestamp := time.Now().Format("20060102_150405")
	htmlFilename := filepath.Join(outputDir, fmt.Sprintf("page_%s.html", timestamp))

	// Write the HTML content to file
	if err := os.WriteFile(htmlFilename, []byte(content), 0644); err != nil {
		log.Fatalf("Error writing HTML file: %v", err)
	}

	// Create output structure
	output := Output{
		URL:           url,
		ElapsedTime:   elapsedTime.Seconds(),
		ContentLength: len(content),
		OutputFile:    htmlFilename,
	}

	if *takeScreenshot {
		output.ScreenshotFiles = screenshots
	}

	// Marshal the output to JSON with indentation
	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	// Print the JSON output
	fmt.Println(string(jsonOutput))
	log.Println("Program completed successfully")
}
