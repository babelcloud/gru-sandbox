package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	restful "github.com/emicklei/go-restful/v3"

	"github.com/babelcloud/gru-sandbox/packages/api-server/config"
	"github.com/babelcloud/gru-sandbox/packages/api-server/handlers"
	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/common"
	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/cron"
	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/format"
	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/log"
	"github.com/babelcloud/gru-sandbox/packages/api-server/models"
)

func main() {
	logger := log.New()

	defer func() {
		if r := recover(); r != nil {
			os.Exit(1)
		}
	}()

	// Initialize configuration
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Fatal("Failed to initialize configuration: %v", err)
	}

	// Initialize file configuration
	fileConfig := config.NewFileConfig().(*config.FileConfig)
	if err := fileConfig.Initialize(logger); err != nil {
		logger.Fatal("Failed to initialize file configuration: %v", err)
	}

	// Initialize handlers
	boxHandler, err := handlers.InitBoxHandler(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize box handler: %v", err)
	}

	fileHandler, err := handlers.NewFileHandler(fileConfig)
	if err != nil {
		logger.Fatal("Failed to initialize file handler: %v", err)
	}

	// Initialize version handler
	versionHandler := handlers.NewVersionHandler()

	// Initialize and start cron manager
	cronManager := cron.NewManager(logger, boxHandler, fileHandler)
	cronManager.Start()
	defer cronManager.Stop()

	// Create REST API container
	container := restful.NewContainer()

	// Create WebService
	ws := new(restful.WebService)
	ws.Path("/api/v1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	// Add version route
	ws.Route(ws.GET("/version").To(versionHandler.GetVersion).
		Doc("get server version information").
		Returns(200, "OK", map[string]string{}).
		Returns(500, "Internal Server Error", models.BoxError{}))

	// Add routes
	ws.Route(ws.GET("/boxes").To(boxHandler.ListBoxes).
		Doc("list all boxes").
		Returns(200, "OK", []models.Box{}).
		Returns(500, "Internal Server Error", models.BoxError{}))

	ws.Route(ws.GET("/boxes/{id}").To(boxHandler.GetBox).
		Doc("get a box by ID").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Returns(200, "OK", models.Box{}).
		Returns(404, "Not Found", models.BoxError{}).
		Returns(500, "Internal Server Error", models.BoxError{}))

	ws.Route(ws.POST("/boxes").To(boxHandler.CreateBox).
		Doc("create a box").
		Reads(models.BoxCreateRequest{}).
		Returns(201, "Created", models.Box{}).
		Returns(400, "Bad Request", models.BoxError{}).
		Returns(500, "Internal Server Error", models.BoxError{}))

	ws.Route(ws.DELETE("/boxes/{id}").To(boxHandler.DeleteBox).
		Doc("delete a box").
		Reads(models.BoxDeleteRequest{}).
		Returns(200, "OK", models.BoxDeleteResponse{}).
		Returns(404, "Not Found", models.BoxError{}).
		Returns(500, "Internal Server Error", models.BoxError{}))

	ws.Route(ws.DELETE("/boxes").To(boxHandler.DeleteBoxes).
		Doc("delete all boxes").
		Reads(models.BoxesDeleteRequest{}).
		Returns(200, "OK", models.BoxesDeleteResponse{}).
		Returns(500, "Internal Server Error", models.BoxError{}))

	ws.Route(ws.POST("/boxes/reclaim").To(boxHandler.ReclaimBoxes).
		Doc("reclaim inactive boxes").
		Returns(200, "OK", models.BoxReclaimResponse{}).
		Returns(500, "Internal Server Error", models.BoxError{}))

	ws.Route(ws.POST("/boxes/{id}/exec").To(boxHandler.ExecBox).
		Doc("execute a command in a box").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Reads(models.BoxExecRequest{}).
		Consumes(restful.MIME_JSON).
		Produces(models.MediaTypeMultiplexedStream, models.MediaTypeRawStream).
		Returns(200, "OK", models.BoxExecResponse{}).
		Returns(400, "Bad Request", models.BoxError{}).
		Returns(404, "Not Found", models.BoxError{}).
		Returns(500, "Internal Server Error", models.BoxError{}))

	ws.Route(ws.POST("/boxes/{id}/run").To(boxHandler.RunBox).
		Doc("run a command in a box").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Reads(models.BoxRunRequest{}).
		Returns(200, "OK", models.BoxRunResponse{}).
		Returns(400, "Bad Request", models.BoxError{}).
		Returns(404, "Not Found", models.BoxError{}).
		Returns(500, "Internal Server Error", models.BoxError{}))

	ws.Route(ws.POST("/boxes/{id}/start").To(boxHandler.StartBox).
		Doc("start a stopped box").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Returns(200, "OK", models.BoxStartResponse{}).
		Returns(400, "Bad Request", models.BoxError{}).
		Returns(404, "Not Found", models.BoxError{}).
		Returns(500, "Internal Server Error", models.BoxError{}))

	ws.Route(ws.POST("/boxes/{id}/stop").To(boxHandler.StopBox).
		Doc("stop a running box").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Returns(200, "OK", models.BoxStopResponse{}).
		Returns(400, "Bad Request", models.BoxError{}).
		Returns(404, "Not Found", models.BoxError{}).
		Returns(500, "Internal Server Error", models.BoxError{}))

	// Add archive routes for box file operations
	ws.Route(ws.HEAD("/boxes/{id}/archive").To(boxHandler.HeadArchive).
		Doc("get metadata about files in box").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.QueryParameter("path", "path to get metadata from").DataType("string").Required(true)).
		Returns(200, "OK", nil).
		Returns(400, "Bad Request", models.BoxError{}).
		Returns(404, "Not Found", models.BoxError{}).
		Returns(500, "Internal Server Error", models.BoxError{}))

	ws.Route(ws.GET("/boxes/{id}/archive").To(boxHandler.GetArchive).
		Doc("get files from box as tar archive").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.QueryParameter("path", "path to get files from").DataType("string").Required(true)).
		Produces("application/x-tar").
		Returns(200, "OK", nil).
		Returns(400, "Bad Request", models.BoxError{}).
		Returns(404, "Not Found", models.BoxError{}).
		Returns(500, "Internal Server Error", models.BoxError{}))

	ws.Route(ws.PUT("/boxes/{id}/archive").To(boxHandler.ExtractArchive).
		Doc("extract tar archive to box").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.QueryParameter("path", "path to extract files to").DataType("string").Required(true)).
		Consumes("application/x-tar").
		Returns(200, "OK", nil).
		Returns(400, "Bad Request", models.BoxError{}).
		Returns(404, "Not Found", models.BoxError{}).
		Returns(500, "Internal Server Error", models.BoxError{}))

	// Add box handler middleware
	ws.Filter(func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		req.SetAttribute("boxHandler", boxHandler)
		chain.ProcessFilter(req, resp)
	})

	// File routes
	ws.Route(ws.HEAD("/files/{path:*}").To(fileHandler.HeadFile).
		Doc("get file metadata").
		Param(ws.PathParameter("path", "path to the file").DataType("string")).
		Returns(200, "OK", models.FileStat{}).
		Returns(400, "Bad Request", models.FileError{}).
		Returns(404, "Not Found", models.FileError{}).
		Returns(500, "Internal Server Error", models.FileError{}))

	ws.Route(ws.GET("/files/{path:*}").To(fileHandler.GetFile).
		Doc("get file content").
		Param(ws.PathParameter("path", "path to the file").DataType("string")).
		Returns(200, "OK", nil).
		Returns(400, "Bad Request", models.FileError{}).
		Returns(404, "Not Found", models.FileError{}).
		Returns(500, "Internal Server Error", models.FileError{}))

	ws.Route(ws.POST("/files").To(fileHandler.HandleFileOperation).
		Doc("handle file operations like reclaim and share").
		Param(ws.QueryParameter("operation", "operation to perform (reclaim or share)").DataType("string").Required(true)).
		Reads(models.FileShareRequest{}).
		Returns(200, "OK", nil).
		Returns(400, "Bad Request", models.FileError{}).
		Returns(404, "Not Found", models.FileError{}).
		Returns(500, "Internal Server Error", models.FileError{}))

	container.Add(ws)

	// Add CORS filter
	cors := restful.CrossOriginResourceSharing{
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedDomains: []string{"*"},
	}
	container.Filter(cors.Filter)

	// Add request logging filter
	container.Filter(func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		logger := log.New()

		// Print request line with query parameters
		url := req.Request.URL.Path
		if req.Request.URL.RawQuery != "" {
			url += "?" + req.Request.URL.RawQuery
		}
		logger.Info("%s %s %s", req.Request.Method, url, req.Request.Proto)

		// Print headers in debug mode
		if logger.IsDebugEnabled() && len(req.Request.Header) > 0 {
			headers := make([]string, 0, len(req.Request.Header))
			for name, values := range req.Request.Header {
				headers = append(headers, fmt.Sprintf("%s: %s", name, values[0]))
			}
			logger.Debug("Headers: %s", strings.Join(headers, ", "))
		}

		// Add debug logging for request routing
		logger.Debug("Request route: %s", req.SelectedRoutePath())
		logger.Debug("Request parameters: %v", req.PathParameters())

		// Process the request
		chain.ProcessFilter(req, resp)

		// Log response status
		logger.Debug("Response status: %d", resp.StatusCode())
	})

	// Get server port from configuration
	port := config.GetServerPort()

	// Get all local IPs
	ips := common.GetLocalIPs()
	logger.Info("Server will listen on port %d", port)
	logger.Info("Accessible URLs:")
	for _, ip := range ips {
		logger.Info("  http://%s:%d", ip, port)
	}

	// Log API endpoints using WebService route information
	logger.Info("API endpoints:")
	for _, route := range ws.Routes() {
		format.LogAPIEndpoint(logger, format.APIEndpoint{
			Method:      route.Method,
			Path:        route.Path,
			Description: route.Doc,
		})
	}

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: container}
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("Server failed: %v", err)
	}
}
