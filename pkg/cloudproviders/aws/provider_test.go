package aws

import (
	"context"
	mockaws "gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders/aws/mocks"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	aws "k8s.io/cloud-provider-aws/pkg/providers/v1"
	"testing"
)

func TestCreateAwsCloudProvider(t *testing.T) {
	ctx := context.TODO()

	ret, err := CreateAwsCloudProvider(ctx)
	assert.NoError(t, err)
	//assert.Equal(t, dummyRegion, ret.Region)
	assert.NotNil(t, ret)
}

func TestTerminateNode(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ec2Client := mockaws.NewMockEC2CLIENT(mockCtrl)
	instanceId := aws.InstanceID("i-123")
	expectedInput := ec2.TerminateInstancesInput{
		InstanceIds: []string{
			string(instanceId),
		},
	}

	ec2Client.EXPECT().TerminateInstances(gomock.Any(), &expectedInput).Return(nil, nil).Times(1)

	cloudProvider := AwsCloudProvider{
		Ec2Client: ec2Client,
	}
	res, err := cloudProvider.TerminateNode(context.TODO(), "aws://nonexistant/i-123")
	assert.NoError(t, err)
	assert.Equal(t, TerminationEventActionSucceeded, res)
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
