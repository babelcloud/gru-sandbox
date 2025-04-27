package service_test

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/babelcloud/gbox/packages/api-server/config"
	model "github.com/babelcloud/gbox/packages/api-server/pkg/browser"
)

// --- Mock BoxService --- (Moved to test_helpers.go)
/*
type mockBoxService struct {
	dynamicPort int // Store the dynamic port for GetExternalPort
}

func (m *mockBoxService) Create(ctx context.Context, params *boxModel.BoxCreateParams) (*boxModel.Box, error) {
// ... methods ...
}
var _ boxSvc.BoxService = (*mockBoxService)(nil)
*/

// ------------------------

// FindFreePort finds an available TCP port and returns it (Moved to test_helpers.go)
/*
func FindFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
// ... function body ...
	return l.Addr().(*net.TCPAddr).Port, nil
}
*/

// ------------------------

// Helper function to get the file URL for the test page (Moved to test_helpers.go)
// Note: This depends on runtime.Caller(0) and must be called from test_helpers.go
/*
func getTestPageURL(t *testing.T) string {
// ... function body ...
}
*/

// setupPlaywrightPage sets up playwright and navigates to the vision test page directly.
// Returns the page and a cleanup function. Used for tests directly interacting with Playwright API.
func setupPlaywrightPage(t *testing.T) (playwright.Page, func()) {
	t.Helper()

	// Run playwright locally
	pw, err := playwright.Run()
	require.NoError(t, err, "could not start playwright run task (ensure drivers are installed locally: npx playwright install)")

	// Launch browser locally
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false), // Run non-headless for debugging
	})
	require.NoError(t, err, "could not launch local browser")

	// Create context and page
	context, err := browser.NewContext()
	require.NoError(t, err, "could not create browser context")
	page, err := context.NewPage()
	require.NoError(t, err, "could not create page")

	// Navigate to the local test HTML file
	testURL := getTestPageURL(t)
	fmt.Printf("Navigating test page to %s...\n", testURL)
	_, err = page.Goto(testURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad, // Wait for page to load
		Timeout:   playwright.Float(10000),      // 10 second timeout for local file
	})
	require.NoError(t, err, "could not navigate to test page %s", testURL)
	fmt.Printf("Navigation to %s complete.\n", testURL)

	// Cleanup closes local playwright resources
	cleanup := func() {
		require.NoError(t, page.Close(), "failed to close page")
		require.NoError(t, context.Close(), "failed to close context")
		require.NoError(t, browser.Close(), "failed to close browser")
		require.NoError(t, pw.Stop(), "failed to stop playwright run task")
	}

	return page, cleanup
}

// setupServiceWithVisionTestPage sets up a BrowserService connected to a local Playwright server
// and navigates a page within that service to the vision-test.html file.
// (Moved to test_helpers.go)
/*
func setupServiceWithVisionTestPage(t *testing.T) (*service.BrowserService, string, string, string, playwright.Page, func()) {
	t.Helper()
// ... function body ...
	return browserService, testBoxID, testContextID, testPageID, pageInstance, cleanup
}
*/

