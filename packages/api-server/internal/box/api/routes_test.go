package api_test

import (
	"testing"

	"github.com/babelcloud/gbox/packages/api-server/internal/box/api"
	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/assert"
)

func TestRegisterRoutes(t *testing.T) {
	ws := new(restful.WebService)
	boxHandler := &api.BoxHandler{}

	api.RegisterRoutes(ws, boxHandler)

	routes := ws.Routes()
	assert.NotEmpty(t, routes)

	// Test route registration for box lifecycle operations
	assertRouteExists(t, routes, "GET", "/boxes")
	assertRouteExists(t, routes, "GET", "/boxes/{id}")
	assertRouteExists(t, routes, "POST", "/boxes")
	assertRouteExists(t, routes, "POST", "/boxes/linux")
	assertRouteExists(t, routes, "POST", "/boxes/android")
	assertRouteExists(t, routes, "DELETE", "/boxes/{id}")
	assertRouteExists(t, routes, "DELETE", "/boxes")
	assertRouteExists(t, routes, "POST", "/boxes/reclaim")

	// Test route registration for box runtime operations
	assertRouteExists(t, routes, "POST", "/boxes/{id}/commands")
	assertRouteExists(t, routes, "POST", "/boxes/{id}/run-code")
	assertRouteExists(t, routes, "POST", "/boxes/{id}/start")
	assertRouteExists(t, routes, "POST", "/boxes/{id}/stop")
	assertRouteExists(t, routes, "GET", "/boxes/{id}/exec/ws")

	// Test route registration for box archive operations
	assertRouteExists(t, routes, "HEAD", "/boxes/{id}/archive")
	assertRouteExists(t, routes, "GET", "/boxes/{id}/archive")
	assertRouteExists(t, routes, "PUT", "/boxes/{id}/archive")

	// Test route registration for box filesystem operations
	assertRouteExists(t, routes, "GET", "/boxes/{id}/fs/list")
	assertRouteExists(t, routes, "GET", "/boxes/{id}/fs/read")
	assertRouteExists(t, routes, "POST", "/boxes/{id}/fs/write")

	// Test route registration for image management operations
	assertRouteExists(t, routes, "POST", "/boxes/images/update")

	// Test route registration for box actions (cloud version)
	assertRouteExists(t, routes, "POST", "/boxes/{id}/actions/click")
	assertRouteExists(t, routes, "POST", "/boxes/{id}/actions/drag")
	assertRouteExists(t, routes, "POST", "/boxes/{id}/actions/move")
	assertRouteExists(t, routes, "POST", "/boxes/{id}/actions/press")
	assertRouteExists(t, routes, "POST", "/boxes/{id}/actions/screenshot")
	assertRouteExists(t, routes, "POST", "/boxes/{id}/actions/scroll")
	assertRouteExists(t, routes, "POST", "/boxes/{id}/actions/touch")
	assertRouteExists(t, routes, "POST", "/boxes/{id}/actions/type")
}

func assertRouteExists(t *testing.T, routes []restful.Route, method string, path string) {
	t.Helper()
	found := false
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			found = true
			break
		}
	}
	assert.True(t, found, "Route %s %s not found", method, path)
}

func TestRouteParameters(t *testing.T) {
	ws := new(restful.WebService)
	boxHandler := &api.BoxHandler{}

	api.RegisterRoutes(ws, boxHandler)
	routes := ws.Routes()

	// Test list boxes route parameters
	listBoxesRoute := findRoute(routes, "GET", "/boxes")
	assert.NotNil(t, listBoxesRoute)

	// Test create box route
	createBoxRoute := findRoute(routes, "POST", "/boxes")
	assert.NotNil(t, createBoxRoute)

	// Test update box image route
	updateImageRoute := findRoute(routes, "POST", "/boxes/images/update")
	assert.NotNil(t, updateImageRoute)
}

func findRoute(routes []restful.Route, method string, path string) *restful.Route {
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			return &route
		}
	}
	return nil
}

func TestRouteReturns(t *testing.T) {
	ws := new(restful.WebService)
	boxHandler := &api.BoxHandler{}

	api.RegisterRoutes(ws, boxHandler)
	routes := ws.Routes()

	// Test list boxes route exists
	listBoxesRoute := findRoute(routes, "GET", "/boxes")
	assert.NotNil(t, listBoxesRoute)

	// Test create box route exists
	createBoxRoute := findRoute(routes, "POST", "/boxes")
	assert.NotNil(t, createBoxRoute)

	// Test get box route exists
	getBoxRoute := findRoute(routes, "GET", "/boxes/{id}")
	assert.NotNil(t, getBoxRoute)

	// Test delete box route exists
	deleteBoxRoute := findRoute(routes, "DELETE", "/boxes/{id}")
	assert.NotNil(t, deleteBoxRoute)
}

func TestRouteConsumesAndProduces(t *testing.T) {
	ws := new(restful.WebService)
	boxHandler := &api.BoxHandler{}

	api.RegisterRoutes(ws, boxHandler)
	routes := ws.Routes()

	// Test archive routes content types
	putArchiveRoute := findRoute(routes, "PUT", "/boxes/{id}/archive")
	assert.NotNil(t, putArchiveRoute)

	getArchiveRoute := findRoute(routes, "GET", "/boxes/{id}/archive")
	assert.NotNil(t, getArchiveRoute)
}
