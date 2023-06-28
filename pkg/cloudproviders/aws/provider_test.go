package aws

import (
	"context"
	"errors"
	mockaws "gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders/aws/mocks"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	autoscalingtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elasticloadbalancingtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elasticloadbalancingv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateAwsCloudProvider(t *testing.T) {
	ctx := context.TODO()

	ret, err := CreateCloudProvider(ctx)
	assert.NoError(t, err)
	//assert.Equal(t, dummyRegion, ret.Region)
	assert.NotNil(t, ret)
	assert.NotNil(t, ret.AsgClient)
	assert.NotNil(t, ret.ElbClient)
	assert.NotNil(t, ret.ElbClient)
	assert.NotNil(t, ret.Ec2Client)
}

func TestTerminatNode(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ec2Client := mockaws.NewMockEC2CLIENT(mockCtrl)

	instanceId := "i-12312313"

	expectedInput := ec2.TerminateInstancesInput{
		InstanceIds: []string{
			instanceId,
		},
	}

	ec2Client.EXPECT().TerminateInstances(gomock.Any(), &expectedInput).Return(nil, nil).Times(1)

	cloudProvider := AwsCloudProvider{
		Ec2Client: ec2Client,
	}

	res, err := cloudProvider.TerminateNode(context.TODO(), "aws://nonexistant/"+instanceId)
	assert.NoError(t, err)
	assert.Equal(t, TerminationEventActionSucceeded, res)
}

func TestPrepareTerminationNodeNotInLB(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	elbClient := mockaws.NewMockELBCLIENT(mockCtrl)
	elbv2Client := mockaws.NewMockELBV2CLIENT(mockCtrl)
	asgClient := mockaws.NewMockASGCLIENT(mockCtrl)

	instanceId := "i-12312313"

	expectedAsgInput := autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []string{
			instanceId,
		},
	}
	expectedAsgOutput := autoscaling.DescribeAutoScalingInstancesOutput{
		AutoScalingInstances: []autoscalingtypes.AutoScalingInstanceDetails{},
	}

	asgClient.EXPECT().DescribeAutoScalingInstances(gomock.Any(), &expectedAsgInput).Return(&expectedAsgOutput, nil).Times(1)

	cloudProvider := AwsCloudProvider{
		AsgClient:   asgClient,
		Elbv2Client: elbv2Client,
		ElbClient:   elbClient,
	}

	res, err := cloudProvider.PrepareTermination(context.TODO(), "aws://nonexistant/"+instanceId)
	assert.NoError(t, err)
	assert.Equal(t, PrepareTerminationEventActionSucceeded, res)
}

