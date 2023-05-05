package kind

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
)

type KindCloudProvider struct {
}

func CreateAwsCloudProvider(ctx context.Context) (KindCloudProvider, error) {
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
		return "InstanceTerminationFailed", err
	}
	cmd = exec.Command("docker", "rm", matches[1])
	err = cmd.Run()
	if err != nil {
		return "InstanceTerminationFailed", err
	}
	return "InstanceTerminated", nil
}
