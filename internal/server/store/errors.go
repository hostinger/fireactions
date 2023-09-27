package store

import (
	"fmt"
)

type ErrNotFound struct {
	ID   string
	Type string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s (ID: %s) doesn't exist", e.Type, e.ID)
}

func (e ErrNotFound) Is(target error) bool {
	_, ok := target.(ErrNotFound)
	return ok
}

func (e ErrNotFound) As(target interface{}) bool {
	_, ok := target.(*ErrNotFound)
	return ok
}
