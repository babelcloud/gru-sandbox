package service_test

import (
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"

	service "github.com/babelcloud/gbox/packages/api-server/internal/browser/service"
	"github.com/playwright-community/playwright-go"
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

// setupTestServiceWithPage sets up playwright, starts a playwright server in the background,
// creates a BrowserService with a mock BoxService configured with the server's port,
// uses service methods to create a context and a page, then sets its content directly.
// (Moved to test_helpers.go)
/*
func setupTestServiceWithPage(t *testing.T) (*service.BrowserService, string, string, string, playwright.Page, func()) {
	t.Helper()
// ... function body ...
	return browserService, testBoxID, testContextID, testPageID, pageInstance, cleanup
}
*/

func TestGetPage(t *testing.T) {
	// Get pageInstance from setup
	// NOTE: This now uses the shared setupServiceWithVisionTestPage from test_helpers.go
	svc, boxID, contextID, pageID, pageInstance, cleanup := setupServiceWithVisionTestPage(t)
	defer cleanup()

	// Get expected values directly from the page instance loaded by the setup function (for most cases)
	expectedURL := pageInstance.URL()
	require.NotEmpty(t, expectedURL, "Page URL should not be empty")
	// expectedTitle := "Vision Action Test Page" // Removed as it's fetched dynamically now

	// Define simple HTML for specific markdown test
	simpleHTMLForMarkdown := `<h2>Test Header</h2><p>Test para.</p>`
	converter := md.NewConverter("", true, nil)
	expectedMarkdown, err := converter.ConvertString(simpleHTMLForMarkdown)
	require.NoError(t, err, "Failed to convert simple test HTML to Markdown")

	testCases := []struct {
		name               string
		withContent        bool
		mimeType           string // Passed to GetPage
		isMarkdownTest     bool   // Flag to indicate special handling
		expectError        bool
		expectContent      *string // Only used for markdown test case
		expectMimeType     *string // Expected in result if withContent=true
		checkErrorContains string
	}{
		{
			name:           "Without Content",
			withContent:    false,
			mimeType:       "text/html",
			isMarkdownTest: false,
			expectError:    false,
			expectContent:  nil,
			expectMimeType: nil,
		},
		{
			name:           "With Content HTML", // This will get content from vision-test.html
			withContent:    true,
			mimeType:       "text/html",
			isMarkdownTest: false,
			expectError:    false,
			expectContent:  nil, // Not checking exact HTML content here
			expectMimeType: func() *string { s := "text/html"; return &s }(),
		},
		{
			name:           "With Content Markdown", // This will set its own content
			withContent:    true,
			mimeType:       "text/markdown",
			isMarkdownTest: true,
			expectError:    false,
			expectContent:  &expectedMarkdown,
			expectMimeType: func() *string { s := "text/markdown"; return &s }(),
		},
		{
			name:           "With Content Invalid MimeType", // This will get content from vision-test.html
			withContent:    true,
			mimeType:       "application/json",
			isMarkdownTest: false,
			expectError:    false,
			expectContent:  nil, // Not checking exact HTML content here
			expectMimeType: func() *string { s := "text/html"; return &s }(),
		},
		{
			name:               "Page Not Found",
			withContent:        false,
			mimeType:           "text/html",
			isMarkdownTest:     false,
			expectError:        true,
			checkErrorContains: service.ErrPageNotFound.Error(),
		},
		{
			name:               "Context Not Found",
			withContent:        false,
			mimeType:           "text/html",
			isMarkdownTest:     false,
			expectError:        true,
			checkErrorContains: service.ErrContextNotFound.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			currentBoxID := boxID
			currentContextID := contextID
			currentPageID := pageID
			currentPageInstance := pageInstance // Use the instance from setup

			if tc.name == "Page Not Found" {
				currentPageID = "non-existent-page"
			} else if tc.name == "Context Not Found" {
				currentContextID = "non-existent-context"
			}

			// Special handling for the markdown test case
			if tc.isMarkdownTest {
				t.Logf("Setting specific HTML content for markdown test...")
				err := currentPageInstance.SetContent(simpleHTMLForMarkdown, playwright.PageSetContentOptions{WaitUntil: playwright.WaitUntilStateLoad})
				require.NoError(t, err, "Failed to set content for markdown test")
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
				// Check URL and Title (Title might change if content was set specifically for markdown test)
				if tc.isMarkdownTest {
					// Title might be empty or different after SetContent, don't assert expectedTitle
					currentActualURL := currentPageInstance.URL() // URL might be 'about:blank' or similar after SetContent
					assert.Equal(t, currentActualURL, result.URL)
				} else {
					// Fetch current title directly before asserting
					currentActualTitle, titleErr := currentPageInstance.Title()
					assert.NoError(t, titleErr, "Failed to get page title during assertion")
					assert.Equal(t, expectedURL, result.URL)
					assert.Equal(t, currentActualTitle, result.Title)
				}

				if tc.withContent {
					require.NotNil(t, tc.expectMimeType, "Test case error: expectMimeType cannot be nil when withContent is true")
					assert.NotNil(t, result.Content)
					assert.NotEmpty(t, *result.Content, "Content should not be empty when requested")
					assert.NotNil(t, result.ContentType)
					if result.ContentType != nil {
						assert.Equal(t, *tc.expectMimeType, *result.ContentType)
					}
					// Only check exact content for the specific markdown test case
					if tc.isMarkdownTest {
						require.NotNil(t, tc.expectContent, "Test case error: expectContent cannot be nil for markdown test")
						assert.Equal(t, *tc.expectContent, *result.Content)
					}
				} else {
					assert.Nil(t, result.Content)
					assert.Nil(t, result.ContentType)
				}
			}
		})
	}
}
