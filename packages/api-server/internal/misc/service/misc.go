package service

import (
	"runtime"
	"time"

	"github.com/babelcloud/gbox/packages/api-server/internal/misc/model"
)

var (
	// Version is the version of the server
	Version = "dev"
	// BuildTime is the time when the server was built
	BuildTime = "unknown"
	// CommitID is the git commit ID of the server
	CommitID = "unknown"
)

// MiscService handles miscellaneous operations
type MiscService struct{}

// New creates a new MiscService
func New() *MiscService {
	return &MiscService{}
}

// formatBuildTime formats the build time to a readable string
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

// GetVersion returns server version information
func (s *MiscService) GetVersion() *model.VersionInfo {
	return &model.VersionInfo{
		Version:       Version,
		APIVersion:    "v1",
		GoVersion:     runtime.Version(),
		GitCommit:     CommitID,
		BuildTime:     BuildTime,
		FormattedTime: formatBuildTime(),
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
	}
}
