package cron

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/robfig/cron/v3"

	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/log"
	"github.com/babelcloud/gru-sandbox/packages/api-server/types"
)

// Manager manages cron jobs
type Manager struct {
	cron    *cron.Cron
	logger  *log.Logger
	handler types.BoxHandler
}

// NewManager creates a new cron manager
func NewManager(logger *log.Logger, handler types.BoxHandler) *Manager {
	return &Manager{
		cron:    cron.New(cron.WithLogger(cron.VerbosePrintfLogger(logger))),
		logger:  logger,
		handler: handler,
	}
}

// Start starts the cron manager
func (m *Manager) Start() {
	// Add reclaim job - run every 10 minutes
	_, err := m.cron.AddFunc("*/10 * * * *", m.reclaimBoxes)
	if err != nil {
		m.logger.Fatal("Failed to add reclaim job: %v", err)
	}

	m.cron.Start()
	m.logger.Info("Cron manager started")
}

// Stop stops the cron manager
func (m *Manager) Stop() {
	m.cron.Stop()
	m.logger.Info("Cron manager stopped")
}

// reclaimBoxes runs the box reclamation job
func (m *Manager) reclaimBoxes() {
	m.logger.Info("Running scheduled box reclamation")
	req := restful.NewRequest(&http.Request{})
	resp := restful.NewResponse(&discardResponseWriter{})
	m.handler.ReclaimBoxes(req, resp)
}

// discardResponseWriter implements http.ResponseWriter but discards all writes
type discardResponseWriter struct {
	header http.Header
}

func (w *discardResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *discardResponseWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (w *discardResponseWriter) WriteHeader(statusCode int) {
	// We don't need to do anything with the status code
}
