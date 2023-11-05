package rootfs

import "syscall"

// MountFunc is a function that mounts a filesystem.
type MountFunc func(source string, target string) error

// UnmountFunc is a function that unmounts a filesystem.
type UnmountFunc func(target string) error

func realMountFunc(source string, target string) error {
	return syscall.Mount(source, target, "ext4", 0, "")
}

func realUnmountFunc(target string) error {
	return syscall.Unmount(target, 0)
}
