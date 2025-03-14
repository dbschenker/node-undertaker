package node

import (
	"context"
	"github.com/dbschenker/node-undertaker/pkg/cloudproviders/kwok"
	"github.com/dbschenker/node-undertaker/pkg/nodeundertaker/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"strings"
	"testing"
)

func TestReportEvent(t *testing.T) {
	namespace := "test"
	nodeName := "test-node"
	action := "DummyAction"
	reason := "DummyReason"
	hostname := "dummy-host"
	reasonDesc := ""
	cfg := config.Config{
		K8sClient: fake.NewClientset(),
		Namespace: namespace,
		Hostname:  hostname,
	}
	lvl := logrus.ErrorLevel
	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	}
	node := CreateNode(&nodev1)
	ReportEvent(context.TODO(), &cfg, lvl, node, action, reason, reasonDesc, "")

	events, err := cfg.K8sClient.EventsV1().Events(namespace).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, events)
	assert.Len(t, events.Items, 1)
	ev := events.Items[0]
	assert.True(t, strings.HasPrefix(ev.ObjectMeta.Name, "node-undertaker."))
	assert.Equal(t, namespace, ev.ObjectMeta.Namespace)
	assert.Equal(t, action, ev.Action)
	assert.Equal(t, reason, ev.Reason)
	assert.Equal(t, ReportingController, ev.ReportingController)
	assert.Equal(t, hostname, ev.ReportingInstance)
	assert.Equal(t, "Warning", ev.Type)
	assert.NotEmpty(t, ev.Note)
}

func TestReportEventReasonDesc(t *testing.T) {
	namespace := "test"
	nodeName := "test-node"
	action := "DummyAction"
	reason := "DummyReason"
	hostname := "dummy-host"
	reasonDesc := "test-reason-desc"
	cfg := config.Config{
		K8sClient: fake.NewClientset(),
		Namespace: namespace,
		Hostname:  hostname,
	}
	lvl := logrus.InfoLevel
	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	}
	node := CreateNode(&nodev1)
	ReportEvent(context.TODO(), &cfg, lvl, node, action, reason, reasonDesc, "")

	events, err := cfg.K8sClient.EventsV1().Events(namespace).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, events)
	assert.Len(t, events.Items, 1)
	ev := events.Items[0]
	assert.True(t, strings.HasPrefix(ev.ObjectMeta.Name, "node-undertaker."))
	assert.Equal(t, namespace, ev.ObjectMeta.Namespace)
	assert.Equal(t, action, ev.Action)
	assert.Equal(t, reason, ev.Reason)
	assert.Equal(t, ReportingController, ev.ReportingController)
	assert.Equal(t, hostname, ev.ReportingInstance)
	assert.Equal(t, "Normal", ev.Type)
	assert.Contains(t, ev.Note, reasonDesc)
}

func TestReportEventReasonOverride(t *testing.T) {
	namespace := "test"
	nodeName := "test-node"
	action := "DummyAction"
	reason := "DummyReason"
	hostname := "dummy-host"
	reasonDesc := "test-reason-desc"
	reasonOverride := "override-message"
	cfg := config.Config{
		K8sClient: fake.NewClientset(),
		Namespace: namespace,
		Hostname:  hostname,
	}
	lvl := logrus.WarnLevel
	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	}
	node := CreateNode(&nodev1)
	ReportEvent(context.TODO(), &cfg, lvl, node, action, reason, reasonDesc, reasonOverride)

	events, err := cfg.K8sClient.EventsV1().Events(namespace).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, events)
	assert.Len(t, events.Items, 1)
	ev := events.Items[0]
	assert.True(t, strings.HasPrefix(ev.ObjectMeta.Name, "node-undertaker."))
	assert.Equal(t, namespace, ev.ObjectMeta.Namespace)
	assert.Equal(t, action, ev.Action)
	assert.Equal(t, reason, ev.Reason)
	assert.Equal(t, ReportingController, ev.ReportingController)
	assert.Equal(t, hostname, ev.ReportingInstance)
	assert.Equal(t, "Warning", ev.Type)
	assert.Equal(t, reasonOverride, ev.Note)
}

