package service_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

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
	return &boxModel.Box{ID: "mock-box-" + uuid.NewString()}, nil
}
func (m *mockBoxService) Get(ctx context.Context, boxID string) (*boxModel.Box, error) {
	return &boxModel.Box{ID: boxID, Status: "running"}, nil // Use string status
}
func (m *mockBoxService) Delete(ctx context.Context, boxID string, params *boxModel.BoxDeleteParams) (*boxModel.BoxDeleteResult, error) {
	return &boxModel.BoxDeleteResult{}, nil
}
func (m *mockBoxService) List(ctx context.Context, params *boxModel.BoxListParams) (*boxModel.BoxListResult, error) {
	return &boxModel.BoxListResult{}, nil
}
func (m *mockBoxService) DeleteAll(ctx context.Context, params *boxModel.BoxesDeleteParams) (*boxModel.BoxesDeleteResult, error) {
	return &boxModel.BoxesDeleteResult{}, nil
}
func (m *mockBoxService) Exec(ctx context.Context, boxID string, params *boxModel.BoxExecParams) (*boxModel.BoxExecResult, error) {
	return nil, fmt.Errorf("mockBoxService.Exec not implemented")
}
func (m *mockBoxService) ExtractArchive(ctx context.Context, boxID string, params *boxModel.BoxArchiveExtractParams) error {
	return nil
}
func (m *mockBoxService) GetArchive(ctx context.Context, boxID string, params *boxModel.BoxArchiveGetParams) (*boxModel.BoxArchiveResult, io.ReadCloser, error) {
	return nil, nil, fmt.Errorf("mockBoxService.GetArchive not implemented")
}
func (m *mockBoxService) GetExternalPort(ctx context.Context, boxID string, port int) (int, error) {
	if m.dynamicPort == 0 {
		return 0, fmt.Errorf("mock dynamic port not set")
	}
	return m.dynamicPort, nil
}
func (m *mockBoxService) HeadArchive(ctx context.Context, boxID string, params *boxModel.BoxArchiveHeadParams) (*boxModel.BoxArchiveHeadResult, error) {
	return nil, fmt.Errorf("mockBoxService.HeadArchive not implemented")
}
func (m *mockBoxService) Reclaim(ctx context.Context) (*boxModel.BoxReclaimResult, error) {
	return &boxModel.BoxReclaimResult{}, nil
}
func (m *mockBoxService) Start(ctx context.Context, id string) (*boxModel.BoxStartResult, error) {
	return nil, fmt.Errorf("mockBoxService.Start not implemented")
}
func (m *mockBoxService) Stop(ctx context.Context, id string) (*boxModel.BoxStopResult, error) {
	return nil, fmt.Errorf("mockBoxService.Stop not implemented")
}
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

// ------------------------

// getTestPageURL is needed by setupServiceWithVisionTestPage
// NOTE: This must remain in the same file as setupServiceWithVisionTestPage or be passed
// because runtime.Caller(0) depends on the call stack.
func getTestPageURL(t *testing.T) string {
	_, b, _, ok := runtime.Caller(0)
	require.True(t, ok, "Failed to get caller information")
	basepath := filepath.Dir(b) // Gets the directory of the current test file (test_helpers.go)

	// Construct the path relative to the current test file's directory
	htmlPath := filepath.Join(basepath, "..", "testdata", "vision-test.html")

	// Get absolute path
	absPath, err := filepath.Abs(htmlPath)
	require.NoError(t, err, "Failed to get absolute path for test HTML file")

	// Check if file exists
	_, err = os.Stat(absPath)
	require.NoError(t, err, "Test HTML file not found at: %s", absPath)

	// Convert filesystem path to file:// URL
	fileURL := url.URL{Scheme: "file", Path: filepath.ToSlash(absPath)} // Use ToSlash for cross-platform compatibility
	return fileURL.String()
}

// setupServiceWithVisionTestPage sets up a BrowserService connected to a local Playwright server
// and navigates a page within that service to the vision-test.html file.
// Returns the service instance, BoxID, ContextID, PageID, the page instance, and a cleanup function.
func setupServiceWithVisionTestPage(t *testing.T) (*service.BrowserService, string, string, string, playwright.Page, func()) {
	t.Helper()

	// --- Start Playwright Server ---
	freePort, err := FindFreePort()
	require.NoError(t, err, "Failed to find free port")
	portStr := strconv.Itoa(freePort)
	t.Logf("Starting Playwright server on port %s...", portStr)

	// Ensure npx and playwright are available in PATH
	// TODO: Make playwright version configurable or detect automatically
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
	// We need the mockBoxSvc to believe the box exists and is running for CreateContext
	_, err = mockBoxSvc.Get(context.Background(), testBoxID) // Check if Get works conceptually
	require.NoError(t, err, "mockBoxSvc.Get failed conceptually")

	ctxResult, err := browserService.CreateContext(testBoxID, model.CreateContextParams{})
	require.NoError(t, err, "browserService.CreateContext failed")
	require.NotNil(t, ctxResult)
	testContextID := ctxResult.ContextID

	// Create page and navigate it to the test HTML file
	testURL := getTestPageURL(t) // Use the local helper
	pageResult, err := browserService.CreatePage(testBoxID, testContextID, model.CreatePageParams{
		URL: testURL,
	})
	require.NoError(t, err, "browserService.CreatePage failed")
	require.NotNil(t, pageResult)
	testPageID := pageResult.PageID

	// Get the actual playwright page instance from the service
	pageInstance, err := browserService.GetPageInstance(testBoxID, testContextID, testPageID)
	require.NoError(t, err, "browserService.GetPageInstance failed")
	require.NotNil(t, pageInstance)

	// Ensure navigation completed (wait for load state)
	err = pageInstance.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateLoad,
	})
	require.NoError(t, err, "pageInstance.WaitForLoadState failed for test page")

	t.Logf("Service Setup Complete - BoxID: %s, ContextID: %s, PageID: %s Navigated to: %s", testBoxID, testContextID, testPageID, testURL)

	// Cleanup function
	cleanup := func() {
		t.Logf("Cleaning up - Page: %s, Context: %s", testPageID, testContextID)
		// Close context/page via service first
		closeErr := browserService.ClosePage(testBoxID, testContextID, testPageID)
		if closeErr != nil && !errors.Is(closeErr, service.ErrPageNotFound) {
			assert.NoError(t, closeErr, "failed to close page via service")
		}
		closeErr = browserService.CloseContext(testBoxID, testContextID)
		if closeErr != nil && !errors.Is(closeErr, service.ErrContextNotFound) {
			assert.NoError(t, closeErr, "failed to close context via service")
		}

		// Stop the background Playwright server
		t.Logf("Stopping Playwright server process (PID: %d)...", cmd.Process.Pid)
		if killErr := cmd.Process.Kill(); killErr != nil {
			// Log non-fatal error if kill fails (process might have already exited)
			if !errors.Is(killErr, os.ErrProcessDone) {
				t.Logf("Warning: Failed to kill playwright server process (PID: %d): %v", cmd.Process.Pid, killErr)
			}
		}
		_ = cmd.Wait() // Wait for the process to release resources
		t.Logf("Playwright server process stopped.")
	}

	return browserService, testBoxID, testContextID, testPageID, pageInstance, cleanup
} 