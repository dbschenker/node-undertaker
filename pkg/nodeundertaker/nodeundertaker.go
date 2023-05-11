package nodeundertaker

import (
	"context"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker/flags"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders/aws"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders/kind"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders/kwok"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/kubeclient"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/nodeupdatehandler"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/observability"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"os"
	"os/signal"
	"syscall"
)

// Execute executes node-undertaker logic
func Execute() error {
	err := setupLogLevel()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cancelOnSigterm(cancel)

	// initialize config
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}
	cloudProvider, err := getCloudProvider(ctx)
	if err != nil {
		return err
	}
	err = cloudProvider.ValidateConfig()
	if err != nil {
		return err
	}
	cfg.CloudProvider = cloudProvider

	// k8s ClientSet
	k8sClient, namespace, err := kubeclient.GetClient()
	if err != nil {
		return err
	}
	cfg.K8sClient = k8sClient
	if cfg.Namespace == "" {
		log.Infof("Using autodetected namespace: %s", namespace)
		cfg.Namespace = namespace
	}
	if cfg.LeaseLockNamespace == "" {
		log.Infof("Using autodetected namespace for lease lock: %s", namespace)
		cfg.LeaseLockNamespace = namespace
	}

	//observability (logging & monitoring http server setup)
	observabilityServer := observability.GetDefaultObservabilityServer(cfg)
	observabilityServer.SetupRoutes()

	// start logic
	kubeclient.LeaderElection(ctx, cfg, func(ctx1 context.Context) {
		err = startLogic(ctx1, cfg, nodeupdatehandler.GetDefaultUpdateHandlerFuncs(ctx1, cfg), observabilityServer)
		if err != nil {
			log.Errorf("couldn't start properly, due to %v", err)
		}
	})
	return nil
}

func setupLogLevel() error {
	lvl, err := log.ParseLevel(viper.GetString(flags.LogLevelFlag))
	if err != nil {
		return err
	}
	log.SetLevel(lvl)
	return nil
}

func startLogic(ctx context.Context, cfg *config.Config, handlerFuncs cache.ResourceEventHandlerFuncs, observabilityserver observability.OBSERVABILITYSERVER) error {
	g, ctx := errgroup.WithContext(ctx)

	factory := informers.NewSharedInformerFactoryWithOptions(cfg.K8sClient, cfg.InformerResync)
	nodeInformer := factory.Core().V1().Nodes()
	informer := nodeInformer.Informer()

	g.Go(func() error {
		factory.Start(ctx.Done())
		return nil
	})

	if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
		return fmt.Errorf("Timed out waiting for caches to sync")
	}

	_, err := informer.AddEventHandler(handlerFuncs)
	if err != nil {
		return err
	}

	g.Go(func() error { return observabilityserver.StartServer(ctx) })
	return g.Wait()
}

func getCloudProvider(ctx context.Context) (cloudproviders.CLOUDPROVIDER, error) {
	switch cloudProviderName := viper.GetString(flags.CloudProviderFlag); cloudProviderName {
	case "aws":
		cloudProvider, err := aws.CreateCloudProvider(ctx)
		return cloudProvider, err
	case "kind":
		cloudProvider, err := kind.CreateCloudProvider(ctx)
		return cloudProvider, err

	case "kwok":
		cloudProvider, err := kwok.CreateCloudProvider(ctx)
		return cloudProvider, err
	default:
		return nil, fmt.Errorf("Unknown cloud provider: %s", cloudProviderName)
	}

}

func cancelOnSigterm(cancel func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		log.Info("Received termination, signaling shutdown")
		cancel()
	}()
}
