package aws

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateConfigOk(t *testing.T) {
	cloudProvider := AwsCloudProvider{}
	result := cloudProvider.ValidateConfig()
	assert.NoError(t, result)
}
