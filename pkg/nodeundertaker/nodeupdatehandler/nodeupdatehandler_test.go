package nodeupdatehandler

import (
	"context"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	nodepkg "gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/node"
	mocknode "gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/node/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"
)

// node not grown up - should do nothing
func TestNodeUpdateInternalNotGrownUp(t *testing.T) {
	nodeName := "test-node1"
	mockCtrl := gomock.NewController(t)
	node := mocknode.NewMockNODE(mockCtrl)
	node.EXPECT().GetName().Return(nodeName)
	cfg := config.Config{}
	node.EXPECT().IsGrownUp(gomock.Any()).Return(false).Times(1)

	nodeUpdateInternal(context.TODO(), &cfg, node)
}

// node grown up & with recent lease & no label - should do nothing
func TestNodeUpdateInternalHealthyNoLabel(t *testing.T) {
	//nodeName := "test-node1"
	hasFreshLease := true
	nodeLabel := nodepkg.NodeHealthy
	var hasFreshLeaseErr error = nil
	mockCtrl := gomock.NewController(t)
	node := mocknode.NewMockNODE(mockCtrl)
	//node.EXPECT().GetName().Return(nodeName).AnyTimes()

	node.EXPECT().IsGrownUp(gomock.Any()).Return(true).Times(1)
	node.EXPECT().HasFreshLease(gomock.Any(), gomock.Any()).Return(hasFreshLease, hasFreshLeaseErr).Times(1)
	node.EXPECT().GetLabel().Return(nodeLabel).Times(1)

	cfg := config.Config{}
	nodeUpdateInternal(context.TODO(), &cfg, node)
}

// node grown up & with recent lease & has label - should remove label, taint and annotation
func TestNodeUpdateInternalHealthyUnhealthyLabel(t *testing.T) {
	nodeName := "test-node1"
	namespaceName := "dummy-ns"
	hasFreshLease := true
	nodeLabel := nodepkg.NodeDraining
	var hasFreshLeaseErr error = nil
	var saveErr error = nil
	mockCtrl := gomock.NewController(t)
	node := mocknode.NewMockNODE(mockCtrl)

	node.EXPECT().GetName().Return(nodeName).AnyTimes()
	node.EXPECT().GetKind().Return("Node").AnyTimes()
	node.EXPECT().GetNamespace().Return(metav1.NamespaceAll).AnyTimes()

	node.EXPECT().IsGrownUp(gomock.Any()).Return(true).Times(1)
	node.EXPECT().HasFreshLease(gomock.Any(), gomock.Any()).Return(hasFreshLease, hasFreshLeaseErr).Times(1)
	node.EXPECT().GetLabel().Return(nodeLabel).Times(1)

	node.EXPECT().Untaint().Times(1)
	node.EXPECT().RemoveActionTimestamp().Times(1)
	node.EXPECT().Save(gomock.Any(), gomock.Any()).Return(saveErr).Times(1)
	node.EXPECT().RemoveLabel().Times(1)

	cfg := config.Config{K8sClient: fake.NewSimpleClientset(), Namespace: namespaceName}

	nodeUpdateInternal(context.TODO(), &cfg, node)
	events, evErr := cfg.K8sClient.EventsV1().Events(namespaceName).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, evErr)
	assert.Len(t, events.Items, 1)
}

// node grown up & with old lease & has no label - should add label & produce event
func TestNodeUpdateInternalUnhealthyNoLabel(t *testing.T) {
	nodeName := "test-node1"
	namespaceName := "dummy-ns"
	hasFreshLease := false
	nodeLabel := nodepkg.NodeHealthy
	var hasFreshLeaseErr error = nil
	var saveErr error = nil
	mockCtrl := gomock.NewController(t)
	node := mocknode.NewMockNODE(mockCtrl)

	node.EXPECT().GetName().Return(nodeName).AnyTimes()
	node.EXPECT().GetKind().Return("Node").AnyTimes()
	node.EXPECT().GetNamespace().Return(metav1.NamespaceAll).AnyTimes()

	node.EXPECT().IsGrownUp(gomock.Any()).Return(true).Times(1)
	node.EXPECT().HasFreshLease(gomock.Any(), gomock.Any()).Return(hasFreshLease, hasFreshLeaseErr).Times(1)
	node.EXPECT().GetLabel().Return(nodeLabel).Times(1)

	setLabelCall := node.EXPECT().SetLabel(nodepkg.NodeUnhealthy).Times(1)
	node.EXPECT().Save(gomock.Any(), gomock.Any()).Return(saveErr).Times(1).After(setLabelCall)

	cfg := config.Config{K8sClient: fake.NewSimpleClientset(), Namespace: namespaceName}

	nodeUpdateInternal(context.TODO(), &cfg, node)
	events, evErr := cfg.K8sClient.EventsV1().Events(namespaceName).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, evErr)
	assert.Len(t, events.Items, 1)
}

