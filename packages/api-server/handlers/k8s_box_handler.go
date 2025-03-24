package handlers

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/emicklei/go-restful/v3"
	"github.com/google/uuid"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/babelcloud/gru-sandbox/packages/api-server/config"
	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/common"
	"github.com/babelcloud/gru-sandbox/packages/api-server/models"
	"github.com/babelcloud/gru-sandbox/packages/api-server/types"
	"k8s.io/apimachinery/pkg/api/errors"
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

// K8sBoxHandler handles box-related operations in Kubernetes
type K8sBoxHandler struct {
	client *kubernetes.Clientset
	config *rest.Config
}

// NewK8sBoxHandler creates a new K8sBoxHandler
func NewK8sBoxHandler(cfg *config.K8sConfig) (types.BoxHandler, error) {
	// Initialize Kubernetes client
	client, err := kubernetes.NewForConfig(cfg.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	return &K8sBoxHandler{
		client: client,
		config: cfg.Config,
	}, nil
}

// ListBoxes returns all boxes
func (h *K8sBoxHandler) ListBoxes(req *restful.Request, resp *restful.Response) {
	deployments, err := h.client.AppsV1().Deployments(tenantNamespace).List(req.Request.Context(), metav1.ListOptions{
		LabelSelector: labelName + "=gbox",
	})
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	boxes := make([]models.Box, 0)
	for _, deployment := range deployments.Items {
		box := models.Box{
			ID:     deployment.Labels[labelInstance],
			Image:  deployment.Spec.Template.Spec.Containers[0].Image,
			Status: string(deployment.Status.AvailableReplicas),
		}
		boxes = append(boxes, box)
	}
	resp.WriteEntity(boxes)
}

// CreateBox creates a new box
func (h *K8sBoxHandler) CreateBox(req *restful.Request, resp *restful.Response) {
	var boxReq models.BoxCreateRequest
	if err := req.ReadEntity(&boxReq); err != nil {
		resp.WriteError(http.StatusBadRequest, err)
		return
	}

	boxID := uuid.New().String()
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
	if boxReq.Cmd != "" {
		annotations[annotationCmd] = boxReq.Cmd
	}
	if len(boxReq.Args) > 0 {
		annotations[annotationArgs] = joinArgs(boxReq.Args)
	}
	if boxReq.WorkingDir != "" {
		annotations[annotationWorkDir] = boxReq.WorkingDir
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
							Image:      getImage(boxReq.Image),
							Command:    common.GetCommand(boxReq.Cmd),
							Args:       boxReq.Args,
							WorkingDir: boxReq.WorkingDir,
						},
					},
					ImagePullSecrets: getImagePullSecrets(boxReq.ImagePullSecret),
				},
			},
		},
	}

	result, err := h.client.AppsV1().Deployments(tenantNamespace).Create(req.Request.Context(), deployment, metav1.CreateOptions{})
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	box := models.Box{
		ID:     boxID,
		Image:  boxReq.Image,
		Status: string(result.Status.AvailableReplicas),
	}
	resp.WriteHeaderAndEntity(http.StatusCreated, box)
}

// DeleteBox deletes a box by ID
func (h *K8sBoxHandler) DeleteBox(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	if boxID == "" {
		resp.WriteErrorString(http.StatusBadRequest, "Box ID is required")
		return
	}

	err := h.client.AppsV1().Deployments(tenantNamespace).Delete(req.Request.Context(), boxID, metav1.DeleteOptions{})
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteHeader(http.StatusOK)
}

// DeleteAllBoxes deletes all boxes
func (h *K8sBoxHandler) DeleteBoxes(req *restful.Request, resp *restful.Response) {
	// List all deployments with gbox label
	deployments, err := h.client.AppsV1().Deployments(tenantNamespace).List(req.Request.Context(), metav1.ListOptions{
		LabelSelector: labelName + "=gbox",
	})
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Delete all deployments
	for _, deployment := range deployments.Items {
		err := h.client.AppsV1().Deployments(tenantNamespace).Delete(req.Request.Context(), deployment.Name, metav1.DeleteOptions{})
		if err != nil {
			resp.WriteError(http.StatusInternalServerError, err)
			return
		}
	}

	resp.WriteHeader(http.StatusOK)
}

