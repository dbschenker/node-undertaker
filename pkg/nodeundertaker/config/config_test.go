package config

import (
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker/flags"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetConfigNegativeValidation(t *testing.T) {
	viper.Set(flags.DrainDelayFlag, -1)
	_, err := GetConfig()
	assert.Error(t, err)
}

func TestGetConfigOk(t *testing.T) {
	viper.Set(flags.PortFlag, 1)
	ret, err := GetConfig()
	assert.Error(t, err)
	assert.NotNil(t, ret)
}

func TestValidateConfigOk(t *testing.T) {
	cfg := &Config{
		DrainDelay:            1,
		CloudTerminationDelay: 1,
		Port:                  8080,
	}
	err := validateConfig(cfg)
	assert.NoError(t, err)
}

func TestValidateConfigErrDrainDelay(t *testing.T) {
	cfg := &Config{
		DrainDelay:            -1,
		CloudTerminationDelay: 1,
		Port:                  8080,
	}
	err := validateConfig(cfg)
	assert.Error(t, err)
}

func TestValidateConfigErrCloudTerminationDelay(t *testing.T) {
	cfg := &Config{
		DrainDelay:            1,
		CloudTerminationDelay: -1,
		Port:                  8080,
	}
	err := validateConfig(cfg)
	assert.Error(t, err)
}

func TestValidateConfigErrPort(t *testing.T) {
	cfg := &Config{
		DrainDelay:            1,
		CloudTerminationDelay: 1,
		Port:                  0,
	}
	err := validateConfig(cfg)
	assert.Error(t, err)
}