// TestVisionScreenshotOptions focuses on testing the mapping of VisionScreenshotParams
// to playwright.PageScreenshotOptions and calling Playwright directly using the
// local vision-test.html page.
// NOTE: This does NOT test the full service logic (URL/Base64 generation, file saving in service).
func TestVisionScreenshotOptions(t *testing.T) {
	page, cleanup := setupPlaywrightPage(t)
	defer cleanup()

	cfg := config.GetInstance()
	testBoxID := "test-box-options"
	expectedBoxDefaultDir := filepath.Join(cfg.File.Share, testBoxID, "screenshot")
	err := os.MkdirAll(expectedBoxDefaultDir, 0755)
	require.NoError(t, err, "Failed to create default screenshot directory for testing: %s", expectedBoxDefaultDir)

	testCases := []struct {
		name           string
		params         model.VisionScreenshotParams // Use the updated params struct (no Path)
		expectError    bool
		outputPath     *string                                               // Explicit path for Playwright options, can be nil for buffer
		validateOutput func(t *testing.T, outputPath *string, buffer []byte) // Validation function
	}{
		{
			name:        "Default PNG to Buffer",
			params:      model.VisionScreenshotParams{}, // Empty params, OutputFormat defaults to base64 in service, but here means buffer
			expectError: false,
			outputPath:  nil, // Capture buffer
			validateOutput: func(t *testing.T, outputPath *string, buffer []byte) {
				assert.Nil(t, outputPath, "Output path should be nil for buffer capture")
				assert.NotEmpty(t, buffer, "Buffer should not be empty")
				// Basic PNG check (first few bytes)
				assert.True(t, len(buffer) > 8 && string(buffer[:8]) == "\x89PNG\r\n\x1a\n", "Buffer should start with PNG signature")
				// Can optionally decode base64 if comparing with service output
				encoded := base64.StdEncoding.EncodeToString(buffer)
				assert.NotEmpty(t, encoded)
			},
		},
		{
			name: "JPEG with Quality to Buffer",
			params: model.VisionScreenshotParams{
				Type:    playwright.String("jpeg"),
				Quality: playwright.Int(80),
			},
			expectError: false,
			outputPath:  nil, // Capture buffer
			validateOutput: func(t *testing.T, outputPath *string, buffer []byte) {
				assert.Nil(t, outputPath, "Output path should be nil for buffer capture")
				assert.NotEmpty(t, buffer, "Buffer should not be empty")
				// Basic JPEG check (first few bytes)
				assert.True(t, len(buffer) > 2 && string(buffer[:2]) == "\xff\xd8", "Buffer should start with JPEG SOI marker")
			},
		},
		{
			name:   "Save PNG to Specific Path",
			params: model.VisionScreenshotParams{
				// Type defaults to PNG
			},
			expectError: false,
			outputPath:  func() *string { p := filepath.Join(t.TempDir(), "test_save.png"); return &p }(),
			validateOutput: func(t *testing.T, outputPath *string, buffer []byte) {
				require.NotNil(t, outputPath, "Output path should be provided")
				_, err := os.Stat(*outputPath)
				assert.NoError(t, err, "File should exist at the specified path: %s", *outputPath)
				assert.Equal(t, ".png", filepath.Ext(*outputPath), "File extension should be .png")
			},
		},
		{
			name: "Save JPEG to Specific Path",
			params: model.VisionScreenshotParams{
				Type:    playwright.String("jpeg"),
				Quality: playwright.Int(90),
			},
			expectError: false,
			outputPath:  func() *string { p := filepath.Join(t.TempDir(), "test_save.jpeg"); return &p }(),
			validateOutput: func(t *testing.T, outputPath *string, buffer []byte) {
				require.NotNil(t, outputPath, "Output path should be provided")
				_, err := os.Stat(*outputPath)
				assert.NoError(t, err, "File should exist at the specified path: %s", *outputPath)
				assert.Equal(t, ".jpeg", filepath.Ext(*outputPath), "File extension should be .jpeg")
			},
		},
		{
			name: "Full Page PNG to Buffer",
			params: model.VisionScreenshotParams{
				FullPage: playwright.Bool(true),
			},
			expectError: false,
			outputPath:  nil,
			validateOutput: func(t *testing.T, outputPath *string, buffer []byte) {
				assert.Nil(t, outputPath)
				assert.NotEmpty(t, buffer)
				assert.True(t, len(buffer) > 8 && string(buffer[:8]) == "\x89PNG\r\n\x1a\n")
				// Note: Validating *if* it's truly full page requires comparison or more complex analysis
			},
		},
		{
			name: "Clip PNG to Buffer",
			params: model.VisionScreenshotParams{
				Clip: &model.Rect{X: 10, Y: 10, Width: 50, Height: 50},
			},
			expectError: false,
			outputPath:  nil,
			validateOutput: func(t *testing.T, outputPath *string, buffer []byte) {
				assert.Nil(t, outputPath)
				assert.NotEmpty(t, buffer)
				assert.True(t, len(buffer) > 8 && string(buffer[:8]) == "\x89PNG\r\n\x1a\n")
				// Note: Validating clip dimensions requires image analysis
			},
		},
		{
			name:   "Invalid Path (Directory as file)",
			params: model.VisionScreenshotParams{
				// No specific params needed, path is set directly in outputPath
			},
			expectError:    true,
			outputPath:     func() *string { dir := t.TempDir(); return &dir }(), // Use the temp dir itself as path
			validateOutput: nil,                                                  // No output to validate on error
		},
		// Add more cases for other options like OmitBackground, Scale, Animations, Caret, Timeout
		// Example for OmitBackground to Buffer:
		{
			name: "Omit Background PNG to Buffer",
			params: model.VisionScreenshotParams{
				OmitBackground: playwright.Bool(true),
			},
			expectError: false,
			outputPath:  nil,
			validateOutput: func(t *testing.T, outputPath *string, buffer []byte) {
				assert.Nil(t, outputPath)
				assert.NotEmpty(t, buffer)
				// Visual inspection or advanced image analysis needed to confirm transparency
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// --- Prepare Playwright Options ---
			screenshotOpts := playwright.PageScreenshotOptions{}
			// Set path only if provided for the test case
			if tc.outputPath != nil {
				screenshotOpts.Path = tc.outputPath
			}

			// Map other options from params
			if tc.params.Type != nil {
				if *tc.params.Type == "png" {
					screenshotOpts.Type = playwright.ScreenshotTypePng
				} else if *tc.params.Type == "jpeg" || *tc.params.Type == "jpg" {
					screenshotOpts.Type = playwright.ScreenshotTypeJpeg
				}
			}
			if tc.params.FullPage != nil {
				screenshotOpts.FullPage = tc.params.FullPage
			}
			if tc.params.Quality != nil {
				// Ensure type is JPEG before applying quality
				isJpeg := false
				if screenshotOpts.Type != nil && *screenshotOpts.Type == *playwright.ScreenshotTypeJpeg {
					isJpeg = true
				}
				if isJpeg {
					screenshotOpts.Quality = tc.params.Quality
				}
			}
			if tc.params.OmitBackground != nil {
				screenshotOpts.OmitBackground = tc.params.OmitBackground
			}
			if tc.params.Timeout != nil {
				screenshotOpts.Timeout = tc.params.Timeout
			}
			if tc.params.Clip != nil {
				screenshotOpts.Clip = &playwright.Rect{X: tc.params.Clip.X, Y: tc.params.Clip.Y, Width: tc.params.Clip.Width, Height: tc.params.Clip.Height}
			}
			if tc.params.Scale != nil {
				if *tc.params.Scale == "css" {
					screenshotOpts.Scale = playwright.ScreenshotScaleCss
				} else if *tc.params.Scale == "device" {
					screenshotOpts.Scale = playwright.ScreenshotScaleDevice
				}
			}
			if tc.params.Animations != nil {
				if *tc.params.Animations == "disabled" {
					screenshotOpts.Animations = playwright.ScreenshotAnimationsDisabled
				} else if *tc.params.Animations == "allow" {
					screenshotOpts.Animations = playwright.ScreenshotAnimationsAllow
				}
			}
			if tc.params.Caret != nil {
				if *tc.params.Caret == "hide" {
					screenshotOpts.Caret = playwright.ScreenshotCaretHide
				} else if *tc.params.Caret == "initial" {
					screenshotOpts.Caret = playwright.ScreenshotCaretInitial
				}
			}
			// --- End Prepare Playwright Options ---

			t.Logf("Running test case: %s", tc.name)
			t.Logf("Playwright Screenshot Options: %+v", screenshotOpts)

			// --- Call Playwright Screenshot Directly ---
			buffer, err := page.Screenshot(screenshotOpts)

			// --- Assertions ---
			if tc.expectError {
				assert.Error(t, err, "Expected an error but got nil")
				if err != nil {
					t.Logf("Received expected error: %v", err)
				}
				// Clean up potentially partially created file if path was specified
				if tc.outputPath != nil {
					_ = os.Remove(*tc.outputPath)
				}
			} else {
				assert.NoError(t, err, "Expected no error, but got: %v", err)
				if err != nil {
					t.FailNow() // Stop test if screenshot failed unexpectedly
				}
				// Validate output (file or buffer)
				if tc.validateOutput != nil {
					tc.validateOutput(t, tc.outputPath, buffer)
				}
			}
		})
	}
}

