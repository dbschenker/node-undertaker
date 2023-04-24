package aws

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateAwsCloudProvider(t *testing.T) {
	ctx := context.TODO()

	ret, err := CreateAwsCloudProvider(ctx)
	assert.NoError(t, err)
	//assert.Equal(t, dummyRegion, ret.Region)
	assert.NotNil(t, ret)
}
