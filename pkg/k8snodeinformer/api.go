package k8snodeinformer

import (
	"context"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/config"
)

//go:generate mockgen -destination=./mocks/api_mocks.go gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/k8snodeinformer K8SNODEINFORMER

type K8SNODEINFORMER interface {
	StartInformer(context.Context, *config.Config) error
}
