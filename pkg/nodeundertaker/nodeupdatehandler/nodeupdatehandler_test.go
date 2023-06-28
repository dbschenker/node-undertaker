package nodeupdatehandler

import (
	"context"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	nodepkg "gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/node"
	mocknode "gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/node/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"
)

func TestOnNodeUpdate(t *testing.T) {
	nodeName := "test-node1"
	namespaceName := "test-dummy-ns"
	creationTime := metav1.Now().Add(-20 * time.Second).UTC()

	nv1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              nodeName,
			Namespace:         namespaceName,
			CreationTimestamp: metav1.NewTime(creationTime),
		},
	}

	cfg := config.Config{
		K8sClient:            fake.NewSimpleClientset(),
		Namespace:            namespaceName,
		NodeInitialThreshold: 1000,
	}

	OnNodeUpdate(context.TODO(), &cfg, &nv1)

	events, evErr := cfg.K8sClient.EventsV1().Events(namespaceName).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, evErr)
	assert.Len(t, events.Items, 0)
}

// unknown node label, node with old lease - should do nothin
func TestUnknownLabel(t *testing.T) {
	nodeName := "test-node1"
	namespaceName := "test-dummy-ns"
	mockCtrl := gomock.NewController(t)
	node := mocknode.NewMockNODE(mockCtrl)

	node.EXPECT().GetName().Return(nodeName).AnyTimes()
	node.EXPECT().GetKind().Return("Node").AnyTimes()

	node.EXPECT().GetLabel().Return("unknown-label").Times(1)
	node.EXPECT().IsGrownUp(gomock.Any()).Return(true).Times(1)
	node.EXPECT().HasFreshLease(gomock.Any(), gomock.Any()).Return(false, nil).Times(1)

	cfg := config.Config{
		K8sClient: fake.NewSimpleClientset(),
		Namespace: namespaceName,
	}

	nodeUpdateInternal(context.TODO(), &cfg, node)

	events, evErr := cfg.K8sClient.EventsV1().Events(namespaceName).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, evErr)
	assert.Len(t, events.Items, 1)
}

func TestNodeUpdateInternalNotAfterInitialDelay(t *testing.T) {
	cfg := config.Config{
		StartupTime:  time.Now().Add(-50 * time.Second),
		InitialDelay: 100,
	}
	n := nodepkg.Node{}
	nodeUpdateInternal(context.TODO(), &cfg, &n)
}

// node not grown up - should do nothing
func TestNodeUpdateInternalNotGrownUp(t *testing.T) {
	nodeName := "test-node1"
	namespaceName := "test-dummy-ns"
	mockCtrl := gomock.NewController(t)
	node := mocknode.NewMockNODE(mockCtrl)
	node.EXPECT().GetName().Return(nodeName)
	node.EXPECT().GetKind().Return("Node").AnyTimes()
	node.EXPECT().IsGrownUp(gomock.Any()).Return(false).Times(1)

	cfg := config.Config{
		K8sClient: fake.NewSimpleClientset(),
		Namespace: namespaceName,
	}

	nodeUpdateInternal(context.TODO(), &cfg, node)

	events, evErr := cfg.K8sClient.EventsV1().Events(namespaceName).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, evErr)
	assert.Len(t, events.Items, 0)
}

// node grown up & with recent lease & no label - should do nothing
func TestNodeUpdateInternalHealthyNoLabel(t *testing.T) {
	nodeName := "test-node1"
	hasFreshLease := true
	nodeLabel := nodepkg.NodeHealthy
	var hasFreshLeaseErr error = nil
	mockCtrl := gomock.NewController(t)
	node := mocknode.NewMockNODE(mockCtrl)
	node.EXPECT().GetName().Return(nodeName).AnyTimes()
	node.EXPECT().GetKind().Return("Node").AnyTimes()

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

	node.EXPECT().IsGrownUp(gomock.Any()).Return(true).Times(1)
	node.EXPECT().HasFreshLease(gomock.Any(), gomock.Any()).Return(hasFreshLease, hasFreshLeaseErr).Times(1)
	node.EXPECT().GetLabel().Return(nodeLabel).Times(1)

	node.EXPECT().GetActionTimestamp().Return(time.Now().Add(-100*time.Second), getTimestampErr).Times(1)

	drainCall := node.EXPECT().StartDrain(gomock.Any(), gomock.Any()).Times(1)
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

	node.EXPECT().IsGrownUp(gomock.Any()).Return(true).Times(1)
	node.EXPECT().HasFreshLease(gomock.Any(), gomock.Any()).Return(hasFreshLease, hasFreshLeaseErr).Times(1)
	node.EXPECT().GetLabel().Return(nodeLabel).Times(1)

	node.EXPECT().GetActionTimestamp().Return(time.Now().Add(-5*time.Second), getTimestampErr).Times(1)

	cfg := config.Config{
		K8sClient:                    fake.NewSimpleClientset(),
		Namespace:                    namespaceName,
		CloudPrepareTerminationDelay: 90,
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
	var saveErr error = nil
	nodeLabel := nodepkg.NodeDraining
	var hasFreshLeaseErr error = nil
	var getTimestampErr error = nil
	mockCtrl := gomock.NewController(t)
	node := mocknode.NewMockNODE(mockCtrl)

	node.EXPECT().GetName().Return(nodeName).AnyTimes()
	node.EXPECT().GetKind().Return("Node").AnyTimes()

	node.EXPECT().IsGrownUp(gomock.Any()).Return(true).Times(1)
	node.EXPECT().HasFreshLease(gomock.Any(), gomock.Any()).Return(hasFreshLease, hasFreshLeaseErr).Times(1)
	node.EXPECT().GetLabel().Return(nodeLabel).Times(1)

	getTimestampCall := node.EXPECT().GetActionTimestamp().Return(time.Now().Add(-100*time.Second), getTimestampErr).Times(1)
	setLabelCall := node.EXPECT().SetLabel(nodepkg.NodePreparingTermination)
	setTimestampCall := node.EXPECT().SetActionTimestamp(gomock.Any()).Times(1).After(getTimestampCall)
	node.EXPECT().Save(gomock.Any(), gomock.Any()).Return(saveErr).Times(1).After(setLabelCall).After(setTimestampCall)

	cfg := config.Config{
		K8sClient:                    fake.NewSimpleClientset(),
		Namespace:                    namespaceName,
		CloudPrepareTerminationDelay: 90,
	}

	nodeUpdateInternal(context.TODO(), &cfg, node)
	events, evErr := cfg.K8sClient.EventsV1().Events(namespaceName).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, evErr)
	assert.Len(t, events.Items, 1)
}

