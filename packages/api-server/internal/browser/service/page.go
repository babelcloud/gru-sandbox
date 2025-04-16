package service

import (
	"fmt"

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

// getCurrentPages gets the state of all currently open pages within a managed context.
func (s *BrowserService) getCurrentPages(mc *ManagedContext) []model.Page { // Renamed function and return type
	// Ensure mc is not nil before proceeding
	if mc == nil {
		// Consider logging this case if it shouldn't happen
		return []model.Page{}
	}

	mc.mu.RLock() // Read-lock the context's page map
	defer mc.mu.RUnlock()

	// Determine the active page ID safely
	activePageID := ""
	if mc.ActivePage != nil { // Assuming ActivePage field exists on managedContext
		activePageID = mc.ActivePage.ID
	}

	pages := make([]model.Page, 0, len(mc.Pages))

	for pageID, mp := range mc.Pages { // Iterate using ID as well for logging clarity
		// Ensure the managed page and its Playwright instance are valid and open
		if mp != nil && mp.Instance != nil && !mp.Instance.IsClosed() {
			pwPage := mp.Instance

			title, err := pwPage.Title()
			if err != nil {
				fmt.Printf("WARN: Failed to get title for page %s: %v. Using empty title.\n", pageID, err)
				title = "" // Use default value on error
			}

			url := pwPage.URL()
			// URL() in playwright-go v0.4.0+ doesn't return an error, but checking defensively
			// if url == "" {
			//    fmt.Printf("WARN: Got empty URL for page %s\n", pageID)
			// }

			isActive := mp.ID == activePageID

			pages = append(pages, model.Page{
				Title:   title,
				URL:     url,
				Favicon: "", // Placeholder - Getting favicon requires extra JS evaluation
				State: model.PageState{
					Loading: false, // Placeholder - Reliably getting loading state is complex
					Active:  isActive,
				},
			})
		} else {
			// Optional: Log if a page in the map is nil or closed unexpectedly
			fmt.Printf("DEBUG: Skipping nil or closed page entry with ID %s in context %s\n", pageID, mc.ID)
		}
	}
	return pages
}
