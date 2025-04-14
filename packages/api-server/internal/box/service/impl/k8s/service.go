package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/babelcloud/gbox/packages/api-server/config"
	"github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/babelcloud/gbox/packages/api-server/internal/box/service"
	"github.com/babelcloud/gbox/packages/api-server/internal/tracker"
	"github.com/babelcloud/gbox/packages/api-server/pkg/id"
	"github.com/babelcloud/gbox/packages/api-server/pkg/logger"
)

const (
	tenantNamespace = "gbox-tenant"
	defaultImage    = "bash:latest"

	// Label keys following Kubernetes recommended labels
	labelPrefix    = "app.kubernetes.io"
	labelName      = labelPrefix + "/name"       // The name of the application
	labelInstance  = labelPrefix + "/instance"   // A unique name identifying the instance of an application
	labelVersion   = labelPrefix + "/version"    // The current version of the application
	labelComponent = labelPrefix + "/component"  // The component within the architecture
	labelPartOf    = labelPrefix + "/part-of"    // The name of a higher level application this one is part of
	labelManagedBy = labelPrefix + "/managed-by" // The tool being used to manage the operation of an application

	// Annotation keys
	annotationPrefix  = "gbox.gru.ai"
	annotationCmd     = annotationPrefix + "/cmd"
	annotationArgs    = annotationPrefix + "/args"
	annotationWorkDir = annotationPrefix + "/working-dir"
)

// Service implements the box service interface using Kubernetes
type Service struct {
	client        *kubernetes.Clientset
	config        *rest.Config
	logger        *logger.Logger
	accessTracker tracker.AccessTracker
}

// NewService creates a new Kubernetes service instance
func NewService(tracker tracker.AccessTracker) (*Service, error) {
	log := logger.New()
	cfg := config.GetInstance()
	kubeConfig := cfg.Cluster.K8s.Config

	log.Info("Initializing Kubernetes service with config: %s", kubeConfig)

	// Build Kubernetes config from kubeconfig file
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		log.Error("Failed to build Kubernetes config: %v", err)
		return nil, fmt.Errorf("failed to build Kubernetes config: %v", err)
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Error("Failed to create Kubernetes client: %v", err)
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	log.Info("Kubernetes service initialized successfully")
	return &Service{
		client:        client,
		config:        restConfig,
		logger:        log,
		accessTracker: tracker,
	}, nil
}

// List returns all boxes
func (s *Service) List(ctx context.Context, params *model.BoxListParams) (*model.BoxListResult, error) {
	s.logger.Debug("Listing all boxes in namespace: %s", tenantNamespace)

	deployments, err := s.client.AppsV1().Deployments(tenantNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelName + "=gbox",
	})
	if err != nil {
		s.logger.Error("Failed to list deployments: %v", err)
		return nil, fmt.Errorf("failed to list deployments: %v", err)
	}

	boxes := make([]model.Box, 0)
	for _, deployment := range deployments.Items {
		boxes = append(boxes, model.Box{
			ID:     deployment.Labels[labelInstance],
			Image:  deployment.Spec.Template.Spec.Containers[0].Image,
			Status: string(deployment.Status.AvailableReplicas),
		})
	}

	s.logger.Debug("Found %d boxes", len(boxes))
	return &model.BoxListResult{
		Boxes: boxes,
		Count: len(boxes),
	}, nil
}

