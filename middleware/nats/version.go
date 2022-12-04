package nats

const (
	instrumentationName = "github.com/tel-io/instrumentation/subMiddleware/nats"
)

// Version is the current release version of the otelnats instrumentation.
func Version() string {
	return "1.00.0"
	// This string is updated by the pre_release.sh script during release
}

// SemVersion is the semantic version to be supplied to tracer/meter creation.
func SemVersion() string {
	return "semver:" + Version()
}
