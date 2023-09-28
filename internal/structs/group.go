package structs

// Group struct.
type Group struct {
	Name string `mapstructure:"name"`
}

// String returns the string representation of a Group.
func (g *Group) String() string {
	return g.Name
}

func (g *Group) Equals(other *Group) bool {
	return g.Name == other.Name
}