// Create creates a new box
func (s *Service) Create(ctx context.Context, req *model.BoxCreateParams) (*model.Box, error) {
	boxID := id.GenerateBoxID()
	s.logger.Info("Creating new box with ID: %s", boxID)
	s.accessTracker.Update(boxID)

	labels := map[string]string{
		labelName:      "gbox",           // The application name
		labelInstance:  boxID,            // Unique instance identifier
		labelVersion:   "v1",             // Version of the box
		labelComponent: "sandbox",        // Component type
		labelPartOf:    "gru-sandbox",    // Part of the gru-sandbox system
		labelManagedBy: "gru-api-server", // Managed by this API server
	}

	// Prepare annotations
	annotations := map[string]string{}

	// Add shell configuration to annotations
	if req.Cmd != "" {
		annotations[annotationCmd] = req.Cmd
	}
	if len(req.Args) > 0 {
		annotations[annotationArgs] = joinArgs(req.Args)
	}
	if req.WorkingDir != "" {
		annotations[annotationWorkDir] = req.WorkingDir
	}

	// Create deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        boxID,
			Namespace:   tenantNamespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					labelName:     "gbox",
					labelInstance: boxID,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:       "box",
							Image:      getImage(req.Image),
							Command:    []string{req.Cmd},
							Args:       req.Args,
							Env:        getEnvVars(req.Env),
							WorkingDir: req.WorkingDir,
						},
					},
					ImagePullSecrets: getImagePullSecrets(req.ImagePullSecret),
				},
			},
		},
	}

	result, err := s.client.AppsV1().Deployments(tenantNamespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		s.logger.Error("Failed to create deployment: %v", err)
		return nil, fmt.Errorf("failed to create deployment: %v", err)
	}

	s.logger.Info("Box created successfully with ID: %s", boxID)
	return &model.Box{
		ID:     boxID,
		Image:  req.Image,
		Status: string(result.Status.AvailableReplicas),
	}, nil
}

// Delete deletes a box by ID
func (s *Service) Delete(ctx context.Context, id string, req *model.BoxDeleteParams) (*model.BoxDeleteResult, error) {
	if id == "" {
		return nil, fmt.Errorf("box ID is required")
	}
	s.accessTracker.Remove(id)

	err := s.client.AppsV1().Deployments(tenantNamespace).Delete(ctx, id, metav1.DeleteOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to delete deployment: %v", err)
	}

	return &model.BoxDeleteResult{
		Message: "Box deleted successfully",
	}, nil
}

// DeleteAll deletes all boxes
func (s *Service) DeleteAll(ctx context.Context, req *model.BoxesDeleteParams) (*model.BoxesDeleteResult, error) {
	// List all deployments with gbox label
	deployments, err := s.client.AppsV1().Deployments(tenantNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelName + "=gbox",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %v", err)
	}

	// Delete all deployments
	var deletedIDs []string
	for _, deployment := range deployments.Items {
		err := s.client.AppsV1().Deployments(tenantNamespace).Delete(ctx, deployment.Name, metav1.DeleteOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to delete deployment %s: %v", deployment.Name, err)
		}
		deletedIDs = append(deletedIDs, deployment.Labels[labelInstance])
		s.accessTracker.Remove(deployment.Labels[labelInstance])
	}

	return &model.BoxesDeleteResult{
		Count:   len(deletedIDs),
		Message: "Boxes deleted successfully",
		IDs:     deletedIDs,
	}, nil
}

// Get returns a box by ID
func (s *Service) Get(ctx context.Context, id string) (*model.Box, error) {
	if id == "" {
		return nil, fmt.Errorf("box ID is required")
	}
	s.accessTracker.Update(id)

	// Get pod details
	pod, err := s.client.CoreV1().Pods(tenantNamespace).Get(ctx, id, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("box not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get box: %v", err)
	}

	// Map pod status to box status
	var status string
	switch pod.Status.Phase {
	case corev1.PodRunning:
		status = "running"
	case corev1.PodPending:
		status = "pending"
	case corev1.PodFailed:
		status = "failed"
	case corev1.PodSucceeded:
		status = "succeeded"
	default:
		status = "unknown"
	}

	// Create box model
	return &model.Box{
		ID:     id,
		Status: status,
		Image:  pod.Spec.Containers[0].Image,
	}, nil
}

// Exec executes a command in a box
func (s *Service) Exec(ctx context.Context, id string, req *model.BoxExecParams) (*model.BoxExecResult, error) {
	if id == "" {
		return nil, fmt.Errorf("box ID is required")
	}
	s.accessTracker.Update(id)

	// Get the pod name for the deployment
	pods, err := s.client.CoreV1().Pods(tenantNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=gbox,%s=%s", labelName, labelInstance, id),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("box not found: %s", id)
	}

	pod := pods.Items[0]
	if pod.Status.Phase != corev1.PodRunning {
		return nil, fmt.Errorf("box is not running: %s", id)
	}

	// Create remote command executor
	execURL, err := url.Parse(fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/exec", tenantNamespace, pod.Name))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v", err)
	}

	// Get the REST config from the client
	exec, err := remotecommand.NewSPDYExecutor(s.config, "POST", execURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %v", err)
	}

	// Create stream options
	streamOptions := remotecommand.StreamOptions{
		Stdin:             req.Conn,
		Stdout:            req.Conn,
		Stderr:            req.Conn,
		TerminalSizeQueue: nil, // We don't need terminal size queue for now
		Tty:               req.TTY,
	}

	// Start streaming with context
	err = exec.Stream(streamOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to stream: %v", err)
	}

	return &model.BoxExecResult{
		ExitCode: 0, // TODO: Get actual exit code from pod
	}, nil
}

