package fireactions

import "fmt"

var (
	// Version is the version of the Fireactions.
	Version = "0.0.0"

	// Date is the date when the binary was built.
	Date = "1970-01-01T00:00:00Z"

	// Commit is the Git SHA of the commit that was built.
	Commit = ""
)

// String returns a string representation of the Fireactions version.
func String() string {
	return fmt.Sprintf("%s (Built on %s from Git SHA %s)\n", Version, Date, Commit)
}