func TestReportEventReasonOverrideTooLong(t *testing.T) {
	namespace := "test"
	nodeName := "test-node"
	action := "DummyAction"
	reason := "DummyReason"
	hostname := "dummy-host"
	reasonDesc := "test-reason-desc"
	reasonOverride := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
	reasonOverride = reasonOverride + reasonOverride + reasonOverride //should be over 1024 chars
	cfg := config.Config{
		K8sClient: fake.NewClientset(),
		Namespace: namespace,
		Hostname:  hostname,
	}
	lvl := logrus.WarnLevel
	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	}
	node := CreateNode(&nodev1)
	ReportEvent(context.TODO(), &cfg, lvl, node, action, reason, reasonDesc, reasonOverride)

	events, err := cfg.K8sClient.EventsV1().Events(namespace).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, events)
	assert.Len(t, events.Items, 1)
	ev := events.Items[0]
	assert.True(t, strings.HasPrefix(ev.ObjectMeta.Name, "node-undertaker."))
	assert.Equal(t, namespace, ev.ObjectMeta.Namespace)
	assert.Equal(t, action, ev.Action)
	assert.Equal(t, reason, ev.Reason)
	assert.Equal(t, ReportingController, ev.ReportingController)
	assert.Equal(t, hostname, ev.ReportingInstance)
	assert.Equal(t, "Warning", ev.Type)
	assert.NotEqual(t, reasonOverride, ev.Note)
	assert.Len(t, ev.Note, 1024)
}

func TestReportEventUnsupportedLevel(t *testing.T) {
	namespace := "test"
	nodeName := "test-node"
	action := "DummyAction"
	reason := "DummyReason"
	hostname := "dummy-host"
	reasonDesc := "test-reason-desc"
	reasonOverride := "override-message"
	cfg := config.Config{
		K8sClient: fake.NewClientset(),
		Namespace: namespace,
		Hostname:  hostname,
	}
	lvl := logrus.DebugLevel
	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	}
	node := CreateNode(&nodev1)
	ReportEvent(context.TODO(), &cfg, lvl, node, action, reason, reasonDesc, reasonOverride)

	events, err := cfg.K8sClient.EventsV1().Events(namespace).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, events)
	assert.Len(t, events.Items, 0)
}

func TestReportEventWithKwok(t *testing.T) {
	namespace := "test"
	nodeName := "test-node"
	action := "DummyAction"
	reason := "DummyReason"
	hostname := "dummy-host"
	reasonDesc := ""

	ctx := context.TODO()

	clientset, err := kwok.StartCluster(t, ctx)
	require.NoError(t, err)

	cfg := config.Config{
		K8sClient: clientset,
		Namespace: namespace,
		Hostname:  hostname,
	}

	_, err = clientset.CoreV1().Namespaces().Create(ctx, &v1.Namespace{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}, metav1.CreateOptions{})
	assert.NoError(t, err)

	lvl := logrus.ErrorLevel
	nodev1 := v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	}
	node := CreateNode(&nodev1)
	ReportEvent(ctx, &cfg, lvl, node, action, reason, reasonDesc, "")

	events, err := cfg.K8sClient.EventsV1().Events(namespace).List(ctx, metav1.ListOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, events)
	assert.Len(t, events.Items, 1)
	ev := events.Items[0]
	assert.True(t, strings.HasPrefix(ev.ObjectMeta.Name, "node-undertaker."))
	assert.Equal(t, namespace, ev.ObjectMeta.Namespace)
	assert.Equal(t, action, ev.Action)
	assert.Equal(t, reason, ev.Reason)
	assert.Equal(t, ReportingController, ev.ReportingController)
	assert.Equal(t, hostname, ev.ReportingInstance)
	assert.Equal(t, "Warning", ev.Type)
	assert.NotEmpty(t, ev.Note)
}
