package kwok

import (
	"context"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"testing"
)

func TestCreateNode(t *testing.T) {
	ctx := context.TODO()

	clientset, err := StartCluster(t, ctx)
	require.NoError(t, err)

	cfg := config.Config{
		K8sClient: clientset,
	}
	kwokProvider, err := CreateCloudProvider(ctx, &cfg)
	require.NoError(t, err)

	nodeName := fmt.Sprintf("kwok-test-create-node-%s", rand.String(20))

	err = kwokProvider.CreateNode(ctx, nodeName)
	assert.NoError(t, err)
	ret, err := kwokProvider.K8sClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, ret)
	assert.Equal(t, nodeName, ret.ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("kwok://%s", nodeName), ret.Spec.ProviderID)
}
