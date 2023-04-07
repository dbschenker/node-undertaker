package nodeundertaker

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
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
