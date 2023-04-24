package nodeundertaker

import (
	"context"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	mock_observability "gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/observability/mocks"
	"github.com/golang/mock/gomock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"testing"
	"time"
)

func TestGetCloudProviderNoProvider(t *testing.T) {
	ctx := context.TODO()
	cloudProvider, err := getCloudProvider(ctx)

	assert.Nil(t, cloudProvider)
	assert.Error(t, err)
}

func TestGetCloudProviderUnknownProvider(t *testing.T) {
	ctx := context.TODO()
	viper.Set("cloud-provider", "unknown")
	cloudProvider, err := getCloudProvider(ctx)

	assert.Nil(t, cloudProvider)
	assert.Error(t, err)
}

func TestGetCloudProviderOk(t *testing.T) {
	ctx := context.TODO()
	viper.Set("cloud-provider", "aws")
	cloudProvider, err := getCloudProvider(ctx)

	assert.NotNil(t, cloudProvider)
	assert.NoError(t, err)
}

func TestStartLogicOk(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	errorMsg := "Error happened"

	observability := mock_observability.NewMockOBSERVABILITYSERVER(mockCtrl)
	observability.EXPECT().StartServer(gomock.Any()).Times(1).DoAndReturn(
		func(context context.Context) error {
			select {
			case <-context.Done():
				return fmt.Errorf(errorMsg)
			case <-time.After(1 * time.Second):
				return nil
			}
		})

	ctx := context.TODO()
	cfg := config.Config{}
	cfg.K8sClient = fake.NewSimpleClientset()
	resourceHandlerFuncs := cache.ResourceEventHandlerFuncs{}

	res := startLogic(ctx, &cfg, resourceHandlerFuncs, observability)
	assert.NoError(t, res)
}

func TestStartLogicNok(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	errorMsg := "Error happened"

	observability := mock_observability.NewMockOBSERVABILITYSERVER(mockCtrl)
	observability.EXPECT().StartServer(gomock.Any()).Times(1).DoAndReturn(
		func(context context.Context) error {
			return fmt.Errorf(errorMsg)
		})

	ctx := context.TODO()
	cfg := config.Config{}
	cfg.K8sClient = fake.NewSimpleClientset()

	resourceHandlerFuncs := cache.ResourceEventHandlerFuncs{}

	var res error
	assert.NotPanics(t,
		func() {
			res = startLogic(ctx, &cfg, resourceHandlerFuncs, observability)
		},
	)
	assert.EqualError(t, res, errorMsg)
}
