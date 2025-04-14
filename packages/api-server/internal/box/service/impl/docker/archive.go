package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/docker/docker/api/types"
)

// GetArchive implements Service.GetArchive
func (s *Service) GetArchive(ctx context.Context, id string, req *model.BoxArchiveGetParams) (*model.BoxArchiveResult, io.ReadCloser, error) {
	// Update access time when getting archive
	s.accessTracker.Update(id)

	containerInfo, err := s.getContainerByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	reader, stat, err := s.client.CopyFromContainer(ctx, containerInfo.ID, req.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to copy from container: %w", err)
	}

	response := &model.BoxArchiveResult{
		Name:  stat.Name,
		Size:  stat.Size,
		Mode:  uint32(stat.Mode),
		Mtime: stat.Mtime.Format(time.RFC3339),
	}

	return response, reader, nil
}

// HeadArchive implements Service.HeadArchive
func (s *Service) HeadArchive(ctx context.Context, id string, req *model.BoxArchiveHeadParams) (*model.BoxArchiveHeadResult, error) {
	// Update access time when checking archive info
	s.accessTracker.Update(id)

	containerInfo, err := s.getContainerByID(ctx, id)
	if err != nil {
		return nil, err
	}

	stat, err := s.client.ContainerStatPath(ctx, containerInfo.ID, req.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	response := &model.BoxArchiveHeadResult{
		Name:  stat.Name,
		Size:  stat.Size,
		Mode:  uint32(stat.Mode),
		Mtime: stat.Mtime.Format(time.RFC3339),
	}

	return response, nil
}

// ExtractArchive implements Service.ExtractArchive
func (s *Service) ExtractArchive(ctx context.Context, id string, req *model.BoxArchiveExtractParams) error {
	// Update access time when putting archive
	s.accessTracker.Update(id)

	containerInfo, err := s.getContainerByID(ctx, id)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(req.Content)
	err = s.client.CopyToContainer(ctx, containerInfo.ID, req.Path, reader, types.CopyToContainerOptions{})
	if err != nil {
		return fmt.Errorf("failed to copy to container: %w", err)
	}

	return nil
}
