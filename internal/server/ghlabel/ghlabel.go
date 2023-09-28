package ghlabel

import (
	"fmt"
	"regexp"
)

// Label is a GitHub Actions job label.
type Label struct {
	Group  string
	Flavor string
}

// LabelOpt is a function that modifies a Label.
type LabelOpt func(*Label)

// New returns a new Label from the specified string and LabelOpts.
func New(s string, opts ...LabelOpt) (*Label, error) {
	regexp := regexp.MustCompile(`^([a-zA-Z0-9]+)(?:-([a-zA-Z0-9-]+))?$`)
	matches := regexp.FindStringSubmatch(s)

	if len(matches) == 0 {
		return nil, fmt.Errorf("regexp %s did not match", regexp)
	}

	l := &Label{
		Group:  matches[1],
		Flavor: matches[2],
	}

	for _, opt := range opts {
		opt(l)
	}

	return l, nil
}

// WithDefaultFlavor sets the default flavor for the Label.
func WithDefaultFlavor(flavor string) LabelOpt {
	f := func(l *Label) {
		if l.Flavor == "" {
			l.Flavor = flavor
		}
	}

	return f
}
