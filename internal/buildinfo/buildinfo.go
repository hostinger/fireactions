package buildinfo

var (
	// Version is the current version of the application.
	Version = "0.0.0"
)

// GetVersion returns the current version of the application.
func GetVersion() string {
	return Version
}
