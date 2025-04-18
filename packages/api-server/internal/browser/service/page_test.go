package service_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strconv"
	"testing"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	boxSvc "github.com/babelcloud/gbox/packages/api-server/internal/box/service" // Import BoxService interface
	service "github.com/babelcloud/gbox/packages/api-server/internal/browser/service"
	boxModel "github.com/babelcloud/gbox/packages/api-server/pkg/box"
	model "github.com/babelcloud/gbox/packages/api-server/pkg/browser"
)

// --- Mock BoxService --- (Needed for BrowserService dependency)
type mockBoxService struct {
	dynamicPort int // Store the dynamic port for GetExternalPort
}

func (m *mockBoxService) Create(ctx context.Context, params *boxModel.BoxCreateParams) (*boxModel.Box, error) {
	// Simulate successful box creation for testing purposes if needed by BrowserService
	// Returning a minimal box object
	return &boxModel.Box{ID: "mock-box-" + uuid.NewString()}, nil
}
func (m *mockBoxService) Get(ctx context.Context, boxID string) (*boxModel.Box, error) {
	// Simulate finding the box
	return &boxModel.Box{ID: boxID}, nil
}

// Update Delete signature to match interface
func (m *mockBoxService) Delete(ctx context.Context, boxID string, params *boxModel.BoxDeleteParams) (*boxModel.BoxDeleteResult, error) {
	// Simulate successful deletion
	return &boxModel.BoxDeleteResult{}, nil
}

// Update List signature to match interface
func (m *mockBoxService) List(ctx context.Context, params *boxModel.BoxListParams) (*boxModel.BoxListResult, error) {
	// Return empty result struct
	return &boxModel.BoxListResult{}, nil
}

// Update DeleteAll signature to match interface
func (m *mockBoxService) DeleteAll(ctx context.Context, params *boxModel.BoxesDeleteParams) (*boxModel.BoxesDeleteResult, error) {
	// Return empty result struct
	return &boxModel.BoxesDeleteResult{}, nil
}

// Add missing Exec method
func (m *mockBoxService) Exec(ctx context.Context, boxID string, params *boxModel.BoxExecParams) (*boxModel.BoxExecResult, error) {
	return nil, fmt.Errorf("mockBoxService.Exec not implemented")
}

// Use correct ExtractArchive signature based on interface and pkg/box/archive.go
func (m *mockBoxService) ExtractArchive(ctx context.Context, boxID string, params *boxModel.BoxArchiveExtractParams) error {
	return nil // Simulate success for the mock
}

// Correct GetArchive signature based on linter feedback
func (m *mockBoxService) GetArchive(ctx context.Context, boxID string, params *boxModel.BoxArchiveGetParams) (*boxModel.BoxArchiveResult, io.ReadCloser, error) {
	return nil, nil, fmt.Errorf("mockBoxService.GetArchive not implemented")
}

// Update GetExternalPort to return the dynamic port stored in the mock
func (m *mockBoxService) GetExternalPort(ctx context.Context, boxID string, port int) (int, error) {
	if m.dynamicPort == 0 {
		return 0, fmt.Errorf("mock dynamic port not set")
	}
	return m.dynamicPort, nil
}

// Add missing HeadArchive method (assuming signature from archive.go)
func (m *mockBoxService) HeadArchive(ctx context.Context, boxID string, params *boxModel.BoxArchiveHeadParams) (*boxModel.BoxArchiveHeadResult, error) {
	return nil, fmt.Errorf("mockBoxService.HeadArchive not implemented")
}

// Update Reclaim signature based on interface
func (m *mockBoxService) Reclaim(ctx context.Context) (*boxModel.BoxReclaimResult, error) {
	// Return empty result struct and nil error
	return &boxModel.BoxReclaimResult{}, nil
}

// Add missing Start method
func (m *mockBoxService) Start(ctx context.Context, id string) (*boxModel.BoxStartResult, error) {
	return nil, fmt.Errorf("mockBoxService.Start not implemented")
}

// Add missing Stop method
func (m *mockBoxService) Stop(ctx context.Context, id string) (*boxModel.BoxStopResult, error) {
	return nil, fmt.Errorf("mockBoxService.Stop not implemented")
}

// Add missing Run method
func (m *mockBoxService) Run(ctx context.Context, id string, params *boxModel.BoxRunParams) (*boxModel.BoxRunResult, error) {
	return nil, fmt.Errorf("mockBoxService.Run not implemented")
}

