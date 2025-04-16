package model

// CreateContextParams represents the request parameters for creating a browser context.
type CreateContextParams struct {
	ViewportWidth  int      `json:"viewport_width,omitempty"`
	ViewportHeight int      `json:"viewport_height,omitempty"`
	UserAgent      string   `json:"user_agent,omitempty"`
	Locale         string   `json:"locale,omitempty"`
	Timezone       string   `json:"timezone,omitempty"`
	Permissions    []string `json:"permissions,omitempty"`
}

// CreateContextResult represents the response from creating a browser context.
type CreateContextResult struct {
	ContextID string `json:"context_id"`
}

// NewCreateContextResult creates a new CreateContextResult.
func NewCreateContextResult(contextID string) CreateContextResult {
	return CreateContextResult{
		ContextID: contextID,
	}
}
