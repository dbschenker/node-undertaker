package kwok

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/rand"
	"testing"
)

func TestCreateCloudProvider(t *testing.T) {
	_, err := CreateCloudProvider(context.TODO())
	assert.NoError(t, err)
}

func TestValidateConfig(t *testing.T) {
	cp, err := CreateCloudProvider(context.TODO())
	require.NoError(t, err)
	err = cp.ValidateConfig()
	assert.NoError(t, err)
}

func TestTerminateNode(t *testing.T) {
	ctx := context.TODO()
	cp, _ := CreateCloudProvider(ctx)
	clientset, err := StartCluster(t, ctx)
	require.NoError(t, err)
	cp.K8sClient = clientset

	nodeName := fmt.Sprintf("kwok-test-terminate-node-%s", rand.String(20))

	err = cp.CreateNode(ctx, nodeName)
	assert.NoError(t, err)

	ret, err := cp.TerminateNode(ctx, fmt.Sprintf("kwok://%s", nodeName))
	assert.NoError(t, err)
	assert.Equal(t, "InstanceTerminated", ret)

}
