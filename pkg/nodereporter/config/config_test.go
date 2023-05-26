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
	nodeName := "test-node1"

	viper.Set(flags.NamespaceFlag, namespace)
	viper.Set(flags.NodeNameFlag, nodeName)
	viper.Set(flags.FrequencyFlag, 90)
	viper.Set(flags.TimeoutFlag, 30)

	ret, err := GetConfig()

	assert.NoError(t, err)
	assert.NotNil(t, ret)
	assert.Equal(t, namespace, ret.Namespace)
	assert.Equal(t, nodeName, ret.NodeName)
}

func TestValidateConfigOk(t *testing.T) {
	cfg := &Config{
		NodeName:  "some",
		Timeout:   10,
		Frequency: 20,
		LeaseTime: 30,
	}
	err := validateConfig(cfg)
	assert.NoError(t, err)
}

func TestValidateConfigErrNodeName(t *testing.T) {
	cfg := &Config{
		Timeout:   10,
		Frequency: 20,
		LeaseTime: 30,
	}
	err := validateConfig(cfg)
	assert.Error(t, err)
}

func TestValidateConfigErrTimeout(t *testing.T) {
	cfg := &Config{
		Timeout:   -1,
		Frequency: 20,
		LeaseTime: 30,
		NodeName:  "asdads",
	}
	err := validateConfig(cfg)
	assert.Error(t, err)
}

func TestValidateConfigErrFrequency(t *testing.T) {
	cfg := &Config{
		Timeout:   10,
		Frequency: -1,
		LeaseTime: 30,
		NodeName:  "asdasd",
	}
	err := validateConfig(cfg)
	assert.Error(t, err)
}

func TestValidateConfigErrLease(t *testing.T) {
	cfg := &Config{
		Timeout:   10,
		Frequency: 20,
		LeaseTime: -1,
		NodeName:  "asdasd",
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
