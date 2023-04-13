package aws

import (
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker/flags"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateAwsCloudProvider(t *testing.T) {
	dummyRegion := "eu-dummy-1z"
	viper.Set(flags.AwsRegionFlag, dummyRegion)

	dummySqs := "http://dummy.aws.com/queueuName"
	viper.Set(flags.AwsSqsUrlFlag, dummySqs)

	ret := CreateAwsCloudProvider()
	assert.Equal(t, dummyRegion, ret.Region)
	assert.Equal(t, dummySqs, ret.SqsUrl)
}
