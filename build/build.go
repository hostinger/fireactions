package build

import "fmt"

var (
	// Version is the version of the binary.
	Version string

	// GitCommit is the git commit of the binary.
	GitCommit string

	// BuildDate is the date of the build.
	BuildDate string
)

func Info() string {
	return fmt.Sprintf("%s (%s) built on %s", Version, GitCommit, BuildDate)
}