// node grown up & with old lease & has unhealthy label - should add timestamp, taint & change label + save + produce event
func TestNodeUpdateInternalUnhealthyUnhealthyLabel(t *testing.T) {
	nodeName := "test-node1"
	namespaceName := "dummy-ns"
	hasFreshLease := false
	nodeLabel := nodepkg.NodeUnhealthy
	var hasFreshLeaseErr error = nil
	var saveErr error = nil
	mockCtrl := gomock.NewController(t)
	node := mocknode.NewMockNODE(mockCtrl)

	node.EXPECT().GetName().Return(nodeName).AnyTimes()
	node.EXPECT().GetKind().Return("Node").AnyTimes()
	node.EXPECT().GetNamespace().Return(metav1.NamespaceAll).AnyTimes()

	node.EXPECT().IsGrownUp(gomock.Any()).Return(true).Times(1)
	node.EXPECT().HasFreshLease(gomock.Any(), gomock.Any()).Return(hasFreshLease, hasFreshLeaseErr).Times(1)
	node.EXPECT().GetLabel().Return(nodeLabel).Times(1)

	setTimestampCall := node.EXPECT().SetActionTimestamp(gomock.Any()).Times(1)
	setLabelCall := node.EXPECT().SetLabel(nodepkg.NodeTainted).Times(1)
	taintCall := node.EXPECT().Taint().Times(1)
	node.EXPECT().Save(gomock.Any(), gomock.Any()).Return(saveErr).Times(1).After(setLabelCall).After(setTimestampCall).After(taintCall)

	cfg := config.Config{K8sClient: fake.NewSimpleClientset(), Namespace: namespaceName}

	nodeUpdateInternal(context.TODO(), &cfg, node)
	events, evErr := cfg.K8sClient.EventsV1().Events(namespaceName).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, evErr)
	assert.Len(t, events.Items, 1)
}

// node grown up & with old lease & label=tainted + timetamp less than threshold - should do nothing
func TestNodeUpdateInternalUnhealthyTaintedLabelRecent(t *testing.T) {
	nodeName := "test-node1"
	namespaceName := "dummy-ns"
	hasFreshLease := false
	nodeLabel := nodepkg.NodeTainted
	var hasFreshLeaseErr error = nil
	var getTimestampErr error = nil
	mockCtrl := gomock.NewController(t)
	node := mocknode.NewMockNODE(mockCtrl)

	node.EXPECT().GetName().Return(nodeName).AnyTimes()
	node.EXPECT().GetKind().Return("Node").AnyTimes()
	node.EXPECT().GetNamespace().Return(metav1.NamespaceAll).AnyTimes()

	node.EXPECT().IsGrownUp(gomock.Any()).Return(true).Times(1)
	node.EXPECT().HasFreshLease(gomock.Any(), gomock.Any()).Return(hasFreshLease, hasFreshLeaseErr).Times(1)
	node.EXPECT().GetLabel().Return(nodeLabel).Times(1)

	node.EXPECT().GetActionTimestamp().Return(time.Now().Add(-5*time.Second), getTimestampErr).Times(1)

	cfg := config.Config{
		K8sClient:  fake.NewSimpleClientset(),
		Namespace:  namespaceName,
		DrainDelay: 90,
	}

	nodeUpdateInternal(context.TODO(), &cfg, node)
	events, evErr := cfg.K8sClient.EventsV1().Events(namespaceName).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, evErr)
	assert.Len(t, events.Items, 0)
}

