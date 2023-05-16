package node

import (
	"context"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders/kwok"
	mockcloudproviders "gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders/mocks"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	coordinationv1 "k8s.io/api/coordination/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"
)

func TestNodeIsGrownUp(t *testing.T) {
	cfg := config.Config{NodeInitialThreshold: 5}
	creationTime := metav1.Now().Add(-20 * time.Second).UTC()

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

func TestSetActionTimestampNone(t *testing.T) {
	nodeName := "node1"
	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	}
	node := CreateNode(&nodev1)
	tret, err := node.GetActionTimestamp()
	assert.Error(t, err)
	assert.True(t, tret.After(time.Now().Add(-time.Hour)))
	assert.True(t, tret.Before(time.Now()))
}

func TestSetActionTimestampExists(t *testing.T) {
	nodeName := "node1"
	tnow := time.Now().Truncate(time.Second).UTC()
	ti := tnow.Format(time.RFC3339)
	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
			Annotations: map[string]string{
				TimestampAnnotation: ti,
			},
		},
	}
	node := CreateNode(&nodev1)
	tret, err := node.GetActionTimestamp()
	assert.NoError(t, err)
	assert.Equal(t, tnow, tret)
}

func TestSetActionTimestampWrongFormat(t *testing.T) {
	nodeName := "node1"
	ti := "test string"
	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
			Annotations: map[string]string{
				TimestampAnnotation: ti,
			},
		},
	}
	node := CreateNode(&nodev1)
	_, err := node.GetActionTimestamp() // don't care about the date
	assert.Error(t, err)
}

func TestFindLeaseOk(t *testing.T) {
	nodeName := "node1"
	namespace := "example-lease-ns"
	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	}
	lease := coordinationv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodeName,
			Namespace: namespace,
		},
	}
	cfg := config.Config{
		K8sClient: fake.NewSimpleClientset(),
		Namespace: namespace,
	}
	_, err := cfg.K8sClient.CoordinationV1().Leases(namespace).Create(context.TODO(), &lease, metav1.CreateOptions{})
	require.NoError(t, err)

	node := CreateNode(&nodev1)
	leaseret, err := node.findLease(context.TODO(), &cfg)
	assert.NoError(t, err)
	assert.Equal(t, lease, *leaseret)
}

func TestFindLeaseMissing(t *testing.T) {
	nodeName := "node1"
	namespace := "example-lease-ns"
	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	}

	cfg := config.Config{
		K8sClient: fake.NewSimpleClientset(),
		Namespace: namespace,
	}

	node := CreateNode(&nodev1)
	leaseret, err := node.findLease(context.TODO(), &cfg)
	assert.Error(t, err)
	assert.Equal(t, metav1.StatusReasonNotFound, errors.ReasonForError(err))
	assert.Nil(t, leaseret)
}

func TestHasFreshLeaseOk(t *testing.T) {
	nodeName := "node1"
	namespace := "example-lease-ns"
	leaseDuration := int32(90)
	renewTime := metav1.NewMicroTime(time.Now().Add(-10 * time.Second))

	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	}
	lease := coordinationv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodeName,
			Namespace: namespace,
		},
		Spec: coordinationv1.LeaseSpec{
			LeaseDurationSeconds: &leaseDuration,
			RenewTime:            &renewTime,
		},
	}
	cfg := config.Config{
		K8sClient: fake.NewSimpleClientset(),
		Namespace: namespace,
	}
	_, err := cfg.K8sClient.CoordinationV1().Leases(namespace).Create(context.TODO(), &lease, metav1.CreateOptions{})
	require.NoError(t, err)

	node := CreateNode(&nodev1)
	ret, err := node.HasFreshLease(context.TODO(), &cfg)

	assert.NoError(t, err)
	assert.True(t, ret)
}

func TestHasFreshLeaseNok(t *testing.T) {
	nodeName := "node1"
	namespace := "example-lease-ns"
	leaseDuration := int32(90)
	renewTime := metav1.NewMicroTime(time.Now().Add(-1000 * time.Second))

	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	}
	lease := coordinationv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodeName,
			Namespace: namespace,
		},
		Spec: coordinationv1.LeaseSpec{
			LeaseDurationSeconds: &leaseDuration,
			RenewTime:            &renewTime,
		},
	}
	cfg := config.Config{
		K8sClient: fake.NewSimpleClientset(),
		Namespace: namespace,
	}
	_, err := cfg.K8sClient.CoordinationV1().Leases(namespace).Create(context.TODO(), &lease, metav1.CreateOptions{})
	require.NoError(t, err)

	node := CreateNode(&nodev1)
	ret, err := node.HasFreshLease(context.TODO(), &cfg)

	assert.NoError(t, err)
	assert.False(t, ret)
}

