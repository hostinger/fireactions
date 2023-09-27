package build

import "fmt"

var (
	// GitCommit is the git commit of the binary.
	GitCommit string

	// GitTag is the git tag of the binary.
	GitTag string

	// BuildDate is the date of the build.
	BuildDate string
)

// Info returns the build information.
func Info() string {
	return fmt.Sprintf("%s (%s) built on %s", GitTag, GitCommit, BuildDate)
}