// Run runs a command in a box
func (s *Service) Run(ctx context.Context, id string, req *model.BoxRunParams) (*model.BoxRunResult, error) {
	// TODO: Implement run operation for K8s
	return nil, fmt.Errorf("run operation not implemented for K8s")
}

// Start starts a stopped box
func (s *Service) Start(ctx context.Context, id string) (*model.BoxStartResult, error) {
	// TODO: Implement Kubernetes pod start
	return nil, fmt.Errorf("Kubernetes start not implemented")
}

// Stop stops a running box
func (s *Service) Stop(ctx context.Context, id string) (*model.BoxStopResult, error) {
	// TODO: Implement Kubernetes pod stop
	return nil, fmt.Errorf("Kubernetes stop not implemented")
}

// Reclaim reclaims inactive boxes
func (s *Service) Reclaim(ctx context.Context) (*model.BoxReclaimResult, error) {
	// TODO: Implement Kubernetes box reclamation
	return nil, fmt.Errorf("Kubernetes box reclamation not implemented")
}

// GetArchive gets files from box as tar archive
func (s *Service) GetArchive(ctx context.Context, id string, req *model.BoxArchiveGetParams) (*model.BoxArchiveResult, io.ReadCloser, error) {
	if req.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}
	s.accessTracker.Update(id)

	// Get the pod name for the deployment
	pods, err := s.client.CoreV1().Pods(tenantNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=gbox,%s=%s", labelName, labelInstance, id),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list pods: %v", err)
	}

	if len(pods.Items) == 0 {
		return nil, nil, fmt.Errorf("box not found: %s", id)
	}

	pod := pods.Items[0]
	if pod.Status.Phase != corev1.PodRunning {
		return nil, nil, fmt.Errorf("box is not running: %s", id)
	}

	// Create command to get file/directory
	cmd := []string{"tar", "czf", "-", req.Path}
	exec, err := s.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(tenantNamespace).
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: cmd,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec).
		Stream(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create exec: %v", err)
	}

	return &model.BoxArchiveResult{
		Name:  req.Path,
		Size:  0, // TODO: Get actual size
		Mode:  0644,
		Mtime: time.Now().Format(time.RFC3339),
	}, exec, nil
}

// BoxArchiveHeadResult represents the result of a box archive head operation
type BoxArchiveHeadResult struct {
	Mode uint32 `json:"mode"` // File mode
	Size int64  `json:"size"` // File size
}

// ArchiveHead returns the metadata for a file in a box
func (s *Service) ArchiveHead(ctx context.Context, boxID string, path string) (*BoxArchiveHeadResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}
	s.accessTracker.Update(boxID)

	// Get the pod name for the deployment
	pods, err := s.client.CoreV1().Pods(tenantNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=gbox,%s=%s", labelName, labelInstance, boxID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("box not found: %s", boxID)
	}

	pod := pods.Items[0]
	if pod.Status.Phase != corev1.PodRunning {
		return nil, fmt.Errorf("box is not running: %s", boxID)
	}

	// Create command to get file/directory metadata
	cmd := []string{"stat", "-f", "%N:%z:%m:%a:%U:%G", path}
	exec, err := s.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(tenantNamespace).
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: cmd,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec).
		Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %v", err)
	}
	defer exec.Close()

	// Read the output
	output, err := io.ReadAll(exec)
	if err != nil {
		return nil, fmt.Errorf("failed to read exec output: %v", err)
	}

	// Parse the output
	parts := strings.Split(string(output), ":")
	if len(parts) != 6 {
		return nil, fmt.Errorf("invalid stat output: %s", string(output))
	}

	// Convert mode to uint32
	mode, err := strconv.ParseUint(parts[3], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mode: %v", err)
	}

	return &BoxArchiveHeadResult{
		Mode: uint32(mode),
		Size: parseInt64(parts[1]),
	}, nil
}

