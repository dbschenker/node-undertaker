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
	nodeUpdateInternal(ctx, cfg, n)
}

func nodeUpdateInternal(ctx context.Context, cfg *config.Config, n nodepkg.NODE) {
	if !n.IsGrownUp(cfg) {
		log.Debugf("Node %s is not old enough - might be not fully initialized.", n.GetName())
		return
	}

	// check if lease is fresh
	fresh, err := n.HasFreshLease(ctx, cfg)
	if err != nil {
		log.Errorf("Node %s update failed: %v", n.GetName(), err)
		return
	}

	nodeLabel := n.GetLabel()

	if nodeLabel == nodepkg.NodeTerminating {
		reason, err := n.Terminate(ctx, cfg)
		if err != nil {
			nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "Termination", reason, err.Error(), "")
		}
		nodepkg.ReportEvent(ctx, cfg, log.InfoLevel, n, "Termination", reason, "", "")
		return
	}

	if fresh {
		if nodeLabel != "" {
			n.Untaint()
			n.RemoveActionTimestamp()
			n.RemoveLabel()
			err := n.Save(ctx, cfg)
			if err != nil {
				log.Errorf("Received error while saving node %s: %v", n.GetName(), err)
				nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "Untaint", "Untaint failed", err.Error(), "")
				return
			}
			nodepkg.ReportEvent(ctx, cfg, log.InfoLevel, n, "Untaint", "Untainted", "", "")
		} else {
			log.Debugf("%s/%s: has fresh lease", n.GetKind(), n.GetName())
		}
	} else { // node has old lease
		switch label := nodeLabel; label {
		case nodepkg.NodeHealthy:
			n.SetLabel(nodepkg.NodeUnhealthy)
			err := n.Save(ctx, cfg)
			if err != nil {
				log.Errorf("Received error while saving node %s: %v", n.GetName(), err)
				nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "Label", "Label unhealthy failed", err.Error(), "")
				return
			}
			nodepkg.ReportEvent(ctx, cfg, log.InfoLevel, n, "LabeledUnhealthy", "Labeled unhealthy", "", "")
		case nodepkg.NodeUnhealthy:
			n.Taint()
			n.SetActionTimestamp(time.Now())
			n.SetLabel(nodepkg.NodeTainted)
			err := n.Save(ctx, cfg)
			if err != nil {
				log.Errorf("Received error while saving node %s: %v", n.GetName(), err)
				nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "Tainted", "Failed", err.Error(), "")
				return
			}
			nodepkg.ReportEvent(ctx, cfg, log.InfoLevel, n, "Taint", "Tainted", "", "")
		case nodepkg.NodeTainted:
			nodeModificationTimestamp, err := n.GetActionTimestamp()
			if err != nil {
				log.Errorf("Node %s: timestamp is not parsed properly: %v", n.GetName(), err)
				return
			}
			timestampShouldBeBefore := time.Now().Add(-time.Duration(cfg.DrainDelay) * time.Second)
			if nodeModificationTimestamp.After(timestampShouldBeBefore) {
				log.Infof("%s/%s: tainted less than %d seconds ago", n.GetKind(), n.GetName(), cfg.DrainDelay)
				return
			}

			n.StartDrain(ctx, cfg)
			n.SetActionTimestamp(time.Now())
			n.SetLabel(nodepkg.NodeDraining)
			err = n.Save(ctx, cfg)
			if err != nil {
				log.Errorf("Received error while saving node %s: %v", n.GetName(), err)
				nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "Drain", "Drain Failed", err.Error(), "")
				return
			}
			nodepkg.ReportEvent(ctx, cfg, log.InfoLevel, n, "Drain", "Drain started", "", "")
		case nodepkg.NodeDraining:
			nodeModificationTimestamp, err := n.GetActionTimestamp()
			if err != nil {
				log.Errorf("Node %s: timestamp is not parsed properly: %v", n.GetName(), err)
				return
			}
			timestampShouldBeBefore := time.Now().Add(-time.Duration(cfg.CloudTerminationDelay) * time.Second)
			if nodeModificationTimestamp.After(timestampShouldBeBefore) {
				log.Infof("%s/%s: drained less than %d seconds ago", n.GetKind(), n.GetName(), cfg.CloudTerminationDelay)
				return
			}

			n.SetActionTimestamp(time.Now())
			n.SetLabel(nodepkg.NodeTerminating)
			err = n.Save(ctx, cfg)
			if err != nil {
				log.Errorf("Received error while saving node %s: %v", n.GetName(), err)
				nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "Label", "Label Terminating Failed", err.Error(), "")
				return
			}

			nodepkg.ReportEvent(ctx, cfg, log.InfoLevel, n, "Termination", "Marked for termination", "", "")
		//case nodepkg.NodeTerminating: Shouldn't be handled here
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
