package service_test

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/babelcloud/gbox/packages/api-server/config"
	model "github.com/babelcloud/gbox/packages/api-server/pkg/browser"
)

// Shared screenshot directory for the current test run
var (
	testRunScreenshotDir  string
	initScreenshotDirOnce sync.Once
)

// getTestRunScreenshotDir ensures a unique directory is created once per test run
// and returns the path to that directory.
func getTestRunScreenshotDir(t *testing.T) string {
	t.Helper() // Mark this as a test helper
	initScreenshotDirOnce.Do(func() {
		// Use a timestamp combined with the base name for the directory name
		timestamp := time.Now().Format("20060102-150405") // YYYYMMDD-HHMMSS format
		dirName := fmt.Sprintf("screenshots_%s", timestamp)
		baseOutputDir := filepath.Join("..", "..", "..", ".test-output") // Parent directory
		testRunScreenshotDir = filepath.Join(baseOutputDir, dirName)     // Combine parent and new dir name
		err := os.MkdirAll(testRunScreenshotDir, 0755)
		// Use require inside Do might be tricky for test lifecycle, Fatalf is safer here
		if err != nil {
			t.Fatalf("Failed to create shared screenshot directory '%s': %v", testRunScreenshotDir, err)
		}
		t.Logf("Created shared screenshot directory for test run: %s", testRunScreenshotDir)
	})
	if testRunScreenshotDir == "" {
		t.Fatal("Shared screenshot directory was not initialized")
	}
	return testRunScreenshotDir
}

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
		Headless: playwright.Bool(true), // Run headless
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
		Timeout:   playwright.Float(10000),       // 10 second timeout for local file
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

	// --- Get shared output directory for this test run ---
	testScreenshotDir := getTestRunScreenshotDir(t)

	// --- 1. Define Target and Get Coordinates ---
	buttonSelector := "#click-btn"
	buttonLocator := page.Locator(buttonSelector)
	bbox, err := buttonLocator.BoundingBox()
	require.NoError(t, err, "Failed to get bounding box for %s", buttonSelector)
	require.NotNil(t, bbox, "Bounding box should not be nil for %s", buttonSelector)

	clickX := bbox.X + bbox.Width/2
	clickY := bbox.Y + bbox.Height/2

	// --- 2. Take Pre-click Screenshot ---
	preClickPath := filepath.Join(testScreenshotDir, "click_before.png") // Use shared dir and descriptive name
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
	postClickPath := filepath.Join(testScreenshotDir, "click_after.png") // Use shared dir and descriptive name
	_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: &postClickPath})
	require.NoError(t, err, "Failed to take post-click screenshot")
	t.Logf("Post-click screenshot saved to: %s", postClickPath)
}

