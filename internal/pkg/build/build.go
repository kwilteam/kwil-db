package build

// Default build-time variable.
// These values are overridden via ldflags
var (
	Version   = "unknown-system"
	GitCommit = "unknown-commit"
	BuildTime = "unknown-buildtime"
)
