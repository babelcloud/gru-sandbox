package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/babelcloud/gbox/packages/cli/config"
)

// These variables will be set at build time via -ldflags
var (
	// Version represents the application version (from git tags)
	Version = "dev"
	// BuildTime is the time when the binary was built
	BuildTime = "unknown"
	// CommitID is the git commit hash
	CommitID = "unknown"
)

// formatBuildTime returns a nicely formatted build time
func formatBuildTime() string {
	if BuildTime == "unknown" {
		return BuildTime
	}

	t, err := time.Parse(time.RFC3339, BuildTime)
	if err != nil {
		return BuildTime
	}

	return t.Format("Mon Jan 2 15:04:05 2006")
}

// ClientInfo returns structured client version information
func ClientInfo() map[string]string {
	return map[string]string{
		"Version":       Version,
		"APIVersion":    "v1",
		"GoVersion":     runtime.Version(),
		"GitCommit":     CommitID,
		"BuildTime":     BuildTime,
		"FormattedTime": formatBuildTime(),
		"OS":            runtime.GOOS,
		"Arch":          runtime.GOARCH,
	}
}

// GetServerInfo retrieves version information from the API server
func GetServerInfo() (map[string]string, error) {
	apiURL := config.GetAPIURL()
	url := fmt.Sprintf("%s/api/v1/version", apiURL)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to API server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API server returned status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from API server: %v", err)
	}

	var serverInfo map[string]string
	if err := json.Unmarshal(body, &serverInfo); err != nil {
		return nil, fmt.Errorf("failed to parse server version information: %v", err)
	}

	return serverInfo, nil
}