// ExtractArchive extracts tar archive to box
func (s *Service) ExtractArchive(ctx context.Context, id string, req *model.BoxArchiveExtractParams) error {
	if req.Path == "" {
		return fmt.Errorf("path is required")
	}
	s.accessTracker.Update(id)

	// Get the pod name for the deployment
	pods, err := s.client.CoreV1().Pods(tenantNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=gbox,%s=%s", labelName, labelInstance, id),
	})
	if err != nil {
		return fmt.Errorf("failed to list pods: %v", err)
	}

	if len(pods.Items) == 0 {
		return fmt.Errorf("box not found: %s", id)
	}

	pod := pods.Items[0]
	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("box is not running: %s", id)
	}

	// Create command to extract archive
	cmd := []string{"tar", "xzf", "-", "-C", req.Path}
	exec, err := s.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(tenantNamespace).
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: cmd,
			Stdin:   true,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec).
		Stream(ctx)
	if err != nil {
		return fmt.Errorf("failed to create exec: %v", err)
	}
	defer exec.Close()

	// Create a buffer to store the output
	var stdout, stderr bytes.Buffer

	// Create remote command executor
	executor, err := remotecommand.NewSPDYExecutor(s.config, "POST", s.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(tenantNamespace).
		Name(pod.Name).
		SubResource("exec").
		URL())
	if err != nil {
		return fmt.Errorf("failed to create executor: %v", err)
	}

	// Execute the command with stdin from request body
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:  bytes.NewReader(req.Content),
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		return fmt.Errorf("failed to execute command: %v", err)
	}

	return nil
}

// HeadArchive returns the metadata for a file in a box
func (s *Service) HeadArchive(ctx context.Context, id string, req *model.BoxArchiveHeadParams) (*model.BoxArchiveHeadResult, error) {
	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}
	s.accessTracker.Update(id)

	// Get the pod name for the deployment
	pods, err := s.client.CoreV1().Pods(tenantNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=gbox,%s=%s", labelName, labelInstance, id),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("box not found: %s", id)
	}

	pod := pods.Items[0]
	if pod.Status.Phase != corev1.PodRunning {
		return nil, fmt.Errorf("box is not running: %s", id)
	}

	// Create command to get file/directory metadata
	cmd := []string{"stat", "-f", "%N:%z:%m:%a:%U:%G", req.Path}
	exec, err := s.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(tenantNamespace).
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: cmd,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec).
		Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %v", err)
	}
	defer exec.Close()

	// Read the output
	output, err := io.ReadAll(exec)
	if err != nil {
		return nil, fmt.Errorf("failed to read exec output: %v", err)
	}

	// Parse the output
	parts := strings.Split(string(output), ":")
	if len(parts) != 6 {
		return nil, fmt.Errorf("invalid stat output: %s", string(output))
	}

	// Convert mode to uint32
	mode, err := strconv.ParseUint(parts[3], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mode: %v", err)
	}

	return &model.BoxArchiveHeadResult{
		Mode: uint32(mode),
		Size: parseInt64(parts[1]),
	}, nil
}

// Helper functions
func getImage(image string) string {
	if image == "" {
		return defaultImage
	}
	return image
}

func getImagePullSecrets(secret string) []corev1.LocalObjectReference {
	if secret == "" {
		return nil
	}
	return []corev1.LocalObjectReference{{Name: secret}}
}

func joinArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	// Convert args array to JSON string to preserve spaces and special characters
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return ""
	}
	return string(argsJSON)
}

func int32Ptr(i int32) *int32 {
	return &i
}

func parseInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

func getEnvVars(env map[string]string) []corev1.EnvVar {
	if env == nil {
		return nil
	}
	vars := make([]corev1.EnvVar, 0, len(env))
	for k, v := range env {
		vars = append(vars, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return vars
}

func init() {
	service.Register("k8s", func(tracker tracker.AccessTracker) (service.BoxService, error) {
		return NewService(tracker)
	})
}
