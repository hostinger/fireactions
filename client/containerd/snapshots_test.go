package containerd

import (
	"context"
	"testing"

	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/snapshots"
	"github.com/hostinger/fireactions/mocks/containerd"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCreateSnapshot(t *testing.T) {
	t.Run("ReturnsMountsOnExistingSnapshot", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		image := containerd.NewMockImage(ctrl)
		snapshotService := containerd.NewMockSnapshotter(ctrl)

		client.EXPECT().GetImage(gomock.Any(), "test-image").Return(image, nil)
		image.EXPECT().RootFS(gomock.Any()).Return(nil, nil)
		client.EXPECT().SnapshotService(gomock.Any()).Return(snapshotService)
		snapshotService.EXPECT().Stat(gomock.Any(), "test-snapshot").Return(snapshots.Info{}, nil)
		snapshotService.EXPECT().Mounts(gomock.Any(), "test-snapshot").Return([]mount.Mount{}, nil)

		_, err := CreateSnapshot(context.Background(), client, "test-image", "test-snapshotter", "test-snapshot")

		assert.NoError(t, err)
	})

	t.Run("ReturnsErrorOnMountsFailure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		image := containerd.NewMockImage(ctrl)
		snapshotService := containerd.NewMockSnapshotter(ctrl)

		client.EXPECT().GetImage(gomock.Any(), "test-image").Return(image, nil)
		image.EXPECT().RootFS(gomock.Any()).Return(nil, nil)
		client.EXPECT().SnapshotService(gomock.Any()).Return(snapshotService)
		snapshotService.EXPECT().Stat(gomock.Any(), "test-snapshot").Return(snapshots.Info{}, nil)
		snapshotService.EXPECT().Mounts(gomock.Any(), "test-snapshot").Return(nil, assert.AnError)

		_, err := CreateSnapshot(context.Background(), client, "test-image", "test-snapshotter", "test-snapshot")

		assert.Error(t, err)
	})

	t.Run("ReturnsErrorOnPrepareFailure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		image := containerd.NewMockImage(ctrl)
		snapshotService := containerd.NewMockSnapshotter(ctrl)

		client.EXPECT().GetImage(gomock.Any(), "test-image").Return(image, nil)
		image.EXPECT().RootFS(gomock.Any()).Return(nil, nil)
		client.EXPECT().SnapshotService(gomock.Any()).Return(snapshotService)
		snapshotService.EXPECT().Stat(gomock.Any(), "test-snapshot").Return(snapshots.Info{}, errdefs.ErrNotFound)
		snapshotService.EXPECT().Prepare(gomock.Any(), "test-snapshot", gomock.Any()).Return(nil, assert.AnError)

		_, err := CreateSnapshot(context.Background(), client, "test-image", "test-snapshotter", "test-snapshot")

		assert.Error(t, err)
	})

	t.Run("ReturnsErrorOnStatFailure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		image := containerd.NewMockImage(ctrl)
		snapshotService := containerd.NewMockSnapshotter(ctrl)

		client.EXPECT().GetImage(gomock.Any(), "test-image").Return(image, nil)
		image.EXPECT().RootFS(gomock.Any()).Return(nil, nil)
		client.EXPECT().SnapshotService(gomock.Any()).Return(snapshotService)
		snapshotService.EXPECT().Stat(gomock.Any(), "test-snapshot").Return(snapshots.Info{}, assert.AnError)

		_, err := CreateSnapshot(context.Background(), client, "test-image", "test-snapshotter", "test-snapshot")

		assert.Error(t, err)
	})

	t.Run("ReturnsErrorOnRootFSFailure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		image := containerd.NewMockImage(ctrl)

		client.EXPECT().GetImage(gomock.Any(), "test-image").Return(image, nil)
		image.EXPECT().RootFS(gomock.Any()).Return(nil, assert.AnError)

		_, err := CreateSnapshot(context.Background(), client, "test-image", "test-snapshotter", "test-snapshot")

		assert.Error(t, err)
	})

	t.Run("ReturnsErrorOnGetImageFailure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)

		client.EXPECT().GetImage(gomock.Any(), "test-image").Return(nil, assert.AnError)

		_, err := CreateSnapshot(context.Background(), client, "test-image", "test-snapshotter", "test-snapshot")

		assert.Error(t, err)
	})
}
