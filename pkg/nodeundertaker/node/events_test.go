package node

import (
	"context"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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
		K8sClient: fake.NewSimpleClientset(),
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
	assert.Equal(t, "Error", ev.Type)
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
		K8sClient: fake.NewSimpleClientset(),
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
		K8sClient: fake.NewSimpleClientset(),
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

func TestReportEventUnsupportedLevel(t *testing.T) {
	namespace := "test"
	nodeName := "test-node"
	action := "DummyAction"
	reason := "DummyReason"
	hostname := "dummy-host"
	reasonDesc := "test-reason-desc"
	reasonOverride := "override-message"
	cfg := config.Config{
		K8sClient: fake.NewSimpleClientset(),
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
