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

	mockClient := mocks.NewClient(ctrl)
	mockResponse := &fireactions.Response{}
	mockMicroVMs := &fireactions.MicroVMs{}
	mockClient.EXPECT().ListMicroVMs(gomock.Any(), gomock.Any()).Return(mockMicroVMs, mockResponse, nil)
	client = mockClient

	cmd := newMicrovmsListCmd()
	if err := cmd.Flags().Set("pool", "pool-name"); err != nil {
		t.Fatalf("failed to set flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{})
	assert.Nil(t, err)
}

func TestMicrovmsList_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewClient(ctrl)
	mockClient.EXPECT().ListMicroVMs(gomock.Any(), gomock.Any()).Return(nil, nil, fmt.Errorf("mock API failure"))
	client = mockClient

	cmd := newMicrovmsListCmd()
	if err := cmd.Flags().Set("pool", "pool-name"); err != nil {
		t.Fatalf("failed to set flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{})
	assert.ErrorContains(t, err, "mock API failure")
}
