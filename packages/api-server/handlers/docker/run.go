package docker

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/emicklei/go-restful/v3"

	"github.com/babelcloud/gru-sandbox/packages/api-server/config"
	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/common"
	"github.com/babelcloud/gru-sandbox/packages/api-server/models"
)

// readDockerStream reads from a Docker stream and returns stdout and stderr content
func readDockerStream(reader io.Reader) (string, string, error) {
	header := make([]byte, 8)
	var stdout, stderr strings.Builder

	for {
		// Read header
		_, err := io.ReadFull(reader, header)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", "", fmt.Errorf("error reading stream header: %v", err)
		}

		// Parse header
		streamType := header[0]
		// Skip 3 bytes reserved for future use
		size := binary.BigEndian.Uint32(header[4:])

		// Read payload
		payload := make([]byte, size)
		_, err = io.ReadFull(reader, payload)
		if err != nil {
			return "", "", fmt.Errorf("error reading stream payload: %v", err)
		}

		// Write to appropriate output based on stream type
		switch streamType {
		case 1: // stdout
			stdout.Write(payload)
		case 2: // stderr
			stderr.Write(payload)
		}
	}

	return stdout.String(), stderr.String(), nil
}

// collectOutput collects output from a reader with line limit
func collectOutput(reader io.Reader, stdoutLimit, stderrLimit int) (string, string) {
	stdout, stderr, err := readDockerStream(reader)
	if err != nil {
		log.Printf("Error reading Docker stream: %v", err)
		return "", ""
	}

	// Process stdout with line limit
	var stdoutLines []string
	if stdoutLimit >= 0 {
		scanner := bufio.NewScanner(strings.NewReader(stdout))
		for scanner.Scan() && len(stdoutLines) < stdoutLimit {
			stdoutLines = append(stdoutLines, scanner.Text())
		}
		stdout = strings.Join(stdoutLines, "\n")
	}

	// Process stderr with line limit
	var stderrLines []string
	if stderrLimit >= 0 {
		scanner := bufio.NewScanner(strings.NewReader(stderr))
		for scanner.Scan() && len(stderrLines) < stderrLimit {
			stderrLines = append(stderrLines, scanner.Text())
		}
		stderr = strings.Join(stderrLines, "\n")
	}

	return stdout, stderr
}

