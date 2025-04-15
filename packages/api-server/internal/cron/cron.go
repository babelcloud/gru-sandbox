package cron

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"

	boxservice "github.com/babelcloud/gbox/packages/api-server/internal/box/service"
	fileservice "github.com/babelcloud/gbox/packages/api-server/internal/file/service"
	"github.com/babelcloud/gbox/packages/api-server/pkg/logger"
)

const (
	// Timeout for box reclamation
	boxReclaimTimeout = 5 * time.Minute
	// Timeout for file reclamation
	fileReclaimTimeout = 10 * time.Minute
)

// Manager manages cron jobs
type Manager struct {
	cron        *cron.Cron
	logger      *logger.Logger
	boxService  boxservice.BoxService
	fileService *fileservice.FileService
}

// NewManager creates a new cron manager
func NewManager(logger *logger.Logger, boxService boxservice.BoxService, fileService *fileservice.FileService) *Manager {
	return &Manager{
		cron:        cron.New(cron.WithLogger(cron.DefaultLogger)),
		logger:      logger,
		boxService:  boxService,
		fileService: fileService,
	}
}

// Start starts the cron manager
func (m *Manager) Start() {
	// Add reclaim jobs
	_, err := m.cron.AddFunc("*/10 * * * *", m.reclaimBoxes)
	if err != nil {
		m.logger.Fatal("Failed to add box reclaim job: %v", err)
	}

	// Run file reclamation daily at midnight
	_, err = m.cron.AddFunc("0 0 * * *", m.reclaimFiles)
	if err != nil {
		m.logger.Fatal("Failed to add file reclaim job: %v", err)
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
	ctx, cancel := context.WithTimeout(context.Background(), boxReclaimTimeout)
	defer cancel()

	_, err := m.boxService.Reclaim(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			m.logger.Error("Box reclamation timed out after %v", boxReclaimTimeout)
		} else {
			m.logger.Error("Failed to reclaim boxes: %v", err)
		}
	}
}

// reclaimFiles runs the file reclamation job
func (m *Manager) reclaimFiles() {
	m.logger.Info("Running scheduled file reclamation")
	ctx, cancel := context.WithTimeout(context.Background(), fileReclaimTimeout)
	defer cancel()

	_, err := m.fileService.ReclaimFiles(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			m.logger.Error("File reclamation timed out after %v", fileReclaimTimeout)
		} else {
			m.logger.Error("Failed to reclaim files: %v", err)
		}
	}
}