// node grown up &with old lease & label=preparing_termination - should prepare termination and label: termination_prepared
func TestNodeUpdateInternalPrepareTermination(t *testing.T) {
	nodeName := "test-node1"
	namespaceName := "dummy-ns"
	hasFreshLease := false
	terminationAction := "CloudInstanceTerminated"
	var terminationErr error = nil
	nodeLabel := nodepkg.NodePreparingTermination
	var hasFreshLeaseErr error = nil
	mockCtrl := gomock.NewController(t)
	node := mocknode.NewMockNODE(mockCtrl)

	node.EXPECT().GetName().Return(nodeName).AnyTimes()
	node.EXPECT().GetKind().Return("Node").AnyTimes()

	node.EXPECT().IsGrownUp(gomock.Any()).Return(true).Times(1)
	node.EXPECT().HasFreshLease(gomock.Any(), gomock.Any()).Return(hasFreshLease, hasFreshLeaseErr).Times(1)
	node.EXPECT().GetLabel().Return(nodeLabel).Times(1)
	setLabelCall := node.EXPECT().SetLabel(nodepkg.NodeTerminationPrepared).Return().Times(1)
	setTimestampCall := node.EXPECT().SetActionTimestamp(gomock.Any()).Return().Times(1)
	node.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil).Times(1).After(setLabelCall).After(setTimestampCall)

	node.EXPECT().PrepareTermination(gomock.Any(), gomock.Any()).Return(terminationAction, terminationErr).Times(1)

	cfg := config.Config{
		K8sClient: fake.NewSimpleClientset(),
		Namespace: namespaceName,
	}

	nodeUpdateInternal(context.TODO(), &cfg, node)
	events, evErr := cfg.K8sClient.EventsV1().Events(namespaceName).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, evErr)
	assert.Len(t, events.Items, 1)
}

// node grown up &with old lease & label=prepared_termination + timestamp is older than CloudPrepareTerminationDelay - should prepare termination and label: terminating
func TestNodeUpdateInternalPreparedTerminationOld(t *testing.T) {
	nodeName := "test-node1"
	namespaceName := "dummy-ns"
	hasFreshLease := false
	var getTimestampErr error = nil
	nodeLabel := nodepkg.NodeTerminationPrepared
	var hasFreshLeaseErr error = nil
	mockCtrl := gomock.NewController(t)
	node := mocknode.NewMockNODE(mockCtrl)

	node.EXPECT().GetName().Return(nodeName).AnyTimes()
	node.EXPECT().GetKind().Return("Node").AnyTimes()

	node.EXPECT().IsGrownUp(gomock.Any()).Return(true).Times(1)
	node.EXPECT().HasFreshLease(gomock.Any(), gomock.Any()).Return(hasFreshLease, hasFreshLeaseErr).Times(1)
	node.EXPECT().GetLabel().Return(nodeLabel).Times(1)
	getTimestampCall := node.EXPECT().GetActionTimestamp().Return(time.Now().Add(-100*time.Second), getTimestampErr).Times(1)
	setLabelCall := node.EXPECT().SetLabel(nodepkg.NodeTerminating).Return().Times(1)
	node.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil).Times(1).After(setLabelCall).After(getTimestampCall)

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

