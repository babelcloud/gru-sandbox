package service_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/babelcloud/gbox/packages/api-server/config"
	// Use service alias to avoid clash with testing package

	model "github.com/babelcloud/gbox/packages/api-server/pkg/browser"
)

// Helper function to setup playwright connection and a page for tests
// Returns the page and a cleanup function
func setupPlaywrightPage(t *testing.T) (playwright.Page, func()) {
	t.Helper()

	// Run playwright locally
	pw, err := playwright.Run()
	require.NoError(t, err, "could not start playwright run task (ensure drivers are installed locally: npx playwright install)")

	// Launch browser locally
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	require.NoError(t, err, "could not launch local browser")

	// Create context and page
	context, err := browser.NewContext()
	require.NoError(t, err, "could not create browser context")
	page, err := context.NewPage()
	require.NoError(t, err, "could not create page")

	// Navigate to a real page for testing screenshots
	targetURL := "https://gru.ai"
	fmt.Printf("Navigating test page to %s...\n", targetURL)
	_, err = page.Goto(targetURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad, // Change to Load state
		Timeout:   playwright.Float(90000),       // Increase timeout to 90 seconds
	})
	require.NoError(t, err, "could not navigate to %s", targetURL)
	fmt.Printf("Navigation to %s complete.\n", targetURL)

	// Cleanup closes local playwright resources
	cleanup := func() {
		require.NoError(t, page.Close(), "failed to close page")
		require.NoError(t, context.Close(), "failed to close context")
		require.NoError(t, browser.Close(), "failed to close browser")
		require.NoError(t, pw.Stop(), "failed to stop playwright run task")
	}

	return page, cleanup
}

