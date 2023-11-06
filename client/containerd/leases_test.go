package containerd

import (
	"context"
	"testing"

	"github.com/containerd/containerd/leases"
	"github.com/hostinger/fireactions/mocks/containerd"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewContextWithLease(t *testing.T) {
	t.Run("ReturnsNoErrorOnNotExistingLease", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		clientLeaseManager := containerd.NewMockLeasesManager(ctrl)

		client.EXPECT().LeasesService().Return(clientLeaseManager)
		clientLeaseManager.EXPECT().List(gomock.Any(), gomock.Any()).Return([]leases.Lease{}, nil)
		clientLeaseManager.EXPECT().Create(gomock.Any(), gomock.Any()).Return(leases.Lease{}, nil)

		_, err := NewContextWithLease(context.Background(), client, "test-lease")

		assert.NoError(t, err)
	})

	t.Run("ReturnsNoErrorOnExistingLease", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		clientLeaseManager := containerd.NewMockLeasesManager(ctrl)

		client.EXPECT().LeasesService().Return(clientLeaseManager)
		clientLeaseManager.EXPECT().List(gomock.Any(), gomock.Any()).Return([]leases.Lease{
			{
				ID: "another-lease",
			},
			{
				ID: "test-lease",
			},
		}, nil)

		_, err := NewContextWithLease(context.Background(), client, "test-lease")

		assert.NoError(t, err)
	})

	t.Run("ReturnsErrorOnCreateFailure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		clientLeaseManager := containerd.NewMockLeasesManager(ctrl)

		client.EXPECT().LeasesService().Return(clientLeaseManager)
		clientLeaseManager.EXPECT().List(gomock.Any(), gomock.Any()).Return([]leases.Lease{}, nil)
		clientLeaseManager.EXPECT().Create(gomock.Any(), gomock.Any()).Return(leases.Lease{}, assert.AnError)

		_, err := NewContextWithLease(context.Background(), client, "test-lease")

		assert.Error(t, err)
	})
}

func TestCreateLease(t *testing.T) {
	t.Run("ReturnsNewLease", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		clientLeaseManager := containerd.NewMockLeasesManager(ctrl)

		client.EXPECT().LeasesService().Return(clientLeaseManager)
		clientLeaseManager.EXPECT().List(gomock.Any(), gomock.Any()).Return([]leases.Lease{}, nil)
		clientLeaseManager.EXPECT().Create(gomock.Any(), gomock.Any()).Return(leases.Lease{}, nil)

		_, err := CreateLease(context.Background(), client, "test-lease")

		assert.NoError(t, err)
	})

	t.Run("ReturnsOldLease", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		clientLeaseManager := containerd.NewMockLeasesManager(ctrl)

		client.EXPECT().LeasesService().Return(clientLeaseManager)
		clientLeaseManager.EXPECT().List(gomock.Any(), gomock.Any()).Return([]leases.Lease{
			{
				ID: "another-lease",
			},
			{
				ID: "test-lease",
			},
		}, nil)

		lease, err := CreateLease(context.Background(), client, "test-lease")

		assert.NoError(t, err)
		assert.Equal(t, "test-lease", lease.ID)
	})

	t.Run("ReturnsErrorOnLeaseCreateFailure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		clientLeaseManager := containerd.NewMockLeasesManager(ctrl)

		client.EXPECT().LeasesService().Return(clientLeaseManager)
		clientLeaseManager.EXPECT().List(gomock.Any(), gomock.Any()).Return([]leases.Lease{}, nil)
		clientLeaseManager.EXPECT().Create(gomock.Any(), gomock.Any()).Return(leases.Lease{}, assert.AnError)

		_, err := CreateLease(context.Background(), client, "test-lease")

		assert.Error(t, err)
	})

	t.Run("ReturnsErrorOnLeaseListFailure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		clientLeaseManager := containerd.NewMockLeasesManager(ctrl)

		client.EXPECT().LeasesService().Return(clientLeaseManager)
		clientLeaseManager.EXPECT().List(gomock.Any(), gomock.Any()).Return([]leases.Lease{}, assert.AnError)

		_, err := CreateLease(context.Background(), client, "test-lease")

		assert.Error(t, err)
	})
}

func TestDeleteLease(t *testing.T) {
	t.Run("ReturnsNoErrorOnExistingLease", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		clientLeaseManager := containerd.NewMockLeasesManager(ctrl)

		client.EXPECT().LeasesService().Return(clientLeaseManager)
		clientLeaseManager.EXPECT().List(gomock.Any(), gomock.Any()).Return([]leases.Lease{
			{
				ID: "another-lease",
			},
			{
				ID: "test-lease",
			},
		}, nil)
		clientLeaseManager.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)

		err := DeleteLease(context.Background(), client, "test-lease")

		assert.NoError(t, err)
	})

	t.Run("ReturnsNoErrorOnNotExistingLease", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		clientLeaseManager := containerd.NewMockLeasesManager(ctrl)

		client.EXPECT().LeasesService().Return(clientLeaseManager)
		clientLeaseManager.EXPECT().List(gomock.Any(), gomock.Any()).Return([]leases.Lease{
			{
				ID: "another-lease",
			},
		}, nil)

		err := DeleteLease(context.Background(), client, "test-lease")

		assert.NoError(t, err)
	})

	t.Run("ReturnsErrorOnLeaseDeleteFailure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		clientLeaseManager := containerd.NewMockLeasesManager(ctrl)

		client.EXPECT().LeasesService().Return(clientLeaseManager)
		clientLeaseManager.EXPECT().List(gomock.Any(), gomock.Any()).Return([]leases.Lease{
			{
				ID: "another-lease",
			},
			{
				ID: "test-lease",
			},
		}, nil)
		clientLeaseManager.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(assert.AnError)

		err := DeleteLease(context.Background(), client, "test-lease")

		assert.Error(t, err)
	})

	t.Run("ReturnsErrorOnLeaseListFailure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := containerd.NewMockClient(ctrl)
		clientLeaseManager := containerd.NewMockLeasesManager(ctrl)

		client.EXPECT().LeasesService().Return(clientLeaseManager)
		clientLeaseManager.EXPECT().List(gomock.Any(), gomock.Any()).Return([]leases.Lease{}, assert.AnError)

		err := DeleteLease(context.Background(), client, "test-lease")

		assert.Error(t, err)
	})
}
