package commands

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hostinger/fireactions/commands/mocks"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRunReloadCmd_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	mockClient.EXPECT().Reload(gomock.Any()).Return(nil, nil)
	client = mockClient

	err := newReloadCmd().RunE(&cobra.Command{}, []string{})
	assert.Nil(t, err)
}

func TestRunReloadCmd_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockfireactionsClient(ctrl)
	mockClient.EXPECT().Reload(gomock.Any()).Return(nil, assert.AnError)
	client = mockClient

	err := newReloadCmd().RunE(&cobra.Command{}, []string{})
	assert.Equal(t, assert.AnError, err)
}
