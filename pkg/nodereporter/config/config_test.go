package config

import (
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-reporter/flags"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestGetConfigNegativeValidation(t *testing.T) {
	_, err := GetConfig()
	assert.Error(t, err)
}

func TestGetConfigOk(t *testing.T) {
	namespace := "ns1"

	viper.Set(flags.NamespaceFlag, namespace)

	ret, err := GetConfig()

	assert.NoError(t, err)
	assert.NotNil(t, ret)
	assert.Equal(t, namespace, ret.Namespace)
}

func TestValidateConfigOk(t *testing.T) {
	cfg := &Config{}
	err := validateConfig(cfg)
	assert.NoError(t, err)
}

func TestValidateConfigErrDrainDelay(t *testing.T) {
	cfg := &Config{}
	err := validateConfig(cfg)
	assert.Error(t, err)
}

func TestValidateConfigErrCloudTerminationDelay(t *testing.T) {
	cfg := &Config{}
	err := validateConfig(cfg)
	assert.Error(t, err)
}

func TestValidateConfigErrNodeInitialThreshold(t *testing.T) {
	cfg := &Config{}
	err := validateConfig(cfg)
	assert.Error(t, err)
}

func TestValidateConfigErrLeaseName(t *testing.T) {
	cfg := &Config{}
	err := validateConfig(cfg)
	assert.Error(t, err)
}

func TestSetK8sClient(t *testing.T) {
	client := fake.NewSimpleClientset()
	currentNamespace := "test"
	cfg := Config{}
	cfg.SetK8sClient(client, currentNamespace)
	assert.Equal(t, currentNamespace, cfg.Namespace)
	assert.Equal(t, client, cfg.K8sClient)
}

func TestSetK8sClient1(t *testing.T) {
	client := fake.NewSimpleClientset()
	currentNamespace := "test"
	appNamespace := "app-ns"
	cfg := Config{
		Namespace: appNamespace,
	}

	cfg.SetK8sClient(client, currentNamespace)
	assert.Equal(t, appNamespace, cfg.Namespace)
	assert.Equal(t, client, cfg.K8sClient)
}
