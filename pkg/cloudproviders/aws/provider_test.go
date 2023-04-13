package aws

import (
	"context"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker/flags"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateAwsCloudProvider(t *testing.T) {
	ctx := context.TODO()

	dummySqs := "http://dummy.aws.com/queueuName"
	viper.Set(flags.AwsSqsUrlFlag, dummySqs)

	ret, err := CreateAwsCloudProvider(ctx)
	assert.NoError(t, err)
	//assert.Equal(t, dummyRegion, ret.Region)
	assert.Equal(t, dummySqs, ret.SqsUrl)
}
