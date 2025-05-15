package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"

	"github.com/babelcloud/gbox/packages/api-server/config"
	model "github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/babelcloud/gbox/packages/api-server/pkg/logger"
)

// Constants for image actions
const (
	ActionPull   = "pull"   // Pull the image
	ActionDelete = "delete" // Delete the image
	ActionKeep   = "keep"   // Keep the image as is
)

// DefaultImageName is the default image name used when none is specified
const DefaultImageName = "babelcloud/gbox-playwright"

// Logger instance
var log = logger.New()

// Image status tracking data structure
type imageStatus struct {
	pulling   bool
	done      chan struct{}
	succeeded bool
}

// Global image status tracker
var (
	// Mutex for protecting image status access
	imageLock sync.Mutex
	// Tracks images currently being pulled
	imageStatuses = make(map[string]*imageStatus)
)

// CheckImageExists checks if an image exists locally
func (s *Service) CheckImageExists(ctx context.Context, params *model.BoxCreateParams) (bool, string) {
	image := params.Image
	if image == "" {
		image = DefaultImageName
	}

	// Use CheckImageTag to get standardized image name
	imageWithTag := config.CheckImageTag(image)

	// Parse repository and tag
	repo, tag, ok := parseImageTag(imageWithTag)
	if !ok {
		return false, imageWithTag
	}

	// Recombine to ensure consistent format
	imageWithTag = fmt.Sprintf("%s:%s", repo, tag)

	// Try to check if image exists
	_, _, err := s.client.ImageInspectWithRaw(ctx, imageWithTag)
	return err == nil, imageWithTag
}

// EnsureImagePulling ensures an image is being pulled, starting the pull process if not already in progress
func (s *Service) EnsureImagePulling(ctx context.Context, imageName string) {
	imageLock.Lock()
	status, exists := imageStatuses[imageName]

	// If image is not being pulled, start pulling
	if !exists || !status.pulling {
		log.Infof("Starting async image pull for %s", imageName)
		newStatus := &imageStatus{
			pulling: true,
			done:    make(chan struct{}),
		}
		imageStatuses[imageName] = newStatus
		imageLock.Unlock()

		// Start async pull
		go func() {
			// Create a new context with cancel function to avoid parent context cancellation affecting the pull
			pullCtx, cancel := context.WithCancel(context.Background())
			defer cancel()

			log.Infof("Async image pull for %s started", imageName)

			// Perform image pull - no progress output as this is a background operation
			result := s.pullImage(pullCtx, imageName)

			// Update pull status
			imageLock.Lock()
			if status, ok := imageStatuses[imageName]; ok {
				status.pulling = false
				status.succeeded = result.success
				close(status.done)
			}
			imageLock.Unlock()

			if result.success {
				log.Infof("Async image pull for %s completed successfully", imageName)
			} else {
				log.Warnf("Async image pull for %s failed: %s", imageName, result.message)
			}
		}()
	} else {
		// Image already being pulled, release lock and return
		log.Infof("Image %s is already being pulled, skipping duplicate pull", imageName)
		imageLock.Unlock()
	}
}

// WaitForImagePull returns a channel that is closed when image pull completes
func (s *Service) WaitForImagePull(imageName string) <-chan struct{} {
	imageLock.Lock()
	defer imageLock.Unlock()

	status, exists := imageStatuses[imageName]
	if !exists {
		// If no pull record exists, create a closed channel to indicate completion
		log.Infof("No active pull found for image %s, returning completed channel", imageName)
		ch := make(chan struct{})
		close(ch)
		return ch
	}

	if !status.pulling {
		// Pull has already completed, check if successful
		if status.succeeded {
			log.Infof("Image %s pull already completed successfully", imageName)
		} else {
			log.Warnf("Image %s pull completed but failed", imageName)
		}
	} else {
		log.Infof("Waiting for image %s pull to complete", imageName)
	}

	return status.done
}

// prepareImageUpdate does common setup for image update operations
func (s *Service) prepareImageUpdate(ctx context.Context, params *model.ImageUpdateParams) (*model.ImageUpdateResponse, string, string, []image.Summary, error) {
	response := &model.ImageUpdateResponse{
		Images: []model.ImageInfo{},
	}

	// if imageReference is not set, use default image name
	image := params.ImageReference
	if image == "" {
		image = DefaultImageName
	}

	// use CheckImageTag to parse image tag
	imageWithTag := config.CheckImageTag(image)

	// parse repo and tag from imageWithTag
	repo, tag, ok := parseImageTag(imageWithTag)
	if !ok {
		response.ErrorMessage = fmt.Sprintf("Can't find latest tag for image %s, you may provide a image not supported by gbox or server side image tag is not set", image)
		return response, "", "", nil, nil
	}

	// First list all relevant images to ensure we don't miss any outdated ones
	filterArgs := filters.NewArgs()
	filterArgs.Add("reference", repo+":*")

	images, err := s.client.ImageList(ctx, types.ImageListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		response.ErrorMessage = fmt.Sprintf("Failed to list local images: %v", err)
		return response, "", "", nil, err
	}

	return response, repo, tag, images, nil
}

