package commands

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/commands/mocks"
	"github.com/stretchr/testify/assert"
)

func TestMicrovmsList_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	mockResponse := &fireactions.Response{}
	mockMicroVMs := &fireactions.MicroVMs{}
	mockClient.EXPECT().ListMicroVMs(gomock.Any(), "pool-name").Return(mockMicroVMs, mockResponse, nil)
	client = mockClient

	cmd := newMicrovmsListCmd()
	err := cmd.RunE(cmd, []string{"pool-name"})
	assert.Nil(t, err)
}

func TestMicrovmsList_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	mockClient.EXPECT().ListMicroVMs(gomock.Any(), "pool-name").Return(nil, nil, fmt.Errorf("mock API failure"))
	client = mockClient

	cmd := newMicrovmsListCmd()
	err := cmd.RunE(cmd, []string{"pool-name"})
	assert.ErrorContains(t, err, "mock API failure")
}

func TestMicrovmLoginCommand_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	mockClient.EXPECT().GetMicroVM(gomock.Any(), "vmid-123").Return(&fireactions.MicroVM{
		IPAddr: "192.168.1.10",
	}, nil, nil)
	client = mockClient

	cmd := newMicrovmLoginCmd()
	err := cmd.RunE(cmd, []string{"vmid-123"})
	assert.NoError(t, err)
}

func TestMicrovmLoginCommand_NoIP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	mockClient.EXPECT().GetMicroVM(gomock.Any(), "vmid-123").Return(&fireactions.MicroVM{
		// Empty IP address
		IPAddr: "",
	}, nil, nil)
	client = mockClient

	cmd := newMicrovmLoginCmd()
	err := cmd.RunE(cmd, []string{"vmid-123"})
	assert.ErrorContains(t, err, "does not have an IP address assigned")
}

func TestMicrovmLoginCommand_APIFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	mockClient.EXPECT().GetMicroVM(gomock.Any(), "vmid-123").Return(nil, nil, fmt.Errorf("API error"))
	client = mockClient

	cmd := newMicrovmLoginCmd()
	err := cmd.RunE(cmd, []string{"vmid-123"})
	assert.ErrorContains(t, err, "API error")
}