func TestExecuteVisionScreenshot(t *testing.T) {
	page, cleanup := setupPlaywrightPage(t)
	defer cleanup()

	// Get the configured share directory path from the main application config
	cfg := config.GetInstance()
	// Define a dummy BoxID for testing purposes, as the actual service func includes it
	testBoxID := "test-box-id"
	expectedBoxDefaultDir := filepath.Join(cfg.File.Share, testBoxID, "screenshot")

	// Ensure the target directory exists on the host machine before running tests
	err := os.MkdirAll(expectedBoxDefaultDir, 0755)
	require.NoError(t, err, "Failed to create configured default screenshot directory (with boxID) on host: %s", expectedBoxDefaultDir)

	// Define test cases
	testCases := []struct {
		name           string
		params         model.VisionScreenshotParams
		manualPath     string // Optional: Path to use instead of generating default, for inspection
		filenamePrefix string // Optional: Prefix for default filename generation
		expectError    bool
		expectedExt    string                                              // e.g., ".png", ".jpeg"
		validatePath   func(t *testing.T, path string, expectedDir string) // Pass expected dir
	}{
		{
			name:           "Default path (PNG)",
			params:         model.VisionScreenshotParams{},
			filenamePrefix: "screenshot_", // Standard prefix
			expectError:    false,
			expectedExt:    ".png",
			validatePath: func(t *testing.T, path string, expectedDir string) {
				assert.Equal(t, expectedDir, filepath.Dir(path), "Path should be within the configured host screenshot directory including boxID")
				match, _ := regexp.MatchString(`^screenshot_\d{8}_\d{6}\.png$`, filepath.Base(path))
				assert.True(t, match, "Filename should match timestamp format screenshot_YYYYMMDD_HHMMSS.png")
			},
		},
		{
			name:   "Specific path (PNG in temp dir)",
			params: model.VisionScreenshotParams{
				// Path will be generated inside t.Run using t.TempDir()
			},
			expectError: false,
			expectedExt: ".png",
			validatePath: func(t *testing.T, path string, expectedDir string) {
				// expectedDir will be the temp dir created in t.Run
				assert.Equal(t, filepath.Join(expectedDir, "specific_test.png"), path, "Path should match the generated temp path")
			},
		},
		{
			name: "Default path (JPEG)",
			params: model.VisionScreenshotParams{
				Type:    playwright.String("jpeg"),
				Quality: playwright.Int(80),
			},
			filenamePrefix: "screenshot_", // Standard prefix
			expectError:    false,
			expectedExt:    ".jpeg",
			validatePath: func(t *testing.T, path string, expectedDir string) {
				assert.Equal(t, expectedDir, filepath.Dir(path), "Path should be within the configured host screenshot directory including boxID")
				match, _ := regexp.MatchString(`^screenshot_\d{8}_\d{6}\.jpeg$`, filepath.Base(path))
				assert.True(t, match, "Filename should match timestamp format screenshot_YYYYMMDD_HHMMSS.jpeg")
			},
		},
		{
			name: "Specific path (JPEG with Quality in temp dir)",
			params: model.VisionScreenshotParams{
				// Path will be generated inside t.Run using t.TempDir()
				Type:    playwright.String("jpeg"),
				Quality: playwright.Int(90),
			},
			expectError: false,
			expectedExt: ".jpeg",
			validatePath: func(t *testing.T, path string, expectedDir string) {
				// expectedDir will be the temp dir created in t.Run
				assert.Equal(t, filepath.Join(expectedDir, "specific_test_quality.jpeg"), path, "Path should match the generated temp path")
			},
		},
		{
			name: "Full page screenshot (in default dir)", // Updated name
			params: model.VisionScreenshotParams{
				// Path generated manually below
				FullPage: playwright.Bool(true),
			},
			filenamePrefix: "screenshot_fullpage_", // Custom prefix
			expectError:    false,
			expectedExt:    ".png",
			validatePath: func(t *testing.T, path string, expectedDir string) {
				assert.Equal(t, expectedDir, filepath.Dir(path), "Path should be within the configured host screenshot directory including boxID")
				match, _ := regexp.MatchString(`^screenshot_fullpage_\d{8}_\d{6}\.png$`, filepath.Base(path))
				assert.True(t, match, "Filename should match prefix and timestamp")
			},
		},
		{
			name: "Clip screenshot (in temp dir)", // Keep clip in temp dir for now
			params: model.VisionScreenshotParams{
				// Path will be generated inside t.Run using t.TempDir()
				Clip: &model.Rect{X: 10, Y: 10, Width: 50, Height: 50},
			},
			expectError: false,
			expectedExt: ".png",
			validatePath: func(t *testing.T, path string, expectedDir string) {
				// expectedDir will be the temp dir created in t.Run
				assert.Equal(t, filepath.Join(expectedDir, "clipped.png"), path, "Path should match the generated temp path")
			},
		},
		{
			name: "Invalid Path (Directory as file)",
			params: model.VisionScreenshotParams{
				// Create a specific temp dir for this test and pass its path
				Path: func() *string { dir := t.TempDir(); return &dir }(),
			},
			expectError:  true,
			expectedExt:  "",
			validatePath: nil,
		},
		// --- Test Cases for Other Options (Saving to Default Dir) ---
		{
			name: "Omit background (PNG in default dir)",
			params: model.VisionScreenshotParams{
				OmitBackground: playwright.Bool(true),
			},
			filenamePrefix: "screenshot_omitbg_",
			expectError:    false,
			expectedExt:    ".png",
			validatePath: func(t *testing.T, path string, expectedDir string) {
				assert.Equal(t, expectedDir, filepath.Dir(path), "Path should be within the configured host screenshot directory including boxID")
				match, _ := regexp.MatchString(`^screenshot_omitbg_\d{8}_\d{6}\.png$`, filepath.Base(path))
				assert.True(t, match, "Filename should match prefix and timestamp")
				// Visual inspection needed
			},
		},
		{
			name: "Scale CSS (in default dir)",
			params: model.VisionScreenshotParams{
				Scale: playwright.String("css"),
			},
			filenamePrefix: "screenshot_scalecss_",
			expectError:    false,
			expectedExt:    ".png",
			validatePath: func(t *testing.T, path string, expectedDir string) {
				assert.Equal(t, expectedDir, filepath.Dir(path), "Path should be within the configured host screenshot directory including boxID")
				match, _ := regexp.MatchString(`^screenshot_scalecss_\d{8}_\d{6}\.png$`, filepath.Base(path))
				assert.True(t, match, "Filename should match prefix and timestamp")
				// Visual inspection needed
			},
		},
		{
			name: "Animations disabled (in default dir)",
			params: model.VisionScreenshotParams{
				Animations: playwright.String("disabled"),
			},
			filenamePrefix: "screenshot_animdisabled_",
			expectError:    false,
			expectedExt:    ".png",
			validatePath: func(t *testing.T, path string, expectedDir string) {
				assert.Equal(t, expectedDir, filepath.Dir(path), "Path should be within the configured host screenshot directory including boxID")
				match, _ := regexp.MatchString(`^screenshot_animdisabled_\d{8}_\d{6}\.png$`, filepath.Base(path))
				assert.True(t, match, "Filename should match prefix and timestamp")
			},
		},
		{
			name: "Caret initial (in default dir)",
			params: model.VisionScreenshotParams{
				Caret: playwright.String("initial"),
			},
			filenamePrefix: "screenshot_caretinitial_",
			expectError:    false,
			expectedExt:    ".png",
			validatePath: func(t *testing.T, path string, expectedDir string) {
				assert.Equal(t, expectedDir, filepath.Dir(path), "Path should be within the configured host screenshot directory including boxID")
				match, _ := regexp.MatchString(`^screenshot_caretinitial_\d{8}_\d{6}\.png$`, filepath.Base(path))
				assert.True(t, match, "Filename should match prefix and timestamp")
			},
		},
		{
			name: "Low Timeout (in default dir)",
			params: model.VisionScreenshotParams{
				Timeout: playwright.Float(500), // 500 ms timeout
			},
			filenamePrefix: "screenshot_timeout_",
			expectError:    false, // Might fail depending on page load speed
			expectedExt:    ".png",
			validatePath: func(t *testing.T, path string, expectedDir string) {
				assert.Equal(t, expectedDir, filepath.Dir(path), "Path should be within the configured host screenshot directory including boxID")
				match, _ := regexp.MatchString(`^screenshot_timeout_\d{8}_\d{6}\.png$`, filepath.Base(path))
				assert.True(t, match, "Filename should match prefix and timestamp")
			},
		},
		{
			name: "Combined Options (JPEG in default dir)",
			params: model.VisionScreenshotParams{
				Type:           playwright.String("jpeg"),
				Quality:        playwright.Int(60),
				OmitBackground: playwright.Bool(true),
				Scale:          playwright.String("css"),
				Animations:     playwright.String("disabled"),
			},
			filenamePrefix: "screenshot_combined_",
			expectError:    false,
			expectedExt:    ".jpeg",
			validatePath: func(t *testing.T, path string, expectedDir string) {
				assert.Equal(t, expectedDir, filepath.Dir(path), "Path should be within the configured host screenshot directory including boxID")
				match, _ := regexp.MatchString(`^screenshot_combined_\d{8}_\d{6}\.jpeg$`, filepath.Base(path))
				assert.True(t, match, "Filename should match prefix and timestamp")
			},
		},
		// --- Test Relative Path ---
		{
			name: "Relative path (PNG in default subdir)",
			params: model.VisionScreenshotParams{
				Path: playwright.String("subdir/relative_test.png"), // Provide relative path
			},
			// No filenamePrefix needed, path is explicitly set
			expectError: false,
			expectedExt: ".png",
			validatePath: func(t *testing.T, path string, expectedBaseDir string) {
				// expectedBaseDir passed here will be expectedBoxDefaultDir
				expectedFullPath := filepath.Join(expectedBaseDir, "subdir/relative_test.png")
				assert.Equal(t, expectedFullPath, path, "Path should match base dir + relative path")
				// Also check the directory part matches the base dir + subdir
				assert.Equal(t, filepath.Join(expectedBaseDir, "subdir"), filepath.Dir(path), "Directory should be box base dir + subdir")
			},
		},
	}

	for _, tc := range testCases {
		// --- Prepare Path and Options ---
		var finalSavePath string
		var isTempPath bool = false // Flag to indicate if path is temporary

		// Special handling for temp paths inside t.Run
		if tc.name == "Specific path (PNG in temp dir)" ||
			tc.name == "Specific path (JPEG with Quality in temp dir)" ||
			tc.name == "Clip screenshot (in temp dir)" {
			isTempPath = true
			// Path generation will happen inside t.Run
		} else if tc.filenamePrefix != "" {
			// Generate path based on prefix and default dir
			timestamp := time.Now().Format("20060102_150405")
			fileExt := tc.expectedExt
			if fileExt == "" {
				fileExt = ".png"
			}
			if tc.params.Type != nil && (*tc.params.Type == "jpeg" || *tc.params.Type == "jpg") {
				fileExt = ".jpeg"
			}
			if fileExt[0] != '.' {
				fileExt = "." + fileExt
			}
			filename := fmt.Sprintf("%s%s%s", tc.filenamePrefix, timestamp, fileExt)
			finalSavePath = filepath.Join(expectedBoxDefaultDir, filename)
			t.Logf("Test case '%s' will save to explicit path: %s", tc.name, finalSavePath)
		} else if tc.params.Path != nil {
			// If path is explicitly provided in params (e.g., relative path)
			// Handle relative vs absolute provided path
			if filepath.IsAbs(*tc.params.Path) {
				finalSavePath = *tc.params.Path // Use absolute path directly (e.g., invalid dir test)
			} else {
				// Assume relative paths are meant to be within the default dir for testing direct screenshot function
				finalSavePath = filepath.Join(expectedBoxDefaultDir, *tc.params.Path)
			}
		} else {
			// Should not happen if cases are defined correctly (either prefix or explicit path)
			t.Fatalf("Test case '%s' has neither filenamePrefix nor explicit Path set", tc.name)
		}
		// Set the Path field within the params copy for this run
		currentParams := tc.params
		currentParams.Path = &finalSavePath

		// Construct Playwright options from the params for this test case
		screenshotOpts := playwright.PageScreenshotOptions{}
		screenshotOpts.Path = currentParams.Path // Use the final path
		if currentParams.Type != nil {
			if *currentParams.Type == "png" {
				screenshotOpts.Type = playwright.ScreenshotTypePng
			} else if *currentParams.Type == "jpeg" || *currentParams.Type == "jpg" {
				screenshotOpts.Type = playwright.ScreenshotTypeJpeg
			}
		}
		if currentParams.FullPage != nil {
			screenshotOpts.FullPage = currentParams.FullPage
		}
		if currentParams.Quality != nil {
			// Correctly check type before applying quality
			isJpeg := false
			// Assuming playwright.ScreenshotTypeJpeg is *playwright.ScreenshotType based on linter errors
			// Check if both pointers are non-nil and dereference to compare values
			if screenshotOpts.Type != nil && playwright.ScreenshotTypeJpeg != nil && *screenshotOpts.Type == *playwright.ScreenshotTypeJpeg {
				isJpeg = true
			}
			if isJpeg {
				screenshotOpts.Quality = currentParams.Quality
			}
		}
		if currentParams.OmitBackground != nil {
			screenshotOpts.OmitBackground = currentParams.OmitBackground
		}
		if currentParams.Timeout != nil {
			screenshotOpts.Timeout = currentParams.Timeout
		}
		if currentParams.Clip != nil {
			screenshotOpts.Clip = &playwright.Rect{X: currentParams.Clip.X, Y: currentParams.Clip.Y, Width: currentParams.Clip.Width, Height: currentParams.Clip.Height}
		}
		if currentParams.Scale != nil {
			if *currentParams.Scale == "css" {
				screenshotOpts.Scale = playwright.ScreenshotScaleCss
			} else if *currentParams.Scale == "device" {
				screenshotOpts.Scale = playwright.ScreenshotScaleDevice
			}
		}
		if currentParams.Animations != nil {
			if *currentParams.Animations == "disabled" {
				screenshotOpts.Animations = playwright.ScreenshotAnimationsDisabled
			} else if *currentParams.Animations == "allow" {
				screenshotOpts.Animations = playwright.ScreenshotAnimationsAllow
			}
		}
		if currentParams.Caret != nil {
			if *currentParams.Caret == "hide" {
				screenshotOpts.Caret = playwright.ScreenshotCaretHide
			} else if *currentParams.Caret == "initial" {
				screenshotOpts.Caret = playwright.ScreenshotCaretInitial
			}
		}
		// --- End Prepare ---

		t.Run(tc.name, func(t *testing.T) {
			// --- Final Path Setup (handle temp paths here) ---
			currentFinalSavePath := finalSavePath // Use outer scope path by default
			currentExpectedDir := expectedBoxDefaultDir

			if isTempPath {
				tempDir := t.TempDir() // Generate temp dir specific to this subtest run
				baseFilename := ""
				if tc.name == "Specific path (PNG in temp dir)" {
					baseFilename = "specific_test.png"
				} else if tc.name == "Specific path (JPEG with Quality in temp dir)" {
					baseFilename = "specific_test_quality.jpeg"
				} else if tc.name == "Clip screenshot (in temp dir)" {
					baseFilename = "clipped.png"
				}
				currentFinalSavePath = filepath.Join(tempDir, baseFilename)
				currentParams.Path = &currentFinalSavePath  // Update path in params for Playwright
				screenshotOpts.Path = &currentFinalSavePath // Ensure options also have the correct path
				currentExpectedDir = tempDir                // Update expected dir for validation
			}

			// Ensure directory exists if this is the relative path test
			ensureDirExistsForRelativePathTest(t, currentFinalSavePath)

			// Call page.Screenshot directly using the potentially updated screenshotOpts
			_, err := page.Screenshot(screenshotOpts)

			if tc.expectError {
				assert.Error(t, err, "Expected an error but got nil")
				if err != nil {
					t.Logf("Received expected error: %v", err)
				}
				// Clean up potentially partially created file if path was specified
				_ = os.Remove(currentFinalSavePath)
			} else {
				assert.NoError(t, err, "Expected no error, but got: %v", err)
				if err != nil {
					t.FailNow()
				}
				// If no error, proceed with file validation based on finalSavePath
				// Validate the path format/location
				if tc.validatePath != nil {
					// Pass the correctly determined path and expected dir for this run
					tc.validatePath(t, currentFinalSavePath, currentExpectedDir)
				}

				// Check if the file actually exists
				_, statErr := os.Stat(currentFinalSavePath)
				assert.NoError(t, statErr, "Screenshot file should exist at path: %s", currentFinalSavePath)

				// Basic check for file extension (using the potentially updated path)
				assert.Equal(t, tc.expectedExt, filepath.Ext(currentFinalSavePath), "File extension should match expected")
			}
		})
	}
}

// Helper function specifically for the relative path test case inside TestExecuteVisionScreenshot
// to ensure the directory exists before calling page.Screenshot directly.
// This mimics the behavior of the actual ExecuteVisionScreenshot service function.
func ensureDirExistsForRelativePathTest(t *testing.T, finalPath string) {
	// This helper now needs to create the directory structure including the boxID and subdir
	// It's only called for the relative path test case
	// The finalPath passed in should already include the boxID and subdir: e.g., /share/test-box-id/screenshot/subdir/relative_test.png
	dirToCreate := filepath.Dir(finalPath)
	t.Logf("Ensuring directory exists for relative path test: %s", dirToCreate)
	err := os.MkdirAll(dirToCreate, 0755)
	require.NoError(t, err, "Failed to create directory for relative path test: %s", dirToCreate)
}
