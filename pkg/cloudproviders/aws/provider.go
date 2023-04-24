package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type AwsCloudProvider struct {
	Ec2Client EC2CLIENT
}

func CreateAwsCloudProvider(ctx context.Context) (AwsCloudProvider, error) {
	ret := AwsCloudProvider{}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return ret, err
	}
	ret.Ec2Client = ec2.NewFromConfig(cfg)
	return ret, nil
}

func (p AwsCloudProvider) TerminateNode(ctx context.Context, cloudProviderNodeId string) error {
	input := ec2.TerminateInstancesInput{
		InstanceIds: []string{
			cloudProviderNodeId,
		},
	}

	_, err := p.Ec2Client.TerminateInstances(ctx, &input)

	if err != nil {
		return err
	}

	return nil
}
