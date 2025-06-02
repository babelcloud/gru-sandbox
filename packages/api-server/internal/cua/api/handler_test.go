package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/babelcloud/gbox/packages/api-server/config"
	"github.com/babelcloud/gbox/packages/api-server/internal/cua/api"
	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockResponseWriter struct {
	headers http.Header
	body    bytes.Buffer
	status  int
}

func newMockResponseWriter() *mockResponseWriter {
	return &mockResponseWriter{
		headers: make(http.Header),
	}
}

func (m *mockResponseWriter) Header() http.Header {
	return m.headers
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	return m.body.Write(b)
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.status = statusCode
}

func (m *mockResponseWriter) Flush() {}

func TestExecuteTask(t *testing.T) {
	tests := []struct {
		name           string
		setupConfig    func()
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful execution",
			setupConfig: func() {
				config.GetInstance().Cua = config.CuaServerConfig{
					Host: "localhost",
					Port: 8080,
				}
			},
			requestBody: api.CuaExecuteParams{
				OpenAIAPIKey: "test-key",
				Task:         "test task",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing parameters",
			setupConfig: func() {
				config.GetInstance().Cua = config.CuaServerConfig{
					Host: "localhost",
					Port: 8080,
				}
			},
			requestBody: api.CuaExecuteParams{
				OpenAIAPIKey: "",
				Task:         "",
			},
			expectedStatus: http.StatusOK,
			expectedError:  "MissingParameters",
		},
		{
			name: "invalid request body",
			setupConfig: func() {
				config.GetInstance().Cua = config.CuaServerConfig{
					Host: "localhost",
					Port: 8080,
				}
			},
			requestBody:    "invalid json",
			expectedStatus: http.StatusOK,
			expectedError:  "InvalidRequest",
		},
		{
			name: "missing cua server config",
			setupConfig: func() {
				config.GetInstance().Cua = config.CuaServerConfig{}
			},
			requestBody: api.CuaExecuteParams{
				OpenAIAPIKey: "test-key",
				Task:         "test task",
			},
			expectedStatus: http.StatusOK,
			expectedError:  "ConfigurationError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup config
			if tt.setupConfig != nil {
				tt.setupConfig()
			}

			// Create handler
			handler := api.NewCuaHandler()

			// Create test request
			var reqBody []byte
			var err error

			switch v := tt.requestBody.(type) {
			case string:
				reqBody = []byte(v)
			default:
				reqBody, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			httpRequest, err := http.NewRequest("POST", "/execute", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			httpRequest.Header.Set("Content-Type", "application/json")

			// Create mock response writer
			mockWriter := newMockResponseWriter()
			resp := restful.NewResponse(mockWriter)

			// Create request
			req := restful.NewRequest(httpRequest)

			// Execute handler
			handler.ExecuteTask(req, resp)

			// Verify response
			assert.Equal(t, tt.expectedStatus, mockWriter.status)

			if tt.expectedError != "" {
				assert.Contains(t, mockWriter.body.String(), tt.expectedError)
			}
		})
	}
}
