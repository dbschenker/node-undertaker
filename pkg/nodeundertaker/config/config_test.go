package config

import (
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker/flags"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestGetConfigNegativeValidation(t *testing.T) {
	viper.Set(flags.DrainDelayFlag, -1)
	_, err := GetConfig()
	assert.Error(t, err)
}

func TestGetConfigOk(t *testing.T) {
	viper.Set(flags.PortFlag, 1)
	ret, err := GetConfig()
	assert.Error(t, err)
	assert.NotNil(t, ret)
}

func TestValidateConfigOk(t *testing.T) {
	cfg := &Config{
		DrainDelay:            1,
		CloudTerminationDelay: 1,
		Port:                  8080,
		LeaseLockName:         "test",
	}
	err := validateConfig(cfg)
	assert.NoError(t, err)
}

func TestValidateConfigErrDrainDelay(t *testing.T) {
	cfg := &Config{
		DrainDelay:            -1,
		CloudTerminationDelay: 1,
		Port:                  8080,
		LeaseLockName:         "test",
	}
	err := validateConfig(cfg)
	assert.Error(t, err)
}

func TestValidateConfigErrCloudTerminationDelay(t *testing.T) {
	cfg := &Config{
		DrainDelay:            1,
		CloudTerminationDelay: -1,
		Port:                  8080,
		LeaseLockName:         "test",
	}
	err := validateConfig(cfg)
	assert.Error(t, err)
}

func TestValidateConfigErrNodeInitialThreshold(t *testing.T) {
	cfg := &Config{
		DrainDelay:            1,
		CloudTerminationDelay: 1,
		NodeInitialThreshold:  -1,
		Port:                  8080,
		LeaseLockName:         "test",
	}
	err := validateConfig(cfg)
	assert.Error(t, err)
}

func TestValidateConfigErrLeaseName(t *testing.T) {
	cfg := &Config{
		DrainDelay:            1,
		CloudTerminationDelay: 1,
		Port:                  8080,
		LeaseLockName:         "",
	}
	err := validateConfig(cfg)
	assert.Error(t, err)
}

func TestSetK8sClient(t *testing.T) {
	client := fake.NewSimpleClientset()
	currentNamespace := "test"
	cfg := Config{}
	cfg.SetK8sClient(client, currentNamespace)
	assert.Equal(t, currentNamespace, cfg.Namespace)
	assert.Equal(t, currentNamespace, cfg.LeaseLockNamespace)
	assert.Equal(t, client, cfg.K8sClient)
}

func TestSetK8sClient1(t *testing.T) {
	client := fake.NewSimpleClientset()
	currentNamespace := "test"
	leaseLockNs := "lease-lock-ns"
	appNamespace := "app-ns"
	cfg := Config{
		LeaseLockNamespace: leaseLockNs,
		Namespace:          appNamespace,
	}

	cfg.SetK8sClient(client, currentNamespace)
	assert.Equal(t, appNamespace, cfg.Namespace)
	assert.Equal(t, leaseLockNs, cfg.LeaseLockNamespace)
	assert.Equal(t, client, cfg.K8sClient)
}
