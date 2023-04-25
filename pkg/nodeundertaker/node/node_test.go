package node

import (
	"context"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"testing"
	"time"
)

func TestNodeIsGrownUp(t *testing.T) {
	cfg := config.Config{NodeInitialThreshold: 5}
	creationTime := metav1.Now().Add(-20 * time.Second)

	v1node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "dummy",
			CreationTimestamp: metav1.NewTime(creationTime),
		},
	}

	node := CreateNode(&v1node)

	res := node.IsGrownUp(&cfg)
	assert.True(t, res)
}

func TestNodeIsGrownUpNot(t *testing.T) {
	cfg := config.Config{NodeInitialThreshold: 90}
	creationTime := metav1.Now()

	v1node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "dummy",
			CreationTimestamp: creationTime,
		},
	}

	node := CreateNode(&v1node)

	res := node.IsGrownUp(&cfg)
	assert.False(t, res)
}

func TestGetLabelOk(t *testing.T) {
	labelValue := "test"
	v1node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dummy",
			Labels: map[string]string{
				Label: labelValue,
			},
		},
	}
	n := Node{
		Node:    &v1node,
		changed: false,
	}
	ret := n.GetLabel()
	assert.Equal(t, labelValue, ret)
}

func TestGetLabelEmpty(t *testing.T) {
	labelValue := ""
	v1node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dummy",
			Labels: map[string]string{
				Label: labelValue,
			},
		},
	}
	n := CreateNode(&v1node)
	ret := n.GetLabel()
	assert.Equal(t, labelValue, ret)
}

func TestGetLabelNone(t *testing.T) {
	expectedLabelValue := ""
	v1node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "dummy",
			Labels: map[string]string{},
		},
	}
	n := CreateNode(&v1node)
	ret := n.GetLabel()
	assert.Equal(t, expectedLabelValue, ret)
}

func TestSetLabelOk(t *testing.T) {
	labelValue := "test"
	v1node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dummy",
		},
	}
	n := CreateNode(&v1node)
	n.SetLabel(labelValue)

	ret, exists := n.ObjectMeta.Labels[Label]
	assert.Equal(t, labelValue, ret)
	assert.True(t, exists)
	assert.True(t, n.changed)
}

func TestSetLabelEmpty(t *testing.T) {
	labelValue := "test"
	v1node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "dummy",
			Labels: map[string]string{},
		},
	}
	n := CreateNode(&v1node)
	n.SetLabel(labelValue)

	ret, exists := n.ObjectMeta.Labels[Label]
	assert.Equal(t, labelValue, ret)
	assert.True(t, exists)
	assert.True(t, n.changed)
}

func TestSetLabelOverwrite(t *testing.T) {
	expectedLabelValue := "new-value"
	v1node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dummy",
			Labels: map[string]string{
				Label: "old-value",
			},
		},
	}
	n := CreateNode(&v1node)
	n.SetLabel(expectedLabelValue)

	ret, exists := n.ObjectMeta.Labels[Label]
	assert.Equal(t, expectedLabelValue, ret)
	assert.True(t, exists)
	assert.True(t, n.changed)
}

func TestSaveNoChange(t *testing.T) {
	nodeName := "node1"
	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	}

	cfg := config.Config{
		K8sClient: fake.NewSimpleClientset(),
	}
	_, err := cfg.K8sClient.CoreV1().Nodes().Create(context.TODO(), &nodev1, metav1.CreateOptions{})
	require.NoError(t, err)

	nodes, err := cfg.K8sClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, err)
	assert.Len(t, nodes.Items, 1)
	assert.Equal(t, nodeName, nodes.Items[0].Name)
	assert.Empty(t, nodes.Items[0].Spec.ProviderID)

	node := CreateNode(&nodev1)
	node.Spec.ProviderID = "test"

	err = node.Save(context.TODO(), &cfg)
	assert.NoError(t, err)

	nodes, err = cfg.K8sClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, err)
	assert.Len(t, nodes.Items, 1)
	assert.Equal(t, nodeName, nodes.Items[0].Name)
	assert.Empty(t, nodes.Items[0].Spec.ProviderID)
}

