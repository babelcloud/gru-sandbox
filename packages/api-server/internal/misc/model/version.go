package model

// VersionInfo represents server version information
type VersionInfo struct {
	// Version is the version of the server
	Version string `json:"version"`
	// APIVersion is the version of the API
	APIVersion string `json:"apiVersion"`
	// GoVersion is the version of Go used to build the server
	GoVersion string `json:"goVersion"`
	// GitCommit is the git commit ID of the server
	GitCommit string `json:"gitCommit"`
	// BuildTime is the time when the server was built
	BuildTime string `json:"buildTime"`
	// FormattedTime is the formatted build time
	FormattedTime string `json:"formattedTime"`
	// OS is the operating system the server is running on
	OS string `json:"os"`
	// Arch is the architecture the server is running on
	Arch string `json:"arch"`
}