func TestPrepareTerminationNodeInMultipleLB(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ec2Client := mockaws.NewMockEC2CLIENT(mockCtrl)
	elbClient := mockaws.NewMockELBCLIENT(mockCtrl)
	elbv2Client := mockaws.NewMockELBV2CLIENT(mockCtrl)
	asgClient := mockaws.NewMockASGCLIENT(mockCtrl)

	instanceId := "i-12312313"
	asgName := "asg-name-1"

	expectedAsgInput := autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []string{
			instanceId,
		},
	}
	expectedAsgOutput := autoscaling.DescribeAutoScalingInstancesOutput{
		AutoScalingInstances: []autoscalingtypes.AutoScalingInstanceDetails{
			{
				InstanceId:           &instanceId,
				AutoScalingGroupName: &asgName,
			},
		},
	}

	asgClient.EXPECT().DescribeAutoScalingInstances(gomock.Any(), &expectedAsgInput).Return(&expectedAsgOutput, nil).Times(1)

	trafficSourceType1 := "elb"
	trafficSourceState1 := "Added"
	trafficSourceIdentifier1 := "lb-name1"
	trafficSourceType2 := "elbv2"
	trafficSourceState2 := "InService"
	trafficSourceIdentifier2 := "arn:aws:elbv2"
	trafficSourceType3 := "elb"
	trafficSourceState3 := "Removing"
	trafficSourceIdentifier3 := "lb-name1"

	expectedTrafficSourcesInput := autoscaling.DescribeTrafficSourcesInput{
		AutoScalingGroupName: &asgName,
	}
	expectedTrafficSourcesOutput := autoscaling.DescribeTrafficSourcesOutput{
		TrafficSources: []autoscalingtypes.TrafficSourceState{
			{
				Type:       &trafficSourceType1,
				State:      &trafficSourceState1,
				Identifier: &trafficSourceIdentifier1,
			},
			{
				Type:       &trafficSourceType2,
				State:      &trafficSourceState2,
				Identifier: &trafficSourceIdentifier2,
			},
			{
				Type:       &trafficSourceType3,
				State:      &trafficSourceState3,
				Identifier: &trafficSourceIdentifier3,
			},
		},
	}
	asgClient.EXPECT().DescribeTrafficSources(gomock.Any(), &expectedTrafficSourcesInput).Return(&expectedTrafficSourcesOutput, nil).Times(1)

	expectedInput1 := elasticloadbalancing.DeregisterInstancesFromLoadBalancerInput{
		LoadBalancerName: &trafficSourceIdentifier1,
		Instances: []elasticloadbalancingtypes.Instance{
			{InstanceId: &instanceId},
		},
	}
	elbClient.EXPECT().DeregisterInstancesFromLoadBalancer(gomock.Any(), &expectedInput1).Return(nil, nil).Times(1)

	expectedInput2 := elasticloadbalancingv2.DeregisterTargetsInput{
		TargetGroupArn: &trafficSourceIdentifier2,
		Targets: []elasticloadbalancingv2types.TargetDescription{
			{Id: &instanceId},
		},
	}
	elbv2Client.EXPECT().DeregisterTargets(gomock.Any(), &expectedInput2).Return(nil, nil).Times(1)

	cloudProvider := AwsCloudProvider{
		Ec2Client:   ec2Client,
		AsgClient:   asgClient,
		Elbv2Client: elbv2Client,
		ElbClient:   elbClient,
	}

	res, err := cloudProvider.PrepareTermination(context.TODO(), "aws://nonexistant/"+instanceId)
	assert.NoError(t, err)
	assert.Equal(t, PrepareTerminationEventActionSucceeded, res)
}

func TestTerminateNodeWrongProviderId(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ec2Client := mockaws.NewMockEC2CLIENT(mockCtrl)

	cloudProvider := AwsCloudProvider{
		Ec2Client: ec2Client,
	}
	res, err := cloudProvider.TerminateNode(context.TODO(), "test123")
	assert.Error(t, err)
	assert.Equal(t, TerminationEventActionFailed, res)
}

func TestPrepareTerminationNodeWrongProviderId(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ec2Client := mockaws.NewMockEC2CLIENT(mockCtrl)

	cloudProvider := AwsCloudProvider{
		Ec2Client: ec2Client,
	}
	res, err := cloudProvider.PrepareTermination(context.TODO(), "test123")
	assert.Error(t, err)
	assert.Equal(t, PrepareTerminationEventActionFailed, res)
}

func TestGetAsgForInstanceNone(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	asgClient := mockaws.NewMockASGCLIENT(mockCtrl)
	instanceId := "i-123"
	expectedInput := autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []string{
			instanceId,
		},
	}
	expectedOutput := autoscaling.DescribeAutoScalingInstancesOutput{
		AutoScalingInstances: []autoscalingtypes.AutoScalingInstanceDetails{},
	}
	asgClient.EXPECT().DescribeAutoScalingInstances(gomock.Any(), &expectedInput).Return(&expectedOutput, nil).Times(1)

	p := AwsCloudProvider{
		AsgClient: asgClient,
	}

	ret, err := p.getAsgForInstance(context.TODO(), instanceId)
	assert.Nil(t, ret)
	assert.NoError(t, err)
}