func TestHasFreshLeaseNolease(t *testing.T) {
	nodeName := "node1"
	namespace := "example-lease-ns"

	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	}

	cfg := config.Config{
		K8sClient: fake.NewSimpleClientset(),
		Namespace: namespace,
	}

	node := CreateNode(&nodev1)
	ret, err := node.HasFreshLease(context.TODO(), &cfg)

	assert.NoError(t, err)
	assert.False(t, ret)
}

func TestRemoveLabelOk(t *testing.T) {
	v1node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dummy",
			Labels: map[string]string{
				Label: "old-value",
			},
		},
	}
	n := CreateNode(&v1node)
	n.RemoveLabel()

	ret, exists := n.ObjectMeta.Labels[Label]
	assert.Equal(t, "", ret)
	assert.False(t, exists)
	assert.True(t, n.changed)
}

func TestRemoveLabelNotExisting(t *testing.T) {
	v1node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dummy",
			Labels: map[string]string{
				"test": "old-value",
			},
		},
	}
	n := CreateNode(&v1node)
	n.RemoveLabel()

	ret, exists := n.ObjectMeta.Labels[Label]
	assert.Equal(t, "", ret)
	assert.False(t, exists)
	assert.False(t, n.changed)
}

func TestRemoveActionTimestampOk(t *testing.T) {
	v1node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dummy",
			Annotations: map[string]string{
				TimestampAnnotation: "old-value",
			},
		},
	}
	n := CreateNode(&v1node)
	n.RemoveActionTimestamp()

	ret, exists := n.ObjectMeta.Annotations[TimestampAnnotation]
	assert.Equal(t, "", ret)
	assert.False(t, exists)
	assert.True(t, n.changed)
}

func TestRemoveActionTimestampNotExisting(t *testing.T) {
	v1node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dummy",
			Annotations: map[string]string{
				"test": "old-value",
			},
		},
	}
	n := CreateNode(&v1node)
	n.RemoveActionTimestamp()

	ret, exists := n.ObjectMeta.Annotations[TimestampAnnotation]
	assert.Equal(t, "", ret)
	assert.False(t, exists)
	assert.False(t, n.changed)
}

func TestTerminate(t *testing.T) {
	termianteAction := "TestAction"
	mockCtrl := gomock.NewController(t)
	cloudProvider := mockcloudproviders.NewMockCLOUDPROVIDER(mockCtrl)
	cloudProvider.EXPECT().TerminateNode(gomock.Any(), gomock.Any()).Return(termianteAction, nil).Times(1)

	cfg := config.Config{
		CloudProvider: cloudProvider,
	}
	v1node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dummy",
		},
	}
	n := CreateNode(&v1node)
	res, err := n.Terminate(context.TODO(), &cfg)
	assert.NoError(t, err)
	assert.Equal(t, termianteAction, res)
}

func TestGetName(t *testing.T) {
	expectedName := "dummy123"
	v1node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: expectedName,
		},
	}
	n := CreateNode(&v1node)
	ret := n.GetName()
	assert.Equal(t, expectedName, ret)
}

func TestGetKind(t *testing.T) {
	expectedName := "dummy123"
	expectedKind := "Node"

	v1node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: expectedName,
		},
	}
	n := CreateNode(&v1node)
	ret := n.GetKind()
	assert.Equal(t, expectedKind, ret)
}

