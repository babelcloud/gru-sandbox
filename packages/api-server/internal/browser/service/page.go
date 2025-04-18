package service

import (
	"errors"
	"fmt"

	// "strings" // Removed as unused after merge

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"

	model "github.com/babelcloud/gbox/packages/api-server/pkg/browser"
)

// findManagedPage locates a ManagedPage by its ID using the global index.
func (s *BrowserService) findManagedPage(pageID string) (*ManagedPage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	mp, exists := s.pageMap[pageID]
	if !exists {
		return nil, ErrPageNotFound
	}
	return mp, nil
}

// CreatePage creates a new page, stores it, and returns its UUID.
func (s *BrowserService) CreatePage(boxID, contextID string, params model.CreatePageParams) (*model.CreatePageResult, error) {
	mc, err := s.findManagedContext(contextID)
	if err != nil {
		return nil, err
	}
	if mc.ParentBrowser.BoxID != boxID {
		return nil, fmt.Errorf("context %s does not belong to box %s", contextID, boxID)
	}

	newPageInstance, err := mc.Instance.NewPage()
	if err != nil {
		return nil, fmt.Errorf("failed to create Playwright page in context %s: %w", contextID, err)
	}

	gotoOpts := playwright.PageGotoOptions{}
	if params.Timeout > 0 {
		gotoOpts.Timeout = playwright.Float(float64(params.Timeout))
	}
	if params.WaitUntil != "" {
		gotoOpts.WaitUntil = &params.WaitUntil
	}

	_, err = newPageInstance.Goto(params.URL, gotoOpts)
	if err != nil {
		newPageInstance.Close() // Close the page if navigation fails
		return nil, fmt.Errorf("failed to navigate page to %s: %w", params.URL, err)
	}
	title, _ := newPageInstance.Title()

	// Create and store managed page
	pageID := uuid.New().String()
	mp := &ManagedPage{
		ID:            pageID,
		Instance:      newPageInstance,
		ParentContext: mc,
	}

	s.mu.Lock()  // Lock global pageMap
	mc.mu.Lock() // Lock context's Pages map
	s.pageMap[pageID] = mp
	mc.Pages[pageID] = mp
	mc.mu.Unlock()
	s.mu.Unlock()

	// Attach listener for page close
	newPageInstance.Once("close", func() {
		fmt.Printf("INFO: Page %s close event\n", pageID)
		s.handlePageClose(mp)
	})

	result := model.NewCreatePageResult(pageID, params.URL, title)
	return &result, nil
}

// handlePageClose removes the ManagedPage from its parent context and the global map.
func (s *BrowserService) handlePageClose(mp *ManagedPage) {
	if mp == nil || mp.ParentContext == nil {
		fmt.Printf("WARN: handlePageClose called with nil page or parent context\n")
		return
	}
	parentContext := mp.ParentContext

	s.mu.Lock()             // Lock global map
	parentContext.mu.Lock() // Lock context map

	delete(s.pageMap, mp.ID)           // Remove from global index
	delete(parentContext.Pages, mp.ID) // Remove from parent context

	parentContext.mu.Unlock()
	s.mu.Unlock()
	fmt.Printf("INFO: Removed page %s from context %s\n", mp.ID, parentContext.ID)
}

// ListPages lists all managed pages in a specific context.
func (s *BrowserService) ListPages(boxID, contextID string) (*model.ListPagesResult, error) {
	mc, err := s.findManagedContext(contextID)
	if err != nil {
		return nil, err
	}
	if mc.ParentBrowser.BoxID != boxID {
		return nil, fmt.Errorf("context %s does not belong to box %s", contextID, boxID)
	}

	mc.mu.RLock() // Read lock context's page map
	defer mc.mu.RUnlock()

	pageIDs := make([]string, 0, len(mc.Pages))
	for pageID := range mc.Pages {
		pageIDs = append(pageIDs, pageID)
	}

	result := model.NewListPagesResult(pageIDs)
	return &result, nil
}

// ClosePage closes a specific managed page.
func (s *BrowserService) ClosePage(boxID, contextID, pageID string) error {
	mp, err := s.findManagedPage(pageID)
	if err != nil {
		return err // Includes ErrPageNotFound
	}

	// Verify ownership
	if mp.ParentContext == nil || mp.ParentContext.ID != contextID || mp.ParentContext.ParentBrowser.BoxID != boxID {
		return fmt.Errorf("page %s does not belong to context %s or box %s", pageID, contextID, boxID)
	}

	// Close the underlying Playwright page. The listener will handle map removal.
	if err := mp.Instance.Close(); err != nil {
		fmt.Printf("WARN: Failed to explicitly close page %s: %v\n", pageID, err)
		s.handlePageClose(mp) // Force map cleanup
		return fmt.Errorf("failed to close page %s: %w", pageID, err)
	}
	return nil
}

// GetPage retrieves details about a specific page, optionally including its content.
func (s *BrowserService) GetPage(boxID, contextID, pageID string, withContent bool, mimeType string) (*model.GetPageResult, error) {
	targetPage, err := s.GetPageInstance(boxID, contextID, pageID)
	if err != nil {
		// Wrap error for clarity if not found
		if errors.Is(err, ErrPageNotFound) || errors.Is(err, ErrContextNotFound) || errors.Is(err, ErrBoxNotFound) {
			return nil, fmt.Errorf("failed to get page instance: %w", err)
		}
		return nil, fmt.Errorf("internal error getting page instance: %w", err)
	}

	// Get URL (assuming it doesn't return an error, based on playwright-go common patterns)
	pageURL := targetPage.URL()

	// Get Title
	pageTitle, err := targetPage.Title()
	if err != nil {
		// Title() *can* return an error according to playwright-go docs, handle it
		pageTitle = "<error retrieving title>"
		fmt.Printf("Warning: Failed to get title for page %s: %v\n", pageID, err)
	}

	result := &model.GetPageResult{
		PageID: pageID,
		URL:    pageURL,
		Title:  pageTitle,
	}

	if withContent {
		// Get raw HTML content
		htmlContent, err := targetPage.Content()
		if err != nil {
			return nil, fmt.Errorf("failed to get page content for %s: %w", pageID, err)
		}

		finalContent := htmlContent
		finalContentType := "text/html" // Default to HTML initially

		// Convert to Markdown if requested
		if mimeType == "text/markdown" {
			converter := md.NewConverter("", true, nil)
			markdownContent, err := converter.ConvertString(htmlContent)
			if err != nil {
				// Log the error and return an error indicating conversion failure
				fmt.Printf("Error converting HTML to Markdown for page %s: %v\n", pageID, err)
				return nil, fmt.Errorf("markdown conversion failed for page %s: %w", pageID, err)
			}
			finalContent = markdownContent
			finalContentType = "text/markdown"
		}

		result.Content = &finalContent
		result.ContentType = &finalContentType
	}

	return result, nil
}

// --- Vision Actions --- (Placeholder - these would be moved here too if implemented)
