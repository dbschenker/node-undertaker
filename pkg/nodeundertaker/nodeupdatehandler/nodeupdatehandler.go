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
	if !isAfterInitialDelay(cfg) {
		log.Debugf("Node udertaker is not running at least %d seconds", cfg.InitialDelay)
		return
	}
	if !n.IsGrownUp(cfg) {
		log.Debugf("%s/%s: is not old enough (%d seconds) - might be not fully initialized.", n.GetKind(), n.GetName(), cfg.NodeInitialThreshold)
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
	} else if nodeLabel == nodepkg.NodePreparingTermination {
		reason, err := n.PrepareTermination(ctx, cfg)
		if err != nil {
			nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "Prepare Termination", reason, err.Error(), "")
			return
		}

		n.SetActionTimestamp(time.Now())
		n.SetLabel(nodepkg.NodeTerminationPrepared)
		err = n.Save(ctx, cfg)
		if err != nil {
			log.Errorf("Received error while saving node %s: %v", n.GetName(), err)
			nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "Prepare Termination", "Prepare Termination failed", err.Error(), "")
			return
		}

		nodepkg.ReportEvent(ctx, cfg, log.InfoLevel, n, "Termination prepared", reason, "", "")
		return
	} else if nodeLabel == nodepkg.NodeTerminationPrepared {
		nodeModificationTimestamp, err := n.GetActionTimestamp()
		if err != nil {
			log.Errorf("Node %s: timestamp is not parsed properly: %v", n.GetName(), err)
			return
		}
		timestampShouldBeBefore := time.Now().Add(-time.Duration(cfg.CloudTerminationDelay) * time.Second)
		if nodeModificationTimestamp.After(timestampShouldBeBefore) {
			log.Infof("%s/%s: prepared for termintaion less than %d seconds ago", n.GetKind(), n.GetName(), cfg.CloudTerminationDelay)
			return
		}

		n.SetLabel(nodepkg.NodeTerminating)
		err = n.Save(ctx, cfg)
		if err != nil {
			log.Errorf("Received error while saving node %s: %v", n.GetName(), err)
			nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "Label", "Label terminating failed", err.Error(), "")
			return
		}

		nodepkg.ReportEvent(ctx, cfg, log.InfoLevel, n, "LabelTerminating", "Labeled terminating", "", "")
		return
	}

	if fresh {
		if nodeLabel != nodepkg.NodeHealthy {
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
				nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "Drain", "Drain Start Failed", err.Error(), "")
				return
			}
			nodepkg.ReportEvent(ctx, cfg, log.InfoLevel, n, "Drain", "Drain started", "", "")
		case nodepkg.NodeDraining:
			nodeModificationTimestamp, err := n.GetActionTimestamp()
			if err != nil {
				log.Errorf("Node %s: timestamp is not parsed properly: %v", n.GetName(), err)
				return
			}
			timestampShouldBeBefore := time.Now().Add(-time.Duration(cfg.CloudPrepareTerminationDelay) * time.Second)
			if nodeModificationTimestamp.After(timestampShouldBeBefore) {
				log.Infof("%s/%s: drained less than %d seconds ago", n.GetKind(), n.GetName(), cfg.CloudPrepareTerminationDelay)
				return
			}

			n.SetActionTimestamp(time.Now())
			n.SetLabel(nodepkg.NodePreparingTermination)
			err = n.Save(ctx, cfg)
			if err != nil {
				log.Errorf("Received error while saving node %s: %v", n.GetName(), err)
				nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "Label", "Label Prepare Termination Failed", err.Error(), "")
				return
			}

			nodepkg.ReportEvent(ctx, cfg, log.InfoLevel, n, "Prepare Termination", "Instance preparing for termination", "", "")
		default:
			nodepkg.ReportEvent(ctx, cfg, log.ErrorLevel, n, "NodeUpdate", "Node Update Failed", fmt.Sprintf("unknown label value found: %s", label), "")
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

func isAfterInitialDelay(cfg *config.Config) bool {
	return cfg.StartupTime.Add(time.Duration(cfg.InitialDelay) * time.Second).Before(time.Now())
}
