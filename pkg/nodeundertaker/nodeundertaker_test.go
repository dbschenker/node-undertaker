package nodeundertaker

import (
	"context"
	"errors"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker/flags"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/kubeclient"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	mock_observability "gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/observability/mocks"
	"github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"testing"
	"time"
)

func TestGetCloudProviderNoProvider(t *testing.T) {
	ctx := context.TODO()
	cfg := config.Config{}
	cloudProvider, err := getCloudProvider(ctx, &cfg)

	assert.Nil(t, cloudProvider)
	assert.Error(t, err)
}

func TestGetCloudProviderUnknownProvider(t *testing.T) {
	ctx := context.TODO()
	cfg := config.Config{}
	viper.Set("cloud-provider", "unknown")
	cloudProvider, err := getCloudProvider(ctx, &cfg)

	assert.Nil(t, cloudProvider)
	assert.Error(t, err)
}

func TestGetCloudProviderOk(t *testing.T) {
	ctx := context.TODO()
	cfg := config.Config{}
	viper.Set("cloud-provider", "aws")
	cloudProvider, err := getCloudProvider(ctx, &cfg)

	assert.NotNil(t, cloudProvider)
	assert.NoError(t, err)
}

func TestGetCloudProviderKindOk(t *testing.T) {
	ctx := context.TODO()
	cfg := config.Config{}
	viper.Set("cloud-provider", "kind")
	cloudProvider, err := getCloudProvider(ctx, &cfg)

	assert.NotNil(t, cloudProvider)
	assert.NoError(t, err)
}

func TestGetCloudProviderKwokOk(t *testing.T) {
	ctx := context.TODO()
	cfg := config.Config{}
	viper.Set("cloud-provider", "kwok")
	cloudProvider, err := getCloudProvider(ctx, &cfg)

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

func TestCancelOnSigterm(t *testing.T) {
	counter := 0
	c := func() {
		counter += 1
	}
	cancelOnSigterm(c)
	assert.Equal(t, 0, counter)
}

func TestSetupLogLevelNok(t *testing.T) {
	err := setupLogging()
	assert.Error(t, err)
}

func TestSetupLogLevelOk(t *testing.T) {
	originalLvl := log.GetLevel()
	viper.Set(flags.LogLevelFlag, "error")
	viper.Set(flags.LogFormatFlag, flags.LogFormatJson)
	err := setupLogging()

	assert.NoError(t, err)
	assert.Equal(t, log.ErrorLevel, log.GetLevel())
	//cleanup
	log.SetLevel(originalLvl)
}

func TestExecuteWithContext(t *testing.T) {
	viper.Set(flags.LeaseLockNameFlag, "test-lease")
	viper.Set(flags.PortFlag, 0) //use random port
	viper.Set(flags.CloudProviderFlag, "kwok")

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	err := executeWithContext(ctx, kubeclient.GetFakeClient, cancel)
	assert.NoError(t, err)
}

func TestExecuteWithContextK8sErr(t *testing.T) {
	viper.Set(flags.LeaseLockNameFlag, "test-lease")
	viper.Set(flags.PortFlag, 0) //use random port
	viper.Set(flags.CloudProviderFlag, "kwok")

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	err := executeWithContext(ctx, func() (kubernetes.Interface, string, error) {
		return nil, "", errors.New("test error")
	}, cancel)
	assert.Error(t, err)
}
