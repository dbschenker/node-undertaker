package kwok

import (
	"context"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/rand"
	"testing"
)

func TestCreateCloudProvider(t *testing.T) {
	cfg := config.Config{}
	_, err := CreateCloudProvider(context.TODO(), &cfg)
	assert.NoError(t, err)
}

func TestValidateConfig(t *testing.T) {
	ctx := context.TODO()
	cfg := config.Config{}
	cp, _ := CreateCloudProvider(ctx, &cfg)
	err := cp.ValidateConfig()
	assert.NoError(t, err)
}

func TestTerminateNode(t *testing.T) {
	ctx := context.TODO()
	clientset, err := StartCluster(t, ctx)
	require.NoError(t, err)

	cfg := config.Config{
		K8sClient:             clientset,
		CloudTerminationDelay: 30,
	}
	cp, _ := CreateCloudProvider(ctx, &cfg)

	nodeName := fmt.Sprintf("kwok-test-terminate-node-%s", rand.String(20))

	err = cp.CreateNode(ctx, nodeName)
	assert.NoError(t, err)

	ret, err := cp.TerminateNode(ctx, fmt.Sprintf("kwok://%s", nodeName))
	assert.NoError(t, err)
	assert.Equal(t, "Instance Terminated", ret)

}
