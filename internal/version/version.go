package version

// Build-time variables injected via -ldflags during compilation
var (
	// Version is the current semantic version of the application.
	Version = "1.0.0"

	// GitCommit is the git commit SHA at build time.
	GitCommit = "dev"

	// BuildDate is the timestamp when the binary was built.
	BuildDate = ""
)

// Get returns the semantic version string.
func Get() string {
	return Version
}

// Info returns a human-readable version string including commit if available.
func Info() string {
	if GitCommit != "" && GitCommit != "dev" {
		return Version + " (" + GitCommit + ")"
	}
	return Version
}

// BuildInfo returns a map of build metadata.
func BuildInfo() map[string]string {
	return map[string]string{
		"version":   Version,
		"gitCommit": GitCommit,
		"buildDate": BuildDate,
	}
}
