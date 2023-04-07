package nodeundertaker

import (
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker/flags"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders/aws"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/observability"
	"github.com/spf13/viper"
)

func Execute() error {
	config, err := config.GetConfig()
	if err != nil {
		return err
	}
	cloudProvider, err := getCloudProvider()
	if err != nil {
		return err
	}
	err = cloudProvider.ValidateConfig()
	if err != nil {
		return err
	}
	config.CloudProvider = cloudProvider

	// do more init

	// start logic

	err = observability.StartServer(config)
	if err != nil {
		return err
	}

	return nil
}

func getCloudProvider() (cloudproviders.CloudProvider, error) {
	switch cloudProviderName := viper.GetString(flags.CloudProviderFlag); cloudProviderName {
	case "aws":
		cloudProvider := aws.CreateAwsCloudProvider()
		return cloudProvider, nil

	default:
		return nil, fmt.Errorf("Unknown cloud provider: %s", cloudProviderName)
	}

}
