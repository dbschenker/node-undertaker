package nodeupdatehandler

import (
	"context"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	nodepkg "gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/node"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"time"
)

func OnNodeUpdate(ctx context.Context, cfg *config.Config, n *v1.Node) {
	node := nodepkg.CreateNode(n)

	if !node.IsGrownUp(cfg) {
		log.Debugf("Node %s is not old enough - might be not fully initialized.", node.ObjectMeta.Name)
		return
	}

	// check if lease is fresh
	fresh, err := node.HasFreshLease(ctx, cfg)
	if err != nil {
		log.Errorf("Node %s update failed: %v", node.ObjectMeta.Name, err)
		return
	}

	if fresh {
		if node.GetLabel() != "" {
			node.Untaint()
			node.RemoveActionTimestamp()
			node.RemoveLabel()
			err := node.Save(ctx, cfg)
			if err != nil {
				log.Errorf("Received error while saving node %s: %v", node.ObjectMeta.Name, err)
				//TODO produce event
				return
			}
			//TODO produce event
		}
	} else { // node has old lease
		switch label := node.GetLabel(); label {
		case nodepkg.NodeHealthy:
			//TODO produce event
			panic("TODO")
		case nodepkg.NodeUnhealthy:
			//TODO produce event
			panic("TODO")
		case nodepkg.NodeTainted:
			//TODO produce event
			panic("TODO")
		case nodepkg.NodeDraining:
			// check last action timestamp
			err := node.Terminate(ctx, cfg)
			if err != nil {
				log.Errorf("Termination of node %s failed due to: %v", node.ObjectMeta.Name, err)
				//TODO produce event
			}
			log.Infof("Termination of node %s succeeded", node.ObjectMeta.Name)
			//TODO produce event
		default:
			log.Errorf("Unknown node label: %s for node %s", label, node.ObjectMeta.Name)
			//TODO produce event
			return
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
