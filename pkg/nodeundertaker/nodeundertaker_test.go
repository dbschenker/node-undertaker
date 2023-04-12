package nodeundertaker

import (
	"context"
	"fmt"
	mock_k8snodeinformer "gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/k8snodeinformer/mocks"
	mockNHNH "gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodehealthnotificationhandler/mocks"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	"github.com/golang/mock/gomock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetCloudProviderNoProvider(t *testing.T) {
	cloudProvider, err := getCloudProvider()

	assert.Nil(t, cloudProvider)
	assert.Error(t, err)
}

func TestGetCloudProviderUnknownProvider(t *testing.T) {
	viper.Set("cloud-provider", "unknown")
	cloudProvider, err := getCloudProvider()

	assert.Nil(t, cloudProvider)
	assert.Error(t, err)
}

func TestGetCloudProviderOk(t *testing.T) {
	viper.Set("cloud-provider", "aws")
	cloudProvider, err := getCloudProvider()

	assert.NotNil(t, cloudProvider)
	assert.NoError(t, err)
}

func TestStartLogicOk(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	nodeHealthNotificationHandler := mockNHNH.NewMockNODEHEALTHNOTIFICATIONHANDLER(mockCtrl)

	nodeHealthNotificationHandler.EXPECT().HandleHealthMessages(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
		func(context context.Context, cfg *config.Config) error {
			select {
			case <-context.Done():
				return fmt.Errorf("finished with error")
			case <-time.After(1 * time.Second):
				return nil
			}
		})

	k8sNodeInformer := mock_k8snodeinformer.NewMockK8SNODEINFORMER(mockCtrl)
	k8sNodeInformer.EXPECT().StartInformer(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
		func(context context.Context, cfg *config.Config) error {
			select {
			case <-context.Done():
				return fmt.Errorf("finished with error")
			case <-time.After(1 * time.Second):
				return nil
			}
		})

	ctx := context.TODO()
	cfg := config.Config{}
	res := startLogic(ctx, &cfg, nodeHealthNotificationHandler, k8sNodeInformer)
	assert.NoError(t, res)
}

func TestStartLogicNok(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	nodeHealthNotificationHandler := mockNHNH.NewMockNODEHEALTHNOTIFICATIONHANDLER(mockCtrl)

	nodeHealthNotificationHandler.EXPECT().HandleHealthMessages(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
		func(context context.Context, cfg *config.Config) error {
			select {
			case <-context.Done():
				return fmt.Errorf("terminated prepaturely")
			case <-time.After(10 * time.Second):
				panic("shouldn't happen")
			}
		})

	errorMsg := "Error happened"

	k8sNodeInformer := mock_k8snodeinformer.NewMockK8SNODEINFORMER(mockCtrl)
	k8sNodeInformer.EXPECT().StartInformer(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
		func(context context.Context, cfg *config.Config) error {
			return fmt.Errorf(errorMsg)
		})

	ctx := context.TODO()
	cfg := config.Config{}
	var res error
	assert.NotPanics(t,
		func() {
			res = startLogic(ctx, &cfg, nodeHealthNotificationHandler, k8sNodeInformer)
		},
	)
	assert.EqualError(t, res, errorMsg)
}
