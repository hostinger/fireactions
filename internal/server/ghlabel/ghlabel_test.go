package ghlabel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_Valid(t *testing.T) {
	s := "group1-2cpu-4gb-ubuntu20.04-5.4"

	l, err := New(s)
	if err != nil {
		t.Fatalf("error creating label: %s", err.Error())
	}

	assert.Equal(t, "group1", l.Group)
	assert.Equal(t, uint64(2), l.CPU)
	assert.Equal(t, uint64(4), l.RAM)
	assert.Equal(t, "ubuntu20.04", l.OS)
	assert.Equal(t, "5.4", l.Kernel)
}

func TestNew_Valid_NoKernel(t *testing.T) {
	s := "group1-2cpu-4gb-ubuntu20.04"

	l, err := New(s)
	if err != nil {
		t.Fatalf("error creating label: %s", err.Error())
	}

	assert.Equal(t, "group1", l.Group)
	assert.Equal(t, uint64(2), l.CPU)
	assert.Equal(t, uint64(4), l.RAM)
	assert.Equal(t, "ubuntu20.04", l.OS)
	assert.Equal(t, "", l.Kernel)
}

func TestNew_Valid_NoKernelAndNoOS(t *testing.T) {
	s := "group1-2cpu-4gb"

	l, err := New(s)
	if err != nil {
		t.Fatalf("error creating label: %s", err.Error())
	}

	assert.Equal(t, "group1", l.Group)
	assert.Equal(t, uint64(2), l.CPU)
	assert.Equal(t, uint64(4), l.RAM)
	assert.Equal(t, "", l.OS)
	assert.Equal(t, "", l.Kernel)
}

func TestNew_Valid_WithLabelOpts(t *testing.T) {
	s := "group1-2cpu-4gb"

	l, err := New(s, WithDefaultKernel("5.5"), WithDefaultOS("ubuntu20.04"))
	if err != nil {
		t.Fatalf("error creating label: %s", err.Error())
	}

	assert.Equal(t, "group1", l.Group)
	assert.Equal(t, uint64(2), l.CPU)
	assert.Equal(t, uint64(4), l.RAM)
	assert.Equal(t, "ubuntu20.04", l.OS)
	assert.Equal(t, "5.5", l.Kernel)
}

func TestNew_Invalid(t *testing.T) {
	labels := []string{
		"group1-1.5cpu-4gb",
		"group1-2cpu-4.5gb",
	}

	for _, label := range labels {
		_, err := New(label)

		assert.Error(t, err)
	}
}