// TestVisionDoubleClick verifies the ExecuteVisionDoubleClick service method.
func TestVisionDoubleClick(t *testing.T) {
	// Use the new setup function that provides a configured service and page
	svc, boxID, contextID, pageID, page, cleanup := setupServiceWithVisionTestPage(t)
	defer cleanup()

	// --- Get shared output directory for this test run ---
	testScreenshotDir := getTestRunScreenshotDir(t)

	// --- 1. Define Target and Get Coordinates ---
	buttonSelector := "#dblclick-btn" // Target the double-click button
	buttonLocator := page.Locator(buttonSelector)
	bbox, err := buttonLocator.BoundingBox()
	require.NoError(t, err, "Failed to get bounding box for %s", buttonSelector)
	require.NotNil(t, bbox, "Bounding box should not be nil for %s", buttonSelector)

	clickX := bbox.X + bbox.Width/2
	clickY := bbox.Y + bbox.Height/2

	// --- 2. Take Pre-double-click Screenshot ---
	preDblClickPath := filepath.Join(testScreenshotDir, "dblclick_before.png") // Use shared dir and descriptive name
	_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: &preDblClickPath})
	require.NoError(t, err, "Failed to take pre-double-click screenshot")
	t.Logf("Pre-double-click screenshot saved to: %s", preDblClickPath)

	// --- 3. Setup Service (Handled by setupServiceWithVisionTestPage) ---
	// svc, boxID, contextID, pageID are available

	// --- 4. Execute Double Click Action ---
	params := model.VisionDoubleClickParams{
		X: int(clickX),
		Y: int(clickY),
		// Button and Delay options could be added here if needed
	}
	dblClickResult := svc.ExecuteVisionDoubleClick(boxID, contextID, pageID, params)

	// --- 5. Assert Service Call Success ---
	require.NotNil(t, dblClickResult, "ExecuteVisionDoubleClick returned nil result")
	resultData, ok := dblClickResult.(model.VisionDoubleClickResult)
	require.True(t, ok, "ExecuteVisionDoubleClick returned unexpected type: %T", dblClickResult)
	assert.True(t, resultData.Success, "ExecuteVisionDoubleClick result.Success should be true")

	// --- 6. Verify URL Hash Change ---
	expectedHash := "#doubleClick-dblclick_btn" // See vision-test.html
	require.Eventually(t, func() bool {
		currentURL := page.URL()
		return strings.HasSuffix(currentURL, expectedHash)
	}, 5*time.Second, 100*time.Millisecond, "URL hash did not become '%s', current URL: %s", expectedHash, page.URL())
	t.Logf("URL hash successfully updated to: %s", expectedHash)

	// --- 7. Take Post-double-click Screenshot ---
	postDblClickPath := filepath.Join(testScreenshotDir, "dblclick_after.png") // Use shared dir and descriptive name
	_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: &postDblClickPath})
	require.NoError(t, err, "Failed to take post-double-click screenshot")
	t.Logf("Post-double-click screenshot saved to: %s", postDblClickPath)
}

// TestVisionType verifies the ExecuteVisionType service method.
func TestVisionType(t *testing.T) {
	// Use the new setup function that provides a configured service and page
	svc, boxID, contextID, pageID, page, cleanup := setupServiceWithVisionTestPage(t)
	defer cleanup()

	// --- Get shared output directory for this test run ---
	testScreenshotDir := getTestRunScreenshotDir(t)

	// --- 1. Define Target and Text ---
	inputSelector := "#type-input"
	typedText := "hello world"

	// --- 2. Take Pre-type Screenshot ---
	preTypePath := filepath.Join(testScreenshotDir, "type_before.png")
	_, err := page.Screenshot(playwright.PageScreenshotOptions{Path: &preTypePath})
	require.NoError(t, err, "Failed to take pre-type screenshot")
	t.Logf("Pre-type screenshot saved to: %s", preTypePath)

	// --- 3. Focus the input element first (good practice for typing) ---
	err = page.Locator(inputSelector).Focus()
	require.NoError(t, err, "Failed to focus input element %s", inputSelector)
	t.Logf("Focused input element: %s", inputSelector)

	// --- 4. Execute Type Action ---
	params := model.VisionTypeParams{
		Text: typedText,
	}
	typeResult := svc.ExecuteVisionType(boxID, contextID, pageID, params)

	// --- 5. Assert Service Call Success ---
	require.NotNil(t, typeResult, "ExecuteVisionType returned nil result")
	resultData, ok := typeResult.(model.VisionTypeResult)
	require.True(t, ok, "ExecuteVisionType returned unexpected type: %T", typeResult)
	assert.True(t, resultData.Success, "ExecuteVisionType result.Success should be true")

	// --- 6. Verify Input Value ---
	inputValue, err := page.Locator(inputSelector).InputValue()
	require.NoError(t, err, "Failed to get input value for %s", inputSelector)
	assert.Equal(t, typedText, inputValue, "Input value should match typed text")
	t.Logf("Input value verified: %s", inputValue)

	// --- 7. Verify URL Hash Change ---
	expectedHashSuffix := strings.ReplaceAll(typedText, " ", "_") // Convert space to underscore for hash
	expectedHash := fmt.Sprintf("#type-%s", expectedHashSuffix)
	require.Eventually(t, func() bool {
		currentURL := page.URL()
		return strings.HasSuffix(currentURL, expectedHash)
	}, 5*time.Second, 100*time.Millisecond, "URL hash did not become '%s', current URL: %s", expectedHash, page.URL())
	t.Logf("URL hash successfully updated to: %s", expectedHash)

	// --- 8. Take Post-type Screenshot ---
	postTypePath := filepath.Join(testScreenshotDir, "type_after.png")
	_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: &postTypePath})
	require.NoError(t, err, "Failed to take post-type screenshot")
	t.Logf("Post-type screenshot saved to: %s", postTypePath)
}

