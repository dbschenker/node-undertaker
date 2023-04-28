package nodeupdatehandler

import (
	"context"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	nodepkg "gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/node"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"time"
)

func OnNodeUpdate(ctx context.Context, cfg *config.Config, nv1 *v1.Node) {
	n := nodepkg.CreateNode(nv1)

	if !n.IsGrownUp(cfg) {
		log.Debugf("Node %s is not old enough - might be not fully initialized.", n.ObjectMeta.Name)
		return
	}

	// check if lease is fresh
	fresh, err := n.HasFreshLease(ctx, cfg)
	if err != nil {
		log.Errorf("Node %s update failed: %v", n.ObjectMeta.Name, err)
		return
	}

	if fresh {
		if n.GetLabel() != "" {
			n.Untaint()
			n.RemoveActionTimestamp()
			n.RemoveLabel()
			err := n.Save(ctx, cfg)
			if err != nil {
				log.Errorf("Received error while saving node %s: %v", n.ObjectMeta.Name, err)
				nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "TaintRemoval", "Failed", err.Error(), "")
				return
			}
			nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "TaintRemoval", "Succeeded", "", "")
		}
	} else { // node has old lease
		switch label := n.GetLabel(); label {
		case nodepkg.NodeHealthy:
			n.SetLabel(nodepkg.NodeUnhealthy)
			err := n.Save(ctx, cfg)
			if err != nil {
				log.Errorf("Received error while saving node %s: %v", n.ObjectMeta.Name, err)
				nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "LabeledUnhealthy", "Failed", err.Error(), "")
				return
			}
			nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "LabeledUnhealthy", "Succeeded", "", "")
		case nodepkg.NodeUnhealthy:
			n.Taint()
			n.SetActionTimestamp(time.Now())
			n.SetLabel(nodepkg.NodeTainted)
			err := n.Save(ctx, cfg)
			if err != nil {
				log.Errorf("Received error while saving node %s: %v", n.ObjectMeta.Name, err)
				nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "Tainted", "Failed", err.Error(), "")
				return
			}
			nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "Tainted", "Succeeded", "", "")
		case nodepkg.NodeTainted:
			nodeModificationTimestamp, err := n.GetActionTimestamp()
			if err != nil {
				log.Errorf("Node %s: timestamp is not parsed properly: %v", n.ObjectMeta.Name, err)
				return
			}
			timestampShouldBeBefore := time.Now().Add(time.Duration(cfg.DrainDelay) * time.Second)
			if nodeModificationTimestamp.After(timestampShouldBeBefore) {
				log.Infof("Node %s tainted too recently", n.ObjectMeta.Name)
				return
			}

			n.Drain()
			n.SetActionTimestamp(time.Now())
			n.SetLabel(nodepkg.NodeDraining)
			err = n.Save(ctx, cfg)
			if err != nil {
				log.Errorf("Received error while saving node %s: %v", n.ObjectMeta.Name, err)
				nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "Drain", "Failed", err.Error(), "")
				return
			}
			nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "Drain", "Started", "", "")

		case nodepkg.NodeDraining:
			nodeModificationTimestamp, err := n.GetActionTimestamp()
			if err != nil {
				log.Errorf("Node %s: timestamp is not parsed properly: %v", n.ObjectMeta.Name, err)
				return
			}
			timestampShouldBeBefore := time.Now().Add(time.Duration(cfg.CloudTerminationDelay) * time.Second)
			if nodeModificationTimestamp.After(timestampShouldBeBefore) {
				log.Infof("Node %s tainted too recently", n.ObjectMeta.Name)
				return
			}

			err = n.Terminate(ctx, cfg)
			if err != nil {
				nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "CloudInstanceTermiantion", "Failed", err.Error(), "")
			}
			nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "CloudInstanceTermiantion", "Succeeded", "", "")
		default:
			nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "NodeUpdate", "Failed", fmt.Sprintf("unknown label value found: %s", label), "")
		}
	}
}

func GetDefaultUpdateHandlerFuncs(ctx context.Context, cfg *config.Config) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			OnNodeUpdate(ctx, cfg, newObj.(*v1.Node))
		},
		AddFunc: func(obj interface{}) {
			OnNodeUpdate(ctx, cfg, obj.(*v1.Node))
		},
		DeleteFunc: nil,
	}
}

func nodeHasFreshLease(cfg *config.Config, node *v1.Node) bool {
	return false
}

func getNodeLabel(cfg *config.Config, node *v1.Node) string {
	return ""
}

func getNodeActionTimestamp(cfg *config.Config, node *v1.Node) time.Time {
	return time.Now()
}
