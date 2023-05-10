package kwok

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"regexp"
)

type KwokCloudProvider struct {
	K8sClient kubernetes.Interface
}

func CreateCloudProvider(ctx context.Context) (KwokCloudProvider, error) {
	ret := KwokCloudProvider{}
	var err error = nil

	return ret, err
}

func (p KwokCloudProvider) ValidateConfig() error {
	return nil
}

func (p KwokCloudProvider) TerminateNode(ctx context.Context, cloudProviderNodeId string) (string, error) {
	re, err := regexp.Compile("^kwok://(.+)$")
	if err != nil {
		return "InstanceTerminationFailed", err
	}
	matches := re.FindStringSubmatch(cloudProviderNodeId)
	if len(matches) != 2 {
		return "InstanceTerminationFailed", fmt.Errorf("couldn't parse providerId: %s", cloudProviderNodeId)
	}

	err = p.K8sClient.CoreV1().Nodes().Delete(ctx, matches[1], metav1.DeleteOptions{})

	if err != nil {
		return "InstanceTerminationFailed", err
	}
	return "InstanceTerminated", nil
}