func TestSaveChange(t *testing.T) {
	nodeName := "node1"
	newProviderId := "test"

	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	}

	cfg := config.Config{
		K8sClient: fake.NewSimpleClientset(),
	}
	_, err := cfg.K8sClient.CoreV1().Nodes().Create(context.TODO(), &nodev1, metav1.CreateOptions{})
	require.NoError(t, err)

	nodes, err := cfg.K8sClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, err)
	assert.Len(t, nodes.Items, 1)
	assert.Equal(t, nodeName, nodes.Items[0].Name)
	assert.Empty(t, nodes.Items[0].Spec.ProviderID)

	node := CreateNode(&nodev1)
	node.Spec.ProviderID = newProviderId
	node.changed = true

	err = node.Save(context.TODO(), &cfg)
	assert.NoError(t, err)

	nodes, err = cfg.K8sClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, err)
	assert.Len(t, nodes.Items, 1)
	assert.Equal(t, nodeName, nodes.Items[0].Name)
	assert.Equal(t, newProviderId, nodes.Items[0].Spec.ProviderID)
}

func TestTaintNoTaints(t *testing.T) {
	nodeName := "node1"
	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	}

	node := CreateNode(&nodev1)
	node.Taint()

	assert.Len(t, node.Spec.Taints, 1)
	assert.Contains(t, node.Spec.Taints, v1.Taint{
		Key: TaintKey, Value: "", Effect: v1.TaintEffectNoSchedule,
	})
	assert.True(t, node.changed)
}

func TestTaintDifferentTaints(t *testing.T) {
	nodeName := "node1"
	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
		Spec: v1.NodeSpec{
			Taints: []v1.Taint{
				v1.Taint{Key: "sample", Value: "different", Effect: v1.TaintEffectPreferNoSchedule},
			},
		},
	}

	node := CreateNode(&nodev1)
	node.Taint()

	assert.Len(t, node.Spec.Taints, 2)
	assert.Contains(t, node.Spec.Taints, v1.Taint{
		Key: TaintKey, Value: TaintValue, Effect: v1.TaintEffectNoSchedule,
	})
	assert.Contains(t, node.Spec.Taints, v1.Taint{
		Key: "sample", Value: "different", Effect: v1.TaintEffectPreferNoSchedule,
	})
	assert.True(t, node.changed)
}

func TestTaintExistingTaint(t *testing.T) {
	nodeName := "node1"
	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
		Spec: v1.NodeSpec{
			Taints: []v1.Taint{
				{
					Key:    TaintKey,
					Value:  TaintValue,
					Effect: v1.TaintEffectNoSchedule,
				},
			},
		},
	}

	node := CreateNode(&nodev1)
	node.Taint()

	assert.Len(t, node.Spec.Taints, 1)
	assert.Contains(t, node.Spec.Taints, v1.Taint{
		Key:    TaintKey,
		Value:  TaintValue,
		Effect: v1.TaintEffectNoSchedule,
	})
	assert.False(t, node.changed)
}

func TestUntaintNoTaint(t *testing.T) {
	nodeName := "node1"
	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
		Spec:       v1.NodeSpec{},
	}

	node := CreateNode(&nodev1)
	node.Untaint()

	assert.Len(t, node.Spec.Taints, 0)
	assert.False(t, node.changed)
}

func TestUntaintExistingTaints(t *testing.T) {
	nodeName := "node1"
	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
		Spec: v1.NodeSpec{
			Taints: []v1.Taint{
				{Key: "sample", Value: "different", Effect: v1.TaintEffectPreferNoSchedule},
				{Key: TaintKey, Value: TaintValue, Effect: v1.TaintEffectNoSchedule},
				{Key: "sample2", Value: "different2", Effect: v1.TaintEffectPreferNoSchedule},
			},
		},
	}

	node := CreateNode(&nodev1)
	node.Untaint()

	assert.Len(t, node.Spec.Taints, 2)
	assert.True(t, node.changed)
	assert.NotContains(t, node.Spec.Taints, v1.Taint{Key: TaintKey, Value: TaintValue, Effect: v1.TaintEffectNoSchedule})
}
