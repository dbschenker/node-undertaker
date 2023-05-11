package config

import (
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker/flags"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"os"
	"time"
)

type Config struct {
	CloudProvider         cloudproviders.CLOUDPROVIDER
	DrainDelay            int
	CloudTerminationDelay int
	NodeInitialThreshold  int
	Port                  int
	K8sClient             kubernetes.Interface
	InformerResync        time.Duration
	Namespace             string
	Hostname              string
	LeaseLockName         string
	LeaseLockNamespace    string
}

func GetConfig() (*Config, error) {
	ret := Config{}
	ret.InformerResync = 60 * time.Second
	ret.DrainDelay = viper.GetInt(flags.DrainDelayFlag)
	ret.CloudTerminationDelay = viper.GetInt(flags.CloudTerminationDelayFlag)
	ret.Port = viper.GetInt(flags.PortFlag)
	ret.Namespace = viper.GetString(flags.NamespaceFlag)
	ret.LeaseLockNamespace = viper.GetString(flags.LeaseLockNamespaceFlag)
	ret.LeaseLockName = viper.GetString(flags.LeaseLockNameFlag)

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	ret.Hostname = hostname

	return &ret, validateConfig(&ret)
}

func validateConfig(cfg *Config) error {
	if cfg.DrainDelay < 0 {
		return fmt.Errorf("%s can't be lower than zero", flags.DrainDelayFlag)
	}
	if cfg.CloudTerminationDelay < 0 {
		return fmt.Errorf("%s can't be lower than zero", flags.CloudTerminationDelayFlag)
	}
	if cfg.NodeInitialThreshold < 0 {
		return fmt.Errorf("%s can't be lower than zero", flags.NodeInitialThresholdFlag)
	}

	if cfg.LeaseLockName == "" {
		return fmt.Errorf("%s can't be empty", flags.LeaseLockNameFlag)
	}

	if cfg.Port <= 0 {
		return fmt.Errorf("%s can't be lower or equal than zero", flags.PortFlag)
	}

	return nil
}
