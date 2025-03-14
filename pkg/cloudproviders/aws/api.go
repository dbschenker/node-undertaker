package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
)

//go:generate mockgen -destination=./mocks/api_mocks.go github.com/dbschenker/node-undertaker/pkg/cloudproviders/aws EC2CLIENT,ELBCLIENT,ELBV2CLIENT,ASGCLIENT

type EC2CLIENT interface {
	TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
}

type ELBCLIENT interface {
	DeregisterInstancesFromLoadBalancer(ctx context.Context, params *elasticloadbalancing.DeregisterInstancesFromLoadBalancerInput, optFns ...func(*elasticloadbalancing.Options)) (*elasticloadbalancing.DeregisterInstancesFromLoadBalancerOutput, error)
}

type ELBV2CLIENT interface {
	DeregisterTargets(ctx context.Context, params *elasticloadbalancingv2.DeregisterTargetsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DeregisterTargetsOutput, error)
}

type ASGCLIENT interface {
	DescribeTrafficSources(ctx context.Context, params *autoscaling.DescribeTrafficSourcesInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeTrafficSourcesOutput, error)
	DescribeAutoScalingInstances(ctx context.Context, params *autoscaling.DescribeAutoScalingInstancesInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingInstancesOutput, error)
}