func TestDrain(t *testing.T) {
	// setup
	ctx := context.TODO()
	clientset, err := kwok.StartCluster(t, ctx)
	require.NoError(t, err)

	cfg := config.Config{
		K8sClient:             clientset,
		CloudTerminationDelay: 30,
		Namespace:             v1.NamespaceDefault,
	}

	kwokProvider, err := kwok.CreateCloudProvider(ctx, &cfg)
	require.NoError(t, err)

	nodeName := fmt.Sprintf("kwok-test-drain-node-%s", rand.String(20))
	replicaCount := 3
	deploymentName := "test-deployment1"

	err = kwokProvider.CreateNode(ctx, nodeName)
	assert.NoError(t, err)

	err = createDeployment(t, ctx, clientset, deploymentName, v1.NamespaceDefault, "pause", int32(replicaCount))
	require.NoError(t, err)
	err = waitForDeploymentPodsReady(ctx, clientset, 60*time.Second, deploymentName, v1.NamespaceDefault, replicaCount)
	require.NoError(t, err)

	nodePodsBefore, err := clientset.CoreV1().Pods(v1.NamespaceDefault).List(ctx, metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	assert.NoError(t, err)
	assert.Len(t, nodePodsBefore.Items, replicaCount)

	// block node from rescheduling pods
	nodev1, err := kwokProvider.K8sClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	require.NoError(t, err)
	node := CreateNode(nodev1)
	node.Taint()
	err = node.Save(ctx, &cfg)
	require.NoError(t, err)

	// drain
	node.StartDrain(ctx, &cfg)
	assert.NoError(t, err)

	time.Sleep(time.Duration(cfg.CloudTerminationDelay+20) * time.Second) //sleep longer than drain takes

	err = waitForDeploymentPodsReady(ctx, clientset, 60*time.Second, deploymentName, v1.NamespaceDefault, 0)
	require.NoError(t, err)

	ret, err := kwokProvider.K8sClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, ret)
	assert.Equal(t, nodeName, nodev1.ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("kwok://%s", nodeName), ret.Spec.ProviderID)

	nodePodsAfter, err := clientset.CoreV1().Pods(v1.NamespaceDefault).List(ctx, metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	assert.NoError(t, err)
	assert.Len(t, nodePodsAfter.Items, 0)
}

func TestDrainWithBlockingPDB(t *testing.T) {
	// setup
	ctx := context.TODO()

	clientset, err := kwok.StartCluster(t, ctx)
	require.NoError(t, err)

	cfg := config.Config{
		K8sClient:             clientset,
		CloudTerminationDelay: 30,
		Namespace:             v1.NamespaceDefault,
	}

	kwokProvider, err := kwok.CreateCloudProvider(ctx, &cfg)
	require.NoError(t, err)

	nodeName := fmt.Sprintf("kwok-test-drain-node-%s", rand.String(20))
	replicaCount := 3
	deploymentName := "test-deployment1"

	err = kwokProvider.CreateNode(ctx, nodeName)
	assert.NoError(t, err)

	err = createDeployment(t, ctx, clientset, deploymentName, v1.NamespaceDefault, "pause", int32(replicaCount))
	require.NoError(t, err)
	err = waitForDeploymentPodsReady(ctx, clientset, 60*time.Second, deploymentName, v1.NamespaceDefault, replicaCount)
	require.NoError(t, err)

	pdpbMaxUnavail := intstr.FromInt(0)
	pdb := policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pdb",
			Namespace: v1.NamespaceDefault,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MaxUnavailable: &pdpbMaxUnavail,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deploymentName,
				},
			},
		},
	}
	_, err = clientset.PolicyV1().PodDisruptionBudgets(v1.NamespaceDefault).Create(ctx, &pdb, metav1.CreateOptions{})
	require.NoError(t, err)

	nodePodsBefore, err := clientset.CoreV1().Pods(v1.NamespaceDefault).List(ctx, metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	assert.NoError(t, err)
	assert.Len(t, nodePodsBefore.Items, replicaCount)

	// block node from rescheduling pods
	nodev1, err := kwokProvider.K8sClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	require.NoError(t, err)
	node := CreateNode(nodev1)
	node.Taint()
	err = node.Save(ctx, &cfg)
	require.NoError(t, err)

	// drain
	node.StartDrain(ctx, &cfg)

	time.Sleep(time.Duration(cfg.CloudTerminationDelay+20) * time.Second) //sleep longer than drain takes

	ret, err := kwokProvider.K8sClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, ret)
	assert.Equal(t, nodeName, nodev1.ObjectMeta.Name)
	assert.Equal(t, fmt.Sprintf("kwok://%s", nodeName), ret.Spec.ProviderID)

	nodePodsAfter, err := clientset.CoreV1().Pods(v1.NamespaceDefault).List(ctx, metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	assert.NoError(t, err)
	assert.Len(t, nodePodsAfter.Items, replicaCount)
}

// createDeployment creates deployment
func createDeployment(t *testing.T, ctx context.Context, clientset *kubernetes.Clientset, name, namespace, image string, replicas int32) error {
	t.Helper()
	deploy := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  name,
							Image: image,
						},
					},
				},
			},
		},
	}

	_, err := clientset.AppsV1().Deployments(namespace).Create(ctx, &deploy, metav1.CreateOptions{})
	return err
}

func waitForDeploymentPodsReady(ctx context.Context, clientset *kubernetes.Clientset, duration time.Duration, name, namespace string, requiredNumber int) error {
	return wait.PollUntilContextTimeout(ctx, time.Second, duration, true,
		func(context.Context) (bool, error) {
			dep, err := clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return false, err
			}
			if dep.Status.ReadyReplicas == int32(requiredNumber) {
				return true, nil
			}
			return false, nil
		},
	)

}

func TestSetActionTimestamp(t *testing.T) {
	v1node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dummy",
		},
	}
	timeNow := time.Now()
	n := CreateNode(&v1node)
	n.SetActionTimestamp(timeNow)

	ret, exists := n.ObjectMeta.Annotations[TimestampAnnotation]
	assert.Equal(t, timeNow.Format(time.RFC3339), ret)
	assert.True(t, exists)
	assert.True(t, n.changed)
}
