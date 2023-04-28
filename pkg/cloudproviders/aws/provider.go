package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awscloudproviderv1 "k8s.io/cloud-provider-aws/pkg/providers/v1"
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
	instanceId, err := awscloudproviderv1.KubernetesInstanceID(cloudProviderNodeId).MapToAWSInstanceID()
	if err != nil {
		return err
	}
	input := ec2.TerminateInstancesInput{
		InstanceIds: []string{
			string(instanceId),
		},
	}

	_, err = p.Ec2Client.TerminateInstances(ctx, &input)

	return err
}
