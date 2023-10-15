package ghlabel

import (
	"fmt"
	"strings"
)

var (
	// ErrEmptyLabel is returned when the label is empty.
	ErrEmptyLabel = fmt.Errorf("label is empty")
)

var defaultSeparator = "."

// Label is a GitHub Actions job label.
type Label struct {
	Prefix string
	Group  string
	Flavor string
}

// New returns a new Label from the specified string and LabelOpts.
func New(s string) (*Label, error) {
	if s == "" {
		return nil, ErrEmptyLabel
	}

	values := strings.SplitN(s, defaultSeparator, 3)

	l := &Label{
		Prefix: values[0],
	}

	if len(values) >= 2 {
		l.Group = values[1]
	}

	if len(values) >= 3 {
		l.Flavor = values[2]
	}

	return l, nil
}

// String returns the string representation of the Label.
func (l *Label) String() string {
	s := l.Prefix

	if l.Group != "" {
		s += fmt.Sprintf("%s%s", defaultSeparator, l.Group)
	}

	if l.Flavor != "" {
		s += fmt.Sprintf("%s%s", defaultSeparator, l.Flavor)
	}

	return s
}
