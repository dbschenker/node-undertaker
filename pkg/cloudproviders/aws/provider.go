package aws

import (
	"context"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker/flags"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/spf13/viper"
)

type AwsCloudProvider struct {
	SqsUrl    string
	SqsClient SQSCLIENT
	Ec2Client EC2CLIENT
}

func CreateAwsCloudProvider(ctx context.Context) (AwsCloudProvider, error) {
	ret := AwsCloudProvider{}
	ret.SqsUrl = viper.GetString(flags.AwsSqsUrlFlag)

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return ret, err
	}
	ret.SqsClient = sqs.NewFromConfig(cfg)
	ret.Ec2Client = ec2.NewFromConfig(cfg)
	return ret, nil
}

func (t AwsCloudProvider) ReceiveMessages(context.Context) ([]cloudproviders.Message, error) {
	err := fmt.Errorf("TODO")
	return []cloudproviders.Message{}, err
}

func (t AwsCloudProvider) CompleteMessage(ctx context.Context, msg cloudproviders.Message) error {

	return fmt.Errorf("TODO")
}