// TestVisionScroll verifies the ExecuteVisionScroll service method.
// It now tests scrolling a specific element (#scroll-area) after moving the mouse into it.
func TestVisionScroll(t *testing.T) {
	// Use the new setup function that provides a configured service and page
	svc, boxID, contextID, pageID, page, cleanup := setupServiceWithVisionTestPage(t)
	defer cleanup()

	// --- Get shared output directory for this test run ---
	testScreenshotDir := getTestRunScreenshotDir(t)

	// --- 1. Define Target Element and Scroll Amount ---
	scrollElementSelector := "#scroll-area" // Assume this is the ID of the scrollable div
	scrollY := 100                          // Scroll down by 100 pixels
	scrollX := 0                            // No horizontal scroll

	// --- 2. Locate Element and Calculate Target Coordinates ---
	scrollLocator := page.Locator(scrollElementSelector)
	bbox, err := scrollLocator.BoundingBox()
	require.NoError(t, err, "Failed to get bounding box for %s", scrollElementSelector)
	require.NotNil(t, bbox, "Bounding box should not be nil for %s", scrollElementSelector)
	targetX := bbox.X + bbox.Width/2
	targetY := bbox.Y + bbox.Height/2

	// --- 3. Move Mouse into the Element ---
	err = page.Mouse().Move(targetX, targetY)
	require.NoError(t, err, "Failed to move mouse to %s (%f, %f)", scrollElementSelector, targetX, targetY)
	t.Logf("Mouse moved into element: %s", scrollElementSelector)
	// Add a small delay to ensure the move completes and potential hover effects trigger
	time.Sleep(100 * time.Millisecond)

	// --- 4. Take Pre-scroll Screenshot (after moving mouse) ---
	preScrollPath := filepath.Join(testScreenshotDir, "scroll_element_before.png") // Update filename
	_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: &preScrollPath})
	require.NoError(t, err, "Failed to take pre-scroll screenshot")
	t.Logf("Pre-scroll screenshot saved to: %s", preScrollPath)

	// --- 5. Execute Scroll Action (at current mouse position) ---
	params := model.VisionScrollParams{
		ScrollX: scrollX,
		ScrollY: scrollY,
		// X/Y are not strictly needed by the current Mouse.Wheel implementation in service,
		// but setting them doesn't hurt.
		X: int(targetX),
		Y: int(targetY),
	}
	scrollResult := svc.ExecuteVisionScroll(boxID, contextID, pageID, params)

	// --- 6. Assert Service Call Success ---
	require.NotNil(t, scrollResult, "ExecuteVisionScroll returned nil result")

	// Check if an error was returned and log it
	if errResult, isError := scrollResult.(model.VisionErrorResult); isError {
		t.Fatalf("ExecuteVisionScroll returned an error: %s", errResult.Error)
	}

	// Proceed with the type assertion now that we know it's not an error
	resultData, ok := scrollResult.(model.VisionScrollResult)
	require.True(t, ok, "ExecuteVisionScroll returned unexpected type: %T (expected VisionScrollResult)", scrollResult)
	assert.True(t, resultData.Success, "ExecuteVisionScroll result.Success should be true")

	// --- 7. Verify Element Scroll Position and URL Hash Change ---
	time.Sleep(100 * time.Millisecond) // Delay for scroll event propagation

	// Get the actual scrollTop of the element after the scroll action.
	// Pass nil as the second argument for Evaluate as required by the signature
	currentScrollYRaw, evalErr := scrollLocator.Evaluate("el => el.scrollTop", nil) // Evaluate element's scrollTop
	require.NoError(t, evalErr, "Failed to evaluate %s.scrollTop", scrollElementSelector)
	require.NotNil(t, currentScrollYRaw, "%s.scrollTop evaluation returned nil", scrollElementSelector)

	// Handle different possible numeric types returned by Evaluate
	var actualScrollY float64
	switch v := currentScrollYRaw.(type) {
	case float64:
		actualScrollY = v
	case int:
		actualScrollY = float64(v)
	case int64:
		actualScrollY = float64(v)
	// Add other potential types like json.Number if necessary
	default:
		t.Fatalf("%s.scrollTop evaluation returned unexpected type: %T", scrollElementSelector, currentScrollYRaw)
	}

	// Use GreaterOrEqual because scroll amount might not be exact
	require.GreaterOrEqual(t, actualScrollY, float64(scrollY)*0.9, "Actual element scrollTop (%f) should be close to the scrolled amount (%d)", actualScrollY, scrollY)

	// Check the URL hash. The JS in vision-test.html seems to replace hyphens in the element ID with underscores.
	elementIDPart := strings.ReplaceAll(scrollElementSelector[1:], "-", "_")        // Replace hyphen with underscore
	expectedHash := fmt.Sprintf("#scroll-%s_%d", elementIDPart, int(actualScrollY)) // Use modified ID part

	require.Eventually(t, func() bool {
		currentURL := page.URL()
		return strings.HasSuffix(currentURL, expectedHash)
	}, 5*time.Second, 100*time.Millisecond, "URL hash did not end with '%s', current URL: %s", expectedHash, page.URL())
	t.Logf("URL hash successfully updated, ending with: %s", expectedHash)

	// --- 8. Take Post-scroll Screenshot ---
	postScrollPath := filepath.Join(testScreenshotDir, "scroll_element_after.png") // Update filename
	_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: &postScrollPath})
	require.NoError(t, err, "Failed to take post-scroll screenshot")
	t.Logf("Post-scroll screenshot saved to: %s", postScrollPath)
}

