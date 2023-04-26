package node

import (
	"context"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	log "github.com/sirupsen/logrus"
	coordinationv1 "k8s.io/api/coordination/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const (
	TaintKey            = "dbschenker.com/node-undertaker"
	TaintValue          = ""
	Label               = "dbschenker.com/node-undertaker"
	TimestampAnnotation = "dbschenker.com/node-undertaker/timestamp"
)

const (
	NodeUnhealthy string = "unhealthy"
	NodeDraining         = "draining"
	NodeTainted          = "tainted"
	NodeHealthy          = ""
)

type Node struct {
	*v1.Node
	changed bool
}

func CreateNode(n *v1.Node) *Node {
	node := Node{
		Node:    n.DeepCopy(),
		changed: false,
	}
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}
	return &node
}

func (n Node) IsGrownUp(cfg *config.Config) bool {
	creationTime := n.ObjectMeta.CreationTimestamp
	before := metav1.NewTime(time.Now().Add(-time.Second * time.Duration(cfg.NodeInitialThreshold)))
	return creationTime.Before(&before)
}

func (n Node) HasFreshLease(ctx context.Context, cfg *config.Config) (bool, error) {
	lease, err := n.findLease(ctx, cfg)
	if errors.IsNotFound(err) {
		log.Warnf("lease not found for node %s: %v", n.Node.ObjectMeta.Name, err)
		return false, nil
	} else if err != nil {
		return false, err
	}

	leaseDuration := time.Duration(*lease.Spec.LeaseDurationSeconds) * time.Second
	isFresh := lease.Spec.RenewTime.Add(leaseDuration).After(time.Now())
	return isFresh, nil
}

func (n Node) GetLabel() string {
	if val, exists := n.Labels[Label]; exists {
		return val
	}
	return ""
}

func (n *Node) RemoveLabel() {
	if _, found := n.ObjectMeta.Labels[Label]; found {
		delete(n.ObjectMeta.Labels, Label)
		n.changed = true
	}
}

func (n *Node) RemoveActionTimestamp() {
	if _, found := n.ObjectMeta.Annotations[TimestampAnnotation]; found {
		delete(n.ObjectMeta.Annotations, TimestampAnnotation)
		n.changed = true
	}
}

func (n *Node) SetLabel(label string) {
	n.ObjectMeta.Labels[Label] = label
	n.changed = true
}

func (n Node) SetActionTimestamp(t time.Time) {
	n.changed = true
	n.ObjectMeta.Annotations[TimestampAnnotation] = t.Format(time.RFC3339)
	return
}

func (n Node) GetActionTimestamp() (time.Time, error) {
	if val, ok := n.ObjectMeta.Annotations[TimestampAnnotation]; ok {
		ret, err := time.Parse(time.RFC3339, val)
		return ret, err
	}
	return time.Now(), fmt.Errorf("node %s doesn't have annotation: %s", n.ObjectMeta.Name, TimestampAnnotation)
}

func (n *Node) Taint() {
	taint := v1.Taint{
		Key:    TaintKey,
		Value:  TaintValue,
		Effect: v1.TaintEffectNoSchedule,
	}

	for i := range n.Spec.Taints {
		if n.Spec.Taints[i] == taint {
			return
		}
	}
	n.Spec.Taints = append(n.Spec.Taints, taint)
	n.changed = true
}

func (n *Node) Untaint() {
	taint := v1.Taint{
		Key:    TaintKey,
		Value:  TaintValue,
		Effect: v1.TaintEffectNoSchedule,
	}

	// assume that there is only taint with same set of parameters (api sever should guard this)
	for i := range n.Spec.Taints {
		if n.Spec.Taints[i] == taint {
			n.Spec.Taints = append(n.Spec.Taints[:i], n.Spec.Taints[i+1])
			n.changed = true
			return
		}
	}
}

func (n Node) Drain() {
	panic("not implemented")
}

// Terminate deletes node from cloud provider
func (n Node) Terminate(ctx context.Context, cfg *config.Config) error {
	return cfg.CloudProvider.TerminateNode(ctx, n.Spec.ProviderID)
}

func (n Node) Save(ctx context.Context, cfg *config.Config) error {
	if n.changed {
		_, err := cfg.K8sClient.CoreV1().Nodes().Update(ctx, n.Node, metav1.UpdateOptions{})
		//TODO maybe Patch instead of Update will work better
		return err
	}
	return nil
}

func (n Node) findLease(ctx context.Context, cfg *config.Config) (*coordinationv1.Lease, error) {
	return cfg.K8sClient.CoordinationV1().Leases(cfg.Namespace).Get(ctx, n.ObjectMeta.Name, metav1.GetOptions{ResourceVersion: "0"})
}