package ghlabel

import (
	"fmt"
	"regexp"
	"strconv"
)

// Label is a GitHub Actions job label.
type Label struct {
	Group  string
	RAM    uint64
	CPU    uint64
	Kernel string
	OS     string
}

// LabelOpt is a function that modifies a Label.
type LabelOpt func(*Label)

// New returns a new Label from the specified string and LabelOpts.
func New(s string, opts ...LabelOpt) (*Label, error) {
	regexp := regexp.MustCompile(`(?P<GROUP>.*)-(?P<CPU>\d+)cpu-(?P<RAM>\d+)gb-?(?P<OS>[a-zA-Z0-9\.]*)-?(?P<KERNEL>[a-zA-Z0-9\.]*)`)
	matches := regexp.FindStringSubmatch(s)

	if len(matches) == 0 {
		return nil, fmt.Errorf("regexp %s did not match", regexp)
	}

	cpu, _ := strconv.ParseUint(matches[2], 10, 64)
	ram, _ := strconv.ParseUint(matches[3], 10, 64)

	l := &Label{
		Group:  matches[1],
		RAM:    ram,
		CPU:    cpu,
		Kernel: matches[5],
		OS:     matches[4],
	}

	for _, opt := range opts {
		opt(l)
	}

	return l, nil
}

// WithDefaultKernel sets the default kernel for a label.
func WithDefaultKernel(kernel string) LabelOpt {
	f := func(l *Label) {
		l.Kernel = kernel
	}

	return f
}

// WithDefaultOS sets the default OS for a label.
func WithDefaultOS(os string) LabelOpt {
	f := func(l *Label) {
		l.OS = os
	}

	return f
}