var _ boxSvc.BoxService = (*mockBoxService)(nil)

// ------------------------

// FindFreePort finds an available TCP port and returns it
func FindFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// setupTestServiceWithPage sets up playwright, starts a playwright server in the background,
// creates a BrowserService with a mock BoxService configured with the server's port,
// uses service methods to create a context and a page, then sets its content directly.
// Returns the service instance, BoxID, ContextID, PageID, the page instance, and a cleanup function.
func setupTestServiceWithPage(t *testing.T) (*service.BrowserService, string, string, string, playwright.Page, func()) {
	t.Helper()

	// --- Start Playwright Server ---
	freePort, err := FindFreePort()
	require.NoError(t, err, "Failed to find free port")
	portStr := strconv.Itoa(freePort)
	t.Logf("Starting Playwright server on port %s...", portStr)

	// Ensure npx and playwright are available in PATH
	cmd := exec.Command("npx", "playwright@1.51.1", "run-server", "--port", portStr)
	err = cmd.Start() // Start in background
	require.NoError(t, err, "Failed to start playwright run-server command")
	t.Logf("Playwright server process started (PID: %d)", cmd.Process.Pid)

	// Wait for the server port to become available
	serverAddr := fmt.Sprintf("localhost:%d", freePort)
	maxWait := 90 * time.Second
	checkInterval := 200 * time.Millisecond
	startTime := time.Now()
	portReady := false
	for time.Since(startTime) < maxWait {
		conn, err := net.DialTimeout("tcp", serverAddr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			portReady = true
			t.Logf("Playwright server port %d is ready.", freePort)
			break
		}
		time.Sleep(checkInterval)
	}
	if !portReady {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("Playwright server port %d did not become available within %v", freePort, maxWait)
	}

	// --- Service Setup ---
	mockBoxSvc := &mockBoxService{dynamicPort: freePort} // Set the dynamic port
	browserService, err := service.NewBrowserService(mockBoxSvc)
	require.NoError(t, err, "NewBrowserService failed")
	require.NotNil(t, browserService)

	// --- Use Service Methods for Setup ---
	testBoxID := "test-box-" + uuid.NewString()
	_, err = mockBoxSvc.Create(context.Background(), &boxModel.BoxCreateParams{ /* ... */ })
	require.NoError(t, err, "mockBoxSvc.Create failed indirectly")

	ctxResult, err := browserService.CreateContext(testBoxID, model.CreateContextParams{})
	require.NoError(t, err, "browserService.CreateContext failed")
	require.NotNil(t, ctxResult)
	testContextID := ctxResult.ContextID

	// Create page initially (e.g., navigating to about:blank)
	initialURL := "about:blank"
	pageResult, err := browserService.CreatePage(testBoxID, testContextID, model.CreatePageParams{
		URL: initialURL, // Start with a blank page
	})
	require.NoError(t, err, "browserService.CreatePage failed")
	require.NotNil(t, pageResult)
	testPageID := pageResult.PageID

	// Get the actual playwright page instance from the service
	pageInstance, err := browserService.GetPageInstance(testBoxID, testContextID, testPageID)
	require.NoError(t, err, "browserService.GetPageInstance failed")
	require.NotNil(t, pageInstance)

	// Now set the desired content directly
	htmlContent := `<!DOCTYPE html><html><head><title>Test Page Title</title></head><body><h1>Hello World</h1><p>This is a test paragraph.</p></body></html>`
	err = pageInstance.SetContent(htmlContent, playwright.PageSetContentOptions{WaitUntil: playwright.WaitUntilStateLoad})
	require.NoError(t, err, "pageInstance.SetContent failed")

	t.Logf("Service Setup Complete - BoxID: %s, ContextID: %s, PageID: %s", testBoxID, testContextID, testPageID)

	// Cleanup function
	cleanup := func() {
		t.Logf("Cleaning up - Page: %s, Context: %s", testPageID, testContextID)
		// Close context/page via service first
		err = browserService.ClosePage(testBoxID, testContextID, testPageID)
		if err != nil && !errors.Is(err, service.ErrPageNotFound) {
			assert.NoError(t, err, "failed to close page via service")
		}
		err = browserService.CloseContext(testBoxID, testContextID)
		if err != nil && !errors.Is(err, service.ErrContextNotFound) {
			assert.NoError(t, err, "failed to close context via service")
		}

		// Stop the background Playwright server
		t.Logf("Stopping Playwright server process (PID: %d)...", cmd.Process.Pid)
		if err := cmd.Process.Kill(); err != nil {
			t.Logf("Warning: Failed to kill playwright server process (PID: %d): %v", cmd.Process.Pid, err)
		} else {
			t.Logf("Playwright server process stopped.")
		}
		_ = cmd.Wait() // Wait for the process to release resources
	}

	// Return pageInstance instead of URL
	return browserService, testBoxID, testContextID, testPageID, pageInstance, cleanup
}

