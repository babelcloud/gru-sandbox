package main

import (
	"fmt"
	"net/http"
	"os"

	restful "github.com/emicklei/go-restful/v3"

	"github.com/babelcloud/gru-sandbox/packages/api-server/config"
	"github.com/babelcloud/gru-sandbox/packages/api-server/handlers"
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

	// Get cluster configuration
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Fatal("Failed to get cluster config: %v", err)
	}

	if err := cfg.Initialize(logger); err != nil {
		logger.Fatal("%v", err)
	}

	// Initialize box handler
	boxHandler, err := handlers.InitBoxHandler(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize box handler: %v", err)
	}

	// Initialize and start cron manager
	cronManager := cron.NewManager(logger, boxHandler)
	cronManager.Start()
	defer cronManager.Stop()

	// Create REST API container
	container := restful.NewContainer()

	// Create WebService
	ws := new(restful.WebService)
	ws.Path("/api/v1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	// Add routes
	ws.Route(ws.GET("/boxes").To(boxHandler.ListBoxes).
		Doc("list all boxes").
		Returns(200, "OK", []models.Box{}))

	ws.Route(ws.GET("/boxes/{id}").To(boxHandler.GetBox).
		Doc("get a box by ID").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Returns(200, "OK", models.Box{}))

	ws.Route(ws.POST("/boxes").To(boxHandler.CreateBox).
		Doc("create a box").
		Reads(models.BoxCreateRequest{}).
		Returns(201, "Created", models.Box{}))

	ws.Route(ws.DELETE("/boxes/{id}").To(boxHandler.DeleteBox).
		Doc("delete a box").
		Reads(models.BoxDeleteRequest{}).
		Returns(200, "OK", models.BoxDeleteResponse{}))

	ws.Route(ws.DELETE("/boxes").To(boxHandler.DeleteBoxes).
		Doc("delete all boxes").
		Reads(models.BoxesDeleteRequest{}).
		Returns(200, "OK", models.BoxesDeleteResponse{}))

	ws.Route(ws.POST("/boxes/{id}/exec").To(boxHandler.ExecBox).
		Doc("execute a command in a box").
		Reads(models.BoxExecRequest{}).
		Produces(models.MediaTypeRawStream, models.MediaTypeMultiplexedStream).
		Returns(200, "OK", models.BoxExecResponse{}))

	ws.Route(ws.POST("/boxes/{id}/run").To(boxHandler.RunBox).
		Doc("run a command in a box and return output").
		Reads(models.BoxRunRequest{}).
		Returns(200, "OK", models.BoxRunResponse{}))

	ws.Route(ws.POST("/boxes/{id}/start").To(boxHandler.StartBox).
		Doc("start a stopped box").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Returns(200, "OK", nil))

	ws.Route(ws.POST("/boxes/{id}/stop").To(boxHandler.StopBox).
		Doc("stop a running box").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Returns(200, "OK", nil))

	ws.Route(ws.POST("/boxes/reclaim").To(boxHandler.ReclaimBoxes).
		Doc("reclaim inactive boxes").
		Returns(200, "OK", models.BoxReclaimResponse{}))

	container.Add(ws)

	// Add CORS filter
	cors := restful.CrossOriginResourceSharing{
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "POST", "DELETE"},
		AllowedDomains: []string{"*"},
	}
	container.Filter(cors.Filter)

	// Get server port from configuration
	port := config.GetServerPort()

	logger.Info("Server will listen on port %d", port)

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
