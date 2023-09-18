package config

import (
	"github.com/dbschenker/node-undertaker/cmd/node-undertaker/flags"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes/fake"
	"os"
	"testing"
)

func TestGetConfigNegativeValidation(t *testing.T) {
	viper.Set(flags.DrainDelayFlag, -1)
	_, err := GetConfig()
	assert.Error(t, err)
}

func TestGetConfigOk(t *testing.T) {
	portValue := 1
	drainDelay := 29
	cloudTerminationDelay := 234
	cloudPrepareTerminationDelay := 544
	namespace := "ns1"
	leaseLockNamespace := "ns2"
	leaseLockName := "lease-lock1"
	hostname, _ := os.Hostname()

	viper.Set(flags.PortFlag, portValue)
	viper.Set(flags.DrainDelayFlag, drainDelay)
	viper.Set(flags.CloudTerminationDelayFlag, cloudTerminationDelay)
	viper.Set(flags.CloudPrepareTerminationDelayFlag, cloudPrepareTerminationDelay)

	viper.Set(flags.LeaseLockNamespaceFlag, leaseLockNamespace)
	viper.Set(flags.NamespaceFlag, namespace)
	viper.Set(flags.LeaseLockNameFlag, leaseLockName)

	ret, err := GetConfig()

	assert.NoError(t, err)
	assert.NotNil(t, ret)
	assert.Positive(t, ret.InformerResync)
	assert.Equal(t, portValue, ret.Port)
	assert.Equal(t, hostname, ret.Hostname)
	assert.Equal(t, leaseLockName, ret.LeaseLockName)
	assert.Equal(t, leaseLockNamespace, ret.LeaseLockNamespace)
	assert.Equal(t, drainDelay, ret.DrainDelay)
	assert.Equal(t, cloudPrepareTerminationDelay, ret.CloudPrepareTerminationDelay)
	assert.Equal(t, cloudTerminationDelay, ret.CloudTerminationDelay)
	assert.Equal(t, namespace, ret.Namespace)
	assert.Nil(t, ret.NodeSelector)
}

func TestGetConfigNodeSelectorNok(t *testing.T) {
	namespaceSelector := "__=9999"
	viper.Set(flags.NodeSelectorFlag, namespaceSelector)

	cfg, err := GetConfig()
	assert.Nil(t, cfg)
	assert.Error(t, err)
}

func TestGetConfigNodeSelectorOk1(t *testing.T) {
	namespaceSelector := "node.undertaker/powered=true"
	viper.Set(flags.NodeSelectorFlag, namespaceSelector)
	viper.Set(flags.LeaseLockNameFlag, "some-value")

	cfg, err := GetConfig()
	assert.NotNil(t, cfg)
	assert.NoError(t, err)
	assert.Len(t, cfg.NodeSelector, 1)
	requirements, _ := cfg.NodeSelector.Requirements()
	assert.Equal(t, selection.Operator("="), requirements[0].Operator())
	assert.Equal(t, "node.undertaker/powered", requirements[0].Key())
	assert.Equal(t, []string{"true"}, requirements[0].Values().List())

	testLabelSets := []labels.Set{
		labels.Set{"node.undertaker/powered": "true"},
		labels.Set{"node.undertaker/powered": "false"},
		labels.Set{"anyother": "false"},
		labels.Set{},
	}
	testLabelSetResults := []bool{true, false, false, false}

	for k := range testLabelSets {
		assert.Equal(t, testLabelSetResults[k], cfg.NodeSelector.Matches(testLabelSets[k]))
	}
}

func TestGetConfigNodeSelectorOk2(t *testing.T) {
	namespaceSelector := "karpenter!=true"

	viper.Set(flags.NodeSelectorFlag, namespaceSelector)
	viper.Set(flags.LeaseLockNameFlag, "some-value")

	cfg, err := GetConfig()
	assert.NotNil(t, cfg)
	assert.NoError(t, err)
	assert.Len(t, cfg.NodeSelector, 1)

	testLabelSets := []labels.Set{
		labels.Set{"karpenter": "false"},
		labels.Set{"karpenter": "true"},
		labels.Set{"anyother": "false"},
		labels.Set{},
	}
	testLabelSetResults := []bool{true, false, true, true}

	for k := range testLabelSets {
		assert.Equal(t, testLabelSetResults[k], cfg.NodeSelector.Matches(testLabelSets[k]))
	}
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

func TestValidateConfigErrInitialDelay(t *testing.T) {
	cfg := &Config{
		DrainDelay:            1,
		CloudTerminationDelay: 1,
		Port:                  8080,
		LeaseLockName:         "test",
		InitialDelay:          -1,
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
	assert.Equal(t, currentNamespace, cfg.NodeLeaseNamespace)
	assert.Equal(t, client, cfg.K8sClient)
}

func TestSetK8sClient1(t *testing.T) {
	client := fake.NewSimpleClientset()
	currentNamespace := "test"
	leaseLockNs := "lease-lock-ns"
	nodeLeaseNs := "node-leases"
	appNamespace := "app-ns"
	cfg := Config{
		LeaseLockNamespace: leaseLockNs,
		Namespace:          appNamespace,
		NodeLeaseNamespace: nodeLeaseNs,
	}

	cfg.SetK8sClient(client, currentNamespace)
	assert.Equal(t, appNamespace, cfg.Namespace)
	assert.Equal(t, leaseLockNs, cfg.LeaseLockNamespace)
	assert.Equal(t, nodeLeaseNs, cfg.NodeLeaseNamespace)
	assert.Equal(t, client, cfg.K8sClient)
}
