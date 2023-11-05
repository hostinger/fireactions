package rootfs

import (
	"io/fs"
	"os"
)

// OSInterface is an interface for OS functions. Primarily used for testing.
type OSInterface interface {
	WriteFile(name string, data []byte, perm fs.FileMode) error
	IsNotExist(err error) bool
	Remove(name string) error
	RemoveAll(path string) error
	Symlink(oldname, newname string) error
	MkdirTemp(dir, pattern string) (string, error)
}

type realOS struct{}

func (o *realOS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (o *realOS) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

func (o *realOS) Remove(name string) error {
	return os.Remove(name)
}

func (o *realOS) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

func (o *realOS) MkdirTemp(dir, pattern string) (string, error) {
	return os.MkdirTemp(dir, pattern)
}

func (o *realOS) RemoveAll(path string) error {
	return os.RemoveAll(path)
}
