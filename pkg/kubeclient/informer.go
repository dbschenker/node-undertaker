package kubeclient

import (
	"context"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func StartInformer(ctx context.Context, cfg *config.Config, handlers cache.ResourceEventHandlerFuncs) error {
	factory := informers.NewSharedInformerFactoryWithOptions(cfg.K8sClient, cfg.InformerResync)
	nodeInformer := factory.Core().V1().Nodes()
	informer := nodeInformer.Informer()

	go factory.Start(ctx.Done())

	if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
		return fmt.Errorf("Timed out waiting for caches to sync")
	}

	_, err := informer.AddEventHandler(handlers)
	if err != nil {
		return err
	}

	return nil
}
