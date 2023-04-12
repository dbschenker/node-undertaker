package cloudproviders

import (
	"context"
	types "gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/pkg/nodeundertaker/struct"
)

type Message struct {
	Node                       types.Node
	CloudProviderMessageHandle string
}

type CloudProvider interface {
	ValidateConfig() error
	ReceiveMessages(context.Context) ([]Message, error)
	CompleteMessage(context.Context, Message) error
}
