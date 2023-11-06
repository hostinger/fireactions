package containerd

import (
	"context"
	"testing"

	"github.com/containerd/containerd/errdefs"
	"github.com/hostinger/fireactions/mocks/containerd"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPullImage(t *testing.T) {
	t.Run("ReturnsNoErrorOnSuccess", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		client.EXPECT().Pull(gomock.Any(), "nginx:latest", gomock.Any()).Return(&containerd.MockImage{}, nil)

		err := PullImage(context.Background(), client, "nginx:latest")

		assert.NoError(t, err)
	})

	t.Run("ReturnsErrorOnPullFailure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		client.EXPECT().Pull(gomock.Any(), "nginx:latest", gomock.Any()).Return(nil, assert.AnError)

		err := PullImage(context.Background(), client, "nginx:latest")

		assert.Error(t, err)
	})

	t.Run("ReturnsErrorOnInvalidImageRef", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		err := PullImage(context.Background(), client, "invalid@#!@")

		assert.Error(t, err)
	})
}

func TestImageExists(t *testing.T) {
	t.Run("ReturnsNoErrorAndTrueOnExistingImage", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		clientImage := containerd.NewMockImage(ctrl)

		client.EXPECT().GetImage(gomock.Any(), "test-image").Return(clientImage, nil)

		exists, err := ImageExists(context.Background(), client, "test-image")

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("ReturnsNoErrorAndFalseOnNotExistingImage", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)

		client.EXPECT().GetImage(gomock.Any(), "test-image").Return(nil, errdefs.ErrNotFound)

		exists, err := ImageExists(context.Background(), client, "test-image")

		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("ReturnsErrorAndFalseOnGetImageFailure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)

		client.EXPECT().GetImage(gomock.Any(), "test-image").Return(nil, assert.AnError)

		exists, err := ImageExists(context.Background(), client, "test-image")

		assert.Error(t, err)
		assert.False(t, exists)
	})
}
