package model_test

import (
	"testing"

	model "github.com/babelcloud/gbox/packages/api-server/pkg/browser"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
)

func TestCreatePageParams(t *testing.T) {
	params := model.CreatePageParams{
		URL:       "https://example.com",
		WaitUntil: "networkidle",
		Timeout:   5000,
	}

	assert.Equal(t, "https://example.com", params.URL)
	assert.Equal(t, playwright.WaitUntilState("networkidle"), params.WaitUntil)
	assert.Equal(t, 5000, params.Timeout)
}

func TestNewCreatePageResult(t *testing.T) {
	result := model.NewCreatePageResult("page1", "https://example.com", "Example Page")

	assert.Equal(t, "page1", result.PageID)
	assert.Equal(t, "https://example.com", result.URL)
	assert.Equal(t, "Example Page", result.Title)
}

func TestNewListPagesResult(t *testing.T) {
	pageIDs := []string{"page1", "page2", "page3"}
	result := model.NewListPagesResult(pageIDs)

	assert.Equal(t, pageIDs, result.PageIDs)

	// Test with empty slice
	emptyResult := model.NewListPagesResult([]string{})
	assert.Empty(t, emptyResult.PageIDs)

	// Test with nil slice
	nilResult := model.NewListPagesResult(nil)
	assert.Nil(t, nilResult.PageIDs)
}

func TestGetPageResult(t *testing.T) {
	content := "<html><body>Test</body></html>"
	contentType := "text/html"

	result := model.GetPageResult{
		PageID:      "page1",
		URL:         "https://example.com",
		Title:       "Example Page",
		Content:     &content,
		ContentType: &contentType,
	}

	assert.Equal(t, "page1", result.PageID)
	assert.Equal(t, "https://example.com", result.URL)
	assert.Equal(t, "Example Page", result.Title)
	assert.Equal(t, content, *result.Content)
	assert.Equal(t, contentType, *result.ContentType)

	// Test with nil optional fields
	resultNoOptional := model.GetPageResult{
		PageID: "page2",
		URL:    "https://example.com",
		Title:  "Example Page",
	}

	assert.Equal(t, "page2", resultNoOptional.PageID)
	assert.Equal(t, "https://example.com", resultNoOptional.URL)
	assert.Equal(t, "Example Page", resultNoOptional.Title)
	assert.Nil(t, resultNoOptional.Content)
	assert.Nil(t, resultNoOptional.ContentType)
}
