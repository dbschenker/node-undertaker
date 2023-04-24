package nodeupdatehandler

import (
	"context"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"time"
)

func OnNodeUpdate(ctx context.Context, cfg *config.Config, node *v1.Node) {
	if !nodeIsGrownUp(cfg, node) {
		log.Infof("Node %s is not old enough - might be not fully initialized.", node.ObjectMeta.Name)
		return
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

// nodeIsGrownUp checks if the node is older than NodeInitialThreshold
func nodeIsGrownUp(cfg *config.Config, node *v1.Node) bool {
	creationTime := node.ObjectMeta.CreationTimestamp
	before := metav1.NewTime(time.Now().Add(-time.Second * time.Duration(cfg.NodeInitialThreshold)))
	return creationTime.Before(&before)
}
