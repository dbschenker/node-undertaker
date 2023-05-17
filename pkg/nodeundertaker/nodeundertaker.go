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
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/observability/metrics"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"os"
	"os/signal"
	"syscall"
)

// Execute executes node-undertaker logic
func Execute() error {
	err := setupLogging()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cancelOnSigterm(cancel)

	return executeWithContext(ctx, kubeclient.GetClient, cancel)

}

func executeWithContext(ctx context.Context, getk8sClient func() (kubernetes.Interface, string, error), cancel func()) error {
	// initialize config
	cfg, err := config.GetConfig()

	// k8s ClientSet
	k8sClient, currentNamespace, err := getk8sClient()
	if err != nil {
		return err
	}
	cfg.SetK8sClient(k8sClient, currentNamespace)

	if err != nil {
		return err
	}
	cloudProvider, err := getCloudProvider(ctx, cfg)
	if err != nil {
		return err
	}
	err = cloudProvider.ValidateConfig()
	if err != nil {
		return err
	}
	cfg.CloudProvider = cloudProvider

	//observability (logging & monitoring http server setup)
	observabilityServer := observability.GetDefaultObservabilityServer(cfg)
	observabilityServer.SetupRoutes()

	// workload
	workload := func(ctx2 context.Context) error {
		kubeclient.LeaderElection(
			ctx2,
			cfg,
			func(ctx3 context.Context) {
				startLogic(ctx2, cfg, nodeupdatehandler.GetDefaultUpdateHandlerFuncs(ctx, cfg), cancel)
			},
			cancel)
		return nil
	}

	// start logic
	err = startServer(ctx, cfg, &observabilityServer, workload, cancel)
	return err
}

func setupLogging() error {
	lvl, err := log.ParseLevel(viper.GetString(flags.LogLevelFlag))
	if err != nil {
		return err
	}
	log.SetLevel(lvl)
	format := viper.GetString(flags.LogFormatFlag)
	switch format {
	case flags.LogFormatText:
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
	case flags.LogFormatJson:
		log.SetFormatter(&log.JSONFormatter{})
	default:
		return fmt.Errorf("unknown log format: %s", format)
	}

	return nil
}

func startServer(ctx context.Context, cfg *config.Config, observabilityServer observability.OBSERVABILITYSERVER, workload func(ctx context.Context) error, cancel func()) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return observabilityServer.StartServer(ctx) })
	g.Go(func() error { return workload(ctx) })

	return g.Wait()
}

func startLogic(ctx context.Context, cfg *config.Config, handlerFuncs cache.ResourceEventHandlerFuncs, cancel func()) {
	factory := informers.NewSharedInformerFactoryWithOptions(cfg.K8sClient, cfg.InformerResync)
	nodeInformer := factory.Core().V1().Nodes()
	informer := nodeInformer.Informer()
	nodeLister := nodeInformer.Lister()
	factory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
		log.Errorf("Timed out waiting for caches to sync")
		cancel()
	}
	_, err := informer.AddEventHandler(handlerFuncs)
	if err != nil {
		log.Errorf("Error occured while adding event handler funcs: %v", err)
		cancel()
	}

	unregisterMetrics := metrics.Initialize(nodeLister)
	select {
	case <-ctx.Done():
		unregisterMetrics()
	}
}

func getCloudProvider(ctx context.Context, cfg *config.Config) (cloudproviders.CLOUDPROVIDER, error) {
	switch cloudProviderName := viper.GetString(flags.CloudProviderFlag); cloudProviderName {
	case "aws":
		cloudProvider, err := aws.CreateCloudProvider(ctx)
		return cloudProvider, err
	case "kind":
		cloudProvider, err := kind.CreateCloudProvider(ctx)
		return cloudProvider, err

	case "kwok":
		cloudProvider, err := kwok.CreateCloudProvider(ctx, cfg)
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