// node grown up &with old lease & label=prepared_termination + timestamp is not older than CloudPrepareTerminationDelay - should prepare termination and label: terminating
func TestNodeUpdateInternalPreparedTerminationRecent(t *testing.T) {
	nodeName := "test-node1"
	namespaceName := "dummy-ns"
	hasFreshLease := false
	nodeLabel := nodepkg.NodeTerminationPrepared
	var hasFreshLeaseErr error = nil
	mockCtrl := gomock.NewController(t)
	node := mocknode.NewMockNODE(mockCtrl)

	node.EXPECT().GetName().Return(nodeName).AnyTimes()
	node.EXPECT().GetKind().Return("Node").AnyTimes()

	node.EXPECT().IsGrownUp(gomock.Any()).Return(true).Times(1)
	node.EXPECT().HasFreshLease(gomock.Any(), gomock.Any()).Return(hasFreshLease, hasFreshLeaseErr).Times(1)
	node.EXPECT().GetLabel().Return(nodeLabel).Times(1)
	node.EXPECT().GetActionTimestamp().Return(time.Now().Add(-10*time.Second), nil).Times(1)

	cfg := config.Config{
		K8sClient:             fake.NewSimpleClientset(),
		Namespace:             namespaceName,
		CloudTerminationDelay: 90,
	}

	nodeUpdateInternal(context.TODO(), &cfg, node)
	events, evErr := cfg.K8sClient.EventsV1().Events(namespaceName).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, evErr)
	assert.Len(t, events.Items, 0) //no node was saved
}

// node grown up & with old lease & label=terminating - should terminate the node in cloud + label + annotate + produce event
func TestNodeUpdateInternalUnhealthyDeletingOldLease(t *testing.T) {
	nodeName := "test-node1"
	namespaceName := "dummy-ns"
	hasFreshLease := false
	terminationAction := "CloudInstanceTerminated"
	var terminationErr error = nil
	nodeLabel := nodepkg.NodeTerminating
	var hasFreshLeaseErr error = nil
	mockCtrl := gomock.NewController(t)
	node := mocknode.NewMockNODE(mockCtrl)

	node.EXPECT().GetName().Return(nodeName).AnyTimes()
	node.EXPECT().GetKind().Return("Node").AnyTimes()

	node.EXPECT().IsGrownUp(gomock.Any()).Return(true).Times(1)
	node.EXPECT().HasFreshLease(gomock.Any(), gomock.Any()).Return(hasFreshLease, hasFreshLeaseErr).Times(1)
	node.EXPECT().GetLabel().Return(nodeLabel).Times(1)

	node.EXPECT().Terminate(gomock.Any(), gomock.Any()).Return(terminationAction, terminationErr).Times(1)

	cfg := config.Config{
		K8sClient: fake.NewSimpleClientset(),
		Namespace: namespaceName,
	}

	nodeUpdateInternal(context.TODO(), &cfg, node)
	events, evErr := cfg.K8sClient.EventsV1().Events(namespaceName).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, evErr)
	assert.Len(t, events.Items, 1)
}

// node grown up & with fresh lease & label=deleting + timestamp less than threshold - should terminate the node in cloud + label + annotate + produce event
func TestNodeUpdateInternalUnhealthyDeletingFreshLease(t *testing.T) {
	nodeName := "test-node1"
	namespaceName := "dummy-ns"
	terminationAction := "CloudInstanceTerminated"

	hasFreshLease := true
	var terminationErr error = nil
	nodeLabel := nodepkg.NodeTerminating
	var hasFreshLeaseErr error = nil
	mockCtrl := gomock.NewController(t)
	node := mocknode.NewMockNODE(mockCtrl)

	node.EXPECT().GetName().Return(nodeName).AnyTimes()
	node.EXPECT().GetKind().Return("Node").AnyTimes()

	node.EXPECT().IsGrownUp(gomock.Any()).Return(true).Times(1)
	node.EXPECT().HasFreshLease(gomock.Any(), gomock.Any()).Return(hasFreshLease, hasFreshLeaseErr).Times(1)
	node.EXPECT().GetLabel().Return(nodeLabel).Times(1)

	node.EXPECT().Terminate(gomock.Any(), gomock.Any()).Return(terminationAction, terminationErr).Times(1)

	cfg := config.Config{
		K8sClient: fake.NewSimpleClientset(),
		Namespace: namespaceName,
	}

	nodeUpdateInternal(context.TODO(), &cfg, node)
	events, evErr := cfg.K8sClient.EventsV1().Events(namespaceName).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, evErr)
	assert.Len(t, events.Items, 1)
}

func TestGetDefaultUpdateHandlerFuncs(t *testing.T) {
	ctx := context.TODO()
	cfg := &config.Config{}

	result := GetDefaultUpdateHandlerFuncs(ctx, cfg)
	assert.Nil(t, result.DeleteFunc)
	assert.NotNil(t, result.AddFunc)
	assert.NotNil(t, result.UpdateFunc)
}

func TestIsAfterInitialDelayOk(t *testing.T) {
	cfg := config.Config{
		StartupTime:  time.Now().Add(-50 * time.Second),
		InitialDelay: 20,
	}
	ret := isAfterInitialDelay(&cfg)
	assert.True(t, ret)
}

func TestIsAfterInitialDelayNok(t *testing.T) {
	cfg := config.Config{
		StartupTime:  time.Now().Add(-50 * time.Second),
		InitialDelay: 100,
	}
	ret := isAfterInitialDelay(&cfg)
	assert.False(t, ret)
}
