package cloudproviders

import "context"

//go:generate mockgen -destination=./mocks/api_mocks.go gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/cloudproviders CLOUDPROVIDER

type CLOUDPROVIDER interface {
	ValidateConfig() error
	TerminateNode(context.Context, string) error
}