// TestVisionDrag verifies the ExecuteVisionDrag service method.
func TestVisionDrag(t *testing.T) {
	// Use the new setup function that provides a configured service and page
	svc, boxID, contextID, pageID, page, cleanup := setupServiceWithVisionTestPage(t)
	defer cleanup()

	// --- Get shared output directory for this test run ---
	testScreenshotDir := getTestRunScreenshotDir(t)

	// --- 1. Define Target and Calculate Coordinates ---
	dragSourceSelector := "#drag-source"
	dragSourceLocator := page.Locator(dragSourceSelector)
	bbox, err := dragSourceLocator.BoundingBox()
	require.NoError(t, err, "Failed to get bounding box for %s", dragSourceSelector)
	require.NotNil(t, bbox, "Bounding box should not be nil for %s", dragSourceSelector)

	startX := int(bbox.X + bbox.Width/2)
	startY := int(bbox.Y + bbox.Height/2)
	endX := startX + 50 // Drag 50px right
	endY := startY + 50 // Drag 50px down

	// --- 2. Take Pre-drag Screenshot ---
	preDragPath := filepath.Join(testScreenshotDir, "drag_before.png")
	_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: &preDragPath})
	require.NoError(t, err, "Failed to take pre-drag screenshot")
	t.Logf("Pre-drag screenshot saved to: %s", preDragPath)

	// --- 3. Execute Drag Action ---
	params := model.VisionDragParams{
		Path: []model.Coordinate{
			{X: startX, Y: startY}, // Start point
			{X: endX, Y: endY},     // End point
		},
	}
	dragResult := svc.ExecuteVisionDrag(boxID, contextID, pageID, params)

	// --- 4. Assert Service Call Success ---
	require.NotNil(t, dragResult, "ExecuteVisionDrag returned nil result")
	resultData, ok := dragResult.(model.VisionDragResult)
	require.True(t, ok, "ExecuteVisionDrag returned unexpected type: %T", dragResult)
	assert.True(t, resultData.Success, "ExecuteVisionDrag result.Success should be true")

	// --- 5. Verify URL Hash Change (indicates dragEnd event fired) ---
	// The JS replaces non-alphanumeric chars in details with '_' -> "#dragEnd-at_X_Y"
	expectedHash := fmt.Sprintf("#dragEnd-at_%d_%d", endX, endY)
	require.Eventually(t, func() bool {
		currentURL := page.URL()
		// The hash might contain extra characters from the replacement logic, check suffix
		return strings.HasSuffix(currentURL, expectedHash)
	}, 5*time.Second, 100*time.Millisecond, "URL hash did not end with '%s', current URL: %s", expectedHash, page.URL())
	t.Logf("URL hash successfully updated, ending with: %s", expectedHash)

	// --- 6. Take Post-drag Screenshot ---
	postDragPath := filepath.Join(testScreenshotDir, "drag_after.png")
	_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: &postDragPath})
	require.NoError(t, err, "Failed to take post-drag screenshot")
	t.Logf("Post-drag screenshot saved to: %s", postDragPath)
}