// UpdateBoxImage implements the BoxService interface method for updating images
func (s *Service) UpdateBoxImage(ctx context.Context, params *model.ImageUpdateParams) (*model.ImageUpdateResponse, error) {
	response, repo, tag, images, err := s.prepareImageUpdate(ctx, params)
	if err != nil || repo == "" {
		return response, err
	}

	imageWithTag := fmt.Sprintf("%s:%s", repo, tag)

	// Process target image
	targetImageInfo, targetImageId := s.processTargetImage(ctx, repo, tag, imageWithTag, params.DryRun)
	response.Images = append(response.Images, targetImageInfo)

	// Process existing images (find outdated versions)
	s.processOutdatedImages(ctx, images, repo, tag, targetImageId, params.DryRun, params.Force, &response.Images)

	return response, nil
}

// UpdateBoxImageWithProgress implements the BoxService interface method for updating images with progress streaming
func (s *Service) UpdateBoxImageWithProgress(ctx context.Context, params *model.ImageUpdateParams, progressWriter io.Writer) (*model.ImageUpdateResponse, error) {
	response, repo, tag, images, err := s.prepareImageUpdate(ctx, params)
	if err != nil || repo == "" {
		return response, err
	}

	imageWithTag := fmt.Sprintf("%s:%s", repo, tag)

	// Process target image with progress reporting
	targetImageInfo, targetImageId := s.processTargetImageWithProgress(ctx, repo, tag, imageWithTag, params.DryRun, progressWriter)
	response.Images = append(response.Images, targetImageInfo)

	// Process existing images (find outdated versions)
	s.processOutdatedImages(ctx, images, repo, tag, targetImageId, params.DryRun, params.Force, &response.Images)

	return response, nil
}

// processTargetImage handles checking, pulling and preparing info for the target image
func (s *Service) processTargetImage(ctx context.Context, repo string, tag string, imageWithTag string, dryRun bool) (model.ImageInfo, string) {
	return s.processTargetImageInternal(ctx, repo, tag, imageWithTag, dryRun, nil)
}

// processTargetImageWithProgress handles checking, pulling and preparing info for the target image with progress reporting
func (s *Service) processTargetImageWithProgress(ctx context.Context, repo string, tag string, imageWithTag string, dryRun bool, progressWriter io.Writer) (model.ImageInfo, string) {
	// Write initial status to client before processing
	initialStatus := model.ProgressUpdate{
		Status:  model.ProgressStatusPrepare,
		Message: fmt.Sprintf("Preparing to pull image: %s", imageWithTag),
	}
	encoder := json.NewEncoder(progressWriter)
	encoder.Encode(initialStatus)

	return s.processTargetImageInternal(ctx, repo, tag, imageWithTag, dryRun, progressWriter)
}

// processTargetImageInternal contains the shared logic between processTargetImage and processTargetImageWithProgress
func (s *Service) processTargetImageInternal(ctx context.Context, repo string, tag string, imageWithTag string, dryRun bool, progressWriter io.Writer) (model.ImageInfo, string) {
	targetImageInfo := model.ImageInfo{
		Repository: repo,
		Tag:        tag,
	}

	// Check if target image exists locally
	targetImage, _, err := s.client.ImageInspectWithRaw(ctx, imageWithTag)

	// Image exists locally
	if err == nil {
		targetImageInfo.Status = model.ImageStatusUpToDate
		targetImageInfo.ImageID = targetImage.ID
		targetImageInfo.Action = ActionKeep
		return targetImageInfo, targetImage.ID
	}

	// Image doesn't exist locally
	targetImageInfo.Status = model.ImageStatusMissing
	targetImageInfo.Action = ActionPull

	if dryRun {
		return targetImageInfo, ""
	}

	// Pull the target image with or without progress
	var pullResult pullResult
	if progressWriter != nil {
		pullResult = s.pullImageWithProgress(ctx, imageWithTag, progressWriter)
	} else {
		pullResult = s.pullImage(ctx, imageWithTag)
	}

	if pullResult.success {
		targetImageInfo.Status = model.ImageStatusUpToDate
		targetImageInfo.ImageID = pullResult.imageID
		return targetImageInfo, pullResult.imageID
	}

	return targetImageInfo, ""
}

// pullResult contains the result of an image pull operation
type pullResult struct {
	success bool
	imageID string
	message string
}

// pullImage pulls an image and returns the result
func (s *Service) pullImage(ctx context.Context, imageWithTag string) pullResult {
	return s.pullImageInternal(ctx, imageWithTag, types.ImagePullOptions{}, nil)
}

