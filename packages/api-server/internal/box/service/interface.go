package service

import (
	"context"
	"fmt"
	"io"

	"github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/babelcloud/gbox/packages/api-server/internal/tracker"
)

// BoxService defines the interface for box operations
type BoxService interface {
	// Box lifecycle operations
	Create(ctx context.Context, params *model.BoxCreateParams) (*model.Box, error)
	Delete(ctx context.Context, id string, params *model.BoxDeleteParams) (*model.BoxDeleteResult, error)
	Start(ctx context.Context, id string) (*model.BoxStartResult, error)
	Stop(ctx context.Context, id string) (*model.BoxStopResult, error)
	DeleteAll(ctx context.Context, params *model.BoxesDeleteParams) (*model.BoxesDeleteResult, error)
	Reclaim(ctx context.Context) (*model.BoxReclaimResult, error)

	// Box query operations
	Get(ctx context.Context, id string) (*model.Box, error)
	List(ctx context.Context, params *model.BoxListParams) (*model.BoxListResult, error)

	// Box run command operations
	Exec(ctx context.Context, id string, params *model.BoxExecParams) (*model.BoxExecResult, error)
	Run(ctx context.Context, id string, params *model.BoxRunParams) (*model.BoxRunResult, error)

	// Box file operations
	GetArchive(ctx context.Context, id string, params *model.BoxArchiveGetParams) (*model.BoxArchiveResult, io.ReadCloser, error)
	HeadArchive(ctx context.Context, id string, params *model.BoxArchiveHeadParams) (*model.BoxArchiveHeadResult, error)
	ExtractArchive(ctx context.Context, id string, req *model.BoxArchiveExtractParams) error
}

// Factory creates a new box service instance, accepting an AccessTracker
type Factory func(tracker tracker.AccessTracker) (BoxService, error)

var implementations = make(map[string]Factory)

// Register registers a box service implementation
func Register(name string, factory Factory) {
	implementations[name] = factory
}

// New creates a new box service based on the implementation name, passing the tracker
func New(name string, tracker tracker.AccessTracker) (BoxService, error) {
	factory, ok := implementations[name]
	if !ok {
		return nil, fmt.Errorf("unknown box service implementation: %s", name)
	}
	// Pass the tracker to the factory function
	return factory(tracker)
}
