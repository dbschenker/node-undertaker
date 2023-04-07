package aws

import (
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker/flags"
	"github.com/spf13/viper"
)

type AwsCloudProvider struct {
	SqsUrl string
}

func CreateAwsCloudProvider() AwsCloudProvider {
	ret := AwsCloudProvider{}
	ret.SqsUrl = viper.GetString(flags.AwsSqsUrlFlag)
	return ret
}
