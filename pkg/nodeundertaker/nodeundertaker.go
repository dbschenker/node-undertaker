package nodeundertaker

import (
	"context"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker/flags"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders/aws"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/kubeclient"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/nodeupdatehandler"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/observability"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

// Execute executes node-undertaker logic
func Execute() error {
	err := setupLogLevel()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
	//observability (logging & monitoring http server setup)
	observabilityServer := observability.GetDefaultObservabilityServer(cfg)
	observabilityServer.SetupRoutes()

	// start logic
	err = startLogic(ctx, cfg, nodeupdatehandler.GetDefaultUpdateHandlerFuncs(ctx, cfg), observabilityServer)
	if err != nil {
		log.Errorf("couldn't start properly")
	}
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
		cloudProvider, err := aws.CreateAwsCloudProvider(ctx)
		return cloudProvider, err

	default:
		return nil, fmt.Errorf("Unknown cloud provider: %s", cloudProviderName)
	}

}
