package rootfs

import (
	"fmt"
	"path/filepath"
)

// RootFS represents the root filesystem of a virtual machine.
type RootFS struct {
	os          OSInterface
	mountPath   string
	unmountFunc UnmountFunc
	mountFunc   MountFunc
}

// Opt is a functional option for RootFS.
type Opt func(*RootFS)

// WithOS sets the OSInterface to use.
func WithOS(os OSInterface) Opt {
	f := func(b *RootFS) {
		b.os = os
	}

	return f
}

// WithMountFunc sets the mount function to use.
func WithMountFunc(mountFunc MountFunc) Opt {
	f := func(b *RootFS) {
		b.mountFunc = mountFunc
	}

	return f
}

// WithUnmountFunc sets the unmount function to use.
func WithUnmountFunc(unmountFunc UnmountFunc) Opt {
	f := func(b *RootFS) {
		b.unmountFunc = unmountFunc
	}

	return f
}

// New creates a new RootFS.
func New(source string, opts ...Opt) (*RootFS, error) {
	b := &RootFS{
		os:          &realOS{},
		unmountFunc: realUnmountFunc,
		mountFunc:   realMountFunc,
	}

	for _, opt := range opts {
		opt(b)
	}

	mountPath, err := b.os.MkdirTemp("", "fireactions")
	if err != nil {
		return nil, err
	}

	err = b.mountFunc(source, mountPath)
	if err != nil {
		return nil, err
	}

	b.mountPath = mountPath
	return b, nil
}

// SetupHostname sets the hostname of the root filesystem.
func (b *RootFS) SetupHostname(hostname string) error {
	return b.os.WriteFile(filepath.Join(b.mountPath, "etc", "hostname"), []byte(fmt.Sprintf("%s\n", hostname)), 0644)
}

// SetupDNS sets up DNS for the root filesystem.
func (b *RootFS) SetupDNS() error {
	resolvConfPath := filepath.Join(b.mountPath, "etc", "resolv.conf")

	var err error
	err = b.os.Remove(resolvConfPath)
	if err != nil && !b.os.IsNotExist(err) {
		return fmt.Errorf("error removing %s: %v", resolvConfPath, err)
	}

	err = b.os.Symlink("../proc/net/pnp", resolvConfPath)
	if err != nil {
		return fmt.Errorf("error creating symlink /etc/resolv.conf -> ../proc/net/pnp: %v", err)
	}

	return nil
}

// Close unmounts the root filesystem and removes the mount path.
func (b *RootFS) Close() error {
	err := b.unmountFunc(b.mountPath)
	if err != nil {
		return err
	}

	return b.os.RemoveAll(b.mountPath)
}