// ensureDirExistsForRelativePathTest is likely no longer needed or needs significant rework
// as the service now handles path generation internally for URL format.
// This helper was designed for tests directly calling page.Screenshot with specific paths.
// func ensureDirExistsForRelativePathTest(t *testing.T, finalPath string) {
// 	dirToCreate := filepath.Dir(finalPath)
// 	t.Logf("Ensuring directory exists for relative path test: %s", dirToCreate)
// 	err := os.MkdirAll(dirToCreate, 0755)
// 	require.NoError(t, err, "Failed to create directory for relative path test: %s", dirToCreate)
// }

// TestVisionClick verifies the ExecuteVisionClick service method.
func TestVisionClick(t *testing.T) {
	// Use the new setup function that provides a configured service and page
	svc, boxID, contextID, pageID, page, cleanup := setupServiceWithVisionTestPage(t)
	defer cleanup()

	// --- Create output directory ---
	screenshotDir := filepath.Join("..", "..", "..", ".test-output", "screenshots") // Relative path to api-server/.test-output/screenshots
	err := os.MkdirAll(screenshotDir, 0755)
	require.NoError(t, err, "Failed to create screenshot directory: %s", screenshotDir)

	// --- 1. Define Target and Get Coordinates ---
	buttonSelector := "#click-btn"
	buttonLocator := page.Locator(buttonSelector)
	bbox, err := buttonLocator.BoundingBox()
	require.NoError(t, err, "Failed to get bounding box for %s", buttonSelector)
	require.NotNil(t, bbox, "Bounding box should not be nil for %s", buttonSelector)

	clickX := bbox.X + bbox.Width/2
	clickY := bbox.Y + bbox.Height/2

	// --- 2. Take Pre-click Screenshot ---
	// Save to project directory instead of temp
	preClickPath := filepath.Join(screenshotDir, fmt.Sprintf("click_before_%s.png", uuid.NewString()))
	_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: &preClickPath})
	require.NoError(t, err, "Failed to take pre-click screenshot")
	t.Logf("Pre-click screenshot saved to: %s", preClickPath)

	// --- 3. Setup Service (Now handled by setupServiceWithVisionTestPage) ---
	// svc, boxID, contextID, pageID are now available from setup

	// --- 4. Execute Click Action ---
	params := model.VisionClickParams{
		X:      int(clickX),
		Y:      int(clickY),
		Button: model.MouseButtonLeft, // Default left click
	}
	clickResult := svc.ExecuteVisionClick(boxID, contextID, pageID, params)

	// --- 5. Assert Service Call Success ---
	require.NotNil(t, clickResult, "ExecuteVisionClick returned nil result")
	resultData, ok := clickResult.(model.VisionClickResult)
	require.True(t, ok, "ExecuteVisionClick returned unexpected type: %T", clickResult)
	assert.True(t, resultData.Success, "ExecuteVisionClick result.Success should be true")

	// --- 6. Verify URL Hash Change ---
	expectedHash := "#click-click_btn"
	// Use require.Eventually for robustness, as hash change might not be instantaneous
	require.Eventually(t, func() bool {
		currentURL := page.URL()
		return strings.HasSuffix(currentURL, expectedHash)
	}, 5*time.Second, 100*time.Millisecond, "URL hash did not become '%s', current URL: %s", expectedHash, page.URL())
	t.Logf("URL hash successfully updated to: %s", expectedHash)

	// --- 7. Take Post-click Screenshot ---
	// Save to project directory instead of temp
	postClickPath := filepath.Join(screenshotDir, fmt.Sprintf("click_after_%s.png", uuid.NewString()))
	_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: &postClickPath})
	require.NoError(t, err, "Failed to take post-click screenshot")
	t.Logf("Post-click screenshot saved to: %s", postClickPath)
}
