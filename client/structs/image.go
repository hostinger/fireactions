package structs

import "time"

// Image represents a virtual machine disk image.
type Image struct {
	Info *ImageInfo

	Path string
}

// ImageInfo represents metadata about an image.
type ImageInfo struct {
	ID        string
	Name      string
	URL       string
	CreatedAt time.Time
	UpdatedAt time.Time
}
