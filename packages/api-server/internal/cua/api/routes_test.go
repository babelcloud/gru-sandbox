package api_test

import (
	"testing"

	"github.com/babelcloud/gbox/packages/api-server/internal/cua/api"
	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/assert"
)

func TestRegisterCuaRoutes(t *testing.T) {
	ws := new(restful.WebService)

	// RegisterCuaRoutes expects *api.CuaHandler, but it's not exported.
	// Since the registration does not invoke the handler, we can use nil
	// as the cuaHandler argument for testing the route registration.
	api.RegisterCuaRoutes(ws, nil)

	// Verify route was added
	routes := ws.Routes()
	assert.Equal(t, 1, len(routes))

	route := routes[0]
	assert.Equal(t, "/cua/execute", route.Path)
	assert.Equal(t, "POST", route.Method)
	assert.Contains(t, route.Produces, "text/event-stream")
	assert.Contains(t, route.Produces, "application/json")

	// Verify docs
	assert.Equal(t, "execute a task using computer use agent", route.Doc)

	// Verify parameters
	assert.IsType(t, api.CuaExecuteParams{}, route.ReadSample)

	// Verify response codes
	responseErrors := make(map[int]restful.ResponseError)
	for _, resp := range route.ResponseErrors {
		responseErrors[resp.Code] = resp
	}

	assert.Contains(t, responseErrors, 200)
	assert.Contains(t, responseErrors, 400)
	assert.Contains(t, responseErrors, 500)

	// The restful.Route.Returns() sets the Message field to the second argument,
	// which for 200 is "OK" (not the third argument, which is the Model).
	// The actual SSE stream documentation is in the Model field for 200.
	assert.Equal(t, "OK", responseErrors[200].Message)
	assert.Equal(t, "SSE stream with execution progress", responseErrors[200].Model)
	assert.IsType(t, api.CuaError{}, responseErrors[400].Model)
	assert.IsType(t, api.CuaError{}, responseErrors[500].Model)
}

func TestCuaExecuteParams(t *testing.T) {
	params := api.CuaExecuteParams{
		OpenAIAPIKey: "test-key",
		Task:         "test task",
	}

	assert.Equal(t, "test-key", params.OpenAIAPIKey)
	assert.Equal(t, "test task", params.Task)

	// Test empty struct
	empty := api.CuaExecuteParams{}
	assert.Equal(t, "", empty.OpenAIAPIKey)
	assert.Equal(t, "", empty.Task)
}

func TestCuaError(t *testing.T) {
	err := api.CuaError{
		Code:    "TEST_ERROR",
		Message: "Test error message",
	}

	assert.Equal(t, "TEST_ERROR", err.Code)
	assert.Equal(t, "Test error message", err.Message)

	// Test empty struct
	empty := api.CuaError{}
	assert.Equal(t, "", empty.Code)
	assert.Equal(t, "", empty.Message)
}