// TestVisionKeyPress verifies the ExecuteVisionKeyPress service method
func TestVisionKeyPress(t *testing.T) {
	// Use the new setup function that provides a configured service and page
	svc, boxID, contextID, pageID, page, cleanup := setupServiceWithVisionTestPage(t)
	defer cleanup()

	// --- Get shared output directory for this test run ---
	testScreenshotDir := getTestRunScreenshotDir(t)

	// --- 1. Define Target Element and Keys ---
	pressAreaSelector := "#press-area"
	keysToPress := []string{"Control+Shift+Alt+Meta+A"}

	// --- 2. Take Pre-press Screenshot ---
	prePressPath := filepath.Join(testScreenshotDir, "keypress_before.png")
	_, err := page.Screenshot(playwright.PageScreenshotOptions{Path: &prePressPath})
	require.NoError(t, err, "Failed to take pre-keypress screenshot")
	t.Logf("Pre-keypress screenshot saved to: %s", prePressPath)

	// --- 3. Focus the target element ---
	pressAreaLocator := page.Locator(pressAreaSelector)
	err = pressAreaLocator.Focus()
	require.NoError(t, err, "Failed to focus element %s", pressAreaSelector)
	t.Logf("Focused element: %s", pressAreaSelector)

	// --- 4. Execute Key Press Action ---
	params := model.VisionKeyPressParams{
		Keys: keysToPress,
	}
	keyPressResult := svc.ExecuteVisionKeyPress(boxID, contextID, pageID, params)

	// --- 5. Assert Service Call Success ---
	require.NotNil(t, keyPressResult, "ExecuteVisionKeyPress returned nil result")
	resultData, ok := keyPressResult.(model.VisionKeyPressResult)
	require.True(t, ok, "ExecuteVisionKeyPress returned unexpected type: %T", keyPressResult)
	assert.True(t, resultData.Success, "ExecuteVisionKeyPress result.Success should be true")

	// --- 6. Verify UI State Change (Last Action Display and URL Hash) ---
	// According to vision-test.html logic, the hash reflects the last key press event.
	// For "Control+Shift+Alt+Meta+A", the JS should generate:
	// Mac:   Display "keyPress (⌃+⌥+⇧+⌘+A)", Hash #keyPress-⌃_⌥_⇧_⌘_A
	// Other: Display "keyPress (Ctrl+Alt+Shift+Meta+A)", Hash #keyPress-Ctrl_Alt_Shift_Meta_A
	// The JS replaces non-alphanumeric chars (excluding symbols ⌃⌥⇧⌘) with '_'
	// Assuming test runs on macOS based on user context
	expectedDisplayText := "(⌃+⌥+⇧+⌘+A)"
	// Define the raw, unescaped hash suffix expected from the JS logic
	expectedRawHashSuffix := "keyPress-⌃_⌥_⇧_⌘_A"
	// URL-encode the raw suffix to match how the browser represents it in the URL fragment
	expectedEncodedHash := "#" + url.PathEscape(expectedRawHashSuffix)

	// Verify the displayed text reflects the combination
	require.Eventually(t, func() bool {
		lastActionText, err := page.Locator("#last-action").TextContent()
		if err != nil {
			t.Logf("Error getting text content: %v", err)
			return false
		}
		// Check if the text contains the representation of the last key press
		return strings.Contains(lastActionText, expectedDisplayText) && strings.Contains(lastActionText, "keyPress") // Updated check
	}, 5*time.Second, 100*time.Millisecond, "Last action display did not update correctly for combo key press. Expected to contain: %s, Got: %s", expectedDisplayText) // Updated message
	t.Logf("Last action display updated correctly.")

	// Verify the URL hash based on the last key press ('Enter')
	require.Eventually(t, func() bool {
		currentURL := page.URL()
		// Compare the suffix with the *encoded* expected hash
		return strings.HasSuffix(currentURL, expectedEncodedHash)
	}, 5*time.Second, 100*time.Millisecond, "URL hash did not end with '%s', current URL: %s", expectedEncodedHash, page.URL()) // Use encoded hash in message
	t.Logf("URL hash successfully updated, ending with: %s", expectedEncodedHash) // Log the encoded hash

	// --- 7. Take Post-press Screenshot ---
	postPressPath := filepath.Join(testScreenshotDir, "keypress_after.png")
	_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: &postPressPath})
	require.NoError(t, err, "Failed to take post-keypress screenshot")
	t.Logf("Post-keypress screenshot saved to: %s", postPressPath)
}

