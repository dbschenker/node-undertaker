package k8snodeinformer

import (
	"context"
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
)

type DefaultK8sNodeInformer struct {
}

func (c DefaultK8sNodeInformer) StartInformer(context.Context, *config.Config) error {
	return fmt.Errorf("Implement err")
}
