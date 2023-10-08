package ghlabel

import (
	"fmt"
	"strings"
)

// Label is a GitHub Actions job label.
type Label struct {
	Group  string
	Flavor string
}

// New returns a new Label from the specified string and LabelOpts.
func New(s string) *Label {
	fields := strings.SplitN(s, ".", 2)

	l := &Label{}

	if len(fields) >= 1 {
		l.Group = fields[0]
	}

	if len(fields) >= 2 {
		l.Flavor = fields[1]
	}

	return l
}

// String returns the string representation of the Label.
func (l *Label) String() string {
	if l.Flavor == "" {
		return l.Group
	}

	return fmt.Sprintf("%s.%s", l.Group, l.Flavor)
}
