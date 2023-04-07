package aws

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateConfigNoUrl(t *testing.T) {
	cloudProvider := AwsCloudProvider{}
	result := cloudProvider.ValidateConfig()
	assert.Error(t, result)
}

func TestValidateConfigWrongUrl(t *testing.T) {
	cloudProvider := AwsCloudProvider{
		SqsUrl: "test123/test",
	}
	result := cloudProvider.ValidateConfig()
	assert.Error(t, result)
}

func TestValidateConfigOk(t *testing.T) {
	cloudProvider := AwsCloudProvider{
		SqsUrl: "https://aws.com/test123/test",
	}
	result := cloudProvider.ValidateConfig()
	assert.NoError(t, result)
}
