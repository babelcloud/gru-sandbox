package version

import (
	"runtime"
	"time"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	CommitID  = "unknown"
)

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

func ServerInfo() map[string]string {
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
