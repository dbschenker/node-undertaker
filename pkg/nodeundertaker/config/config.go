package config

import (
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker/flags"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders"
	"github.com/spf13/viper"
)

type Config struct {
	CloudProvider         cloudproviders.CloudProvider
	DrainDelay            int
	CloudTerminationDelay int
	Port                  int
}

func GetConfig() (*Config, error) {
	ret := Config{}
	ret.DrainDelay = viper.GetInt(flags.DrainDelayFlag)
	ret.CloudTerminationDelay = viper.GetInt(flags.CloudTerminationDelayFlag)
	ret.Port = viper.GetInt(flags.PortFlag)
	return &ret, validateConfig(&ret)
}

func validateConfig(cfg *Config) error {
	if cfg.DrainDelay < 0 {
		return fmt.Errorf("%s can't be lower than zero", flags.DrainDelayFlag)
	}
	if cfg.CloudTerminationDelay < 0 {
		return fmt.Errorf("%s can't be lower than zero", flags.CloudTerminationDelayFlag)
	}
	if cfg.Port <= 0 {
		return fmt.Errorf("%s can't be lower or equal than zero", flags.PortFlag)
	}

	return nil
}
