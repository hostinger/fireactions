package commands

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/commands/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPoolsPauseCommand_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	mockClient.EXPECT().PausePool(gomock.Any(), "pool-name").Return(nil, nil)
	client = mockClient

	cmd := newPoolsPauseCmd()
	err := cmd.RunE(cmd, []string{"pool-name"})
	assert.Nil(t, err)
}

func TestPoolsPauseCommand_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	mockClient.EXPECT().PausePool(gomock.Any(), "pool-name").Return(nil, errors.New("error"))
	client = mockClient

	cmd := newPoolsPauseCmd()
	err := cmd.RunE(cmd, []string{"pool-name"})
	assert.Error(t, err)
}

func TestPoolsResumeCommand_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	mockClient.EXPECT().ResumePool(gomock.Any(), "pool-name").Return(nil, nil)
	client = mockClient

	cmd := newPoolsResumeCmd()
	err := cmd.RunE(cmd, []string{"pool-name"})
	assert.Nil(t, err)
}

func TestPoolsResumeCommand_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	mockClient.EXPECT().ResumePool(gomock.Any(), "pool-name").Return(nil, errors.New("error"))
	client = mockClient

	cmd := newPoolsResumeCmd()
	err := cmd.RunE(cmd, []string{"pool-name"})
	assert.Error(t, err)
}

func TestPoolsScaleCommand_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	mockClient.EXPECT().ScalePool(gomock.Any(), "pool-name").Return(nil, nil)
	client = mockClient

	cmd := newPoolsScaleCmd()
	err := cmd.Flags().Set("replicas", "1")
	if err != nil {
		t.Fatal(err)
	}

	err = newPoolsScaleCmd().RunE(cmd, []string{"pool-name", "1"})
	assert.Nil(t, err)
}

func TestPoolsScaleCommand_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	mockClient.EXPECT().ScalePool(gomock.Any(), "pool-name").Return(nil, errors.New("error"))
	client = mockClient

	cmd := newPoolsScaleCmd()
	err := cmd.Flags().Set("replicas", "1")
	if err != nil {
		t.Fatal(err)
	}

	err = cmd.RunE(cmd, []string{"pool-name"})
	assert.Error(t, err)
}

func TestPoolsShowCommand_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	mockClient.EXPECT().GetPool(gomock.Any(), "pool-name").Return(&fireactions.Pool{}, nil, nil)
	client = mockClient

	cmd := newPoolsShowCmd()
	err := cmd.RunE(cmd, []string{"pool-name"})
	assert.Nil(t, err)
}

func TestPoolsShowCommand_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	mockClient.EXPECT().GetPool(gomock.Any(), "pool-name").Return(nil, nil, errors.New("error"))
	client = mockClient

	cmd := newPoolsShowCmd()
	err := cmd.RunE(cmd, []string{"pool-name"})
	assert.Error(t, err)
}

func TestPoolsListCommand_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	mockClient.EXPECT().ListPools(gomock.Any(), nil).Return([]*fireactions.Pool{}, nil, nil)
	client = mockClient

	cmd := newPoolsListCmd()
	err := cmd.RunE(cmd, []string{})
	assert.Nil(t, err)
}

func TestPoolsListCommand_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	mockClient.EXPECT().ListPools(gomock.Any(), nil).Return(nil, nil, errors.New("error"))
	client = mockClient

	cmd := newPoolsListCmd()
	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}
