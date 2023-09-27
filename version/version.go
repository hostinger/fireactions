package version

import "fmt"

var (
	Version = "0.0.0"
	Date    = "1970-01-01T00:00:00Z"
	Commit  = ""
)

// String returns a string representation of the Fireactions version.
func String() string {
	return fmt.Sprintf("%s (Built on %s from Git SHA %s)\n", Version, Date, Commit)
}
