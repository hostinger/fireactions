package structs

import (
	"fmt"
)

// Image represents a virtual machine disk image.
type Image struct {
	ID   string
	Name string
	URL  string
}

// String returns a string representation of the Image.
func (i Image) String() string {
	return fmt.Sprintf("%s (ID: %s, URL: %s)", i.Name, i.ID, i.URL)
}