// handleRunBox handles the run box operation
func handleRunBox(h *DockerBoxHandler, req *restful.Request, resp *restful.Response) {
	// Parse request body
	var runReq models.BoxRunRequest
	if err := req.ReadEntity(&runReq); err != nil {
		log.Printf("Error reading request body: %v", err)
		writeError(resp, http.StatusBadRequest, "INVALID_REQUEST", fmt.Sprintf("Error reading request body: %v", err))
		return
	}

	// Set default line limits if not specified
	if runReq.StdoutLineLimit == 0 {
		runReq.StdoutLineLimit = 100
	}
	if runReq.StderrLineLimit == 0 {
		runReq.StderrLineLimit = 100
	}

	// Get image name
	img := common.GetImage(runReq.Image)
	log.Printf("Checking image: %q", img)

	// Check if image exists
	_, _, err := h.client.ImageInspectWithRaw(req.Request.Context(), img)
	if err == nil {
		log.Printf("Using existing image: %q", img)
	} else {
		log.Printf("Image %q not found, pulling", img)
		if err := pullImage(h, req, resp, img, runReq.ImagePullSecret); err != nil {
			return
		}
	}

	boxID := common.GenerateBoxID()
	containerName := fmt.Sprintf("gbox-%s", boxID)

	// Prepare labels
	labels := map[string]string{
		GboxNamespace:      config.GetGboxNamespace(),
		GboxLabelID:        boxID,
		GboxLabelName:      "gbox",
		GboxLabelVersion:   "v1",
		GboxLabelComponent: "sandbox",
		GboxLabelManagedBy: "gru-api-server",
	}

	// Add command configuration to labels
	if len(runReq.Cmd) > 0 {
		labels[GboxLabelPrefix+".cmd"] = runReq.Cmd[0]
	}
	if len(runReq.Args) > 0 {
		labels[GboxLabelPrefix+".args"] = common.JoinArgs(runReq.Args)
	}
	if runReq.WorkingDir != "" {
		labels[GboxLabelPrefix+".working-dir"] = runReq.WorkingDir
	}

	// Add custom labels with prefix
	if runReq.ExtraLabels != nil {
		for k, v := range runReq.ExtraLabels {
			labels[GboxExtraLabelPrefix+"."+k] = v
		}
	}

	// Get share directory from config
	fileConfig := config.NewFileConfig().(*config.FileConfig)
	if err := fileConfig.Initialize(nil); err != nil {
		log.Printf("Error initializing file config: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Create share directory for the box
	hostShareDir := filepath.Join(fileConfig.GetHostShareDir(), boxID)
	shareDir := filepath.Join(fileConfig.GetFileShareDir(), boxID)
	if err := os.MkdirAll(shareDir, 0755); err != nil {
		log.Printf("Error creating share directory: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Prepare volume mounts
	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: hostShareDir,
			Target: common.DefaultShareDirPath,
		},
	}

	// Add user-specified mounts
	for _, m := range runReq.Mounts {
		mountType := mount.TypeBind
		switch m.Type {
		case models.MountTypeBind:
			mountType = mount.TypeBind
		case models.MountTypeVolume:
			mountType = mount.TypeVolume
		case models.MountTypeTmpfs:
			mountType = mount.TypeTmpfs
		default:
			log.Printf("Invalid mount type: %s", m.Type)
			resp.WriteError(http.StatusBadRequest, fmt.Errorf("invalid mount type: %s", m.Type))
			return
		}

		// Validate source path for bind mounts
		if m.Type == models.MountTypeBind {
			if !filepath.IsAbs(m.Source) {
				log.Printf("Source path must be absolute: %s", m.Source)
				resp.WriteError(http.StatusBadRequest, fmt.Errorf("source path must be absolute: %s", m.Source))
				return
			}

			// Check if source path exists
			if _, err := os.Stat(m.Source); err != nil {
				log.Printf("Source path does not exist: %s", m.Source)
				resp.WriteError(http.StatusBadRequest, fmt.Errorf("source path does not exist: %s", m.Source))
				return
			}
		}

		// Add mount configuration
		mountConfig := mount.Mount{
			Type:     mountType,
			Source:   m.Source,
			Target:   m.Target,
			ReadOnly: m.ReadOnly,
		}

		// Set consistency if specified
		if m.Consistency != "" {
			switch m.Consistency {
			case "default":
				mountConfig.Consistency = mount.ConsistencyDefault
			case "cached":
				mountConfig.Consistency = mount.ConsistencyCached
			case "delegated":
				mountConfig.Consistency = mount.ConsistencyDelegated
			default:
				log.Printf("Invalid mount consistency: %s", m.Consistency)
				resp.WriteError(http.StatusBadRequest, fmt.Errorf("invalid mount consistency: %s", m.Consistency))
				return
			}
		}

		mounts = append(mounts, mountConfig)
	}

	// Create container
	containerResp, err := h.client.ContainerCreate(
		req.Request.Context(),
		&container.Config{
			Image:      img,
			Cmd:        append(runReq.Cmd, runReq.Args...),
			Env:        common.GetEnvVars(runReq.Env),
			WorkingDir: runReq.WorkingDir,
			Labels:     labels,
		},
		&container.HostConfig{
			Mounts: mounts,
		},
		nil,
		nil,
		containerName,
	)
	if err != nil {
		log.Printf("Error creating container: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Start container
	if err := h.client.ContainerStart(req.Request.Context(), containerResp.ID, types.ContainerStartOptions{}); err != nil {
		log.Printf("Error starting container: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Wait for container to finish
	statusCh, errCh := h.client.ContainerWait(req.Request.Context(), containerResp.ID, container.WaitConditionNotRunning)
	var statusCode int64
	select {
	case err := <-errCh:
		log.Printf("Error waiting for container: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	case status := <-statusCh:
		statusCode = status.StatusCode
	}

	// Get container logs
	logs, err := h.client.ContainerLogs(req.Request.Context(), containerResp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		log.Printf("Error getting container logs: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}
	defer logs.Close()

	// Collect output with line limits
	stdout, stderr := collectOutput(logs, runReq.StdoutLineLimit, runReq.StderrLineLimit)

	// Remove container
	err = h.client.ContainerRemove(req.Request.Context(), containerResp.ID, types.ContainerRemoveOptions{
		Force: true,
	})
	if err != nil {
		log.Printf("Error removing container: %v", err)
		// Don't fail the request if cleanup fails
	}

	// Prepare response
	result := models.BoxRunResponse{
		Box: models.Box{
			ID:     boxID,
			Status: "exited",
			Image:  img,
		},
		ExitCode: int(statusCode),
		Stdout:   stdout,
		Stderr:   stderr,
	}

	resp.WriteAsJson(result)
}
