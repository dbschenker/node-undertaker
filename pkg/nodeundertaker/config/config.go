package config

import (
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker/flags"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"time"
)

type Config struct {
	CloudProvider         cloudproviders.CloudProvider
	DrainDelay            int
	CloudTerminationDelay int
	NodeInitialThreshold  int
	Port                  int
	K8sClient             kubernetes.Interface
	InformerResync        time.Duration
	Namespace             string
}

func GetConfig() (*Config, error) {
	ret := Config{}
	//var err error = nil
	ret.DrainDelay = viper.GetInt(flags.DrainDelayFlag)
	ret.CloudTerminationDelay = viper.GetInt(flags.CloudTerminationDelayFlag)
	ret.Port = viper.GetInt(flags.PortFlag)
	namespace := viper.GetString(flags.NamespaceFlag)
	//if namespace == "" {
	//	namespace, err = autodetectNamespace()
	//	if err != nil {
	//		return &ret, err
	//	}
	//}

	ret.Namespace = namespace
	return &ret, validateConfig(&ret)
}

//
//func autodetectNamespace() (string, error) {
//	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
//
//	return "TODO"
//}

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

	if cfg.Port <= 0 {
		return fmt.Errorf("%s can't be lower or equal than zero", flags.PortFlag)
	}

	return nil
}
