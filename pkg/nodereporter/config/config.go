package config

import (
	"errors"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-reporter/flags"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"net/url"
)

type Config struct {
	K8sClient kubernetes.Interface
	Namespace string
	NodeName  string
	URL       *url.URL
	Timeout   int
	Frequency int
	LeaseTime int
}

func GetConfig() (*Config, error) {
	ret := Config{}

	ret.Namespace = viper.GetString(flags.NamespaceFlag)
	ret.NodeName = viper.GetString(flags.NodeNameFlag)
	rawUrl := viper.GetString(flags.UrlFlag)
	url, err := url.Parse(rawUrl)
	if err != nil {
		return &ret, err
	}
	ret.LeaseTime = viper.GetInt(flags.TimeoutFlag)
	ret.LeaseTime = viper.GetInt(flags.FrequencyFlag)
	ret.LeaseTime = viper.GetInt(flags.LeaseTimeFlag)

	ret.URL = url
	return &ret, validateConfig(&ret)
}

func (cfg *Config) SetK8sClient(k8sClient kubernetes.Interface, namespace string) {
	cfg.K8sClient = k8sClient
	if cfg.Namespace == "" {
		log.Infof("Using autodetected namespace: %s", namespace)
		cfg.Namespace = namespace
	}

}

func validateConfig(cfg *Config) error {
	if cfg.NodeName == "" {
		return errors.New("node name is required")
	}
	if cfg.Frequency <= 0 {
		return fmt.Errorf("frequency should be positive number, got %d", cfg.Frequency)
	}
	if cfg.Timeout <= 0 {
		return fmt.Errorf("timeout should be positive number, got %d", cfg.Timeout)
	}
	if cfg.Timeout <= 0 {
		return fmt.Errorf("timeout should be positive number, got %d", cfg.Timeout)
	}
	return nil
}
