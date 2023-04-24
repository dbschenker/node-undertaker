package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

//go:generate mockgen -destination=./mocks/api_mocks.go gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders/aws EC2CLIENT

type EC2CLIENT interface {
	TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
}
