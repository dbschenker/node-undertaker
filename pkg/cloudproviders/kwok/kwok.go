package kwok

import (
	"context"
	"errors"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"regexp"
)

type KwokCloudProvider struct {
	K8sClient kubernetes.Interface
}

func CreateCloudProvider(ctx context.Context, cfg *config.Config) (KwokCloudProvider, error) {
	log.Warnf("Kwok cloud provider should be used only for development and testing. This provider is not intended for production use.")
	ret := KwokCloudProvider{}
	ret.K8sClient = cfg.K8sClient
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

	if p.K8sClient == nil {
		return "InstanceTerminationFailed", errors.New("K8sclient is nil")
	}

	err = p.K8sClient.CoreV1().Nodes().Delete(ctx, matches[1], metav1.DeleteOptions{})

	if err != nil {
		return "InstanceTerminationFailed", err
	}
	return "InstanceTerminated", nil
}
