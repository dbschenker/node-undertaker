package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	autoscalingtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elasticloadbalancingtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elasticloadbalancingv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	awscloudproviderv1 "k8s.io/cloud-provider-aws/pkg/providers/v1"
)

type AwsCloudProvider struct {
	Ec2Client   EC2CLIENT
	ElbClient   ELBCLIENT
	Elbv2Client ELBV2CLIENT
	AsgClient   ASGCLIENT
}

const (
	TerminationEventActionFailed           = "Instance Termination Failed"
	TerminationEventActionSucceeded        = "Instance Terminated"
	PrepareTerminationEventActionFailed    = "Instance Preparation For Termination Failed"
	PrepareTerminationEventActionSucceeded = "Instance Prepared For Termination "
)

func CreateCloudProvider(ctx context.Context) (AwsCloudProvider, error) {
	ret := AwsCloudProvider{}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return ret, err
	}
	ret.Ec2Client = ec2.NewFromConfig(cfg)
	ret.AsgClient = autoscaling.NewFromConfig(cfg)
	ret.ElbClient = elasticloadbalancing.NewFromConfig(cfg)
	ret.Elbv2Client = elasticloadbalancingv2.NewFromConfig(cfg)
	return ret, nil
}

func (p AwsCloudProvider) TerminateNode(ctx context.Context, cloudProviderNodeId string) (string, error) {
	instanceId, err := awscloudproviderv1.KubernetesInstanceID(cloudProviderNodeId).MapToAWSInstanceID()
	if err != nil {
		return TerminationEventActionFailed, err
	}
	err = p.terminateInstance(ctx, string(instanceId))
	if err != nil {
		return TerminationEventActionFailed, err
	}
	return TerminationEventActionSucceeded, nil
}

func (p AwsCloudProvider) PrepareTermination(ctx context.Context, cloudProviderNodeId string) (string, error) {
	instanceId, err := awscloudproviderv1.KubernetesInstanceID(cloudProviderNodeId).MapToAWSInstanceID()
	if err != nil {
		return PrepareTerminationEventActionFailed, err
	}
	asgName, err := p.getAsgForInstance(ctx, string(instanceId))
	if err != nil {
		return PrepareTerminationEventActionFailed, err
	}
	if asgName != nil {
		ts, err := p.getTrafficSourcesForAsg(ctx, asgName)
		if err != nil {
			return PrepareTerminationEventActionFailed, err
		}
		if len(ts) > 0 {
			err := p.detachInstanceFromTrafficSources(ctx, ts, string(instanceId))
			if err != nil {
				return PrepareTerminationEventActionFailed, err
			}

		}
	}
	return PrepareTerminationEventActionSucceeded, nil
}

func (p AwsCloudProvider) terminateInstance(ctx context.Context, instanceId string) error {
	input := ec2.TerminateInstancesInput{
		InstanceIds: []string{
			string(instanceId),
		},
	}
	_, err := p.Ec2Client.TerminateInstances(ctx, &input)
	return err
}

func (p AwsCloudProvider) getAsgForInstance(ctx context.Context, instanceId string) (*string, error) {
	input := autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []string{
			instanceId,
		},
	}
	output, err := p.AsgClient.DescribeAutoScalingInstances(ctx, &input)
	if err != nil {
		return nil, err
	}
	if len(output.AutoScalingInstances) == 0 {
		return nil, nil
	} else if len(output.AutoScalingInstances) == 1 {
		return output.AutoScalingInstances[0].AutoScalingGroupName, nil
	}

	return nil, fmt.Errorf("AWS autoscaling API returned more than one ASG instance for instanceId: %s", instanceId)
}

func (p AwsCloudProvider) getTrafficSourcesForAsg(ctx context.Context, asgName *string) ([]autoscalingtypes.TrafficSourceState, error) {
	input := autoscaling.DescribeTrafficSourcesInput{
		AutoScalingGroupName: asgName,
	}
	output, err := p.AsgClient.DescribeTrafficSources(ctx, &input)
	if err != nil {
		return []autoscalingtypes.TrafficSourceState{}, err
	}
	ret := []autoscalingtypes.TrafficSourceState{}
	for i := range output.TrafficSources {
		if *output.TrafficSources[i].Type == "elb" || *output.TrafficSources[i].Type == "elbv2" {
			if *output.TrafficSources[i].State != "Removing" && *output.TrafficSources[i].State != "Removed" {
				ret = append(ret, output.TrafficSources[i])
			}
		}
	}

	return ret, err
}

func (p AwsCloudProvider) detachInstanceFromTrafficSources(ctx context.Context, sources []autoscalingtypes.TrafficSourceState, instanceId string) error {
	for i := range sources {
		if *sources[i].Type == "elb" {
			input := elasticloadbalancing.DeregisterInstancesFromLoadBalancerInput{
				LoadBalancerName: sources[i].Identifier,
				Instances: []elasticloadbalancingtypes.Instance{
					{InstanceId: &instanceId},
				},
			}
			_, err := p.ElbClient.DeregisterInstancesFromLoadBalancer(ctx, &input)
			if err != nil {
				return err
			}
		} else if *sources[i].Type == "elbv2" {
			input := elasticloadbalancingv2.DeregisterTargetsInput{
				TargetGroupArn: sources[i].Identifier,
				Targets: []elasticloadbalancingv2types.TargetDescription{
					{Id: &instanceId},
				},
			}
			_, err := p.Elbv2Client.DeregisterTargets(ctx, &input)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
