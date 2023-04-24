package nodeundertaker

import (
	"context"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker/flags"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders/aws"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/k8snodeinformer"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/observability"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

// Execute executes node-undertaker logic
func Execute() error {
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

	// workload processors
	var k8sNodeInformer k8snodeinformer.K8SNODEINFORMER = k8snodeinformer.DefaultK8sNodeInformer{}

	//observability (logging & monitoring http server setup)
	observabilityServer := observability.GetDefaultObservabilityServer(cfg)
	observabilityServer.SetupRoutes()

	// do more init
	//cloud provider clients
	//k8s clientset
	// start logic
	err = startLogic(ctx, cfg, k8sNodeInformer, observabilityServer)
	if err != nil {
		log.Errorf("couldn't start properly")
	}
	return nil
}

func startLogic(ctx context.Context, cfg *config.Config, nodeInformer k8snodeinformer.K8SNODEINFORMER, observabilityserver observability.OBSERVABILITYSERVER) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return nodeInformer.StartInformer(ctx, cfg) })
	g.Go(func() error { return observabilityserver.StartServer(ctx) })
	return g.Wait()
}

func getCloudProvider(ctx context.Context) (cloudproviders.CloudProvider, error) {
	switch cloudProviderName := viper.GetString(flags.CloudProviderFlag); cloudProviderName {
	case "aws":
		cloudProvider, err := aws.CreateAwsCloudProvider(ctx)
		return cloudProvider, err

	default:
		return nil, fmt.Errorf("Unknown cloud provider: %s", cloudProviderName)
	}

}