func TestGetAsgForInstanceOne(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	asgClient := mockaws.NewMockASGCLIENT(mockCtrl)
	instanceId := "i-123"
	asgName := "test-asg-1"
	expectedInput := autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []string{
			instanceId,
		},
	}
	expectedOutput := autoscaling.DescribeAutoScalingInstancesOutput{
		AutoScalingInstances: []autoscalingtypes.AutoScalingInstanceDetails{
			{
				InstanceId:           &instanceId,
				AutoScalingGroupName: &asgName,
			},
		},
	}
	asgClient.EXPECT().DescribeAutoScalingInstances(gomock.Any(), &expectedInput).Return(&expectedOutput, nil).Times(1)

	p := AwsCloudProvider{
		AsgClient: asgClient,
	}

	ret, err := p.getAsgForInstance(context.TODO(), instanceId)
	assert.NotNil(t, ret)
	assert.Equal(t, asgName, *ret)
	assert.NoError(t, err)
}

func TestGetAsgForInstanceErr(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	asgClient := mockaws.NewMockASGCLIENT(mockCtrl)
	instanceId := "i-123"
	asgName := "test-asg-1"
	errorReturned := errors.New("test-error")
	expectedInput := autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []string{
			instanceId,
		},
	}
	expectedOutput := autoscaling.DescribeAutoScalingInstancesOutput{
		AutoScalingInstances: []autoscalingtypes.AutoScalingInstanceDetails{
			{
				InstanceId:           &instanceId,
				AutoScalingGroupName: &asgName,
			},
		},
	}
	asgClient.EXPECT().DescribeAutoScalingInstances(gomock.Any(), &expectedInput).Return(&expectedOutput, errorReturned).Times(1)

	p := AwsCloudProvider{
		AsgClient: asgClient,
	}

	ret, err := p.getAsgForInstance(context.TODO(), instanceId)
	assert.Nil(t, ret)
	assert.Error(t, err)
}

func TestGetAsgForInstanceErrMore(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	asgClient := mockaws.NewMockASGCLIENT(mockCtrl)
	instanceId := "i-123"
	asgName := "test-asg-1"
	expectedInput := autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []string{
			instanceId,
		},
	}
	expectedOutput := autoscaling.DescribeAutoScalingInstancesOutput{
		AutoScalingInstances: []autoscalingtypes.AutoScalingInstanceDetails{
			{
				InstanceId:           &instanceId,
				AutoScalingGroupName: &asgName,
			},
			{
				InstanceId:           &instanceId,
				AutoScalingGroupName: &asgName,
			},
		},
	}
	asgClient.EXPECT().DescribeAutoScalingInstances(gomock.Any(), &expectedInput).Return(&expectedOutput, nil).Times(1)

	p := AwsCloudProvider{
		AsgClient: asgClient,
	}

	ret, err := p.getAsgForInstance(context.TODO(), instanceId)
	assert.Nil(t, ret)
	assert.Error(t, err)
}

func TestGetTrafficSourcesForAsgNone(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	asgClient := mockaws.NewMockASGCLIENT(mockCtrl)
	asgName := "test-asg-1"
	expectedInput := autoscaling.DescribeTrafficSourcesInput{
		AutoScalingGroupName: &asgName,
	}
	expectedOutput := autoscaling.DescribeTrafficSourcesOutput{
		TrafficSources: []autoscalingtypes.TrafficSourceState{},
	}
	asgClient.EXPECT().DescribeTrafficSources(gomock.Any(), &expectedInput).Return(&expectedOutput, nil).Times(1)

	p := AwsCloudProvider{
		AsgClient: asgClient,
	}

	ret, err := p.getTrafficSourcesForAsg(context.TODO(), &asgName)
	assert.Empty(t, ret)
	assert.NoError(t, err)
}

