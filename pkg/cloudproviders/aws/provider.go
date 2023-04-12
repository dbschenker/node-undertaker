package aws

import (
	"context"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker/flags"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders"
	"github.com/spf13/viper"
)

type AwsCloudProvider struct {
	SqsUrl string
	Region string
}

func CreateAwsCloudProvider() AwsCloudProvider {
	ret := AwsCloudProvider{}
	ret.SqsUrl = viper.GetString(flags.AwsSqsUrlFlag)
	ret.Region = viper.GetString(flags.AwsRegionFlag)
	// TODO if region is empty perform an initialization
	return ret
}

func (t AwsCloudProvider) ReceiveMessages(context.Context) ([]cloudproviders.Message, error) {
	err := fmt.Errorf("TODO")
	return []cloudproviders.Message{}, err
}

func (t AwsCloudProvider) CompleteMessage(ctx context.Context, msg cloudproviders.Message) error {

	return fmt.Errorf("TODO")
}
