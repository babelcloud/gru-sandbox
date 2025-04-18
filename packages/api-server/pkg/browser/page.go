package model

import (
	"github.com/playwright-community/playwright-go"
)

// CreatePageParams represents the request parameters for creating a page.
type CreatePageParams struct {
	URL       string                    `json:"url"`
	WaitUntil playwright.WaitUntilState `json:"wait_until,omitempty"` // e.g., "load", "domcontentloaded", "networkidle"
	Timeout   int                       `json:"timeout,omitempty"`    // Timeout in milliseconds
}

// CreatePageResult represents the response from creating a page.
type CreatePageResult struct {
	PageID string `json:"page_id"`
	URL    string `json:"url"`
	Title  string `json:"title"`
}

// NewCreatePageResult creates a new CreatePageResult.
func NewCreatePageResult(pageID, url, title string) CreatePageResult {
	return CreatePageResult{
		PageID: pageID,
		URL:    url,
		Title:  title,
	}
}

// ListPagesResult represents the response for listing pages.
type ListPagesResult struct {
	PageIDs []string `json:"page_ids"`
}

// NewListPagesResult creates a new ListPagesResult.
func NewListPagesResult(pageIDs []string) ListPagesResult {
	return ListPagesResult{
		PageIDs: pageIDs,
	}
}

// --- Get Page --- MGHM

// GetPageResult represents the response for getting page details.
type GetPageResult struct {
	PageID      string  `json:"page_id"`
	URL         string  `json:"url"`
	Title       string  `json:"title"`
	Content     *string `json:"content,omitempty"`     // Content of the page (HTML or Markdown)
	ContentType *string `json:"contentType,omitempty"` // MIME type (e.g., text/html, text/markdown)
}