// ExecBox executes a command in a box
func (h *K8sBoxHandler) ExecBox(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	if boxID == "" {
		resp.WriteErrorString(http.StatusBadRequest, "Box ID is required")
		return
	}

	// Parse request body
	var execReq models.BoxExecRequest
	if err := req.ReadEntity(&execReq); err != nil {
		resp.WriteError(http.StatusBadRequest, err)
		return
	}

	// Get the pod name for the deployment
	pods, err := h.client.CoreV1().Pods(tenantNamespace).List(req.Request.Context(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=gbox,%s=%s", labelName, labelInstance, boxID),
	})
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	if len(pods.Items) == 0 {
		resp.WriteErrorString(http.StatusNotFound, "Box not found")
		return
	}

	pod := pods.Items[0]
	if pod.Status.Phase != corev1.PodRunning {
		resp.WriteErrorString(http.StatusBadRequest, "Box is not running")
		return
	}

	// Hijack the connection
	httpResp := resp.ResponseWriter
	clientConn, _, err := httpResp.(http.Hijacker).Hijack()
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}
	defer clientConn.Close()

	// Write HTTP response headers
	if execReq.TTY {
		// For TTY sessions, use raw stream
		fmt.Fprintf(clientConn, "HTTP/1.1 101 UPGRADED\r\n")
		fmt.Fprintf(clientConn, "Content-Type: %s\r\n", models.MediaTypeRawStream)
		fmt.Fprintf(clientConn, "Connection: Upgrade\r\n")
		fmt.Fprintf(clientConn, "Upgrade: tcp\r\n")
	} else {
		// For non-TTY sessions, use multiplexed stream
		fmt.Fprintf(clientConn, "HTTP/1.1 200 OK\r\n")
		fmt.Fprintf(clientConn, "Content-Type: %s\r\n", models.MediaTypeMultiplexedStream)
	}
	fmt.Fprintf(clientConn, "\r\n")

	// Create remote command executor
	execURL, err := url.Parse(fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/exec", tenantNamespace, pod.Name))
	if err != nil {
		log.Printf("Error parsing URL: %v", err)
		return
	}

	// Get the REST config from the client
	exec, err := remotecommand.NewSPDYExecutor(h.config, "POST", execURL)
	if err != nil {
		log.Printf("Error creating executor: %v", err)
		return
	}

	// Create stream options
	streamOptions := remotecommand.StreamOptions{
		Stdin:             nil,
		Stdout:            nil,
		Stderr:            nil,
		TerminalSizeQueue: nil,
		Tty:               execReq.TTY,
	}

	// Set up streams based on request
	if execReq.Stdin {
		streamOptions.Stdin = clientConn
	}
	if execReq.Stdout {
		if execReq.TTY {
			streamOptions.Stdout = clientConn
		} else {
			streamOptions.Stdout = &multiplexedWriter{writer: clientConn}
		}
	}
	if execReq.Stderr && !execReq.TTY {
		streamOptions.Stderr = &multiplexedWriter{writer: clientConn, stream: models.StreamStderr}
	}

	// Start streaming with context
	err = exec.Stream(streamOptions)
	if err != nil {
		log.Printf("Error executing command: %v", err)
	}
}

// StartBox starts a stopped box
func (h *K8sBoxHandler) StartBox(req *restful.Request, resp *restful.Response) {
	// TODO: Implement Kubernetes pod start
	resp.WriteErrorString(http.StatusNotImplemented, "Kubernetes start not implemented")
}

// StopBox stops a running box
func (h *K8sBoxHandler) StopBox(req *restful.Request, resp *restful.Response) {
	// TODO: Implement Kubernetes pod stop
	resp.WriteErrorString(http.StatusNotImplemented, "Kubernetes stop not implemented")
}

// RunBox handles the run box operation
func (h *K8sBoxHandler) RunBox(req *restful.Request, resp *restful.Response) {
	// TODO: Implement run operation for K8s
	resp.WriteError(http.StatusNotImplemented, fmt.Errorf("run operation not implemented for K8s"))
}

// GetBox implements the BoxHandler interface
func (h *K8sBoxHandler) GetBox(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	if boxID == "" {
		resp.WriteError(http.StatusBadRequest, fmt.Errorf("box ID is required"))
		return
	}

	// Get pod details
	pod, err := h.client.CoreV1().Pods(tenantNamespace).Get(context.Background(), boxID, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			resp.WriteError(http.StatusNotFound, fmt.Errorf("box not found: %s", boxID))
			return
		}
		resp.WriteError(http.StatusInternalServerError, fmt.Errorf("failed to get box: %v", err))
		return
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
	box := models.Box{
		ID:     boxID,
		Status: status,
		Image:  pod.Spec.Containers[0].Image,
	}

	resp.WriteAsJson(box)
}

// ReclaimBoxes implements the BoxHandler interface
func (h *K8sBoxHandler) ReclaimBoxes(req *restful.Request, resp *restful.Response) {
	// TODO: Implement Kubernetes box reclamation
	resp.WriteErrorString(http.StatusNotImplemented, "Kubernetes box reclamation not implemented")
}

// multiplexedWriter implements io.Writer for multiplexed streams
type multiplexedWriter struct {
	writer io.Writer
	stream models.StreamType
}

func (w *multiplexedWriter) Write(p []byte) (n int, err error) {
	header := make([]byte, 8)
	header[0] = byte(w.stream)
	binary.BigEndian.PutUint32(header[4:], uint32(len(p)))

	if _, err := w.writer.Write(header); err != nil {
		return 0, err
	}
	return w.writer.Write(p)
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
