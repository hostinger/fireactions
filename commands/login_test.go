package commands

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/commands/mocks"
	"github.com/stretchr/testify/assert"
)

func TestLoginCmd_WithVMID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	client = mockClient

	cmd := newLoginCmd()
	cmd.SetArgs([]string{"test-vm-1"})

	// We can't actually test SSH execution, but we can verify the command structure
	assert.NotNil(t, cmd)
	assert.Equal(t, "login <vmid>", cmd.Use)
	assert.Equal(t, "SSH into a running VM as root user", cmd.Short)
}

func TestLoginCmd_RequiresVMID(t *testing.T) {
	cmd := newLoginCmd()
	cmd.SetArgs([]string{}) // No VM ID provided

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s)")
}

func TestLoginCmd_VMNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	client = mockClient

	mockClient.EXPECT().GetMicroVM(gomock.Any(), "nonexistent-vm").Return(nil, &fireactions.Response{}, errors.New("VM not found"))

	cmd := newLoginCmd()
	cmd.SetArgs([]string{"nonexistent-vm"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get VM details")
}

func TestLoginCmd_VMNoIPAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	client = mockClient

	vm := &fireactions.MicroVM{
		VMID:   "test-vm-starting",
		IPAddr: "", // No IP yet
	}

	mockClient.EXPECT().GetMicroVM(gomock.Any(), "test-vm-starting").Return(vm, &fireactions.Response{}, nil)

	cmd := newLoginCmd()
	cmd.SetArgs([]string{"test-vm-starting"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not have an IP address yet")
}
