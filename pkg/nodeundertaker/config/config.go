package config

import (
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker/flags"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/labels"
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
	NodeLeaseNamespace    string
	InitialDelay          int
	StartupTime           time.Time
	NodeSelector          labels.Selector
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
	ret.NodeInitialThreshold = viper.GetInt(flags.NodeInitialThresholdFlag)
	ret.NodeLeaseNamespace = viper.GetString(flags.NodeLeaseNamespaceFlag)
	ret.InitialDelay = viper.GetInt(flags.InitialDelayFlag)
	ret.StartupTime = time.Now()

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	ret.Hostname = hostname

	selectors, err := labels.Parse(viper.GetString(flags.NodeSelectorFlag))
	if err != nil {
		return nil, err
	}
	ret.NodeSelector = selectors

	return &ret, validateConfig(&ret)
}

func (cfg *Config) SetK8sClient(k8sClient kubernetes.Interface, namespace string) {
	cfg.K8sClient = k8sClient
	if cfg.Namespace == "" {
		log.Infof("Using autodetected namespace: %s", namespace)
		cfg.Namespace = namespace
	}
	if cfg.LeaseLockNamespace == "" {
		log.Infof("Using autodetected namespace for lease lock: %s", namespace)
		cfg.LeaseLockNamespace = namespace
	}
	if cfg.NodeLeaseNamespace == "" {
		log.Infof("Using autodetected namespace for node leases: %s", namespace)
		cfg.NodeLeaseNamespace = namespace
	}
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

	if cfg.Port < 0 {
		return fmt.Errorf("%s can't be lower than zero", flags.PortFlag)
	}
	if cfg.InitialDelay < 0 {
		return fmt.Errorf("%s can't be lower than zero", flags.InitialDelayFlag)
	}

	return nil
}
