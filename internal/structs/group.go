package structs

import "fmt"

// Group struct.
type Group struct {
	Name    string
	Enabled bool
}

// String returns the string representation of a Group.
func (g *Group) String() string {
	return fmt.Sprintf("%s (Enabled: %t)", g.Name, g.Enabled)
}

// Equals returns true if the Group is equal to the other Group.
func (g *Group) Equals(other *Group) bool {
	return g.Name == other.Name
}
