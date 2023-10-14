package structs

import "time"

// Image represents a virtual machine disk image.
type Image struct {
	Info *ImageInfo

	// Path is the path to the image file.
	Path string
}

// ImageInfo represents metadata about an Image.
type ImageInfo struct {
	ID        string
	Name      string
	URL       string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NodeRegistrationInfo represents information about the Node's registration status. It is used to determine whether
// the Node has been registered with the server.
type NodeRegistrationInfo struct {
	// ID is the ID of the Node, as assigned by the server.
	ID string
}

// IsRegistered returns true if the Node has been registered with the server.
func (n *NodeRegistrationInfo) IsRegistered() bool {
	return n.ID != ""
}
