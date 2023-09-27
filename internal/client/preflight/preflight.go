// Package preflight provides a set of preflight checks that are run before the client is started.
package preflight

type Check interface {
	// Name returns the name of the check.
	Name() string

	// Check runs the check.
	Check() error
}