// node grown up & with old lease & label=tainted + timetamp less than threshold - should drain node + label + update timestamp
func TestNodeUpdateInternalUnhealthyTaintedLabelOld(t *testing.T) {
	nodeName := "test-node1"
	namespaceName := "dummy-ns"
	hasFreshLease := false
	var saveErr error = nil
	nodeLabel := nodepkg.NodeTainted
	var hasFreshLeaseErr error = nil
	var getTimestampErr error = nil
	mockCtrl := gomock.NewController(t)
	node := mocknode.NewMockNODE(mockCtrl)

	node.EXPECT().GetName().Return(nodeName).AnyTimes()
	node.EXPECT().GetKind().Return("Node").AnyTimes()
	node.EXPECT().GetNamespace().Return(metav1.NamespaceAll).AnyTimes()

	node.EXPECT().IsGrownUp(gomock.Any()).Return(true).Times(1)
	node.EXPECT().HasFreshLease(gomock.Any(), gomock.Any()).Return(hasFreshLease, hasFreshLeaseErr).Times(1)
	node.EXPECT().GetLabel().Return(nodeLabel).Times(1)

	node.EXPECT().GetActionTimestamp().Return(time.Now().Add(-100*time.Second), getTimestampErr).Times(1)

	drainCall := node.EXPECT().Drain().Times(1)
	drainingCall := node.EXPECT().SetLabel(nodepkg.NodeDraining).Times(1)
	timestampCall := node.EXPECT().SetActionTimestamp(gomock.Any()).Times(1)
	node.EXPECT().Save(gomock.Any(), gomock.Any()).Return(saveErr).Times(1).After(drainingCall).After(timestampCall).After(drainCall)

	cfg := config.Config{
		K8sClient:  fake.NewSimpleClientset(),
		Namespace:  namespaceName,
		DrainDelay: 90,
	}

	nodeUpdateInternal(context.TODO(), &cfg, node)
	events, evErr := cfg.K8sClient.EventsV1().Events(namespaceName).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, evErr)
	assert.Len(t, events.Items, 1)
}

// node grown up & with old lease & label=draining + timetamp less than threshold - should do nothing
func TestNodeUpdateInternalUnhealthyDrainingLabelRecent(t *testing.T) {
	nodeName := "test-node1"
	namespaceName := "dummy-ns"
	hasFreshLease := false
	nodeLabel := nodepkg.NodeDraining
	var hasFreshLeaseErr error = nil
	var getTimestampErr error = nil
	mockCtrl := gomock.NewController(t)
	node := mocknode.NewMockNODE(mockCtrl)

	node.EXPECT().GetName().Return(nodeName).AnyTimes()
	node.EXPECT().GetKind().Return("Node").AnyTimes()
	node.EXPECT().GetNamespace().Return(metav1.NamespaceAll).AnyTimes()

	node.EXPECT().IsGrownUp(gomock.Any()).Return(true).Times(1)
	node.EXPECT().HasFreshLease(gomock.Any(), gomock.Any()).Return(hasFreshLease, hasFreshLeaseErr).Times(1)
	node.EXPECT().GetLabel().Return(nodeLabel).Times(1)

	node.EXPECT().GetActionTimestamp().Return(time.Now().Add(-5*time.Second), getTimestampErr).Times(1)

	cfg := config.Config{
		K8sClient:             fake.NewSimpleClientset(),
		Namespace:             namespaceName,
		CloudTerminationDelay: 90,
	}

	nodeUpdateInternal(context.TODO(), &cfg, node)
	events, evErr := cfg.K8sClient.EventsV1().Events(namespaceName).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, evErr)
	assert.Len(t, events.Items, 0)
}

// node grown up & with old lease & label=draining + timestamp less than threshold - should terminate the node in cloud + label + annotate + produce event
func TestNodeUpdateInternalUnhealthyDrainingLabelOld(t *testing.T) {
	nodeName := "test-node1"
	namespaceName := "dummy-ns"
	hasFreshLease := false
	var terminationErr error = nil
	nodeLabel := nodepkg.NodeDraining
	var hasFreshLeaseErr error = nil
	var getTimestampErr error = nil
	mockCtrl := gomock.NewController(t)
	node := mocknode.NewMockNODE(mockCtrl)

	node.EXPECT().GetName().Return(nodeName).AnyTimes()
	node.EXPECT().GetKind().Return("Node").AnyTimes()
	node.EXPECT().GetNamespace().Return(metav1.NamespaceAll).AnyTimes()

	node.EXPECT().IsGrownUp(gomock.Any()).Return(true).Times(1)
	node.EXPECT().HasFreshLease(gomock.Any(), gomock.Any()).Return(hasFreshLease, hasFreshLeaseErr).Times(1)
	node.EXPECT().GetLabel().Return(nodeLabel).Times(1)

	node.EXPECT().GetActionTimestamp().Return(time.Now().Add(-100*time.Second), getTimestampErr).Times(1)

	node.EXPECT().Terminate(gomock.Any(), gomock.Any()).Return(terminationErr).Times(1)

	cfg := config.Config{
		K8sClient:             fake.NewSimpleClientset(),
		Namespace:             namespaceName,
		CloudTerminationDelay: 90,
	}

	nodeUpdateInternal(context.TODO(), &cfg, node)
	events, evErr := cfg.K8sClient.EventsV1().Events(namespaceName).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, evErr)
	assert.Len(t, events.Items, 1)
}
