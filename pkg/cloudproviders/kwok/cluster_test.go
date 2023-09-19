package kwok

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestStartCluster(t *testing.T) {
	ctx := context.TODO()
	clientset, err := StartCluster(t, ctx)
	require.NoError(t, err)
	list, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	assert.NoError(t, err)
	assert.Len(t, list.Items, 0)
}