func TestGetPage(t *testing.T) {
	// Get pageInstance from setup
	svc, boxID, contextID, pageID, pageInstance, cleanup := setupTestServiceWithPage(t)
	defer cleanup()

	// Get expected URL directly from the page instance after SetContent
	expectedURL := pageInstance.URL()
	require.NotEmpty(t, expectedURL) // Should be the URL after SetContent (might still be data: or about:blank depending on PW behavior)
	expectedTitle := "Test Page Title"
	expectedHTML := `<!DOCTYPE html><html><head><title>Test Page Title</title></head><body><h1>Hello World</h1><p>This is a test paragraph.</p></body></html>`

	// Use html-to-markdown to get expected markdown
	converter := md.NewConverter("", true, nil)
	expectedMarkdown, err := converter.ConvertString(expectedHTML)
	require.NoError(t, err, "Failed to convert test HTML to Markdown")

	testCases := []struct {
		name               string
		withContent        bool
		mimeType           string // Passed to GetPage
		expectError        bool
		expectContent      *string
		expectMimeType     *string // Expected in result
		checkErrorContains string
	}{
		{
			name:           "Without Content",
			withContent:    false,
			mimeType:       "text/html", // Irrelevant when withContent=false
			expectError:    false,
			expectContent:  nil,
			expectMimeType: nil,
		},
		{
			name:           "With Content HTML",
			withContent:    true,
			mimeType:       "text/html",
			expectError:    false,
			expectContent:  &expectedHTML,
			expectMimeType: func() *string { s := "text/html"; return &s }(),
		},
		{
			name:           "With Content Markdown",
			withContent:    true,
			mimeType:       "text/markdown",
			expectError:    false,
			expectContent:  &expectedMarkdown,
			expectMimeType: func() *string { s := "text/markdown"; return &s }(),
		},
		{
			name:           "With Content Invalid MimeType",
			withContent:    true,
			mimeType:       "application/json", // Service GetPage defaults to HTML if not markdown
			expectError:    false,
			expectContent:  &expectedHTML,
			expectMimeType: func() *string { s := "text/html"; return &s }(),
		},
		{
			name:               "Page Not Found",
			withContent:        false,
			mimeType:           "text/html",
			expectError:        true,
			checkErrorContains: service.ErrPageNotFound.Error(), // Use exported error
		},
		{
			name:               "Context Not Found",
			withContent:        false,
			mimeType:           "text/html",
			expectError:        true,
			checkErrorContains: service.ErrContextNotFound.Error(), // Use exported error
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			currentBoxID := boxID
			currentContextID := contextID
			currentPageID := pageID

			if tc.name == "Page Not Found" {
				currentPageID = "non-existent-page"
			} else if tc.name == "Context Not Found" {
				currentContextID = "non-existent-context"
			}

			result, err := svc.GetPage(currentBoxID, currentContextID, currentPageID, tc.withContent, tc.mimeType)

			if tc.expectError {
				assert.Error(t, err)
				if tc.checkErrorContains != "" {
					assert.ErrorContains(t, err, tc.checkErrorContains)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)

				assert.Equal(t, currentPageID, result.PageID)
				// Check the URL returned by GetPage matches the one from the page instance
				assert.Equal(t, expectedURL, result.URL)
				assert.Equal(t, expectedTitle, result.Title)

				if tc.withContent {
					require.NotNil(t, tc.expectContent, "Test case error: expectContent cannot be nil when withContent is true")
					require.NotNil(t, tc.expectMimeType, "Test case error: expectMimeType cannot be nil when withContent is true")
					assert.NotNil(t, result.Content)
					assert.NotNil(t, result.ContentType)
					if result.Content != nil {
						// Compare content carefully, whitespace/encoding might differ slightly
						assert.Equal(t, *tc.expectContent, *result.Content)
					}
					if result.ContentType != nil {
						assert.Equal(t, *tc.expectMimeType, *result.ContentType)
					}
				} else {
					assert.Nil(t, result.Content)
					assert.Nil(t, result.ContentType)
				}
			}
		})
	}
}
