package store

import (
	"fmt"
)

// ErrNotFound is returned when an entity doesn't exist.
type ErrNotFound struct {
	ID   string
	Type string
}

// Error returns the error message.
func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s (ID: %s) doesn't exist", e.Type, e.ID)
}

// Is returns true if the target is an ErrNotFound.
func (e ErrNotFound) Is(target error) bool {
	_, ok := target.(ErrNotFound)
	return ok
}

// As returns true if the target is an ErrNotFound.
func (e ErrNotFound) As(target interface{}) bool {
	_, ok := target.(*ErrNotFound)
	return ok
}
