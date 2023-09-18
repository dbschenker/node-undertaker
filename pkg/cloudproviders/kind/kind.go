package kind

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"regexp"
)

type KindCloudProvider struct {
}

func CreateCloudProvider(ctx context.Context) (KindCloudProvider, error) {
	log.Warnf("Kind cloud provider should be used only for development and testing. This provider is not intended for production use.")
	ret := KindCloudProvider{}

	return ret, nil
}

func (p KindCloudProvider) ValidateConfig() error {
	return nil
}

func (p KindCloudProvider) TerminateNode(ctx context.Context, cloudProviderNodeId string) (string, error) {
	re, err := regexp.Compile("^kind://[^/]+/kind/(.+)$")
	if err != nil {
		return "InstanceTerminationFailed", err
	}
	matches := re.FindStringSubmatch(cloudProviderNodeId)
	if len(matches) != 2 {
		return "InstanceTerminationFailed", fmt.Errorf("couldn't parse providerId: %s", cloudProviderNodeId)
	}

	cmd := exec.Command("docker", "stop", matches[1])
	err = cmd.Run()
	if err != nil {
		return "Instance Termination Failed", err
	}
	cmd = exec.Command("docker", "rm", matches[1])
	err = cmd.Run()
	if err != nil {
		return "Instance Termination Failed", err
	}
	return "Instance Terminated", nil
}

func (p KindCloudProvider) PrepareTermination(ctx context.Context, cloudProviderNodeId string) (string, error) {
	return "No preparation required", nil
}