// TestVisionMove verifies the ExecuteVisionMove service method.
func TestVisionMove(t *testing.T) {
	// Use the new setup function that provides a configured service and page
	svc, boxID, contextID, pageID, page, cleanup := setupServiceWithVisionTestPage(t)
	defer cleanup()

	// --- Get shared output directory for this test run ---
	testScreenshotDir := getTestRunScreenshotDir(t)

	// --- 1. Define Target Coordinates ---
	// Let's move the mouse to the center of the 'drag-source' element
	targetSelector := "#drag-source"
	targetLocator := page.Locator(targetSelector)
	bbox, err := targetLocator.BoundingBox()
	require.NoError(t, err, "Failed to get bounding box for %s", targetSelector)
	require.NotNil(t, bbox, "Bounding box should not be nil for %s", targetSelector)

	targetX := int(bbox.X + bbox.Width/2)
	targetY := int(bbox.Y + bbox.Height/2)

	// --- 2. Take Pre-move Screenshot ---
	preMovePath := filepath.Join(testScreenshotDir, "move_before.png")
	_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: &preMovePath})
	require.NoError(t, err, "Failed to take pre-move screenshot")
	t.Logf("Pre-move screenshot saved to: %s", preMovePath)

	// --- 3. Execute Move Action ---
	params := model.VisionMoveParams{
		X: targetX,
		Y: targetY,
	}
	moveResult := svc.ExecuteVisionMove(boxID, contextID, pageID, params)

	// --- 4. Assert Service Call Success ---
	require.NotNil(t, moveResult, "ExecuteVisionMove returned nil result")
	resultData, ok := moveResult.(model.VisionMoveResult)
	require.True(t, ok, "ExecuteVisionMove returned unexpected type: %T", moveResult)
	assert.True(t, resultData.Success, "ExecuteVisionMove result.Success should be true")

	// --- 5. Verify URL Hash Change (indicates mousemove event fired at location) ---
	// The JS updates hash like #mouseMove-at_X_Y
	expectedHash := fmt.Sprintf("#mouseMove-at_%d_%d", targetX, targetY)
	require.Eventually(t, func() bool {
		currentURL := page.URL()
		// Use HasSuffix as events might fire rapidly, we care about the final state
		return strings.HasSuffix(currentURL, expectedHash)
	}, 5*time.Second, 100*time.Millisecond, "URL hash did not end with '%s', current URL: %s", expectedHash, page.URL())
	t.Logf("URL hash successfully updated, ending with: %s", expectedHash)

	// --- 6. Take Post-move Screenshot ---
	postMovePath := filepath.Join(testScreenshotDir, "move_after.png")
	_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: &postMovePath})
	require.NoError(t, err, "Failed to take post-move screenshot")
	t.Logf("Post-move screenshot saved to: %s", postMovePath)
}
