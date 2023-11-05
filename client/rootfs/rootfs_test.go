package rootfs

import (
	"errors"
	"fmt"
	"io/fs"
	"testing"

	"github.com/hostinger/fireactions/client/rootfs/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNew(t *testing.T) {
	t.Run("returns no error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mountFunc := func(source string, target string) error {
			return nil
		}

		unmountFunc := func(target string) error {
			return nil
		}

		os := mocks.NewMockOSInterface(ctrl)
		os.EXPECT().MkdirTemp("", "fireactions").Return("/tmp/fireactions", nil)

		rootfs, err := New("test", WithOS(os), WithMountFunc(mountFunc), WithUnmountFunc(unmountFunc))

		assert.Nil(t, err)
		assert.NotNil(t, rootfs)
	})

	t.Run("returns error if os.MkdirTemp fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mountFunc := func(source string, target string) error {
			return nil
		}

		unmountFunc := func(target string) error {
			return nil
		}

		os := mocks.NewMockOSInterface(ctrl)
		os.EXPECT().MkdirTemp("", "fireactions").Return("", errors.New("error"))

		rootfs, err := New("test", WithOS(os), WithMountFunc(mountFunc), WithUnmountFunc(unmountFunc))

		assert.NotNil(t, err)
		assert.Nil(t, rootfs)
	})

	t.Run("returns error if mountFunc fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mountFunc := func(source string, target string) error {
			return errors.New("error")
		}

		unmountFunc := func(target string) error {
			return nil
		}

		os := mocks.NewMockOSInterface(ctrl)
		os.EXPECT().MkdirTemp("", "fireactions").Return("/tmp/fireactions", nil)

		rootfs, err := New("test", WithOS(os), WithMountFunc(mountFunc), WithUnmountFunc(unmountFunc))

		assert.NotNil(t, err)
		assert.Nil(t, rootfs)
	})
}

func TestRootFS_SetupHostname(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mountFunc := func(source string, target string) error {
		return nil
	}

	unmountFunc := func(target string) error {
		return nil
	}

	os := mocks.NewMockOSInterface(ctrl)
	os.EXPECT().MkdirTemp("", "fireactions").Return("/tmp/fireactions", nil)

	rootfs, err := New("test", WithOS(os), WithMountFunc(mountFunc), WithUnmountFunc(unmountFunc))
	assert.Nil(t, err)
	assert.NotNil(t, rootfs)

	os.EXPECT().WriteFile("/tmp/fireactions/etc/hostname", []byte(fmt.Sprintf("%s\n", "test")), fs.FileMode(0644)).Return(nil)

	err = rootfs.SetupHostname("test")
	assert.Nil(t, err)
}

func TestRootFS_SetupDNS(t *testing.T) {
	t.Run("returns no error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mountFunc := func(source string, target string) error {
			return nil
		}

		unmountFunc := func(target string) error {
			return nil
		}

		os := mocks.NewMockOSInterface(ctrl)
		os.EXPECT().MkdirTemp("", "fireactions").Return("/tmp/fireactions", nil)

		rootfs, err := New("test", WithOS(os), WithMountFunc(mountFunc), WithUnmountFunc(unmountFunc))
		assert.Nil(t, err)
		assert.NotNil(t, rootfs)

		os.EXPECT().Remove("/tmp/fireactions/etc/resolv.conf").Return(nil)
		os.EXPECT().Symlink("../proc/net/pnp", "/tmp/fireactions/etc/resolv.conf").Return(nil)

		err = rootfs.SetupDNS()

		assert.Nil(t, err)
	})

	t.Run("returns error if os.Remove fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mountFunc := func(source string, target string) error {
			return nil
		}

		unmountFunc := func(target string) error {
			return nil
		}

		os := mocks.NewMockOSInterface(ctrl)
		os.EXPECT().MkdirTemp("", "fireactions").Return("/tmp/fireactions", nil)

		rootfs, err := New("test", WithOS(os), WithMountFunc(mountFunc), WithUnmountFunc(unmountFunc))
		assert.Nil(t, err)
		assert.NotNil(t, rootfs)

		os.EXPECT().Remove("/tmp/fireactions/etc/resolv.conf").Return(errors.New("error"))
		os.EXPECT().IsNotExist(errors.New("error")).Return(false)

		err = rootfs.SetupDNS()

		assert.NotNil(t, err)
	})

	t.Run("returns error if os.Symlink fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mountFunc := func(source string, target string) error {
			return nil
		}

		unmountFunc := func(target string) error {
			return nil
		}

		os := mocks.NewMockOSInterface(ctrl)
		os.EXPECT().MkdirTemp("", "fireactions").Return("/tmp/fireactions", nil)

		rootfs, err := New("test", WithOS(os), WithMountFunc(mountFunc), WithUnmountFunc(unmountFunc))
		assert.Nil(t, err)
		assert.NotNil(t, rootfs)

		os.EXPECT().Remove("/tmp/fireactions/etc/resolv.conf").Return(nil)
		os.EXPECT().Symlink("../proc/net/pnp", "/tmp/fireactions/etc/resolv.conf").Return(errors.New("error"))

		err = rootfs.SetupDNS()

		assert.NotNil(t, err)
	})
}

func TestRootFS_Close(t *testing.T) {
	t.Run("returns no error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mountFunc := func(source string, target string) error {
			return nil
		}

		unmountFunc := func(target string) error {
			return nil
		}

		os := mocks.NewMockOSInterface(ctrl)
		os.EXPECT().MkdirTemp("", "fireactions").Return("/tmp/fireactions", nil)

		rootfs, err := New("test", WithOS(os), WithMountFunc(mountFunc), WithUnmountFunc(unmountFunc))
		assert.Nil(t, err)
		assert.NotNil(t, rootfs)

		os.EXPECT().RemoveAll("/tmp/fireactions").Return(nil)

		err = rootfs.Close()

		assert.Nil(t, err)
	})

	t.Run("returns error if os.RemoveAll fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mountFunc := func(source string, target string) error {
			return nil
		}

		unmountFunc := func(target string) error {
			return nil
		}

		os := mocks.NewMockOSInterface(ctrl)
		os.EXPECT().MkdirTemp("", "fireactions").Return("/tmp/fireactions", nil)

		rootfs, err := New("test", WithOS(os), WithMountFunc(mountFunc), WithUnmountFunc(unmountFunc))
		assert.Nil(t, err)
		assert.NotNil(t, rootfs)

		os.EXPECT().RemoveAll("/tmp/fireactions").Return(errors.New("error"))

		err = rootfs.Close()

		assert.NotNil(t, err)
	})

	t.Run("returns error if unmountFunc fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mountFunc := func(source string, target string) error {
			return nil
		}

		unmountFunc := func(target string) error {
			return errors.New("error")
		}

		os := mocks.NewMockOSInterface(ctrl)
		os.EXPECT().MkdirTemp("", "fireactions").Return("/tmp/fireactions", nil)

		rootfs, err := New("test", WithOS(os), WithMountFunc(mountFunc), WithUnmountFunc(unmountFunc))
		assert.Nil(t, err)
		assert.NotNil(t, rootfs)

		err = rootfs.Close()

		assert.NotNil(t, err)
	})
}