// pullImageWithProgress pulls an image and streams progress information to the writer
func (s *Service) pullImageWithProgress(ctx context.Context, imageWithTag string, progressWriter io.Writer) pullResult {
	return s.pullImageInternal(ctx, imageWithTag, types.ImagePullOptions{}, progressWriter)
}

// pullImageInternal contains the shared logic for pulling images with optional progress reporting
func (s *Service) pullImageInternal(ctx context.Context, imageWithTag string, pullOptions types.ImagePullOptions, progressWriter io.Writer) pullResult {
	reader, err := s.client.ImagePull(ctx, imageWithTag, pullOptions)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to pull: %v", err)

		// Write error to client if we have a progress writer
		if progressWriter != nil {
			errorStatus := model.ProgressUpdate{
				Status: model.ProgressStatusError,
				Error:  errMsg,
			}
			encoder := json.NewEncoder(progressWriter)
			encoder.Encode(errorStatus)
		}

		return pullResult{
			success: false,
			message: errMsg,
		}
	}
	defer reader.Close()

	// Process the pull based on whether we need to stream progress
	var processErr error
	if progressWriter != nil {
		processErr = ProcessPullProgress(reader, progressWriter)
	} else {
		_, processErr = WaitForResponse(reader)
	}

	if processErr != nil {
		// Write error to client if we have a progress writer
		if progressWriter != nil {
			errorStatus := model.ProgressUpdate{
				Status: model.ProgressStatusError,
				Error:  fmt.Sprintf("Error during pull: %v", processErr),
			}
			encoder := json.NewEncoder(progressWriter)
			encoder.Encode(errorStatus)
		}
		return pullResult{
			success: false,
			message: fmt.Sprintf("Error during pull: %v", processErr),
		}
	}

	// Try to get the ID of the pulled image
	pulledImg, _, err := s.client.ImageInspectWithRaw(ctx, imageWithTag)
	if err != nil {
		successMsg := "Successfully pulled, but couldn't retrieve image ID"

		// Send completion status if we have a progress writer
		if progressWriter != nil {
			completeStatus := model.ProgressUpdate{
				Status:  model.ProgressStatusComplete,
				Message: successMsg,
			}
			encoder := json.NewEncoder(progressWriter)
			encoder.Encode(completeStatus)
		}

		return pullResult{
			success: true,
			message: successMsg,
		}
	}

	// Send completion status with image ID if we have a progress writer
	if progressWriter != nil {
		completeStatus := model.ProgressUpdate{
			Status:  model.ProgressStatusComplete,
			Message: "Successfully pulled image",
			ImageID: pulledImg.ID,
		}
		encoder := json.NewEncoder(progressWriter)
		encoder.Encode(completeStatus)
	}

	return pullResult{
		success: true,
		imageID: pulledImg.ID,
		message: "Successfully pulled",
	}
}

// processOutdatedImages processes existing images and handles outdated ones
func (s *Service) processOutdatedImages(ctx context.Context, images []image.Summary, repo string, currentTag string, targetImageId string, dryRun bool, force bool, resultImages *[]model.ImageInfo) {
	for _, img := range images {
		// Skip target image as we already added it
		if targetImageId != "" && img.ID == targetImageId {
			continue
		}

		for _, imgTag := range img.RepoTags {
			imgRepo, imgTagValue, ok := parseImageTag(imgTag)
			if !ok {
				continue // Skip invalid tags
			}

			// Only process images with same repo but different tags
			if imgRepo == repo && imgTagValue != currentTag {
				outdatedImage := model.ImageInfo{
					ImageID:    img.ID,
					Repository: imgRepo,
					Tag:        imgTagValue,
					Status:     model.ImageStatusOutdated,
					Action:     ActionDelete,
				}

				if !dryRun {
					s.removeOutdatedImage(ctx, img.ID, &outdatedImage, force)
				}

				*resultImages = append(*resultImages, outdatedImage)
			}
		}
	}
}

// removeOutdatedImage attempts to remove an outdated image
func (s *Service) removeOutdatedImage(ctx context.Context, imageID string, imageInfo *model.ImageInfo, force bool) error {
	_, err := s.client.ImageRemove(ctx, imageID, types.ImageRemoveOptions{
		Force:         force,
		PruneChildren: true,
	})

	if err != nil {
		imageInfo.Action = ActionKeep
		return err
	}

	return nil
}

// parseImageTag parses a full image name (with tag) into repository and tag parts
func parseImageTag(imageWithTag string) (repo string, tag string, ok bool) {
	if !strings.Contains(imageWithTag, ":") {
		return "", "", false
	}

	parts := strings.Split(imageWithTag, ":")
	return parts[0], parts[1], true
}