func TestGetTrafficSourcesForAsgOk(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	asgClient := mockaws.NewMockASGCLIENT(mockCtrl)
	asgName := "test-asg-1"
	expectedInput := autoscaling.DescribeTrafficSourcesInput{
		AutoScalingGroupName: &asgName,
	}

	trafficSourceType1 := "elb"
	trafficSourceState1 := "Added"
	trafficSourceIdentifier1 := "lb-name1"
	trafficSourceType2 := "elbv2"
	trafficSourceState2 := "InService"
	trafficSourceIdentifier2 := "arn:aws:elbv2"
	trafficSourceType3 := "elb"
	trafficSourceState3 := "Removing"
	trafficSourceIdentifier3 := "lb-name1"
	expectedOutput := autoscaling.DescribeTrafficSourcesOutput{
		TrafficSources: []autoscalingtypes.TrafficSourceState{
			{
				Type:       &trafficSourceType1,
				State:      &trafficSourceState1,
				Identifier: &trafficSourceIdentifier1,
			},
			{
				Type:       &trafficSourceType2,
				State:      &trafficSourceState2,
				Identifier: &trafficSourceIdentifier2,
			},
			{
				Type:       &trafficSourceType3,
				State:      &trafficSourceState3,
				Identifier: &trafficSourceIdentifier3,
			},
		},
	}
	asgClient.EXPECT().DescribeTrafficSources(gomock.Any(), &expectedInput).Return(&expectedOutput, nil).Times(1)

	p := AwsCloudProvider{
		AsgClient: asgClient,
	}

	ret, err := p.getTrafficSourcesForAsg(context.TODO(), &asgName)
	assert.NoError(t, err)
	assert.Len(t, ret, 2)
	assert.Equal(t, trafficSourceType1, *ret[0].Type)
	assert.Equal(t, trafficSourceIdentifier1, *ret[0].Identifier)
	assert.Equal(t, trafficSourceState1, *ret[0].State)
	assert.Equal(t, trafficSourceType2, *ret[1].Type)
	assert.Equal(t, trafficSourceIdentifier2, *ret[1].Identifier)
	assert.Equal(t, trafficSourceState2, *ret[1].State)
}

func TestDetachInstanceFromTrafficSources(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	elbClient := mockaws.NewMockELBCLIENT(mockCtrl)
	elbv2Client := mockaws.NewMockELBV2CLIENT(mockCtrl)
	p := AwsCloudProvider{
		ElbClient:   elbClient,
		Elbv2Client: elbv2Client,
	}

	instanceId := "i-123124"

	trafficSourceType1 := "elb"
	trafficSourceState1 := "Added"
	trafficSourceIdentifier1 := "lb-name1"
	trafficSourceType2 := "elbv2"
	trafficSourceState2 := "InService"
	trafficSourceIdentifier2 := "arn:aws:elbv:target-group:123322"
	trafficSourceType3 := "elb"
	trafficSourceState3 := "Adding"
	trafficSourceIdentifier3 := "lb-name1"
	sources := []autoscalingtypes.TrafficSourceState{
		{
			Type:       &trafficSourceType1,
			State:      &trafficSourceState1,
			Identifier: &trafficSourceIdentifier1,
		},
		{
			Type:       &trafficSourceType2,
			State:      &trafficSourceState2,
			Identifier: &trafficSourceIdentifier2,
		},
		{
			Type:       &trafficSourceType3,
			State:      &trafficSourceState3,
			Identifier: &trafficSourceIdentifier3,
		},
	}

	expectedInput1 := elasticloadbalancing.DeregisterInstancesFromLoadBalancerInput{
		LoadBalancerName: &trafficSourceIdentifier1,
		Instances: []elasticloadbalancingtypes.Instance{
			{InstanceId: &instanceId},
		},
	}
	elbClient.EXPECT().DeregisterInstancesFromLoadBalancer(gomock.Any(), &expectedInput1).Return(nil, nil).Times(1)

	expectedInput2 := elasticloadbalancingv2.DeregisterTargetsInput{
		TargetGroupArn: &trafficSourceIdentifier2,
		Targets: []elasticloadbalancingv2types.TargetDescription{
			{Id: &instanceId},
		},
	}
	elbv2Client.EXPECT().DeregisterTargets(gomock.Any(), &expectedInput2).Return(nil, nil).Times(1)

	expectedInput3 := elasticloadbalancing.DeregisterInstancesFromLoadBalancerInput{
		LoadBalancerName: &trafficSourceIdentifier3,
		Instances: []elasticloadbalancingtypes.Instance{
			{InstanceId: &instanceId},
		},
	}
	elbClient.EXPECT().DeregisterInstancesFromLoadBalancer(gomock.Any(), &expectedInput3).Return(nil, nil).Times(1)

	err := p.detachInstanceFromTrafficSources(context.TODO(), sources, instanceId)
	assert.NoError(t, err)
}
